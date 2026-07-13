// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/discovery/vfs"
)

type (
	githubFS struct {
		api *apiClient
		org string
	}

	codeSearchResponse struct {
		TotalCount        int              `json:"total_count"`
		IncompleteResults bool             `json:"incomplete_results"`
		Items             []codeSearchItem `json:"items"`
	}

	codeSearchItem struct {
		Name       string `json:"name"`
		Path       string `json:"path"`
		Repository struct {
			Name string `json:"name"`
		} `json:"repository"`
	}
)

func newGitHubFS(api *apiClient, org string) *githubFS {
	return &githubFS{api: api, org: org}
}

func (f *githubFS) Exists(ctx context.Context, path string) (bool, error) {
	path = vfs.NormalizePath(path)
	if path == "" {
		return false, nil
	}

	repoName, filePath, ok := vfs.SplitRepoPath(path)
	if !ok || filePath == "" {
		return false, nil
	}

	segments := append([]string{"contents"}, splitContentPath(filePath)...)

	endpoint, err := f.api.repoEndpoint(f.org, repoName, segments...)
	if err != nil {
		return false, fmt.Errorf("cannot build github contents URL: %w", err)
	}

	var payload map[string]any

	if _, err := f.api.getJSON(ctx, endpoint, &payload); err != nil {
		return false, nil
	}

	return true, nil
}

func (f *githubFS) Read(ctx context.Context, path string) ([]byte, error) {
	path = vfs.NormalizePath(path)
	if path == "" {
		return nil, vfs.ErrNotFound
	}

	repoName, filePath, ok := vfs.SplitRepoPath(path)
	if !ok || filePath == "" {
		return nil, vfs.ErrNotFound
	}

	segments := append([]string{"contents"}, splitContentPath(filePath)...)

	endpoint, err := f.api.repoEndpoint(f.org, repoName, segments...)
	if err != nil {
		return nil, fmt.Errorf("cannot build github contents URL: %w", err)
	}

	var payload contentResponse

	if _, err := f.api.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, vfs.ErrNotFound
	}

	content, ok := decodeGitHubContent(payload.Encoding, payload.Content)
	if !ok {
		return nil, fmt.Errorf("cannot decode github file content")
	}

	return []byte(content), nil
}

func (f *githubFS) Search(ctx context.Context, query vfs.SearchQuery) ([]string, error) {
	return f.api.searchCode(ctx, f.buildSearchQuery(query))
}

func (f *githubFS) buildSearchQuery(query vfs.SearchQuery) string {
	parts := []string{"org:" + f.org}

	if query.Path != "" {
		parts = append(parts, "path:"+query.Path)
	}

	if query.Filename != "" {
		parts = append(parts, "filename:"+query.Filename)
	}

	if query.Extension != "" {
		parts = append(parts, "extension:"+query.Extension)
	}

	return strings.Join(parts, "+")
}

func (c *apiClient) searchCode(ctx context.Context, query string) ([]string, error) {
	endpoint, err := url.Parse(githubAPIBase + "/search/code")
	if err != nil {
		return nil, fmt.Errorf("cannot build github search URL: %w", err)
	}

	q := endpoint.Query()
	q.Set("q", strings.ReplaceAll(query, "+", " "))
	q.Set("per_page", "100")
	endpoint.RawQuery = q.Encode()

	var (
		paths []string
		next  = endpoint.String()
	)

	for page := 0; page < maxPagesPerList && next != ""; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return paths, fmt.Errorf("cannot create github search request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return paths, fmt.Errorf("cannot execute github search request: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()

			if resp.StatusCode == 403 || resp.StatusCode == 422 {
				return paths, fmt.Errorf("github code search unavailable: unexpected status %d", resp.StatusCode)
			}

			return paths, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var pageResp codeSearchResponse

		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			_ = resp.Body.Close()

			return paths, fmt.Errorf("cannot decode github search response: %w", err)
		}

		next = parseLinkNext(resp.Header.Get("Link"))
		_ = resp.Body.Close()

		if len(pageResp.Items) == 0 {
			break
		}

		for _, item := range pageResp.Items {
			if item.Path == "" || item.Repository.Name == "" {
				continue
			}

			paths = append(paths, vfs.RepoPath(item.Repository.Name, item.Path))
		}
	}

	return paths, nil
}
