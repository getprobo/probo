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
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"go.probo.inc/probo/pkg/discovery/vfs"
)

// Workspace is a multi-repository vfs.FS backed by go-git worktrees.
type Workspace struct {
	repos map[string]billy.Filesystem
}

func NewWorkspace() *Workspace {
	return &Workspace{repos: map[string]billy.Filesystem{}}
}

func (w *Workspace) AddRepo(name string, fs billy.Filesystem) {
	if name == "" || fs == nil {
		return
	}

	w.repos[name] = fs
}

func (w *Workspace) RepoCount() int {
	return len(w.repos)
}

func (w *Workspace) Read(ctx context.Context, path string) ([]byte, error) {
	_ = ctx

	repoName, filePath, ok := vfs.SplitRepoPath(path)
	if !ok || filePath == "" {
		return nil, vfs.ErrNotFound
	}

	fs, ok := w.repos[repoName]
	if !ok {
		return nil, vfs.ErrNotFound
	}

	content, err := ReadAll(fs, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, vfs.ErrNotFound
		}

		return nil, err
	}

	return content, nil
}

func (w *Workspace) ReadDir(ctx context.Context, dir string) ([]vfs.Entry, error) {
	_ = ctx

	dir = vfs.NormalizePath(dir)
	if dir == "" {
		return w.readWorkspaceRoot()
	}

	repoName, dirPath, ok := vfs.SplitRepoPath(dir)
	if !ok {
		return nil, vfs.ErrNotFound
	}

	fs, ok := w.repos[repoName]
	if !ok {
		return nil, vfs.ErrNotFound
	}

	if dirPath == "" {
		return readBillyDir(fs, ".")
	}

	return readBillyDir(fs, dirPath)
}

func (w *Workspace) Glob(ctx context.Context, pattern string) ([]string, error) {
	return vfs.GlobFS(ctx, w, pattern)
}

func (w *Workspace) readWorkspaceRoot() ([]vfs.Entry, error) {
	names := make([]string, 0, len(w.repos))
	for name := range w.repos {
		names = append(names, name)
	}

	sort.Strings(names)

	entries := make([]vfs.Entry, 0, len(names))
	for _, name := range names {
		entries = append(entries, vfs.Entry{Name: name, IsDir: true})
	}

	if len(entries) == 0 {
		return nil, vfs.ErrNotFound
	}

	return entries, nil
}

func readBillyDir(fs billy.Filesystem, dir string) ([]vfs.Entry, error) {
	items, err := fs.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, vfs.ErrNotFound
		}

		return nil, err
	}

	entries := make([]vfs.Entry, 0, len(items))
	for _, item := range items {
		name := item.Name()
		if name == "" || name == "." || name == ".." {
			continue
		}

		entries = append(entries, vfs.Entry{
			Name:  name,
			IsDir: item.IsDir(),
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

// PopulateFromMap adds in-memory repositories, primarily for tests.
func (w *Workspace) PopulateFromMap(files map[string]map[string][]byte) {
	for repoName, repoFiles := range files {
		fs := memfsFromMap(repoFiles)
		w.AddRepo(repoName, fs)
	}
}

func memfsFromMap(files map[string][]byte) billy.Filesystem {
	fs := memfs.New()

	for path, content := range files {
		path = strings.TrimPrefix(vfs.NormalizePath(path), "/")
		if path == "" {
			continue
		}

		if err := writeFile(fs, path, content); err != nil {
			continue
		}
	}

	return fs
}

func writeFile(fs billy.Filesystem, path string, content []byte) error {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		if err := fs.MkdirAll(path[:idx], 0o755); err != nil {
			return err
		}
	}

	file, err := fs.Create(path)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	_, err = file.Write(content)

	return err
}
