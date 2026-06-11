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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestZendeskDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/zendesk", "ZENDESK_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("ZENDESK_TOKEN")))

	driver := NewZendeskDriver(client, "acme")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	// The page holds three users; the end-user (carol) is filtered out so
	// only the two staff members remain.
	assert.Len(t, records, 2)

	r := records[0]
	assert.Equal(t, "alice@example.com", r.Email)
	assert.Equal(t, "Alice Example", r.FullName)
	assert.Equal(t, "12345", r.ExternalID)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	assert.True(t, r.IsAdmin)
	assert.Equal(t, "admin", r.Role)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, r.AccountType)
	assert.Equal(t, coredata.MFAStatusEnabled, r.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, r.AuthMethod)
	require.NotNil(t, r.LastLogin)
	require.NotNil(t, r.CreatedAt)

	// Second record exercises the agent (non-admin), MFA-disabled, and
	// never-logged-in (null last_login_at) branches.
	r2 := records[1]
	assert.Equal(t, "bob@example.com", r2.Email)
	assert.Equal(t, "67890", r2.ExternalID)
	require.NotNil(t, r2.Active)
	assert.True(t, *r2.Active)
	assert.False(t, r2.IsAdmin)
	assert.Equal(t, "agent", r2.Role)
	assert.Equal(t, coredata.MFAStatusDisabled, r2.MFAStatus)
	assert.Nil(t, r2.LastLogin)
}

// TestZendeskRecord_FieldMapping covers the field-mapping edge cases that the
// cassette does not: a null 2FA flag stays unknown (not "disabled"), a
// suspended user is inactive even when active is true, and a custom role name
// passes through verbatim.
func TestZendeskRecord_FieldMapping(t *testing.T) {
	t.Parallel()

	rec := zendeskRecord(zendeskUser{
		ID:        42,
		Email:     "dana@example.com",
		Name:      "Dana Example",
		Role:      "Light agent",
		Suspended: true,
		Active:    true,
	})

	assert.Equal(t, "dana@example.com", rec.Email)
	assert.Equal(t, "Dana Example", rec.FullName)
	require.NotNil(t, rec.Active)
	assert.False(t, *rec.Active, "a suspended user must be inactive")
	assert.Equal(t, coredata.MFAStatusUnknown, rec.MFAStatus, "null 2FA must stay unknown")
	assert.Equal(t, "Light agent", rec.Role, "custom role names pass through")
	assert.False(t, rec.IsAdmin)
	assert.Equal(t, "42", rec.ExternalID)
	assert.Nil(t, rec.LastLogin)
	assert.Nil(t, rec.CreatedAt)
}
