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

func TestCrispDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/crisp", "CRISP_API_KEY")
	// Crisp authenticates with HTTP Basic over the "identifier:key" plugin
	// token. The matcher ignores Authorization, so replay needs no credential.
	client := newVCRClient(rec, basicAuthUserPass(os.Getenv("CRISP_API_KEY")))

	driver := NewCrispDriver(client, "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	owner := records[0]
	assert.Equal(t, "5c068745-c7da-4b59-89a0-1b67f3b0d6df", owner.ExternalID)
	assert.Equal(t, "alex@example.com", owner.Email)
	assert.Equal(t, "Alex Martin", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, "Founder", owner.JobTitle)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	// Crisp's operators/list exposes no MFA or account-status signal, so the
	// driver hardcodes MFA Unknown and leaves Active nil for every record.
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Nil(t, owner.Active)

	member := records[1]
	assert.Equal(t, "9a1f3c2e-6b4d-4f8a-bc11-7d2e9f0a1b22", member.ExternalID)
	assert.Equal(t, "jordan@example.com", member.Email)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, "Support Agent", member.JobTitle)
	assert.Equal(t, coredata.MFAStatusUnknown, member.MFAStatus)
	assert.Nil(t, member.Active)
}
