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
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestNeonDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/neon", "NEON_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("NEON_API_KEY")))

	orgID := os.Getenv("NEON_ORG_ID")
	if orgID == "" {
		orgID = "org-cool-breeze-12345678"
	}

	driver := NewNeonDriver(client, orgID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// Four members across two pages in the cassette; the fourth has no
	// email and is dropped.
	require.Len(t, records, 3)

	// Admin with MFA enabled and no deactivated_at. Neon exposes no
	// display name on the members endpoint, so FullName is the email;
	// ExternalID is the stable account UUID (member.user_id).
	assert.Equal(t, "jane.doe@example.com", records[0].Email)
	assert.Equal(t, "jane.doe@example.com", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, "bbbbbbbb-1111-2222-3333-000000000001", records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].CreatedAt)
	assert.Nil(t, records[0].LastLogin)

	// Deactivated member with MFA disabled.
	assert.Equal(t, "john.smith@example.com", records[1].Email)
	assert.Equal(t, []string{"Member"}, records[1].Roles)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, records[1].MFAStatus)
	assert.Equal(t, "bbbbbbbb-1111-2222-3333-000000000002", records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	// Second page: editor with has_mfa omitted (Unknown) and an empty
	// user_id falling back to the membership ID.
	assert.Equal(t, "erin.lee@example.com", records[2].Email)
	assert.Equal(t, []string{"Editor"}, records[2].Roles)
	assert.False(t, records[2].IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
	assert.Equal(t, "aaaaaaaa-1111-2222-3333-000000000003", records[2].ExternalID)
	require.NotNil(t, records[2].Active)
	assert.True(t, *records[2].Active)
}

func TestNeonDriverListAccountsError(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"error":"unauthorized"}`)),
					Header:     make(http.Header),
				}, nil
			},
		),
	}

	driver := NewNeonDriver(client, "org-cool-breeze-12345678")
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestNeonRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    []string
		isAdmin bool
	}{
		{in: "admin", want: []string{"Admin"}, isAdmin: true},
		{in: "member", want: []string{"Member"}, isAdmin: false},
		{in: "editor", want: []string{"Editor"}, isAdmin: false},
		{in: "viewer", want: []string{"Viewer"}, isAdmin: false},
		{in: "collaborator", want: []string{"Collaborator"}, isAdmin: false},
		{in: "future_role", want: []string{"future_role"}, isAdmin: false},
		{in: "", want: []string{}, isAdmin: false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, neonRoles(c.in))
			assert.Equal(t, c.isAdmin, neonIsAdmin(c.in))
		})
	}
}
