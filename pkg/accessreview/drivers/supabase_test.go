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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupabaseDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/supabase", "SUPABASE_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SUPABASE_TOKEN")))

	orgSlug := os.Getenv("SUPABASE_ORG_SLUG")
	if orgSlug == "" {
		orgSlug = "acme-corp"
	}

	driver := NewSupabaseDriver(client, orgSlug, "https://api.supabase.com/v1")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.Roles)
}

func TestSupabaseDriverUnknownOrganization(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/supabase_unknown_organization", "SUPABASE_TOKEN", dropResponseHeaders("Set-Cookie", "X-Gotrue-Id"))
	client := newVCRClient(rec, bearerAuth(os.Getenv("SUPABASE_TOKEN")))

	_, err := NewSupabaseDriver(client, "probo-missing-org", "https://api.supabase.com/v1").ListAccounts(context.Background())

	rejected, ok := errors.AsType[*SettingRejectedError](err)
	require.Truef(t, ok, "expected a rejected slug, got %v", err)
	assert.Equal(t, SupabaseOrganizationNotFound, rejected.Code)
	assert.Equal(t, http.StatusNotFound, rejected.StatusCode)
}

func TestSupabaseDriverInaccessibleOrganization(t *testing.T) {
	t.Parallel()

	// Not recorded: only a real organization answers 403, and its slug would
	// identify it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/organizations/acmeorgslug/members", r.URL.Path)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Forbidden"}`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	_, err := NewSupabaseDriver(client, "acmeorgslug", "https://api.supabase.com/v1").ListAccounts(context.Background())

	rejected, ok := errors.AsType[*SettingRejectedError](err)
	require.Truef(t, ok, "expected a rejected slug, got %v", err)
	assert.Equal(t, SupabaseOrganizationNotAccessible, rejected.Code)
	assert.Equal(t, http.StatusForbidden, rejected.StatusCode)
}

func TestSupabaseDriverEdgePage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<!doctype html><title>Attention Required</title>`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	_, err := NewSupabaseDriver(client, "acmeorgslug", "https://api.supabase.com/v1").ListAccounts(context.Background())
	require.Error(t, err)

	_, ok := errors.AsType[*SettingRejectedError](err)
	assert.False(t, ok, "a page from the edge must not blame the slug or token")
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestSupabaseDriverUnexpectedStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	_, err := NewSupabaseDriver(client, "acmeorgslug", "https://api.supabase.com/v1").ListAccounts(context.Background())
	require.Error(t, err)

	_, ok := errors.AsType[*SettingRejectedError](err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "unexpected status 500")
}
