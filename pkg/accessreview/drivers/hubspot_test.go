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
)

func TestHubSpotDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/hubspot", "HUBSPOT_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HUBSPOT_TOKEN")))
	driver := NewHubSpotDriver(client)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
}

func TestHubSpotDriverArchivedUsers(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/settings/v3/users/roles":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"role-1","name":"Sales Admin"}]}`,
					), nil
				case "/settings/v3/users":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"user-1","email":"active@example.com","firstName":"Active","lastName":"User","roleIds":["role-1"],"superAdmin":false,"isActive":true},{"id":"user-2","email":"","firstName":"Archived","lastName":"User","superAdmin":false,"archived":true}]}`,
					), nil
				default:
					return hubspotResponse(http.StatusNotFound, `{"message":"not found"}`), nil
				}
			},
		),
	}

	driver := NewHubSpotDriver(client)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, []string{"Sales Admin"}, records[0].Roles)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)

	assert.Equal(t, "user-2", records[1].ExternalID)
	assert.Empty(t, records[1].Email)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
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

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func hubspotResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
