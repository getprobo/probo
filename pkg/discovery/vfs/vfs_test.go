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

func TestMemoryOrgFS_ReadAndSearch(t *testing.T) {
	t.Parallel()

	repo := Repo{Owner: "acme", Name: "api"}
	orgFS := NewMemoryOrgFS(
		[]Repo{repo},
		map[string]map[string][]byte{
			"api": {
				"SECURITY.md":               []byte("security@example.com"),
				".github/workflows/ci.yml":  []byte("on: pull_request"),
				".github/dependabot.yml":    []byte("version: 2"),
				"docs/incident-response.md": []byte("incident response"),
			},
		},
	)

	repoFS := orgFS.Open(repo)

	exists, err := repoFS.Exists(context.Background(), "SECURITY.md")
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := repoFS.Read(context.Background(), ".github/workflows/ci.yml")
	require.NoError(t, err)
	assert.Contains(t, string(content), "pull_request")

	index, err := BuildIndex(context.Background(), orgFS)
	require.NoError(t, err)

	assert.True(t, index.Has("api", "SECURITY.md"))
	assert.True(t, index.HasPrefix("api", ".github/workflows"))
	assert.True(t, index.Has("api", ".github/dependabot.yml"))
}
