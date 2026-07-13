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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryFS_ReadDirAndGlob(t *testing.T) {
	t.Parallel()

	fs := NewMemoryFS(map[string][]byte{
		"api/SECURITY.md":               []byte("security@example.com"),
		"api/.github/workflows/ci.yml":  []byte("on: pull_request"),
		"api/.github/dependabot.yml":    []byte("version: 2"),
		"api/docs/incident-response.md": []byte("incident response"),
		"web/README.md":                 []byte("hello"),
	})

	content, err := fs.Read(context.Background(), "api/SECURITY.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "security@")

	entries, err := fs.ReadDir(context.Background(), "api")
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	matches, err := fs.Glob(context.Background(), "*/SECURITY.md")
	require.NoError(t, err)
	assert.Contains(t, matches, "api/SECURITY.md")

	index, err := BuildDiscoveryIndex(context.Background(), fs)
	require.NoError(t, err)

	assert.True(t, index.HasRepoFile("api", "SECURITY.md"))
	assert.True(t, index.HasRepoPrefix("api", ".github/workflows"))
}

func TestHasPath(t *testing.T) {
	t.Parallel()

	fs := NewMemoryFS(map[string][]byte{
		"api/SECURITY.md": []byte("x"),
	})

	assert.True(t, HasPath(context.Background(), fs, "api/SECURITY.md"))
	assert.False(t, HasPath(context.Background(), fs, "api/missing.md"))
}

func TestRepoPathHelpers(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "api/SECURITY.md", RepoPath("api", "SECURITY.md"))

	repo, file, ok := SplitRepoPath("api/.github/workflows/ci.yml")
	assert.True(t, ok)
	assert.Equal(t, "api", repo)
	assert.Equal(t, ".github/workflows/ci.yml", file)
}

func TestMatchGlob(t *testing.T) {
	t.Parallel()

	assert.True(t, MatchGlob("*/SECURITY.md", "api/SECURITY.md"))
	assert.True(t, MatchGlob("*/.github/workflows/*.yml", "api/.github/workflows/ci.yml"))
	assert.False(t, MatchGlob("*/SECURITY.md", "api/docs/SECURITY.md"))
}
