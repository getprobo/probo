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
)

func TestMetabaseDriverListAccounts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/user", r.URL.Path)
		assert.Equal(t, "all", r.URL.Query().Get("status"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": 1,
				"email": "alice@example.com",
				"first_name": "Alice",
				"last_name": "Admin",
				"common_name": "Alice A.",
				"is_active": true,
				"is_superuser": true,
				"last_login": "2026-05-20T10:11:12.345678Z",
				"date_joined": "2026-01-02T03:04:05Z"
			},
			{
				"id": 2,
				"email": "bob@example.com",
				"first_name": "Bob",
				"last_name": "Builder",
				"is_active": false,
				"is_superuser": false,
				"last_login": "",
				"date_joined": "2026-02-03T04:05:06Z"
			},
			{
				"id": 3,
				"email": "",
				"first_name": "No",
				"last_name": "Email",
				"is_active": true,
				"is_superuser": false
			}
		]`))
	}))
	defer srv.Close()

	driver := NewMetabaseDriver(srv.Client(), srv.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice A.", records[0].FullName)
	assert.Equal(t, "Admin", records[0].Role)
	assert.True(t, records[0].IsAdmin)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	assert.Equal(t, "1", records[0].ExternalID)
	require.NotNil(t, records[0].LastLogin)
	require.NotNil(t, records[0].CreatedAt)

	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, "Bob Builder", records[1].FullName)
	assert.Equal(t, "User", records[1].Role)
	assert.False(t, records[1].IsAdmin)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
	assert.Equal(t, "2", records[1].ExternalID)
	assert.Nil(t, records[1].LastLogin)
	require.NotNil(t, records[1].CreatedAt)
}

func TestMetabaseDriverListAccountsUnexpectedStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer srv.Close()

	driver := NewMetabaseDriver(srv.Client(), srv.URL)
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}
