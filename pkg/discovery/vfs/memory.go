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

// MemoryFS is an in-memory FS for tests and local git workspaces.
type MemoryFS struct {
	files map[string][]byte
}

func NewMemoryFS(files map[string][]byte) *MemoryFS {
	if files == nil {
		files = map[string][]byte{}
	}

	normalized := make(map[string][]byte, len(files))
	for path, content := range files {
		normalized[NormalizePath(path)] = content
	}

	return &MemoryFS{files: normalized}
}

func (f *MemoryFS) Exists(ctx context.Context, path string) (bool, error) {
	_ = ctx

	_, ok := f.files[NormalizePath(path)]

	return ok, nil
}

func (f *MemoryFS) Read(ctx context.Context, path string) ([]byte, error) {
	_ = ctx

	content, ok := f.files[NormalizePath(path)]
	if !ok {
		return nil, ErrNotFound
	}

	return append([]byte(nil), content...), nil
}

func (f *MemoryFS) Search(ctx context.Context, query SearchQuery) ([]string, error) {
	_ = ctx

	var matches []string

	for path := range f.files {
		repoName, filePath, ok := SplitRepoPath(path)
		if !ok {
			continue
		}

		if !memoryQueryMatches(query, repoName, filePath) {
			continue
		}

		matches = append(matches, path)
	}

	return matches, nil
}

func memoryQueryMatches(query SearchQuery, repoName, filePath string) bool {
	_ = repoName

	path := NormalizePath(filePath)
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
