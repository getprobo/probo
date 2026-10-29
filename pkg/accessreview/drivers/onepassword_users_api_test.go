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

// The Users API returns the roster under "results" and pages with
// next_page_token / page_token.
func TestOnePasswordUsersAPIDriver(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta1/accounts/ACCOUNT123/users" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("page_token") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"id":"u1","email":"alice@example.com","display_name":"Alice","state":"ACTIVE","create_time":"2024-04-10T08:09:46Z","path":"accounts/ACCOUNT123/users/u1"}],"next_page_token":"CAIQAg"}`))
		case "CAIQAg":
			_, _ = w.Write([]byte(`{"results":[{"id":"u2","email":"bob@example.com","display_name":"Bob","state":"SUSPENDED","create_time":"2024-05-01T12:00:00Z","path":"accounts/ACCOUNT123/users/u2"}]}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	driver := &OnePasswordUsersAPIDriver{
		httpClient: server.Client(),
		baseURL:    server.URL,
		accountID:  "ACCOUNT123",
	}

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice", records[0].FullName)
	assert.Equal(t, "u1", records[0].ExternalID)
	assert.Equal(t, new(true), records[0].Active)
	require.NotNil(t, records[0].CreatedAt)
	assert.Equal(t, time.Date(2024, 4, 10, 8, 9, 46, 0, time.UTC), records[0].CreatedAt.UTC())

	// Second page, reached through next_page_token.
	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, new(false), records[1].Active)
}
