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
	// Repo identifies a repository in an organization or workspace.
	Repo struct {
		Owner string
		Name  string
	}

	// FileRef is a file path inside a repository.
	FileRef struct {
		Repo Repo
		Path string
	}

	// SearchQuery describes an org-wide file search pattern.
	SearchQuery struct {
		Filename  string
		Path      string
		Extension string
	}

	// RepositoryFS reads files from a single repository.
	RepositoryFS interface {
		Exists(ctx context.Context, path string) (bool, error)
		Read(ctx context.Context, path string) ([]byte, error)
	}

	// OrgFS lists repositories and resolves per-repo filesystems.
	OrgFS interface {
		Repositories(ctx context.Context) ([]Repo, error)
		Open(repo Repo) RepositoryFS
		SearchFiles(ctx context.Context, query SearchQuery) ([]FileRef, error)
	}

	// FileIndex caches org-wide file paths keyed by repository name.
	FileIndex struct {
		byRepo map[string]map[string]struct{}
	}
)

func (r Repo) Key() string {
	return r.Owner + "/" + r.Name
}

func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	return path
}

func NewFileIndex() *FileIndex {
	return &FileIndex{byRepo: map[string]map[string]struct{}{}}
}

func (idx *FileIndex) Add(ref FileRef) {
	path := NormalizePath(ref.Path)
	if path == "" {
		return
	}

	repo := ref.Repo.Name
	if idx.byRepo[repo] == nil {
		idx.byRepo[repo] = map[string]struct{}{}
	}

	idx.byRepo[repo][path] = struct{}{}
}

func (idx *FileIndex) Has(repoName, path string) bool {
	paths, ok := idx.byRepo[repoName]
	if !ok {
		return false
	}

	_, ok = paths[NormalizePath(path)]

	return ok
}

func (idx *FileIndex) HasPrefix(repoName, prefix string) bool {
	prefix = NormalizePath(prefix)
	if prefix == "" {
		return false
	}

	for path := range idx.byRepo[repoName] {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}

func (idx *FileIndex) Paths(repoName string) []string {
	paths := idx.byRepo[repoName]
	if len(paths) == 0 {
		return nil
	}

	out := make([]string, 0, len(paths))
	for path := range paths {
		out = append(out, path)
	}

	return out
}

func (idx *FileIndex) HasAny(repoName string, candidates ...string) bool {
	for _, candidate := range candidates {
		if idx.Has(repoName, candidate) {
			return true
		}
	}

	for _, candidate := range candidates {
		if idx.HasPrefix(repoName, candidate) {
			return true
		}
	}

	return false
}

// DiscoveryCatalog returns org-wide search patterns used by discovery scanners.
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

// BuildIndex runs catalog searches and merges results into a FileIndex.
func BuildIndex(ctx context.Context, orgFS OrgFS) (*FileIndex, error) {
	index := NewFileIndex()

	for _, query := range DiscoveryCatalog() {
		files, err := orgFS.SearchFiles(ctx, query)
		if err != nil {
			return index, err
		}

		for _, file := range files {
			index.Add(file)
		}
	}

	return index, nil
}
