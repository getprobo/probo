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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/discovery/vfs"
)

func TestLayeredFS_FallsBackToSecondary(t *testing.T) {
	t.Parallel()

	primary := vfs.NewMemoryFS(map[string][]byte{
		"api/SECURITY.md": []byte("from git"),
	})
	secondary := vfs.NewMemoryFS(map[string][]byte{
		"web/README.md": []byte("from api"),
	})

	fs := newLayeredFS(primary, secondary)

	content, err := fs.Read(context.Background(), "api/SECURITY.md")
	require.NoError(t, err)
	assert.Equal(t, "from git", string(content))

	content, err = fs.Read(context.Background(), "web/README.md")
	require.NoError(t, err)
	assert.Equal(t, "from api", string(content))

	matches, err := fs.Glob(context.Background(), "*/README.md")
	require.NoError(t, err)
	assert.Contains(t, matches, "web/README.md")
}
