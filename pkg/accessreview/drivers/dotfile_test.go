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

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotfileDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/dotfile", "DOTFILE_API_KEY")
	// Dotfile authenticates via the X-DOTFILE-API-KEY header, not Authorization.
	client := newVCRClientWithHeader(rec, "X-DOTFILE-API-KEY", os.Getenv("DOTFILE_API_KEY"))

	driver := NewDotfileDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Owner: active, admin, id-first ExternalID, first+last name joined.
	owner := records[0]
	assert.Equal(t, "a1b2c3d4-0000-4000-8000-000000000001", owner.ExternalID)
	assert.Equal(t, "alice.martin@example.com", owner.Email)
	assert.Equal(t, "Alice Martin", owner.FullName)
	assert.True(t, owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	assert.Equal(t, []string{"owner"}, owner.Roles)

	// Admin: also administrative.
	admin := records[1]
	assert.Equal(t, "bob.durand@example.com", admin.Email)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, []string{"admin"}, admin.Roles)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)

	// Member: suspended (suspended_at set) → inactive, not admin.
	member := records[2]
	assert.Equal(t, "carla.petit@example.com", member.Email)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, []string{"member"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.False(t, *member.Active)
}
