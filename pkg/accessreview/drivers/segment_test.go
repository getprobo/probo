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

func TestSegmentDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/segment", "SEGMENT_API_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SEGMENT_API_TOKEN")))

	driver := NewSegmentDriver(client, "https://api.segmentapis.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// First user: empty name falls back to email; Workspace Owner ⇒ admin.
	owner := records[0]
	assert.Equal(t, "sgJDWk3K21k6LE3tLU9nRK", owner.ExternalID)
	assert.Equal(t, "papi@example.com", owner.Email)
	assert.Equal(t, "papi@example.com", owner.FullName)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, []string{"Workspace Owner"}, owner.Roles)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)

	// Second user: named; resource-scoped read-only role ⇒ not admin.
	member := records[1]
	assert.Equal(t, "i2VTJURQprNfqdwjLFPWYx", member.ExternalID)
	assert.Equal(t, "Sloth", member.FullName)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, []string{"Source Read-only"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	// Pending invite: inactive, no roles, keyed by email.
	invite := records[2]
	assert.Equal(t, "foo@example.com", invite.Email)
	assert.Equal(t, "foo@example.com", invite.ExternalID)
	assert.False(t, invite.IsAdmin)
	assert.Empty(t, invite.Roles)
	require.NotNil(t, invite.Active)
	assert.False(t, *invite.Active)
}
