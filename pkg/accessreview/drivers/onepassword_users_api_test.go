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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveOnePasswordUsers answers the list-users endpoint of ACCOUNT123 with the
// page each page_token names, and 404s anything else.
func serveOnePasswordUsers(t *testing.T, pages map[string]string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				page, ok := pages[r.URL.Query().Get("page_token")]
				if r.URL.Path != "/v1/accounts/ACCOUNT123/users" || !ok {
					w.WriteHeader(http.StatusNotFound)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(page))
			},
		),
	)
	t.Cleanup(server.Close)

	return server
}

func TestOnePasswordUsersAPIDriver_ListAccounts(t *testing.T) {
	t.Parallel()

	t.Run(
		"follows next_page_token across pages",
		func(t *testing.T) {
			t.Parallel()

			server := serveOnePasswordUsers(
				t,
				map[string]string{
					"":       `{"results":[{"id":"u1","email":"alice@example.com","display_name":"Alice","state":"ACTIVE","create_time":"2024-04-10T08:09:46Z","path":"accounts/ACCOUNT123/users/u1"}],"next_page_token":"CAIQAg"}`,
					"CAIQAg": `{"results":[{"id":"u2","email":"bob@example.com","display_name":"Bob","state":"SUSPENDED","create_time":"2024-05-01T12:00:00Z","path":"accounts/ACCOUNT123/users/u2"}]}`,
				},
			)

			driver := NewOnePasswordUsersAPIDriver(server.Client(), server.URL, "ACCOUNT123")

			records, err := driver.ListAccounts(context.Background())
			require.NoError(t, err)
			require.Len(t, records, 2)

			assert.Equal(t, "alice@example.com", records[0].Email)
			assert.Equal(t, "Alice", records[0].FullName)
			assert.Equal(t, "u1", records[0].ExternalID)
			assert.Equal(t, new(true), records[0].Active)
			require.NotNil(t, records[0].CreatedAt)
			assert.Equal(t, time.Date(2024, 4, 10, 8, 9, 46, 0, time.UTC), records[0].CreatedAt.UTC())

			assert.Equal(t, "bob@example.com", records[1].Email)
			assert.Equal(t, new(false), records[1].Active)
		},
	)

	t.Run(
		"empty account",
		func(t *testing.T) {
			t.Parallel()

			// A page with no users is an empty object, with no results key.
			server := serveOnePasswordUsers(t, map[string]string{"": `{}`})

			driver := NewOnePasswordUsersAPIDriver(server.Client(), server.URL, "ACCOUNT123")

			records, err := driver.ListAccounts(context.Background())
			require.NoError(t, err)
			assert.Empty(t, records)
		},
	)

	t.Run(
		"state other than active or suspended is unknown",
		func(t *testing.T) {
			t.Parallel()

			server := serveOnePasswordUsers(
				t,
				map[string]string{
					"": `{"results":[{"id":"u1","email":"alice@example.com"},{"id":"u2","email":"bob@example.com","state":"LOCKED"}]}`,
				},
			)

			driver := NewOnePasswordUsersAPIDriver(server.Client(), server.URL, "ACCOUNT123")

			records, err := driver.ListAccounts(context.Background())
			require.NoError(t, err)
			require.Len(t, records, 2)
			assert.Nil(t, records[0].Active)
			assert.Nil(t, records[1].Active)
		},
	)

	t.Run(
		"requests the largest page",
		func(t *testing.T) {
			t.Parallel()

			var pageSize string

			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						pageSize = r.URL.Query().Get("max_page_size")
						_, _ = w.Write([]byte(`{}`))
					},
				),
			)
			t.Cleanup(server.Close)

			driver := NewOnePasswordUsersAPIDriver(server.Client(), server.URL, "ACCOUNT123")

			_, err := driver.ListAccounts(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "1000", pageSize)
		},
	)

	t.Run(
		"rejected token",
		func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
					},
				),
			)
			t.Cleanup(server.Close)

			driver := NewOnePasswordUsersAPIDriver(server.Client(), server.URL, "ACCOUNT123")

			_, err := driver.ListAccounts(context.Background())
			require.Error(t, err)
			assert.ErrorContains(t, err, "unexpected status 401")
		},
	)
}

func TestOnePasswordUsersURL_EscapesAccountID(t *testing.T) {
	t.Parallel()

	endpoint, err := OnePasswordUsersURL("https://api.1password.com", "ACCOUNT/../../v1/oauth")
	require.NoError(t, err)
	assert.Equal(t, "https://api.1password.com/v1/accounts/ACCOUNT%2F..%2F..%2Fv1%2Foauth/users", endpoint.String())
}

func TestOnePasswordRegionHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		region   string
		baseURL  string
		tokenURL string
	}{
		{region: "US", baseURL: "https://api.1password.com", tokenURL: "https://api.1password.com/v1/oauth/token"},
		{region: "CA", baseURL: "https://api.1password.ca", tokenURL: "https://api.1password.ca/v1/oauth/token"},
		{region: "EU", baseURL: "https://api.1password.eu", tokenURL: "https://api.1password.eu/v1/oauth/token"},
		{region: "ENT", baseURL: "https://api.ent.1password.com", tokenURL: "https://api.ent.1password.com/v1/oauth/token"},
		{region: " eu ", baseURL: "https://api.1password.eu", tokenURL: "https://api.1password.eu/v1/oauth/token"},
		{region: "United States", baseURL: "https://api.1password.com", tokenURL: "https://api.1password.com/v1/oauth/token"},
	}

	for _, tt := range tests {
		t.Run(
			tt.region,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.baseURL, OnePasswordUsersAPIBaseURL(tt.region))
				assert.Equal(t, tt.tokenURL, OnePasswordTokenURL(tt.region))
			},
		)
	}
}

func TestIsOnePasswordRegion(t *testing.T) {
	t.Parallel()

	for _, region := range []string{"US", "CA", "EU", "ENT"} {
		assert.True(t, IsOnePasswordRegion(region), region)
	}

	for _, region := range []string{"", "us", "ent", "AU", "United States"} {
		assert.False(t, IsOnePasswordRegion(region), region)
	}
}
