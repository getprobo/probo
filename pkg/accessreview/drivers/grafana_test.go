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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestGrafanaDriverListAccounts(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		require.Equal(t, "/api/org/users", r.URL.Path)
		require.Equal(t, "100", r.URL.Query().Get("perpage"))

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		switch page {
		case 1:
			users := make([]map[string]any, 0, grafanaUsersPageSize)
			users = append(users, map[string]any{
				"userId":     1,
				"email":      "admin@example.com",
				"name":       "Admin User",
				"role":       "Admin",
				"isDisabled": false,
				"lastSeenAt": "2026-05-20T10:00:00Z",
			})
			for i := 1; i < grafanaUsersPageSize; i++ {
				users = append(users, map[string]any{
					"userId": i + 100,
					"name":   "Ignored User",
					"role":   "Viewer",
				})
			}

			_ = json.NewEncoder(w).Encode(users)
		case 2:
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"userId":     2,
					"login":      "viewer@example.com",
					"name":       "Viewer User",
					"role":       "Viewer",
					"isDisabled": true,
				},
			})
		default:
			t.Fatalf("unexpected page %d", page)
		}
	}))
	t.Cleanup(ts.Close)

	driver := NewGrafanaDriver(ts.Client(), ts.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, "Admin", records[0].Role)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, "1", records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, records[0].AccountType)
	assert.Equal(t, coredata.AccessEntryAuthMethodUnknown, records[0].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	require.NotNil(t, records[0].LastLogin)
	assert.Equal(t, time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), *records[0].LastLogin)

	assert.Equal(t, "viewer@example.com", records[1].Email)
	assert.Equal(t, "Viewer User", records[1].FullName)
	assert.Equal(t, "Viewer", records[1].Role)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, "2", records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
}

func TestGrafanaNameResolver(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		require.Equal(t, "/api/org", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name": "Acme Grafana",
		})
	}))
	t.Cleanup(ts.Close)

	resolver := NewGrafanaNameResolver(ts.Client(), ts.URL)
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Acme Grafana", name)
}
