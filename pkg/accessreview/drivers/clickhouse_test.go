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

func TestClickHouseDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/clickhouse", "CLICKHOUSE_API_KEY")
	// ClickHouse Cloud authenticates with HTTP Basic (keyId:keySecret). The
	// matcher ignores Authorization, so replay needs no auth.
	client := newVCRClient(rec, basicAuthUserPass(os.Getenv("CLICKHOUSE_API_KEY")))

	driver := NewClickHouseDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	admin := records[0]
	assert.Equal(t, "u-0000-0001", admin.ExternalID)
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "Admin User", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.NotNil(t, admin.CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	dev := records[1]
	assert.Equal(t, []string{"Developer"}, dev.Roles)
	assert.False(t, dev.IsAdmin)

	// assignedRoles wins over the deprecated role field; a custom role
	// merely containing "Admin" is not promoted to admin.
	custom := records[2]
	assert.Equal(t, "custom@example.com", custom.FullName)
	assert.Equal(t, []string{"Billing Admin"}, custom.Roles)
	assert.False(t, custom.IsAdmin)
}
