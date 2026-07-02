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
	assert.Equal(t, []string{"Datadog Admin Role"}, r.Roles)
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
	assert.Equal(t, []string{"Datadog Standard Role"}, r2.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, r2.AccountType)
	assert.Equal(t, coredata.MFAStatusDisabled, r2.MFAStatus)
}
