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

func TestQoveryDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/qovery", "QOVERY_API_TOKEN")

	authValue := ""
	if token := os.Getenv("QOVERY_API_TOKEN"); token != "" {
		authValue = "Token " + token
	}

	client := newVCRClient(rec, authValue)

	orgID := os.Getenv("QOVERY_ORG_ID")
	if orgID == "" {
		orgID = "11111111-2222-3333-4444-555555555555"
	}

	driver := NewQoveryDriver(client, orgID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// Three members in the cassette; the third has no email and is dropped.
	require.Len(t, records, 2)

	// Owner: built-in OWNER role → admin; name + both timestamps populated.
	// Qovery member IDs are the IdP subject (e.g. "google-oauth2|<sub>").
	assert.Equal(t, "jane.doe@example.com", records[0].Email)
	assert.Equal(t, "Jane Doe", records[0].FullName)
	assert.Equal(t, "Owner", records[0].Role)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, "google-oauth2|100000000000000000001", records[0].ExternalID)
	require.NotNil(t, records[0].LastLogin)
	require.NotNil(t, records[0].CreatedAt)
	assert.Nil(t, records[0].Active)

	// Developer: empty name falls back to nickname; null last_activity_at
	// leaves LastLogin nil.
	assert.Equal(t, "john.smith@example.com", records[1].Email)
	assert.Equal(t, "john", records[1].FullName)
	assert.Equal(t, "Developer", records[1].Role)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, "google-oauth2|100000000000000000002", records[1].ExternalID)
	assert.Nil(t, records[1].LastLogin)
	require.NotNil(t, records[1].CreatedAt)
}

func TestQoveryDriverListAccountsError(t *testing.T) {
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

	driver := NewQoveryDriver(client, "26ac87db-ae79-4be4-bd33-7f839f0e1647")
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestQoveryRole(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    string
		isAdmin bool
	}{
		{in: "OWNER", want: "Owner", isAdmin: true},
		{in: "ADMIN", want: "Admin", isAdmin: true},
		{in: "DEVELOPER", want: "Developer", isAdmin: false},
		{in: "VIEWER", want: "Viewer", isAdmin: false},
		{in: "future_role", want: "future_role", isAdmin: false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, qoveryRole(c.in))
			assert.Equal(t, c.isAdmin, qoveryIsAdmin(c.in))
		})
	}
}
