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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// TestSigNozDriver replays a cassette (recorded against SigNoz Cloud, then
// anonymized) covering the role/status matrix and exercises both the driver
// (GET /api/v1/user) and the name resolver (GET /api/v2/orgs/me).
func TestSigNozDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/signoz", "SIGNOZ_API_KEY")
	client := newVCRClientWithHeader(rec, "SIGNOZ-API-KEY", os.Getenv("SIGNOZ_API_KEY"))

	baseURL := os.Getenv("SIGNOZ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://signoz.example.com"
	}

	records, err := NewSigNozDriver(client, baseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 7) // 5 users (the no-email one is skipped) + 2 service accounts

	// ADMIN role -> admin.
	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, "00000000-0000-4000-8000-000000000001", records[0].ExternalID)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].CreatedAt)

	// isRoot -> admin even with a non-admin role.
	assert.Equal(t, "owner@example.com", records[1].Email)
	assert.Equal(t, []string{"Viewer"}, records[1].Roles)
	assert.Equal(t, new(true), records[1].IsAdmin)

	// Managed-role display name -> Editor; not admin.
	assert.Equal(t, "editor@example.com", records[2].Email)
	assert.Equal(t, []string{"Editor"}, records[2].Roles)
	assert.Equal(t, new(false), records[2].IsAdmin)
	require.NotNil(t, records[2].Active)
	assert.True(t, *records[2].Active)

	// pending_invite -> inactive.
	assert.Equal(t, "invited@example.com", records[3].Email)
	assert.Equal(t, []string{"Viewer"}, records[3].Roles)
	require.NotNil(t, records[3].Active)
	assert.False(t, *records[3].Active)

	// deleted -> inactive.
	assert.Equal(t, "removed@example.com", records[4].Email)
	assert.Equal(t, []string{"Editor"}, records[4].Roles)
	require.NotNil(t, records[4].Active)
	assert.False(t, *records[4].Active)

	// Service account with signoz-admin -> admin, API-key auth method.
	assert.Equal(t, "ci-deployer@example.com", records[5].Email)
	assert.Equal(t, "ci-deployer", records[5].FullName)
	assert.Equal(t, []string{"Admin"}, records[5].Roles)
	assert.Equal(t, new(true), records[5].IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, records[5].AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, records[5].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[5].MFAStatus)
	assert.Equal(t, "00000000-0000-4000-8000-0000000000b1", records[5].ExternalID)
	require.NotNil(t, records[5].Active)
	assert.True(t, *records[5].Active)
	require.NotNil(t, records[5].CreatedAt)

	// revoked -> inactive, and not admin.
	assert.Equal(t, "revoked-bot@example.com", records[6].Email)
	assert.Equal(t, []string{"Viewer"}, records[6].Roles)
	assert.Equal(t, new(false), records[6].IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, records[6].AccountType)
	require.NotNil(t, records[6].Active)
	assert.False(t, *records[6].Active)

	name, err := NewSigNozNameResolver(client, baseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Example Org", name)
}

func TestSigNozDriverListAccountsEmptyData(t *testing.T) {
	t.Parallel()

	for name, payload := range map[string]string{
		"null data":   `{"status":"success","data":null}`,
		"empty array": `{"status":"success","data":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(payload))
			}))
			defer srv.Close()

			records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
			require.NoError(t, err)
			assert.Empty(t, records)
		})
	}
}

func TestSigNozDriverListAccountsErrorStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":"error"}`))
	}))
	defer srv.Close()

	_, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestSigNozRoles(t *testing.T) {
	t.Parallel()

	for in, want := range map[string][]string{
		"ADMIN":         {"Admin"},
		"signoz-admin":  {"Admin"},
		"EDITOR":        {"Editor"},
		"signoz-editor": {"Editor"},
		"VIEWER":        {"Viewer"},
		"signoz-viewer": {"Viewer"},
		"":              {},
		"  ":            {},
		"custom-role":   {"custom-role"}, // unknown role preserved verbatim
		"superadmin":    {"superadmin"},  // contains "admin" but must NOT be promoted
	} {
		assert.Equalf(t, want, sigNozRoles(in), "role %q", in)
	}
}

func TestSigNozActiveStatus(t *testing.T) {
	t.Parallel()

	active := sigNozActiveStatus("active")
	require.NotNil(t, active)
	assert.True(t, *active)

	for _, status := range []string{"pending_invite", "deleted"} {
		v := sigNozActiveStatus(status)
		require.NotNilf(t, v, "status %q", status)
		assert.Falsef(t, *v, "status %q", status)
	}

	assert.Nil(t, sigNozActiveStatus("something_unexpected"))
}

func TestSigNozNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("falls back to name when displayName is empty", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":{"displayName":"","name":"acme"}}`))
		}))
		defer srv.Close()

		name, err := NewSigNozNameResolver(srv.Client(), srv.URL).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "acme", name)
	})

	t.Run("returns empty without error on terminal failure", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		name, err := NewSigNozNameResolver(srv.Client(), srv.URL).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, name)
	})
}

// A key below signoz-admin cannot list service accounts. The users it can see
// must still form a valid review population.
func TestSigNozDriverServiceAccountsForbidden(t *testing.T) {
	t.Parallel()

	for name, status := range map[string]int{
		"forbidden":    http.StatusForbidden,
		"unauthorized": http.StatusUnauthorized,
		"not found":    http.StatusNotFound,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/service_accounts" {
					w.WriteHeader(status)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"u1","email":"user@example.com","displayName":"User","role":"VIEWER","status":"active"}]}`))
			}))
			defer srv.Close()

			records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
			require.NoError(t, err)
			require.Len(t, records, 1)
			assert.Equal(t, "user@example.com", records[0].Email)
			assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
		})
	}
}

// A failed role lookup must leave admin status unknown, not false.
func TestSigNozDriverServiceAccountRolesUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/user":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
		case "/api/v1/service_accounts":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"sa1","name":"bot","email":"bot@example.com","status":"active"}]}`))
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer srv.Close()

	records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "bot@example.com", records[0].Email)
	assert.Nil(t, records[0].Roles)
	assert.Nil(t, records[0].IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, records[0].AccountType)
}

func TestSigNozDriverServiceAccountRolesContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/user":
			_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
		case "/api/v1/service_accounts":
			// Cancel only once users and the account list have been served, so
			// the failure lands on the role lookup rather than earlier.
			cancel()

			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"sa1","name":"bot","email":"bot@example.com","status":"active"}]}`))
		default:
			_, _ = w.Write([]byte(`{"status":"success","data":{"serviceAccountRoles":[]}}`))
		}
	}))
	defer srv.Close()

	_, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
