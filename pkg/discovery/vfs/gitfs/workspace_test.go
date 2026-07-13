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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/discovery/vfs"
)

func TestWorkspace_ReadDirGlob(t *testing.T) {
	t.Parallel()

	workspace := NewWorkspace()
	workspace.PopulateFromMap(map[string]map[string][]byte{
		"api": {
			"SECURITY.md":              []byte("security@example.com"),
			".github/workflows/ci.yml": []byte("on: pull_request"),
		},
	})

	content, err := workspace.Read(context.Background(), "api/SECURITY.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "security@")

	matches, err := workspace.Glob(context.Background(), "*/SECURITY.md")
	require.NoError(t, err)
	assert.Contains(t, matches, "api/SECURITY.md")

	index, err := vfs.BuildDiscoveryIndex(context.Background(), workspace)
	require.NoError(t, err)
	assert.True(t, index.HasRepoPrefix("api", ".github/workflows"))
}
