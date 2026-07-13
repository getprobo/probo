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
	"strings"
)

type (
	// MemoryOrgFS is an in-memory OrgFS for tests and local git clones.
	MemoryOrgFS struct {
		repos []Repo
		files map[string]map[string][]byte
	}

	memoryRepoFS struct {
		files map[string][]byte
	}
)

func NewMemoryOrgFS(repos []Repo, files map[string]map[string][]byte) *MemoryOrgFS {
	if files == nil {
		files = map[string]map[string][]byte{}
	}

	return &MemoryOrgFS{repos: repos, files: files}
}

func (f *MemoryOrgFS) Repositories(ctx context.Context) ([]Repo, error) {
	_ = ctx

	return append([]Repo(nil), f.repos...), nil
}

func (f *MemoryOrgFS) Open(repo Repo) RepositoryFS {
	return &memoryRepoFS{files: f.files[repo.Name]}
}

func (f *MemoryOrgFS) SearchFiles(ctx context.Context, query SearchQuery) ([]FileRef, error) {
	_ = ctx

	var matches []FileRef

	for _, repo := range f.repos {
		for path := range f.files[repo.Name] {
			if !memoryQueryMatches(query, path) {
				continue
			}

			matches = append(matches, FileRef{Repo: repo, Path: path})
		}
	}

	return matches, nil
}

func (f *memoryRepoFS) Exists(ctx context.Context, path string) (bool, error) {
	_ = ctx

	_, ok := f.files[NormalizePath(path)]

	return ok, nil
}

func (f *memoryRepoFS) Read(ctx context.Context, path string) ([]byte, error) {
	_ = ctx

	content, ok := f.files[NormalizePath(path)]
	if !ok {
		return nil, ErrNotFound
	}

	return append([]byte(nil), content...), nil
}

func memoryQueryMatches(query SearchQuery, path string) bool {
	path = NormalizePath(path)
	base := path

	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		base = path[idx+1:]
	}

	if query.Filename != "" && base != query.Filename {
		return false
	}

	if query.Path != "" {
		prefix := NormalizePath(query.Path)
		if path != prefix && !strings.HasPrefix(path, prefix+"/") {
			return false
		}
	}

	if query.Extension != "" {
		ext := strings.ToLower(query.Extension)
		if !strings.HasSuffix(strings.ToLower(base), "."+ext) {
			return false
		}
	}

	return true
}
