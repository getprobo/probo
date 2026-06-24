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
