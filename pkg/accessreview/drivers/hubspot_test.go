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
	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
}

func TestHubSpotDriverArchivedUsers(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch {
				case req.URL.Path == "/settings/v3/users/roles":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"role-1","name":"Sales Admin"}]}`,
					), nil
				case req.URL.Path == "/settings/v3/users":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"10000001","email":"active@example.com","firstName":"Active","lastName":"User","roleIds":["role-1"],"superAdmin":false},{"id":"10000002","email":"still-listed@example.com","firstName":"Listed","lastName":"Inactive","roleIds":[],"superAdmin":false}]}`,
					), nil
				case req.URL.Path == "/crm/v3/owners" && req.URL.Query().Get("archived") == "true":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"owner-2","email":"still-listed@example.com","firstName":"Listed","lastName":"Inactive","type":"PERSON","archived":true,"userId":null,"userIdIncludingInactive":10000002},{"id":"owner-3","email":"gone@example.com","firstName":"","lastName":"","type":"PERSON","archived":true,"userId":null,"userIdIncludingInactive":10000003}]}`,
					), nil
				default:
					return hubspotResponse(http.StatusNotFound, `{"message":"not found"}`), nil
				}
			},
		),
	}

	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, "active@example.com", records[0].Email)
	assert.Equal(t, []string{"Sales Admin"}, records[0].Roles)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)

	assert.Equal(t, "10000002", records[1].ExternalID)
	assert.Equal(t, "still-listed@example.com", records[1].Email)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	// Archived-only owner: empty HubSpot name falls back to email, Active=false.
	assert.Equal(t, "10000003", records[2].ExternalID)
	assert.Equal(t, "gone@example.com", records[2].Email)
	assert.Equal(t, "gone@example.com", records[2].FullName)
	require.NotNil(t, records[2].Active)
	assert.False(t, *records[2].Active)
}

func TestHubSpotDriverEmailMatchDoesNotDuplicate(t *testing.T) {
	t.Parallel()

	// Settings user ID differs from owner userIdIncludingInactive, but the
	// email matches. The driver must mark the settings row inactive and must
	// not emit a second archived-only row for the same person.
	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch {
				case req.URL.Path == "/settings/v3/users/roles":
					return hubspotResponse(http.StatusOK, `{"results":[]}`), nil
				case req.URL.Path == "/settings/v3/users":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"settings-1","email":"Jane.Doe@Example.com","firstName":"Jane","lastName":"Doe","roleIds":[],"superAdmin":true}]}`,
					), nil
				case req.URL.Path == "/crm/v3/owners" && req.URL.Query().Get("archived") == "true":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"owner-9","email":"jane.doe@example.com","firstName":"","lastName":"","type":"PERSON","archived":true,"userId":null,"userIdIncludingInactive":999}]}`,
					), nil
				default:
					return hubspotResponse(http.StatusNotFound, `{"message":"not found"}`), nil
				}
			},
		),
	}

	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "settings-1", records[0].ExternalID)
	assert.Equal(t, "Jane.Doe@Example.com", records[0].Email)
	assert.Equal(t, []string{"Super Admin"}, records[0].Roles)
	require.NotNil(t, records[0].Active)
	assert.False(t, *records[0].Active)
}

func TestHubSpotDriverOwnersRequired(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/settings/v3/users/roles":
					return hubspotResponse(http.StatusOK, `{"results":[]}`), nil
				case "/settings/v3/users":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"1","email":"a@example.com","firstName":"A","lastName":"B","roleIds":[],"superAdmin":false}]}`,
					), nil
				case "/crm/v3/owners":
					return hubspotResponse(http.StatusForbidden, `{"message":"missing scope"}`), nil
				default:
					return hubspotResponse(http.StatusNotFound, `{"message":"not found"}`), nil
				}
			},
		),
	}

	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fetch hubspot owners")
}

func TestHubSpotDriverSkipsNonPersonOwners(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				switch {
				case req.URL.Path == "/settings/v3/users/roles":
					return hubspotResponse(http.StatusOK, `{"results":[]}`), nil
				case req.URL.Path == "/settings/v3/users":
					return hubspotResponse(http.StatusOK, `{"results":[]}`), nil
				case req.URL.Path == "/crm/v3/owners" && req.URL.Query().Get("archived") == "true":
					return hubspotResponse(
						http.StatusOK,
						`{"results":[{"id":"queue-1","email":"queue@example.com","type":"QUEUE","archived":true,"userId":null,"userIdIncludingInactive":null},{"id":"owner-1","email":"person@example.com","firstName":"P","lastName":"erson","type":"PERSON","archived":true,"userId":null,"userIdIncludingInactive":42}]}`,
					), nil
				default:
					return hubspotResponse(http.StatusNotFound, `{"message":"not found"}`), nil
				}
			},
		),
	}

	driver := NewHubSpotDriver(client, "https://api.hubapi.com")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "42", records[0].ExternalID)
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
