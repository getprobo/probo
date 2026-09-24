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
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

const daytonaTestOrgID = "aaaaaaaa-1111-2222-3333-000000000001"

func TestDaytonaDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/daytona", "DAYTONA_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("DAYTONA_API_KEY")))

	orgID := os.Getenv("DAYTONA_ORG_ID")
	if orgID == "" {
		orgID = daytonaTestOrgID
	}

	driver := NewDaytonaDriver(client, orgID, "https://app.daytona.io/api")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "jane.doe@example.com", owner.Email)
	assert.Equal(t, "Jane Doe", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, "user-owner", owner.ExternalID)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	require.NotNil(t, owner.CreatedAt)

	member := records[1]
	assert.Equal(t, "john.smith@example.com", member.Email)
	assert.Equal(t, "John Smith", member.FullName)
	assert.Equal(t, []string{"Member", "Sandbox Admin"}, member.Roles)
	assert.Equal(t, new(false), member.IsAdmin)
	assert.Equal(t, "user-member", member.ExternalID)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	invite := records[2]
	assert.Equal(t, "erin.lee@example.com", invite.Email)
	assert.Equal(t, "erin.lee@example.com", invite.FullName)
	assert.Equal(t, []string{"Member"}, invite.Roles)
	assert.Equal(t, new(false), invite.IsAdmin)
	assert.Equal(t, "invite-pending", invite.ExternalID)
	require.NotNil(t, invite.Active)
	assert.False(t, *invite.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, invite.MFAStatus)
}

func TestDaytonaDriverListAccountsError(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)),
					Header:     make(http.Header),
				}, nil
			},
		),
	}

	driver := NewDaytonaDriver(client, daytonaTestOrgID, "https://app.daytona.io/api")
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestDaytonaRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		role     string
		assigned []daytonaAssignedRole
		want     []string
		isAdmin  bool
	}{
		{name: "owner", role: "owner", want: []string{"Owner"}, isAdmin: true},
		{name: "member", role: "member", want: []string{"Member"}, isAdmin: false},
		{
			name:     "member with assigned role",
			role:     "member",
			assigned: []daytonaAssignedRole{{Name: "Sandbox Admin"}},
			want:     []string{"Member", "Sandbox Admin"},
			isAdmin:  false,
		},
		{
			name:     "owner assigned role does not duplicate Owner",
			role:     "owner",
			assigned: []daytonaAssignedRole{{Name: "Owner"}},
			want:     []string{"Owner"},
			isAdmin:  true,
		},
		{name: "future role", role: "developer", want: []string{"developer"}, isAdmin: false},
		{name: "empty", role: "", want: []string{}, isAdmin: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, daytonaRoles(c.role, c.assigned))
			assert.Equal(t, c.isAdmin, daytonaIsAdmin(c.role))
		})
	}
}

func TestDaytonaNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("returns organization name", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{
			Transport: roundTripFunc(
				func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "/api/organizations/"+daytonaTestOrgID, req.URL.Path)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"name":"Acme"}`)),
						Header:     make(http.Header),
					}, nil
				},
			),
		}

		name, err := NewDaytonaNameResolver(client, daytonaTestOrgID, "https://app.daytona.io/api").
			ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Acme", name)
	})

	t.Run("gives up on a non-2xx", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{
			Transport: roundTripFunc(
				func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(strings.NewReader(`{"message":"not found"}`)),
						Header:     make(http.Header),
					}, nil
				},
			),
		}

		name, err := NewDaytonaNameResolver(client, daytonaTestOrgID, "https://app.daytona.io/api").
			ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, name)
	})
}
