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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Merging by lowercased email is the whole contract of the Google Analytics
// driver: one user holding an account-level and a property-level binding must
// collapse to a single reviewed account. Cassette fixtures are all already
// lowercase, so the normalization only shows up here.
func TestAddGoogleAnalyticsBinding(t *testing.T) {
	t.Parallel()

	t.Run("merges casing and whitespace variants of one email", func(t *testing.T) {
		t.Parallel()

		members := map[string]*googleAnalyticsMember{}

		addGoogleAnalyticsBinding(members, "Ada@Example.com", []string{"predefinedRoles/viewer"})
		addGoogleAnalyticsBinding(members, "  ada@example.com  ", []string{"predefinedRoles/analyst"})

		require.Len(t, members, 1)

		records := googleAnalyticsRecords(members)
		require.Len(t, records, 1)
		assert.Equal(t, "ada@example.com", records[0].Email)
		assert.Equal(t, []string{"analyst", "viewer"}, records[0].Roles)
		assert.False(t, records[0].IsAdmin)
	})

	t.Run("an admin role anywhere wins", func(t *testing.T) {
		t.Parallel()

		members := map[string]*googleAnalyticsMember{}

		addGoogleAnalyticsBinding(members, "ada@example.com", []string{"predefinedRoles/viewer"})
		addGoogleAnalyticsBinding(members, "ada@example.com", []string{googleAnalyticsAdminRole})

		records := googleAnalyticsRecords(members)
		require.Len(t, records, 1)
		assert.True(t, records[0].IsAdmin)
	})

	t.Run("skips empty emails and blank roles", func(t *testing.T) {
		t.Parallel()

		members := map[string]*googleAnalyticsMember{}

		addGoogleAnalyticsBinding(members, "   ", []string{"predefinedRoles/viewer"})
		addGoogleAnalyticsBinding(members, "ada@example.com", []string{"", "  "})

		require.Len(t, members, 1)

		records := googleAnalyticsRecords(members)
		require.Len(t, records, 1)
		assert.Empty(t, records[0].Roles)
	})
}

// The probe must target the account's accessBindings, not the accounts list,
// so a connection that can read accounts but not bindings reports disconnected.
func TestGoogleAnalyticsAccountBindingsProbeURL(t *testing.T) {
	t.Parallel()

	t.Run("targets the account bindings with a single-item page", func(t *testing.T) {
		t.Parallel()

		got, err := GoogleAnalyticsAccountBindingsProbeURL("123456")
		require.NoError(t, err)
		assert.Equal(
			t,
			"https://analyticsadmin.googleapis.com/v1alpha/accounts/123456/accessBindings?pageSize=1",
			got,
		)
	})

	t.Run("escapes the account ID", func(t *testing.T) {
		t.Parallel()

		got, err := GoogleAnalyticsAccountBindingsProbeURL("12/34")
		require.NoError(t, err)
		assert.Contains(t, got, "/accounts/12%2F34/accessBindings")
	})
}

func TestDotfileIsAdmin(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		role string
		want bool
	}{
		{role: "owner", want: true},
		{role: "admin", want: true},
		{role: "OWNER", want: true},
		{role: "  Admin  ", want: true},
		{role: "member", want: false},
		{role: "", want: false},
	} {
		t.Run(tc.role, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, dotfileIsAdmin(tc.role))
		})
	}
}

func TestSegmentRolesAndAdmin(t *testing.T) {
	t.Parallel()

	t.Run("deduplicates and sorts role names", func(t *testing.T) {
		t.Parallel()

		roles, isAdmin := segmentRolesAndAdmin([]segmentPermission{
			{RoleName: "Source Admin"},
			{RoleName: "Source Admin"},
			{RoleName: "Read-only"},
			{RoleName: "  "},
		})

		assert.Equal(t, []string{"Read-only", "Source Admin"}, roles)
		assert.False(t, isAdmin)
	})

	t.Run("matches the workspace owner role case-insensitively", func(t *testing.T) {
		t.Parallel()

		roles, isAdmin := segmentRolesAndAdmin([]segmentPermission{
			{RoleName: "workspace owner"},
		})

		assert.Equal(t, []string{"workspace owner"}, roles)
		assert.True(t, isAdmin)
	})

	t.Run("no permissions means no roles and no admin", func(t *testing.T) {
		t.Parallel()

		roles, isAdmin := segmentRolesAndAdmin(nil)

		assert.Empty(t, roles)
		assert.False(t, isAdmin)
	})
}
