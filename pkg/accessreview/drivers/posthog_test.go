// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestPostHogDriverListAccounts(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/organizations/@current/members/", r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("limit"))
		assert.Equal(t, "-joined_at", r.URL.Query().Get("order"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{
  "next": "` + srv.URL + `/api/organizations/@current/members/?limit=100&order=-joined_at&offset=2",
  "results": [
    {
      "id": "membership-1",
      "user": {
        "uuid": "user-1",
        "first_name": "Olivia",
        "last_name": "Owner",
        "email": "owner@example.com"
      },
      "level": 15,
      "is_2fa_enabled": true,
      "joined_at": "2025-01-10T12:00:00Z",
      "last_login": "2025-05-01T09:00:00Z"
    },
    {
      "id": "membership-2",
      "user": {
        "uuid": "user-2",
        "first_name": "Maya",
        "last_name": "Member",
        "email": "member@example.com"
      },
      "level": 1,
      "is_2fa_enabled": false
    }
  ]
}`))
		case "2":
			_, _ = w.Write([]byte(`{
  "next": "",
  "results": [
    {
      "id": "membership-3",
      "user": {
        "uuid": "",
        "first_name": "Ari",
        "last_name": "Admin",
        "email": "admin@example.com"
      },
      "level": 8,
      "joined_at": "2024-02-20T00:00:00Z"
    }
  ]
}`))
		default:
			t.Fatalf("unexpected offset query value %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: &hostRewriter{target: srv.URL},
	}

	records, err := NewPostHogDriver(client).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "Olivia Owner", owner.FullName)
	assert.Equal(t, "Owner", owner.Role)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, owner.MFAStatus)
	assert.Equal(t, "user-1", owner.ExternalID)
	require.NotNil(t, owner.CreatedAt)
	require.NotNil(t, owner.LastLogin)

	member := records[1]
	assert.Equal(t, "member@example.com", member.Email)
	assert.Equal(t, "Member", member.Role)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, member.MFAStatus)
	assert.Nil(t, member.CreatedAt)
	assert.Nil(t, member.LastLogin)

	admin := records[2]
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "Admin", admin.Role)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.Equal(t, "membership-3", admin.ExternalID)
	require.NotNil(t, admin.CreatedAt)
}

func TestPostHogRoleFallback(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Owner", posthogRole(15, ""))
	assert.Equal(t, "Admin", posthogRole(8, ""))
	assert.Equal(t, "Member", posthogRole(1, ""))
	assert.Equal(t, "engineering", posthogRole(0, "engineering"))
	assert.Equal(t, "Member", posthogRole(0, ""))
}
