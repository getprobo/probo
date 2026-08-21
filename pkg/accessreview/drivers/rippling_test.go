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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRipplingDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/rippling", "RIPPLING_API_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("RIPPLING_API_TOKEN")))

	driver := NewRipplingDriver(client, "https://rest.ripplingapis.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice Admin", records[0].FullName)
	assert.Equal(t, "user-1", records[0].ExternalID)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, records[0].AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	assert.NotNil(t, records[0].CreatedAt)

	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, "Bob Former", records[1].FullName)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
	assert.Nil(t, records[1].CreatedAt)

	assert.Equal(t, "carol@example.com", records[2].Email)
	assert.Equal(t, "carol", records[2].FullName)
	assert.Nil(t, records[2].Active)
}

func TestRipplingDriverRejectsCrossHostPagination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				_, err := fmt.Fprint(w, `{"results":[],"next_link":"https://example.com/users?cursor=stolen"}`)
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	driver := NewRipplingDriver(server.Client(), server.URL)
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cross-host pagination is not allowed")
}

func TestRipplingDriverDoesNotLeakErrorBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, err := fmt.Fprint(w, `{"detail":"secret provider response"}`)
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	driver := NewRipplingDriver(server.Client(), server.URL)
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
	assert.NotContains(t, err.Error(), "secret provider response")
}
