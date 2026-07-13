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
	// Entry is a single directory listing item.
	Entry struct {
		Name  string
		IsDir bool
	}

	// FS is a read-only workspace filesystem aligned with io/fs semantics.
	// Paths are workspace-rooted, for example "api/SECURITY.md".
	FS interface {
		Read(ctx context.Context, path string) ([]byte, error)
		ReadDir(ctx context.Context, dir string) ([]Entry, error)
		Glob(ctx context.Context, pattern string) ([]string, error)
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

// HasPath reports whether a file or non-empty directory exists at path.
func HasPath(ctx context.Context, fs FS, path string) bool {
	_, err := fs.Read(ctx, path)
	if err == nil {
		return true
	}

	if !errors.Is(err, ErrNotFound) {
		return false
	}

	entries, err := fs.ReadDir(ctx, path)

	return err == nil && len(entries) > 0
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

// DiscoveryGlobCatalog returns glob patterns used to index discovery files.
func DiscoveryGlobCatalog() []string {
	return []string{
		"*/.github/workflows/*.yml",
		"*/.github/workflows/*.yaml",
		"*/SECURITY.md",
		"*/CONTRIBUTING.md",
		"*/.github/dependabot.yml",
		"*/renovate.json",
		"*/.github/renovate.json",
		"*/package-lock.json",
		"*/yarn.lock",
		"*/pnpm-lock.yaml",
		"*/go.sum",
		"*/Gemfile.lock",
		"*/poetry.lock",
		"*/Cargo.lock",
		"*/.env",
		"*/.env.production",
		"*/.env.local",
		"*/DEVELOPMENT.md",
		"*/docs/development.md",
		"*/docs/code-review.md",
		"*/docs/incident-response.md",
		"*/docs/security/README.md",
		"*/SECURITY_GUIDELINES.md",
		"*/.github/ISSUE_TEMPLATE/*",
		"*/Jenkinsfile",
		"*/.circleci/config.yml",
		"*/.gitlab-ci.yml",
	}
}

// BuildDiscoveryIndex walks the workspace once and indexes discovery catalog matches.
func BuildDiscoveryIndex(ctx context.Context, fs FS) (*FileIndex, error) {
	index := NewFileIndex()
	patterns := DiscoveryGlobCatalog()

	err := Walk(ctx, fs, func(path string) error {
		for _, pattern := range patterns {
			if MatchGlob(pattern, path) {
				index.Add(path)

				break
			}
		}

		return nil
	})

	return index, err
}
