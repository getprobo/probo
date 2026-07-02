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
