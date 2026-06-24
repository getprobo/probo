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
