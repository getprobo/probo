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
)

func TestHubSpotDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/hubspot", "HUBSPOT_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HUBSPOT_TOKEN")))
	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Active super-admin with a resolved role.
	active := records[0]
	assert.Equal(t, "10000001", active.ExternalID)
	assert.Equal(t, "john.smith@example.com", active.Email)
	assert.Equal(t, "John Smith", active.FullName)
	assert.Equal(t, []string{"Sales Admin", "Super Admin"}, active.Roles)
	assert.Equal(t, new(true), active.IsAdmin)
	require.NotNil(t, active.Active)
	assert.True(t, *active.Active)
	require.NotNil(t, active.LastLogin)
	assert.Equal(t, time.Date(2025, 2, 7, 14, 12, 13, 544000000, time.UTC), active.LastLogin.UTC())

	// Still listed in Settings Users, but archived as an owner → inactive.
	listedInactive := records[1]
	assert.Equal(t, "10000002", listedInactive.ExternalID)
	assert.Equal(t, "jane.doe@example.com", listedInactive.Email)
	assert.Equal(t, "Jane Doe", listedInactive.FullName)
	assert.Equal(t, new(false), listedInactive.IsAdmin)
	require.NotNil(t, listedInactive.Active)
	assert.False(t, *listedInactive.Active)
	// Failed login is ignored; successful older login wins.
	require.NotNil(t, listedInactive.LastLogin)
	assert.Equal(t, time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC), listedInactive.LastLogin.UTC())

	// Archived-only owner (missing from Settings Users); empty name → email.
	archivedOnly := records[2]
	assert.Equal(t, "10000003", archivedOnly.ExternalID)
	assert.Equal(t, "gone@example.com", archivedOnly.Email)
	assert.Equal(t, "gone@example.com", archivedOnly.FullName)
	require.NotNil(t, archivedOnly.Active)
	assert.False(t, *archivedOnly.Active)
	assert.Nil(t, archivedOnly.LastLogin)
}

func TestHubSpotDriverOwnersRequired(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/hubspot_owners_forbidden", "HUBSPOT_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HUBSPOT_TOKEN")))
	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fetch hubspot owners")
}

func TestHubSpotDriverLoginsOptional(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/hubspot_logins_forbidden", "HUBSPOT_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HUBSPOT_TOKEN")))
	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "10000001", records[0].ExternalID)
	assert.Nil(t, records[0].LastLogin)
}

func TestHubSpotDriverEmailMatchDoesNotDuplicate(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/hubspot_email_match", "HUBSPOT_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HUBSPOT_TOKEN")))
	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "settings-1", records[0].ExternalID)
	assert.Equal(t, "Jane.Doe@example.com", records[0].Email)
	assert.Equal(t, []string{"Super Admin"}, records[0].Roles)
	require.NotNil(t, records[0].Active)
	assert.False(t, *records[0].Active)
}

func TestHubSpotOwnerUserID(t *testing.T) {
	t.Parallel()

	inactiveID := int64(9685555)
	activeID := int64(123)

	assert.Equal(t, "9685555", hubspotOwnerUserID(hubspotOwner{
		UserIDIncludingInactive: &inactiveID,
	}))
	assert.Equal(t, "123", hubspotOwnerUserID(hubspotOwner{
		UserID: &activeID,
	}))
	assert.Equal(t, "", hubspotOwnerUserID(hubspotOwner{}))

	zero := int64(0)
	assert.Equal(t, "", hubspotOwnerUserID(hubspotOwner{
		UserIDIncludingInactive: &zero,
	}))
}

func TestHubSpotApplyLastLogins(t *testing.T) {
	t.Parallel()

	byUserID := map[string]time.Time{
		"1": time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	byEmail := map[string]time.Time{
		"alice@example.com": time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		"bob@example.com":   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	records := []AccountRecord{
		{ExternalID: "1", Email: "alice@example.com"},
		{ExternalID: "2", Email: "Bob@Example.com"},
		{ExternalID: "3", Email: "nobody@example.com"},
	}

	hubspotApplyLastLogins(records, byUserID, byEmail)

	require.NotNil(t, records[0].LastLogin)
	assert.Equal(t, byUserID["1"], records[0].LastLogin.UTC())
	require.NotNil(t, records[1].LastLogin)
	assert.Equal(t, byEmail["bob@example.com"], records[1].LastLogin.UTC())
	assert.Nil(t, records[2].LastLogin)
}

func TestHubSpotParseTime(t *testing.T) {
	t.Parallel()

	t.Run("fractional seconds", func(t *testing.T) {
		t.Parallel()

		got, ok := hubspotParseTime("2025-02-07T14:12:13.544Z")
		require.True(t, ok)
		assert.Equal(t, time.Date(2025, 2, 7, 14, 12, 13, 544000000, time.UTC), got.UTC())
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		_, ok := hubspotParseTime("")
		assert.False(t, ok)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, ok := hubspotParseTime("not-a-time")
		assert.False(t, ok)
	})
}

func TestHubSpotRoles(t *testing.T) {
	t.Parallel()

	roleMap := map[string]string{
		"role-1": "Sales Admin",
		"role-2": "Marketing Admin",
	}

	tests := []struct {
		name string
		user hubspotUser
		want []string
	}{
		{
			name: "multiple role IDs",
			user: hubspotUser{RoleIDs: []string{"role-1", "role-2"}},
			want: []string{"Sales Admin", "Marketing Admin"},
		},
		{
			name: "roleId and roleIds merged without duplicates",
			user: hubspotUser{RoleID: "role-1", RoleIDs: []string{"role-1", "role-2"}},
			want: []string{"Sales Admin", "Marketing Admin"},
		},
		{
			name: "unknown role falls back to user",
			user: hubspotUser{RoleIDs: []string{"missing"}},
			want: []string{"User"},
		},
		{
			name: "unknown role with super admin",
			user: hubspotUser{RoleIDs: []string{"missing"}, SuperAdmin: true},
			want: []string{"Super Admin"},
		},
		{
			name: "known role merged with super admin",
			user: hubspotUser{RoleIDs: []string{"role-1"}, SuperAdmin: true},
			want: []string{"Sales Admin", "Super Admin"},
		},
		{
			name: "multiple roles merged with super admin",
			user: hubspotUser{RoleIDs: []string{"role-1", "role-2"}, SuperAdmin: true},
			want: []string{"Sales Admin", "Marketing Admin", "Super Admin"},
		},
		{
			name: "no roles defaults to user",
			user: hubspotUser{},
			want: []string{"User"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, hubspotRoles(tt.user, roleMap))
		})
	}
}
