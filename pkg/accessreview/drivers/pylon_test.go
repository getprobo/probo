// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestPylonDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/pylon", "PYLON_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("PYLON_API_KEY")))

	driver := NewPylonDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	admin := records[0]
	assert.Equal(t, "user_1", admin.ExternalID)
	assert.Equal(t, "alice@example.com", admin.Email)
	assert.Equal(t, "Alice Admin", admin.FullName)
	// role_id "role_admin" resolved through GET /user-roles.
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	member := records[1]
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)

	// No name → display name falls back to the email; "deactivated" status →
	// Active false.
	deactivated := records[2]
	assert.Equal(t, "carol@example.com", deactivated.FullName)
	assert.Equal(t, []string{"Member"}, deactivated.Roles)
	require.NotNil(t, deactivated.Active)
	assert.False(t, *deactivated.Active)
}

func TestPylonIsAdmin(t *testing.T) {
	t.Parallel()

	// The stable slug is preferred when present.
	assert.True(t, pylonIsAdmin(pylonRole{Slug: "admin", Name: "Admin"}))
	assert.False(t, pylonIsAdmin(pylonRole{Slug: "member", Name: "Member"}))
	// A custom role named like an admin but with a non-admin slug is NOT an
	// admin — the slug wins.
	assert.False(t, pylonIsAdmin(pylonRole{Slug: "billing", Name: "Billing Admin"}))
	// With no slug, the exact (case-insensitive) name is used.
	assert.True(t, pylonIsAdmin(pylonRole{Name: "Admin"}))
	assert.False(t, pylonIsAdmin(pylonRole{Name: "Billing Admin"}))
	// An unresolved role (zero value: role_id not in the catalogue) is not admin.
	assert.False(t, pylonIsAdmin(pylonRole{}))
}
