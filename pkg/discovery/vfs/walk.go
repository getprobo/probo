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
	"sort"
)

// Walk visits every file in fs, passing workspace-rooted paths to fn.
func Walk(ctx context.Context, fs FS, fn func(path string) error) error {
	return walkDir(ctx, fs, "", fn)
}

// GlobFS matches pattern against every file reachable from fs via ReadDir.
func GlobFS(ctx context.Context, fs FS, pattern string) ([]string, error) {
	var matches []string

	err := Walk(ctx, fs, func(path string) error {
		if MatchGlob(pattern, path) {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(matches)

	return matches, nil
}

func walkDir(ctx context.Context, fs FS, dir string, fn func(path string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	entries, err := fs.ReadDir(ctx, dir)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}

		return err
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		child := entry.Name
		if dir != "" {
			child = dir + "/" + entry.Name
		}

		child = NormalizePath(child)

		if entry.IsDir {
			if err := walkDir(ctx, fs, child, fn); err != nil {
				return err
			}

			continue
		}

		if err := fn(child); err != nil {
			return err
		}
	}

	return nil
}
