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
	assert.Equal(t, []string{"admin"}, r.Roles)
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
	assert.Equal(t, []string{"agent"}, r2.Roles)
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
	assert.Equal(t, []string{"Light agent"}, rec.Roles, "custom role names pass through")
	assert.False(t, rec.IsAdmin)
	assert.Equal(t, "42", rec.ExternalID)
	assert.Nil(t, rec.LastLogin)
	assert.Nil(t, rec.CreatedAt)
}
