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
	"sort"
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
	for filePath, content := range files {
		normalized[NormalizePath(filePath)] = content
	}

	return &MemoryFS{files: normalized}
}

func (f *MemoryFS) Read(ctx context.Context, path string) ([]byte, error) {
	_ = ctx

	content, ok := f.files[NormalizePath(path)]
	if !ok {
		return nil, ErrNotFound
	}

	return append([]byte(nil), content...), nil
}

func (f *MemoryFS) ReadDir(ctx context.Context, dir string) ([]Entry, error) {
	_ = ctx

	dir = NormalizePath(dir)

	prefix := dir
	if prefix != "" {
		prefix += "/"
	}

	children := map[string]Entry{}

	for filePath := range f.files {
		if dir != "" && !strings.HasPrefix(filePath, prefix) {
			continue
		}

		rest := filePath
		if dir != "" {
			rest = strings.TrimPrefix(filePath, prefix)
		}

		if rest == "" {
			continue
		}

		parts := strings.SplitN(rest, "/", 2)
		name := parts[0]
		isDir := len(parts) > 1

		entry := children[name]
		entry.Name = name
		entry.IsDir = entry.IsDir || isDir
		children[name] = entry
	}

	if len(children) == 0 {
		return nil, ErrNotFound
	}

	out := make([]Entry, 0, len(children))
	for _, entry := range children {
		out = append(out, entry)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out, nil
}

func (f *MemoryFS) Glob(ctx context.Context, pattern string) ([]string, error) {
	_ = ctx

	var matches []string

	for filePath := range f.files {
		if MatchGlob(pattern, filePath) {
			matches = append(matches, filePath)
		}
	}

	sort.Strings(matches)

	return matches, nil
}
