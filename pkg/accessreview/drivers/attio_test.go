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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAttioDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/attio", "ATTIO_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("ATTIO_TOKEN")))

	records, err := NewAttioDriver(client, "https://api.attio.com").ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	createdAt := time.Date(2022, time.November, 21, 13, 22, 49, 61281000, time.UTC)
	assert.Equal(t, AccountRecord{
		Email:       "ada@example.com",
		FullName:    "Ada Lovelace",
		Roles:       []string{"Admin"},
		Active:      new(true),
		IsAdmin:     new(true),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		CreatedAt:   new(createdAt),
	}, records[0])
	assert.Equal(t, []string{"Member"}, records[1].Roles)
	assert.Equal(t, new(true), records[1].Active)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Equal(t, []string{"Suspended"}, records[2].Roles)
	assert.Equal(t, new(false), records[2].Active)
	assert.Equal(t, new(false), records[2].IsAdmin)
}

func TestAttioNameResolver_ResolveInstanceName(t *testing.T) {
	t.Parallel()

	t.Run("active token", func(t *testing.T) {
		t.Parallel()

		rec := newRecorder(t, "testdata/attio_identity", "ATTIO_TOKEN")
		client := newVCRClient(rec, bearerAuth(os.Getenv("ATTIO_TOKEN")))

		resolver := NewAttioNameResolver(client, "https://api.attio.com")
		name, err := resolver.ResolveInstanceName(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "Acme", name)
	})

	t.Run("inactive token", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"active":false}`))
		}))
		t.Cleanup(server.Close)

		resolver := NewAttioNameResolver(server.Client(), server.URL)
		name, err := resolver.ResolveInstanceName(context.Background())

		require.Error(t, err)
		assert.Empty(t, name)
		assert.ErrorIs(t, err, ErrTerminalNameResolution)
	})

	t.Run("server error remains retryable", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		resolver := NewAttioNameResolver(server.Client(), server.URL)
		_, err := resolver.ResolveInstanceName(context.Background())

		require.Error(t, err)
		assert.False(t, errors.Is(err, ErrTerminalNameResolution))
	})
}
