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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentryDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/sentry", "SENTRY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SENTRY_TOKEN")))

	orgSlug := os.Getenv("SENTRY_ORG_SLUG")
	if orgSlug == "" {
		orgSlug = "acme-corp"
	}

	driver := NewSentryDriver(client, orgSlug)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.Role)
}

func TestSentryDriverListAccountsStaleSlug(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/0/organizations/acme-old/members", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"The requested resource does not exist"}`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	_, err := NewSentryDriver(client, "acme-old").ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"acme-old"`)
	assert.Contains(t, err.Error(), "not accessible")
	assert.Contains(t, err.Error(), "reconnect")
	assert.ErrorIs(t, err, errSentryOrgNotAccessible)
}

func TestSentryDriverListAccountsAutoDiscoversSlug(t *testing.T) {
	t.Parallel()

	const discoveredSlug = "discovered-org"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/0/organizations/":
			assert.Equal(t, "true", r.URL.Query().Get("member"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"slug":"` + discoveredSlug + `","name":"Discovered Org"}]`))
		case "/api/0/organizations/" + discoveredSlug + "/members":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":"42","email":"alice@example.com","name":"Alice","orgRole":"member"}]`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	records, err := NewSentryDriver(client, "").ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "alice@example.com", records[0].Email)
}
