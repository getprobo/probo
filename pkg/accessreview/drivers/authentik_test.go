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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// Replays a cassette covering the account-type, status and factor matrix
// across two pages, plus the name resolver.
func TestAuthentikDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/authentik", "AUTHENTIK_API_TOKEN")
	client := newVCRClient(rec, os.Getenv("AUTHENTIK_API_TOKEN"))

	baseURL := os.Getenv("AUTHENTIK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://authentik.example.com"
	}

	records, err := NewAuthentikDriver(client, baseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 6) // the account with neither email nor username is skipped

	// Superuser with a TOTP device.
	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, []string{"authentik Admins"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, new(true), records[0].Active)
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
	assert.Equal(t, "00000000-0000-4000-8000-000000000001", records[0].ExternalID)
	require.NotNil(t, records[0].LastLogin)
	assert.Equal(t, "2026-08-01T08:30:00.123456Z", records[0].LastLogin.UTC().Format("2006-01-02T15:04:05.000000Z"))
	require.NotNil(t, records[0].CreatedAt)

	// No factor at all: an explicit DISABLED.
	assert.Equal(t, "engineer@example.com", records[1].Email)
	assert.Equal(t, coredata.MFAStatusDisabled, records[1].MFAStatus)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Equal(t, []string{"Application Viewer", "Engineering"}, records[1].Roles)

	// Deactivated accounts are still returned.
	assert.Equal(t, "leaver@example.com", records[2].Email)
	assert.Equal(t, new(false), records[2].Active)

	// Username stands in for the absent email.
	assert.Equal(t, "ak-outpost-proxy", records[3].Email)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, records[3].AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, records[3].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[3].MFAStatus)
	assert.Nil(t, records[3].LastLogin)

	// External user with a WebAuthn device.
	assert.Equal(t, "contractor@example.com", records[4].Email)
	assert.Equal(t, coredata.MFAStatusEnabled, records[4].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[4].AccountType)

	// Second page, and a role that appears under both roles_obj and groups_obj.
	assert.Equal(t, "support@example.com", records[5].Email)
	assert.Equal(t, []string{"Support"}, records[5].Roles)

	// The cassette puts the default brand on the second page.
	name, err := NewAuthentikNameResolver(client, baseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Example Internal", name)
}

// A token that cannot read every device kind, and an authentik without the SMS
// stage, must leave MFA unknown rather than disabled.
func TestAuthentikDriverPartialFactorVisibility(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/authenticators/admin/totp/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"pagination":{"next":0},"results":[{"pk":11,"user":{"pk":1}}]}`))
		case "/api/v3/authenticators/admin/webauthn/":
			w.WriteHeader(http.StatusForbidden)
		case "/api/v3/authenticators/admin/duo/", "/api/v3/authenticators/admin/sms/":
			w.WriteHeader(http.StatusNotFound)
		case "/api/v3/core/users/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"pagination":{"next":0},"results":[{"pk":1,"uuid":"u1","username":"admin","email":"admin@example.com","type":"internal","is_active":true},{"pk":2,"uuid":"u2","username":"engineer","email":"engineer@example.com","type":"internal","is_active":true}]}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	records, err := NewAuthentikDriver(server.Client(), server.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)
}

// An unreadable user list stays a hard error: an empty account set would read
// as "nobody has access".
func TestAuthentikDriverPropagatesFetchFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v3/authenticators/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"pagination":{"next":0},"results":[]}`))

			return
		}

		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := NewAuthentikDriver(server.Client(), server.URL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "core/users")
}
