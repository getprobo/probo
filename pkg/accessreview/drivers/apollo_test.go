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

func TestApolloDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/apollo", "APOLLO_API_KEY")
	// Apollo authenticates via the x-api-key header, not Authorization.
	client := newVCRClientWithHeader(rec, "x-api-key", os.Getenv("APOLLO_API_KEY"))

	driver := NewApolloDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	admin := records[0]
	assert.Equal(t, "5f0000000000000000000001", admin.ExternalID)
	assert.Equal(t, "alice@example.com", admin.Email)
	assert.Equal(t, "Alice Admin", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	rep := records[1]
	assert.Equal(t, []string{"Sales Rep"}, rep.Roles)
	assert.False(t, rep.IsAdmin)

	manager := records[2]
	assert.Equal(t, []string{"Billing and Seat Manager"}, manager.Roles)
	assert.False(t, manager.IsAdmin)

	// No name and no first/last: the display name falls back to the email.
	noName := records[3]
	assert.Equal(t, "dave@example.com", noName.Email)
	assert.Equal(t, "dave@example.com", noName.FullName)
}

func TestApolloIsAdmin(t *testing.T) {
	t.Parallel()

	assert.True(t, apolloIsAdmin("Admin"))
	assert.True(t, apolloIsAdmin("admin"))
	// Exact match only: profiles that merely contain "admin" are not admins.
	assert.False(t, apolloIsAdmin("Master Admin"))
	assert.False(t, apolloIsAdmin("Billing Admin"))
	assert.False(t, apolloIsAdmin("Sales Rep"))
}
