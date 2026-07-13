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
	githubOrgFS struct {
		api   *apiClient
		org   string
		index *vfs.FileIndex
	}

	githubRepoFS struct {
		api   *apiClient
		org   string
		repo  string
		index *vfs.FileIndex
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

func newGitHubOrgFS(api *apiClient, org string) *githubOrgFS {
	return &githubOrgFS{api: api, org: org}
}

func (f *githubOrgFS) Repositories(ctx context.Context) ([]vfs.Repo, error) {
	repos, err := f.listRepos(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]vfs.Repo, 0, len(repos))
	for _, repo := range repos {
		out = append(out, vfs.Repo{Owner: f.org, Name: repo.Name})
	}

	return out, nil
}

func (f *githubOrgFS) Open(repo vfs.Repo) vfs.RepositoryFS {
	return &githubRepoFS{
		api:   f.api,
		org:   f.org,
		repo:  repo.Name,
		index: f.index,
	}
}

func (f *githubOrgFS) SearchFiles(ctx context.Context, query vfs.SearchQuery) ([]vfs.FileRef, error) {
	return f.api.searchCode(ctx, f.buildSearchQuery(query))
}

func (f *githubOrgFS) IndexFiles(ctx context.Context) (*vfs.FileIndex, error) {
	index := vfs.NewFileIndex()

	var firstErr error

	for _, query := range vfs.DiscoveryCatalog() {
		files, err := f.SearchFiles(ctx, query)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}

			continue
		}

		for _, file := range files {
			file.Repo.Owner = f.org
			index.Add(file)
		}
	}

	f.index = index

	return index, firstErr
}

func (f *githubOrgFS) buildSearchQuery(query vfs.SearchQuery) string {
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

func (f *githubOrgFS) listRepos(ctx context.Context) ([]repoListItem, error) {
	endpoint, err := f.api.orgEndpoint(f.org, "repos")
	if err != nil {
		return nil, fmt.Errorf("cannot build github repos URL: %w", err)
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return nil, fmt.Errorf("cannot build github repos URL: %w", err)
	}

	var repos []repoListItem

	if _, err := f.api.getPaginated(ctx, endpoint, &repos); err != nil {
		return nil, fmt.Errorf("cannot list github repos: %w", err)
	}

	return repos, nil
}

func (r *githubRepoFS) Exists(ctx context.Context, path string) (bool, error) {
	path = vfs.NormalizePath(path)
	if path == "" {
		return false, nil
	}

	if r.index != nil && r.index.Has(r.repo, path) {
		return true, nil
	}

	segments := append([]string{"contents"}, splitContentPath(path)...)

	endpoint, err := r.api.repoEndpoint(r.org, r.repo, segments...)
	if err != nil {
		return false, fmt.Errorf("cannot build github contents URL: %w", err)
	}

	var payload map[string]any

	if _, err := r.api.getJSON(ctx, endpoint, &payload); err != nil {
		return false, nil
	}

	return true, nil
}

func (r *githubRepoFS) Read(ctx context.Context, path string) ([]byte, error) {
	path = vfs.NormalizePath(path)
	if path == "" {
		return nil, vfs.ErrNotFound
	}

	segments := append([]string{"contents"}, splitContentPath(path)...)

	endpoint, err := r.api.repoEndpoint(r.org, r.repo, segments...)
	if err != nil {
		return nil, fmt.Errorf("cannot build github contents URL: %w", err)
	}

	var payload contentResponse

	if _, err := r.api.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, vfs.ErrNotFound
	}

	content, ok := decodeGitHubContent(payload.Encoding, payload.Content)
	if !ok {
		return nil, fmt.Errorf("cannot decode github file content")
	}

	return []byte(content), nil
}

func (c *apiClient) searchCode(ctx context.Context, query string) ([]vfs.FileRef, error) {
	endpoint, err := url.Parse(githubAPIBase + "/search/code")
	if err != nil {
		return nil, fmt.Errorf("cannot build github search URL: %w", err)
	}

	q := endpoint.Query()
	q.Set("q", strings.ReplaceAll(query, "+", " "))
	q.Set("per_page", "100")
	endpoint.RawQuery = q.Encode()

	var (
		refs []vfs.FileRef
		next = endpoint.String()
	)

	for page := 0; page < maxPagesPerList && next != ""; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return refs, fmt.Errorf("cannot create github search request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return refs, fmt.Errorf("cannot execute github search request: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()

			if resp.StatusCode == 403 || resp.StatusCode == 422 {
				return refs, fmt.Errorf("github code search unavailable: unexpected status %d", resp.StatusCode)
			}

			return refs, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var pageResp codeSearchResponse

		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			_ = resp.Body.Close()

			return refs, fmt.Errorf("cannot decode github search response: %w", err)
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

			refs = append(refs, vfs.FileRef{
				Repo: vfs.Repo{Name: item.Repository.Name},
				Path: item.Path,
			})
		}
	}

	return refs, nil
}
