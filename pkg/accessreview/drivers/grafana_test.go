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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const grafanaCassetteHost = "grafana.example.com"

// TestGrafanaDriver records with GRAFANA_TOKEN (an Admin service account) and
// GRAFANA_BASE_URL against a throwaway Grafana seeded with the two users
// asserted below, then replays the cassette rebased onto grafanaCassetteHost.
func TestGrafanaDriver(t *testing.T) {
	t.Parallel()

	baseURL := "https://" + grafanaCassetteHost
	if os.Getenv("GRAFANA_TOKEN") != "" {
		baseURL = strings.TrimSuffix(os.Getenv("GRAFANA_BASE_URL"), "/")
		require.NotEmpty(t, baseURL, "GRAFANA_BASE_URL must be set when recording with GRAFANA_TOKEN")
	}

	rec := newRecorder(
		t,
		"testdata/grafana",
		"GRAFANA_TOKEN",
		func(i *cassette.Interaction) error {
			base, err := url.Parse(baseURL)
			if err != nil {
				return fmt.Errorf("cannot parse grafana base URL: %w", err)
			}

			u, err := url.Parse(i.Request.URL)
			if err != nil {
				return fmt.Errorf("cannot parse recorded grafana URL: %w", err)
			}

			u.Scheme = "https"
			u.Host = grafanaCassetteHost
			u.Path = strings.TrimPrefix(u.Path, base.Path)
			i.Request.URL = u.String()
			i.Request.Host = grafanaCassetteHost

			return nil
		},
	)
	client := newVCRClient(rec, bearerAuth(os.Getenv("GRAFANA_TOKEN")))

	driver := NewGrafanaDriver(client, baseURL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, strconv.Itoa(1), records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].LastLogin)

	assert.Equal(t, "viewer@example.com", records[1].Email)
	assert.Equal(t, "Viewer User", records[1].FullName)
	assert.Equal(t, []string{"Viewer"}, records[1].Roles)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Equal(t, strconv.Itoa(2), records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	resolver := NewGrafanaNameResolver(client, baseURL)
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Acme Grafana", name)
}

// TestGrafanaDriver_OrgUsersUnpaged mirrors Grafana's /api/org/users, which
// ignores page and perpage and returns every user on each call.
func TestGrafanaDriver_OrgUsersUnpaged(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32

	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, "/api/org/users", r.URL.Path)
				assert.Empty(t, r.URL.RawQuery)

				users := make([]map[string]any, 150)
				for i := range users {
					users[i] = map[string]any{
						"userId":     i + 1,
						"email":      fmt.Sprintf("user%d@example.com", i+1),
						"login":      fmt.Sprintf("user%d", i+1),
						"role":       "Viewer",
						"lastSeenAt": "2026-10-30T12:00:00Z",
						"isDisabled": false,
					}
				}

				users[0]["email"] = ""
				users[0]["lastSeenAt"] = ""

				w.Header().Set("Content-Type", "application/json")
				assert.NoError(t, json.NewEncoder(w).Encode(users))
			},
		),
	)
	defer srv.Close()

	records, err := NewGrafanaDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 150)
	assert.Equal(t, int32(1), requests.Load())

	assert.Equal(t, "user1", records[0].Email)
	assert.Nil(t, records[0].LastLogin)
	assert.Equal(t, "user150@example.com", records[149].Email)
}
