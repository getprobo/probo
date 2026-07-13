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

package gitfs

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
)

// CloneRepo performs a shallow single-branch clone into memory and returns the
// repository worktree filesystem.
func CloneRepo(
	ctx context.Context,
	repoURL string,
	auth transport.AuthMethod,
	branch string,
) (billy.Filesystem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fs := memfs.New()

	opts := &git.CloneOptions{
		URL:          repoURL,
		Auth:         auth,
		Depth:        1,
		SingleBranch: true,
		Tags:         git.NoTags,
	}

	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	_, err := git.Clone(memory.NewStorage(), fs, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot clone repository %s: %w", repoURL, err)
	}

	return fs, nil
}

// WalkFiles invokes fn for every file in fs, passing repo-relative paths.
func WalkFiles(fs billy.Filesystem, fn func(path string) error) error {
	return walkDir(fs, ".", fn)
}

func walkDir(fs billy.Filesystem, dir string, fn func(path string) error) error {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}

		child := path.Join(dir, name)
		if dir == "." {
			child = name
		}

		if entry.IsDir() {
			if err := walkDir(fs, child, fn); err != nil {
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

// ReadAll reads a file from a billy filesystem.
func ReadAll(fs billy.Filesystem, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	return io.ReadAll(file)
}
