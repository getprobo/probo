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
	"fmt"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/discovery/vfs"
)

type (
	githubFS struct {
		api *apiClient
		org string
	}

	contentsDirItem struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
)

// discoveryGlobQueries maps vfs discovery glob patterns to GitHub code search
// query fragments (without the org: qualifier).
var discoveryGlobQueries = map[string]string{
	"*/.github/workflows/*.yml":   "path:.github/workflows extension:yml",
	"*/.github/workflows/*.yaml":  "path:.github/workflows extension:yaml",
	"*/SECURITY.md":               "filename:SECURITY.md",
	"*/CONTRIBUTING.md":           "filename:CONTRIBUTING.md",
	"*/.github/dependabot.yml":    "path:.github filename:dependabot.yml",
	"*/renovate.json":             "filename:renovate.json",
	"*/.github/renovate.json":     "path:.github filename:renovate.json",
	"*/package-lock.json":         "filename:package-lock.json",
	"*/yarn.lock":                 "filename:yarn.lock",
	"*/pnpm-lock.yaml":            "filename:pnpm-lock.yaml",
	"*/go.sum":                    "filename:go.sum",
	"*/Gemfile.lock":              "filename:Gemfile.lock",
	"*/poetry.lock":               "filename:poetry.lock",
	"*/Cargo.lock":                "filename:Cargo.lock",
	"*/.env":                      "filename:.env",
	"*/.env.production":           "filename:.env.production",
	"*/.env.local":                "filename:.env.local",
	"*/DEVELOPMENT.md":            "filename:DEVELOPMENT.md",
	"*/docs/development.md":       "path:docs filename:development.md",
	"*/docs/code-review.md":       "path:docs filename:code-review.md",
	"*/docs/incident-response.md": "path:docs filename:incident-response.md",
	"*/docs/security/README.md":   "path:docs/security filename:README.md",
	"*/SECURITY_GUIDELINES.md":    "filename:SECURITY_GUIDELINES.md",
	"*/.github/ISSUE_TEMPLATE/*":  "path:.github/ISSUE_TEMPLATE",
	"*/Jenkinsfile":               "filename:Jenkinsfile",
	"*/.circleci/config.yml":      "path:.circleci filename:config.yml",
	"*/.gitlab-ci.yml":            "filename:.gitlab-ci.yml",
}

func newGitHubFS(api *apiClient, org string) *githubFS {
	return &githubFS{api: api, org: org}
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

func (f *githubFS) ReadDir(ctx context.Context, dir string) ([]vfs.Entry, error) {
	dir = vfs.NormalizePath(dir)
	if dir == "" {
		return nil, vfs.ErrNotFound
	}

	repoName, dirPath, ok := vfs.SplitRepoPath(dir)
	if !ok {
		return nil, vfs.ErrNotFound
	}

	segments := []string{"contents"}
	if dirPath != "" {
		segments = append(segments, splitContentPath(dirPath)...)
	}

	endpoint, err := f.api.repoEndpoint(f.org, repoName, segments...)
	if err != nil {
		return nil, fmt.Errorf("cannot build github contents URL: %w", err)
	}

	var items []contentsDirItem

	if _, err := f.api.getJSON(ctx, endpoint, &items); err != nil {
		return nil, vfs.ErrNotFound
	}

	entries := make([]vfs.Entry, 0, len(items))
	for _, item := range items {
		entries = append(entries, vfs.Entry{
			Name:  item.Name,
			IsDir: strings.EqualFold(item.Type, "dir"),
		})
	}

	if len(entries) == 0 {
		return nil, vfs.ErrNotFound
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

func (f *githubFS) Glob(ctx context.Context, pattern string) ([]string, error) {
	query, ok := codeSearchQueryForGlob(f.org, pattern)
	if !ok {
		return nil, nil
	}

	paths, err := f.api.searchCode(ctx, query)
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if vfs.MatchGlob(pattern, path) {
			filtered = append(filtered, path)
		}
	}

	sort.Strings(filtered)

	return filtered, nil
}

func (c *apiClient) searchCode(ctx context.Context, query string) ([]string, error) {
	rawPaths, err := c.searchCodePaths(ctx, strings.ReplaceAll(query, "+", " "))
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(rawPaths))
	for _, path := range rawPaths {
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}

		paths = append(paths, vfs.RepoPath(parts[0], parts[1]))
	}

	return paths, nil
}

func codeSearchQueryForGlob(org, pattern string) (string, bool) {
	fragment, ok := discoveryGlobQueries[pattern]
	if !ok {
		return "", false
	}

	return formatCodeSearchQuery(org, fragment), true
}

func formatCodeSearchQuery(org, fragment string) string {
	fragment = strings.TrimSpace(fragment)
	fragment = strings.TrimPrefix(fragment, "org:"+org)
	fragment = strings.TrimSpace(fragment)

	return "org:" + org + " " + strings.ReplaceAll(fragment, " ", "+")
}
