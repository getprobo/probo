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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

func TestBetterStackDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/better_stack", "BETTER_STACK_TOKEN")

	teamName := os.Getenv("BETTER_STACK_TEAM_NAME")
	if teamName == "" {
		teamName = "acme"
	}

	client := newVCRClient(rec, bearerAuth(os.Getenv("BETTER_STACK_TOKEN")))

	records, err := NewBetterStackDriver(client, teamName).ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Len(t, records, 1)

	r := records[0]
	assert.Equal(t, "alice@example.com", r.Email)
	assert.Equal(t, "Alice Smith", r.FullName)
	assert.Equal(t, "Admin", r.Role)
	assert.True(t, r.IsAdmin)
	assert.Equal(t, "101", r.ExternalID)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	require.NotNil(t, r.CreatedAt)
	assert.Equal(t, coredata.MFAStatusUnknown, r.MFAStatus)
}

func TestBetterStackDriverPagination(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v2/team-members", r.URL.Path)
		assert.Equal(t, "acme", r.URL.Query().Get("team_name"))

		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"data":[{"id":"201","type":"team_member_invitation","attributes":{"email":"charlie@example.com","invited_at":"2023-10-28T12:00:00.000Z","role":"member"}}],"pagination":{"next":null}}`))

			return
		}

		_, _ = w.Write([]byte(`{"data":[{"id":"101","type":"team_member","attributes":{"email":"alice@example.com","first_name":"Alice","last_name":"Smith","created_at":"2023-10-26T10:00:00.000Z","role":"admin"}},{"id":"102","type":"team_member","attributes":{"email":"bob@example.com","first_name":"Bob","last_name":"","created_at":"2023-10-27T11:00:00.000Z","role":"responder"}}],"pagination":{"next":"https://betterstack.com/api/v2/team-members?page=2&team_name=acme"}}`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	records, err := NewBetterStackDriver(client, "acme").ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice Smith", records[0].FullName)
	assert.True(t, *records[0].Active)
	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, "Bob", records[1].FullName)
	assert.Equal(t, "charlie@example.com", records[2].Email)
	require.NotNil(t, records[2].Active)
	assert.False(t, *records[2].Active)
}

func TestBetterStackDriverListAccountsError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":"invalid token"}`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	_, err := NewBetterStackDriver(client, "acme").ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestBetterStackRole(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    string
		isAdmin bool
	}{
		{"admin", "Admin", true},
		{"billing_admin", "Billing admin", false},
		{"team_lead", "Team lead", true},
		{"responder", "Responder", false},
		{"member", "Member", false},
		{"custom", "custom", false},
		{"future_role", "future_role", false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, betterStackRole(c.in))
			assert.Equal(t, c.isAdmin, betterStackIsAdmin(c.in))
		})
	}
}
