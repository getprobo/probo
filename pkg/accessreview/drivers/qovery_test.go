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
	assert.Equal(t, []string{"Owner"}, records[0].Roles)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, "google-oauth2|100000000000000000001", records[0].ExternalID)
	require.NotNil(t, records[0].LastLogin)
	require.NotNil(t, records[0].CreatedAt)
	assert.Nil(t, records[0].Active)

	// Developer: empty name falls back to nickname; null last_activity_at
	// leaves LastLogin nil.
	assert.Equal(t, "john.smith@example.com", records[1].Email)
	assert.Equal(t, "john", records[1].FullName)
	assert.Equal(t, []string{"Developer"}, records[1].Roles)
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

func TestQoveryRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    []string
		isAdmin bool
	}{
		{in: "OWNER", want: []string{"Owner"}, isAdmin: true},
		{in: "ADMIN", want: []string{"Admin"}, isAdmin: true},
		{in: "DEVELOPER", want: []string{"Developer"}, isAdmin: false},
		{in: "VIEWER", want: []string{"Viewer"}, isAdmin: false},
		{in: "future_role", want: []string{"future_role"}, isAdmin: false},
		{in: "", want: []string{}, isAdmin: false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, qoveryRoles(c.in))
			assert.Equal(t, c.isAdmin, qoveryIsAdmin(c.in))
		})
	}
}
