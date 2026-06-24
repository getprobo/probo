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

func TestClerkDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/clerk", "CLERK_SECRET_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("CLERK_SECRET_KEY")))

	driver := NewClerkDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Clerk returns users newest-first (default order_by=-created_at).
	first := records[0]
	assert.Equal(t, "user_3EfkCEWmtIsoMD3rRxIpDsBOPzv", first.ExternalID)
	assert.Equal(t, "c@example.com", first.Email)
	assert.Equal(t, "c c", first.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, first.AccountType)
	require.NotNil(t, first.Active)
	assert.True(t, *first.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, first.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, first.AuthMethod)
	assert.NotNil(t, first.CreatedAt)
	assert.Nil(t, first.LastLogin)

	second := records[1]
	assert.Equal(t, "b@example.com", second.Email)
	assert.Equal(t, "b b", second.FullName)
	require.NotNil(t, second.Active)
	assert.True(t, *second.Active)

	// a@example.com is locked, so it must be reported inactive.
	third := records[2]
	assert.Equal(t, "a@example.com", third.Email)
	assert.Equal(t, "a a", third.FullName)
	require.NotNil(t, third.Active)
	assert.False(t, *third.Active)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, third.AuthMethod)
}

func TestClerkPrimaryEmail(t *testing.T) {
	t.Parallel()

	user := clerkUser{
		PrimaryEmailAddressID: new("eml_primary"),
		EmailAddresses: []struct {
			ID           string `json:"id"`
			EmailAddress string `json:"email_address"`
		}{
			{ID: "eml_secondary", EmailAddress: "secondary@example.com"},
			{ID: "eml_primary", EmailAddress: "primary@example.com"},
		},
	}

	assert.Equal(t, "primary@example.com", clerkPrimaryEmail(user))
}
