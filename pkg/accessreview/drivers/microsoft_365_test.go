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

func TestMicrosoft365Driver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/microsoft_365", "MICROSOFT_365_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("MICROSOFT_365_TOKEN")))

	driver := NewMicrosoft365Driver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	recordsByEmail := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		recordsByEmail[record.Email] = record
	}

	// Alice: Global Administrator with MFA registered (matched by user id).
	alice := recordsByEmail["alice@example.com"]
	assert.Equal(t, "Alice Admin", alice.FullName)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", alice.ExternalID)
	assert.Equal(t, []string{"Global Administrator"}, alice.Roles)
	assert.True(t, alice.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, alice.MFAStatus)
	assert.Equal(t, "Security Engineer", alice.JobTitle)
	require.NotNil(t, alice.Active)
	assert.True(t, *alice.Active)
	require.NotNil(t, alice.CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, alice.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, alice.AccountType)

	// Bob: User Administrator with MFA not registered.
	bob := recordsByEmail["bob@example.com"]
	assert.Equal(t, []string{"User Administrator"}, bob.Roles)
	assert.True(t, bob.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, bob.MFAStatus)
	require.NotNil(t, bob.Active)
	assert.True(t, *bob.Active)

	// Carol: no directory role (defaults to User); MFA matched by
	// case-insensitive UPN fallback when the report id differs.
	carol := recordsByEmail["carol@example.com"]
	assert.Equal(t, []string{"User"}, carol.Roles)
	assert.False(t, carol.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, carol.MFAStatus)

	// Dana: inactive, absent from the MFA registration report → Unknown.
	dana := recordsByEmail["dana@example.com"]
	assert.Equal(t, []string{"User"}, dana.Roles)
	assert.False(t, dana.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, dana.MFAStatus)
	require.NotNil(t, dana.Active)
	assert.False(t, *dana.Active)
}
