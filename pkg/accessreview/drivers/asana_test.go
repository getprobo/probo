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
	require.NotEmpty(t, records)

	byEmail := map[string]AccountRecord{}
	for _, r := range records {
		byEmail[r.Email] = r
	}

	admin := byEmail["admin@example.com"]
	assert.Equal(t, "Ada Admin", admin.FullName)
	assert.Equal(t, "1000000000000001", admin.ExternalID)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	require.NotNil(t, admin.CreatedAt)
	assert.Equal(t, "2020-01-15T12:00:00Z", admin.CreatedAt.UTC().Format(time.RFC3339))

	member := byEmail["member@example.com"]
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)

	guest := byEmail["guest@example.com"]
	assert.Equal(t, []string{"Guest"}, guest.Roles)
	assert.False(t, guest.IsAdmin)

	viewer := byEmail["viewer@example.com"]
	assert.Equal(t, []string{"View only"}, viewer.Roles)
	assert.False(t, viewer.IsAdmin)

	inactive := byEmail["inactive@example.com"]
	require.NotNil(t, inactive.Active)
	assert.False(t, *inactive.Active)
}

func TestAsanaRoles(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		isAdmin    bool
		isGuest    bool
		isViewOnly bool
		want       []string
	}{
		{name: "admin", isAdmin: true, want: []string{"Admin"}},
		{name: "guest", isGuest: true, want: []string{"Guest"}},
		{name: "view only", isViewOnly: true, want: []string{"View only"}},
		{name: "member", want: []string{"Member"}},
		{name: "admin wins over guest", isAdmin: true, isGuest: true, want: []string{"Admin"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, asanaRoles(tc.isAdmin, tc.isGuest, tc.isViewOnly))
		})
	}
}
