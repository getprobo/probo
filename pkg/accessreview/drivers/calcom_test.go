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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestCalComDriver_ListAccounts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/me":
			assert.Equal(t, http.MethodGet, r.Method)
			_, _ = w.Write([]byte(`{"status":"success","data":{"organizationId":42}}`))
		case "/v2/organizations/42/memberships":
			assert.Equal(t, strconv.Itoa(calComPageSize), r.URL.Query().Get("take"))

			skip, err := strconv.Atoi(r.URL.Query().Get("skip"))
			require.NoError(t, err)

			if skip == 0 {
				memberships := make([]map[string]any, calComPageSize)
				for i := range memberships {
					role := "MEMBER"
					if i == 0 {
						role = "OWNER"
					}

					memberships[i] = map[string]any{
						"id":       i + 1000,
						"userId":   i + 2000,
						"accepted": true,
						"role":     role,
						"user": map[string]any{
							"name":  fmt.Sprintf("User %d", i),
							"email": fmt.Sprintf("user%d@example.com", i),
						},
					}
				}

				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": memberships}))

				return
			}

			assert.Equal(t, calComPageSize, skip)
			_, _ = w.Write(
				[]byte(`{"status":"success","data":[{"id":9999,"userId":7777,"accepted":false,"role":"ADMIN","user":{"name":"Pending Admin","email":"admin@example.com"}}]}`),
			)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	driver := NewCalComDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, calComPageSize+1)

	assert.Equal(t, "user0@example.com", records[0].Email)
	assert.Equal(t, "User 0", records[0].FullName)
	assert.Equal(t, []string{"Owner"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].Active)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, "2000", records[0].ExternalID)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)

	pendingAdmin := records[calComPageSize]
	assert.Equal(t, "admin@example.com", pendingAdmin.Email)
	assert.Equal(t, []string{"Admin"}, pendingAdmin.Roles)
	assert.Equal(t, new(false), pendingAdmin.Active)
	assert.Equal(t, new(true), pendingAdmin.IsAdmin)
	assert.Equal(t, "7777", pendingAdmin.ExternalID)
}

func TestCalComDriver_ListAccountsWithoutOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"organizationId":null}}`))
	}))
	t.Cleanup(server.Close)

	driver := NewCalComDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())

	require.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "authenticated user has no organization")
}
