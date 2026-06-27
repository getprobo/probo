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

func TestYousignDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/yousign", "YOUSIGN_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("YOUSIGN_API_KEY")))

	driver := NewYousignDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	admin := records[0]
	assert.Equal(t, "9a93d3b5-fb3b-4abf-9e70-26315b33506c", admin.ExternalID)
	assert.Equal(t, "john.doe@example.com", admin.Email)
	assert.Equal(t, "John Doe", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	assert.Equal(t, "Legal Counsel", admin.JobTitle)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.NotNil(t, admin.CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	// An invited (is_active=false) member is inactive regardless of onboarding
	// status.
	member := records[1]
	assert.Equal(t, "b2f4e1c8-6d3a-4e2b-8f1a-9d5c7e8a0b3f", member.ExternalID)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	require.NotNil(t, member.Active)
	assert.False(t, *member.Active)
	assert.Equal(t, "Sales Manager", member.JobTitle)
}

func TestYousignIsAdmin(t *testing.T) {
	t.Parallel()

	// owner is strictly more privileged than admin, so both are admins; the
	// match is exact and case-insensitive, never a substring.
	assert.True(t, yousignIsAdmin("admin"))
	assert.True(t, yousignIsAdmin("owner"))
	assert.True(t, yousignIsAdmin("Admin"))
	assert.False(t, yousignIsAdmin("member"))
	assert.False(t, yousignIsAdmin(""))
}
