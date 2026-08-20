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
	"go.probo.inc/probo/pkg/coredata"
)

func TestCalendlyDriver_ListAccounts(t *testing.T) {
	t.Parallel()

	const organizationURI = "https://api.calendly.com/organizations/ORG"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			assert.Equal(t, http.MethodGet, r.Method)
			_, _ = w.Write([]byte(`{"resource":{"current_organization":"` + organizationURI + `"}}`))
		case "/organization_memberships":
			assert.Equal(t, organizationURI, r.URL.Query().Get("organization"))
			assert.Equal(t, "100", r.URL.Query().Get("count"))

			if r.URL.Query().Get("page_token") == "" {
				_, _ = w.Write(
					[]byte(`{"collection":[{"uri":"https://api.calendly.com/organization_memberships/M1","role":"owner","user":{"uri":"https://api.calendly.com/users/U1","name":"Owner User","email":"owner@example.com","created_at":"2026-01-02T03:04:05Z"}}],"pagination":{"next_page_token":"next"}}`),
				)

				return
			}

			assert.Equal(t, "next", r.URL.Query().Get("page_token"))
			_, _ = w.Write(
				[]byte(`{"collection":[{"uri":"https://api.calendly.com/organization_memberships/M2","role":"user","user":{"uri":"https://api.calendly.com/users/U2","name":"Member User","email":"member@example.com"}}],"pagination":{"next_page_token":null}}`),
			)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	driver := NewCalendlyDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	assert.Equal(t, "owner@example.com", records[0].Email)
	assert.Equal(t, "Owner User", records[0].FullName)
	assert.Equal(t, []string{"Owner"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Nil(t, records[0].Active)
	assert.Equal(t, "https://api.calendly.com/users/U1", records[0].ExternalID)
	assert.Equal(t, new(createdAt), records[0].CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)

	assert.Equal(t, []string{"User"}, records[1].Roles)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Nil(t, records[1].CreatedAt)
}

func TestCalendlyDriver_ListAccountsWithoutOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resource":{"current_organization":""}}`))
	}))
	t.Cleanup(server.Close)

	driver := NewCalendlyDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())

	require.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "authenticated user has no organization")
}
