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

func TestCursorDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/cursor", "CURSOR_ADMIN_TOKEN")
	// Cursor authenticates via HTTP Basic auth (the admin key as the
	// username). The cassette matcher ignores Authorization, so replay
	// needs no auth; the value matters only when re-recording.
	client := newVCRClient(rec, basicAuth(os.Getenv("CURSOR_ADMIN_TOKEN")))

	driver := NewCursorDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	member := records[0]
	assert.Equal(t, "jane@example.com", member.Email)
	assert.Equal(t, "Jane Doe", member.FullName)
	assert.Equal(t, "Member", member.Role)
	assert.False(t, member.IsAdmin)
	// The Cursor Admin API returns the member id as a string; it is used
	// verbatim as the stable ExternalID.
	assert.Equal(t, "10000001", member.ExternalID)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	owner := records[1]
	assert.Equal(t, "Owner", owner.Role)
	assert.True(t, owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)

	// A removed member (role "removed", isRemoved true) is still returned,
	// flagged inactive rather than dropped, per the AccountRecord contract.
	removed := records[2]
	assert.Equal(t, "Removed", removed.Role)
	assert.False(t, removed.IsAdmin)
	require.NotNil(t, removed.Active)
	assert.False(t, *removed.Active)
}

func TestCursorRole(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    string
		isAdmin bool
	}{
		{"owner", "Owner", true},
		{"free-owner", "Owner", true},
		{"member", "Member", false},
		{"removed", "Removed", false},
		{"unknown_future_role", "unknown_future_role", false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, cursorRole(c.in))
			assert.Equal(t, c.isAdmin, cursorIsAdmin(c.in))
		})
	}
}
