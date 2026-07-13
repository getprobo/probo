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

package vfs

import (
	"context"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("vfs: file not found")

type (
	// FS is a read-only workspace filesystem. Paths are rooted at the workspace,
	// for example "api/SECURITY.md" or ".github/CONTRIBUTING.md".
	FS interface {
		Exists(ctx context.Context, path string) (bool, error)
		Read(ctx context.Context, path string) ([]byte, error)
		Search(ctx context.Context, query SearchQuery) ([]string, error)
	}

	// SearchQuery describes a workspace file search pattern.
	SearchQuery struct {
		Filename  string
		Path      string
		Extension string
	}

	// FileIndex caches discovered workspace file paths.
	FileIndex struct {
		paths map[string]struct{}
	}
)

func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	return path
}

func RepoPath(repoName, filePath string) string {
	filePath = NormalizePath(filePath)
	if filePath == "" {
		return NormalizePath(repoName)
	}

	return NormalizePath(repoName + "/" + filePath)
}

func SplitRepoPath(path string) (repoName string, filePath string, ok bool) {
	path = NormalizePath(path)
	if path == "" {
		return "", "", false
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], "", true
	}

	return parts[0], parts[1], true
}

func NewFileIndex() *FileIndex {
	return &FileIndex{paths: map[string]struct{}{}}
}

func (idx *FileIndex) Add(path string) {
	path = NormalizePath(path)
	if path == "" {
		return
	}

	idx.paths[path] = struct{}{}
}

func (idx *FileIndex) Has(path string) bool {
	_, ok := idx.paths[NormalizePath(path)]

	return ok
}

func (idx *FileIndex) HasRepoFile(repoName, filePath string) bool {
	return idx.Has(RepoPath(repoName, filePath))
}

func (idx *FileIndex) HasRepoPrefix(repoName, prefix string) bool {
	prefix = RepoPath(repoName, prefix)
	if prefix == "" {
		return false
	}

	for path := range idx.paths {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}

func (idx *FileIndex) RepoFiles(repoName string) []string {
	prefix := NormalizePath(repoName) + "/"

	var out []string

	for path := range idx.paths {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		out = append(out, strings.TrimPrefix(path, prefix))
	}

	return out
}

// DiscoveryCatalog returns workspace search patterns used by discovery scanners.
func DiscoveryCatalog() []SearchQuery {
	return []SearchQuery{
		{Path: ".github/workflows", Extension: "yml"},
		{Path: ".github/workflows", Extension: "yaml"},
		{Filename: "SECURITY.md"},
		{Filename: "CONTRIBUTING.md"},
		{Path: ".github", Filename: "dependabot.yml"},
		{Filename: "renovate.json"},
		{Path: ".github", Filename: "renovate.json"},
		{Filename: "package-lock.json"},
		{Filename: "yarn.lock"},
		{Filename: "pnpm-lock.yaml"},
		{Filename: "go.sum"},
		{Filename: "Gemfile.lock"},
		{Filename: "poetry.lock"},
		{Filename: "Cargo.lock"},
		{Filename: ".env"},
		{Filename: ".env.production"},
		{Filename: ".env.local"},
		{Filename: "DEVELOPMENT.md"},
		{Path: "docs", Filename: "development.md"},
		{Path: "docs", Filename: "code-review.md"},
		{Path: "docs", Filename: "incident-response.md"},
		{Path: "docs/security", Filename: "README.md"},
		{Filename: "SECURITY_GUIDELINES.md"},
		{Path: ".github/ISSUE_TEMPLATE"},
		{Filename: "Jenkinsfile"},
		{Path: ".circleci", Filename: "config.yml"},
		{Filename: ".gitlab-ci.yml"},
	}
}

// BuildDiscoveryIndex runs catalog searches and merges results into a FileIndex.
func BuildDiscoveryIndex(ctx context.Context, fs FS) (*FileIndex, error) {
	index := NewFileIndex()

	var firstErr error

	for _, query := range DiscoveryCatalog() {
		paths, err := fs.Search(ctx, query)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}

			continue
		}

		for _, path := range paths {
			index.Add(path)
		}
	}

	return index, firstErr
}
