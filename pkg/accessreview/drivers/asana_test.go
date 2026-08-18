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
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAsanaDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/asana", "ASANA_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("ASANA_TOKEN")))

	workspaceGID := os.Getenv("ASANA_WORKSPACE_GID")
	if workspaceGID == "" {
		workspaceGID = "9999999"
	}

	driver := NewAsanaDriver(client, workspaceGID, "https://app.asana.com/api/1.0")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	admin := records[0]
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "Ada Admin", admin.FullName)
	assert.Equal(t, "1000000000000001", admin.ExternalID)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, admin.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	require.NotNil(t, admin.IsAdmin)
	assert.True(t, *admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	require.NotNil(t, admin.CreatedAt)
	assert.Equal(t, "2026-05-15T13:12:35Z", admin.CreatedAt.UTC().Format(time.RFC3339))
}

func TestAsanaRoles(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		isAdmin    *bool
		isGuest    *bool
		isViewOnly *bool
		want       []string
	}{
		{name: "admin", isAdmin: new(true), isGuest: new(false), isViewOnly: new(false), want: []string{"Admin"}},
		{name: "guest", isAdmin: new(false), isGuest: new(true), isViewOnly: new(false), want: []string{"Guest"}},
		{name: "view only", isAdmin: new(false), isGuest: new(false), isViewOnly: new(true), want: []string{"View only"}},
		{name: "member", isAdmin: new(false), isGuest: new(false), isViewOnly: new(false), want: []string{"Member"}},
		{name: "admin wins over guest", isAdmin: new(true), isGuest: new(true), isViewOnly: new(false), want: []string{"Admin"}},
		// A set flag classifies even when another is missing.
		{name: "partial signal still classifies", isAdmin: new(false), isGuest: new(true), want: []string{"Guest"}},
		// "Member" means none of the three, so a withheld flag blocks it:
		// the account may be exactly the thing Asana did not report.
		{name: "member unreachable without is_admin", isGuest: new(false), isViewOnly: new(false), want: nil},
		{name: "member unreachable without is_guest", isAdmin: new(false), isViewOnly: new(false), want: nil},
		{name: "no signal at all", want: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, asanaRoles(tc.isAdmin, tc.isGuest, tc.isViewOnly))
		})
	}
}

// TestAsanaMembershipAbsentFlagsStayUnknown pins the decode boundary where the
// shipped bug lived: this endpoint returns WorkspaceMembershipCompact and adds
// the privilege flags only because opt_fields asks for them, so a plain bool
// would turn a withheld flag into an observed false.
func TestAsanaMembershipAbsentFlagsStayUnknown(t *testing.T) {
	t.Parallel()

	var page asanaMembershipsPage
	require.NoError(t, json.Unmarshal(
		[]byte(`{"data":[{"gid":"1","user":{"gid":"2","name":"Ada","email":"ada@example.com"}}]}`),
		&page,
	))
	require.Len(t, page.Data, 1)

	m := page.Data[0]
	assert.Nil(t, m.IsAdmin)
	assert.Nil(t, m.IsGuest)
	assert.Nil(t, m.IsViewOnly)
	assert.Nil(t, m.IsActive)
	assert.Nil(t, asanaRoles(m.IsAdmin, m.IsGuest, m.IsViewOnly))
}
