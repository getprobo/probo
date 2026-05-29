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
	"go.probo.inc/probo/pkg/coredata"
)

func TestPostHogDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/posthog", "POSTHOG_PERSONAL_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("POSTHOG_PERSONAL_API_KEY")))

	records, err := NewPostHogDriver(client).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "Olivia Owner", owner.FullName)
	assert.Equal(t, "Owner", owner.Role)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, owner.MFAStatus)
	assert.Equal(t, "user-1", owner.ExternalID)
	require.NotNil(t, owner.CreatedAt)
	require.NotNil(t, owner.LastLogin)

	member := records[1]
	assert.Equal(t, "member@example.com", member.Email)
	assert.Equal(t, "Member", member.Role)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, member.MFAStatus)
	require.NotNil(t, member.CreatedAt)
	assert.Nil(t, member.LastLogin)

	admin := records[2]
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "Admin", admin.Role)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.Equal(t, "membership-3", admin.ExternalID)
	require.NotNil(t, admin.CreatedAt)
}

func TestPostHogNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{
			name:   "200 returns name",
			status: http.StatusOK,
			body:   `{"id":"org-1","name":"Acme Inc","slug":"acme"}`,
			want:   "Acme Inc",
		},
		{
			name:   "401 is terminal (no error, no name)",
			status: http.StatusUnauthorized,
			body:   `{"detail":"Authentication credentials were not provided."}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no error, no name)",
			status: http.StatusNotFound,
			body:   `{"detail":"Not found."}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/organizations/@current/", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewPostHogNameResolver(client).ResolveInstanceName(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPostHogRoleFallback(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Owner", posthogRole(15, ""))
	assert.Equal(t, "Admin", posthogRole(8, ""))
	assert.Equal(t, "Member", posthogRole(1, ""))
	assert.Equal(t, "engineering", posthogRole(0, "engineering"))
	assert.Equal(t, "Member", posthogRole(0, ""))
}
