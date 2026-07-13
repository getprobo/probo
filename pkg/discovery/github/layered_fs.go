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
	"errors"
	"sort"

	"go.probo.inc/probo/pkg/discovery/vfs"
)

type layeredFS struct {
	primary  vfs.FS
	fallback vfs.FS
}

func newLayeredFS(primary, fallback vfs.FS) vfs.FS {
	return &layeredFS{primary: primary, fallback: fallback}
}

func (f *layeredFS) Read(ctx context.Context, path string) ([]byte, error) {
	content, err := f.primary.Read(ctx, path)
	if err == nil {
		return content, nil
	}

	if !errors.Is(err, vfs.ErrNotFound) {
		return nil, err
	}

	return f.fallback.Read(ctx, path)
}

func (f *layeredFS) ReadDir(ctx context.Context, dir string) ([]vfs.Entry, error) {
	entries, err := f.primary.ReadDir(ctx, dir)
	if err == nil {
		return entries, nil
	}

	if !errors.Is(err, vfs.ErrNotFound) {
		return nil, err
	}

	return f.fallback.ReadDir(ctx, dir)
}

func (f *layeredFS) Glob(ctx context.Context, pattern string) ([]string, error) {
	primaryMatches, err := f.primary.Glob(ctx, pattern)
	if err != nil {
		return nil, err
	}

	fallbackMatches, err := f.fallback.Glob(ctx, pattern)
	if err != nil {
		return primaryMatches, nil
	}

	return mergePaths(primaryMatches, fallbackMatches), nil
}

func mergePaths(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	out := make([]string, 0, len(left)+len(right))

	for _, path := range append(left, right...) {
		if _, ok := seen[path]; ok {
			continue
		}

		seen[path] = struct{}{}
		out = append(out, path)
	}

	sort.Strings(out)

	return out
}
