// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRenderDriverListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/render", "RENDER_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("RENDER_API_KEY")))

	// RENDER_OWNER_ID supplies the live workspace id when recording; the
	// default matches the anonymized cassette URL for replay.
	ownerID := os.Getenv("RENDER_OWNER_ID")
	if ownerID == "" {
		ownerID = "tea-000000000000000000000"
	}

	driver := NewRenderDriver(client, ownerID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// Four members in the cassette; the fourth has no email and is dropped.
	require.Len(t, records, 3)

	// Admin: active with MFA enabled. ExternalID is Render's stable "usr-" id,
	// never the email.
	assert.Equal(t, "jane.doe@example.com", records[0].Email)
	assert.Equal(t, "Jane Doe", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, records[0].AuthMethod)
	assert.Equal(t, "usr-000000000000000000a1", records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)

	// Developer: active, MFA disabled, not an admin.
	assert.Equal(t, "john.smith@example.com", records[1].Email)
	assert.Equal(t, "John Smith", records[1].FullName)
	assert.Equal(t, []string{"Developer"}, records[1].Roles)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, records[1].MFAStatus)
	assert.Equal(t, "usr-000000000000000000b2", records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.True(t, *records[1].Active)

	// Workspace viewer: inactive → Active false; empty name falls back to the
	// email so the row is never nameless.
	assert.Equal(t, "sam.viewer@example.com", records[2].Email)
	assert.Equal(t, "sam.viewer@example.com", records[2].FullName)
	assert.Equal(t, []string{"Viewer"}, records[2].Roles)
	assert.False(t, records[2].IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, records[2].MFAStatus)
	assert.Equal(t, "usr-000000000000000000c3", records[2].ExternalID)
	require.NotNil(t, records[2].Active)
	assert.False(t, *records[2].Active)
}

func TestRenderDriverListAccountsError(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)),
					Header:     make(http.Header),
				}, nil
			},
		),
	}

	driver := NewRenderDriver(client, "tea-000000000000000000000")
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
	// The raw third-party body must never leak into the returned error.
	assert.NotContains(t, err.Error(), "unauthorized")
}

func TestRenderRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    []string
		isAdmin bool
	}{
		{in: "ADMIN", want: []string{"Admin"}, isAdmin: true},
		{in: "DEVELOPER", want: []string{"Developer"}, isAdmin: false},
		{in: "WORKSPACE_CONTRIBUTOR", want: []string{"Contributor"}, isAdmin: false},
		{in: "WORKSPACE_BILLING", want: []string{"Billing"}, isAdmin: false},
		{in: "WORKSPACE_VIEWER", want: []string{"Viewer"}, isAdmin: false},
		{in: "future_role", want: []string{"future_role"}, isAdmin: false},
		{in: "", want: []string{}, isAdmin: false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, renderRoles(c.in))
			assert.Equal(t, c.isAdmin, renderIsAdmin(c.in))
		})
	}
}

func TestRenderMFAStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.MFAStatusEnabled, renderMFAStatus(true))
	assert.Equal(t, coredata.MFAStatusDisabled, renderMFAStatus(false))
}

func TestRenderActive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		in        string
		wantSet   bool
		wantValue bool
	}{
		{name: "active", in: "active", wantSet: true, wantValue: true},
		{name: "active uppercase", in: "ACTIVE", wantSet: true, wantValue: true},
		{name: "inactive", in: "inactive", wantSet: true, wantValue: false},
		{name: "inactive mixed case", in: "Inactive", wantSet: true, wantValue: false},
		// Undocumented / missing statuses must leave Active unset (unknown),
		// not fabricate a deactivated signal.
		{name: "empty", in: "", wantSet: false},
		{name: "pending", in: "pending", wantSet: false},
		{name: "suspended", in: "suspended", wantSet: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := renderActive(c.in)
			if !c.wantSet {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, c.wantValue, *got)
		})
	}
}
