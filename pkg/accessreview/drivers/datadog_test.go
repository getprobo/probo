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

func TestDatadogDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/datadog", "DATADOG_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("DATADOG_TOKEN")))

	driver := NewDatadogDriver(client, "datadoghq.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Len(t, records, 2)

	r := records[0]
	assert.Equal(t, "alice@example.com", r.Email)
	assert.Equal(t, "Alice Example", r.FullName)
	assert.Equal(t, "abc-111", r.ExternalID)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	assert.True(t, r.IsAdmin)
	assert.Equal(t, "Datadog Admin Role", r.Role)
	assert.Equal(t, "Security Engineer", r.JobTitle)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, r.AccountType)
	assert.Equal(t, coredata.MFAStatusEnabled, r.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, r.AuthMethod)

	// Second record exercises the inactive, non-admin, and service-account
	// (MFA-disabled) branches.
	r2 := records[1]
	assert.Equal(t, "bob@example.com", r2.Email)
	assert.Equal(t, "abc-222", r2.ExternalID)
	require.NotNil(t, r2.Active)
	assert.False(t, *r2.Active)
	assert.False(t, r2.IsAdmin)
	assert.Equal(t, "Datadog Standard Role", r2.Role)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, r2.AccountType)
	assert.Equal(t, coredata.MFAStatusDisabled, r2.MFAStatus)
}
