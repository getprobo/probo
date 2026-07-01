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
	"go.probo.inc/probo/pkg/coredata"
)

func TestRailwayDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/railway", "RAILWAY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("RAILWAY_TOKEN")))

	driver := NewRailwayDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	admin := records[0]
	assert.Equal(t, "8f7e6d5c-4b3a-2910-8a7b-6c5d4e3f2a1b", admin.ExternalID)
	assert.Equal(t, "ada@example.com", admin.Email)
	assert.Equal(t, "Ada Lovelace", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, admin.MFAStatus)
	// Railway's WorkspaceMember exposes no status/active field.
	assert.Nil(t, admin.Active)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	member := records[1]
	assert.Equal(t, "1b2c3d4e-5f6a-7b8c-9d0e-1f2a3b4c5d6e", member.ExternalID)
	assert.Equal(t, "grace@example.com", member.Email)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, member.MFAStatus)
	assert.Nil(t, member.Active)
}

// TestRailwayRecords drives the cross-workspace aggregation directly (the
// cassette has a single workspace, so it cannot exercise dedup). A member who
// appears in two workspaces yields one record with unioned roles, IsAdmin true
// if any appearance is ADMIN, and MFA enabled if any appearance reports it; a
// member whose two-factor flag is null in every workspace stays Unknown.
func TestRailwayRecords(t *testing.T) {
	t.Parallel()

	enabled := true
	disabled := false

	me := &railwayMe{
		Workspaces: []railwayWorkspace{
			{
				ID:   "ws-a",
				Name: "Alpha",
				Members: []railwayMember{
					{ID: "u-alice", Email: "alice@example.com", Name: "Alice", Role: "ADMIN", TwoFactorAuthEnabled: &enabled},
					{ID: "u-bob", Email: "bob@example.com", Name: "Bob", Role: "MEMBER", TwoFactorAuthEnabled: &disabled},
					{ID: "u-carol", Email: "carol@example.com", Name: "Carol", Role: "VIEWER"},
				},
			},
			{
				ID:   "ws-b",
				Name: "Beta",
				Members: []railwayMember{
					{ID: "u-alice", Email: "alice@example.com", Name: "Alice", Role: "MEMBER", TwoFactorAuthEnabled: &disabled},
					{ID: "u-carol", Email: "carol@example.com", Name: "Carol", Role: "VIEWER"},
				},
			},
		},
	}

	records := railwayRecords(me)
	require.Len(t, records, 3)

	byID := make(map[string]AccountRecord, len(records))
	for _, r := range records {
		byID[r.ExternalID] = r
	}

	alice := byID["u-alice"]
	assert.Equal(t, []string{"Admin", "Member"}, alice.Roles)
	assert.True(t, alice.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, alice.MFAStatus)
	assert.Nil(t, alice.Active)

	bob := byID["u-bob"]
	assert.Equal(t, []string{"Member"}, bob.Roles)
	assert.False(t, bob.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, bob.MFAStatus)

	carol := byID["u-carol"]
	assert.Equal(t, []string{"Viewer"}, carol.Roles)
	assert.False(t, carol.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, carol.MFAStatus)
}
