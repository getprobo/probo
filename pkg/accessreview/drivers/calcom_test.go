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
	"os"
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
			assert.NoError(t, err)

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

				assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": memberships}))

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

func TestCalComDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/calcom", "CAL_COM_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("CAL_COM_API_KEY")))

	driver := NewCalComDriver(client, "https://api.cal.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, "100", records[0].ExternalID)
	assert.Equal(t, []string{"Member", "Owner"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].Active)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, "101", records[1].ExternalID)
	assert.Equal(t, []string{"Member"}, records[1].Roles)
	assert.Equal(t, "102", records[2].ExternalID)
	assert.Equal(t, []string{"Admin"}, records[2].Roles)
}

func TestCalComDriver_ListAccountsSolo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/me":
			_, _ = w.Write(
				[]byte(`{"status":"success","data":{"id":42,"name":" Solo User ","email":" solo@example.com ","organizationId":null}}`),
			)
		case "/v2/teams":
			_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	driver := NewCalComDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "solo@example.com", records[0].Email)
	assert.Equal(t, "Solo User", records[0].FullName)
	assert.Empty(t, records[0].Roles)
	assert.Equal(t, new(true), records[0].Active)
	assert.Nil(t, records[0].IsAdmin)
	assert.Equal(t, "42", records[0].ExternalID)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
}

func TestCalComMembershipRecords_AdminSignal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		roles    []string
		expected *bool
	}{
		{name: "member is known non-admin", roles: []string{"MEMBER"}, expected: new(false)},
		{name: "owner is known admin", roles: []string{"OWNER"}, expected: new(true)},
		{name: "empty role is unknown", roles: []string{""}},
		{name: "unrecognized role is unknown", roles: []string{"SUPER_ADMIN"}},
		{name: "member plus unknown stays unknown", roles: []string{"MEMBER", "SUPER_ADMIN"}},
		{name: "owner plus unknown stays known admin", roles: []string{"OWNER", "SUPER_ADMIN"}, expected: new(true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			memberships := make([]calComMembership, len(tt.roles))
			for i, role := range tt.roles {
				memberships[i].UserID = 42
				memberships[i].User.Email = "user@example.com"
				memberships[i].Role = role
			}

			records := calComMembershipRecords(memberships)
			require.Len(t, records, 1)
			assert.Equal(t, tt.expected, records[0].IsAdmin)
		})
	}
}
