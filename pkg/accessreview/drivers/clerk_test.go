// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

func TestClerkDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/clerk", "CLERK_SECRET_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("CLERK_SECRET_KEY")))

	driver := NewClerkDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	first := records[0]
	assert.Equal(t, "usr_000000000000000000000001", first.ExternalID)
	assert.Equal(t, "jane@example.com", first.Email)
	assert.Equal(t, "Jane Doe", first.FullName)
	require.NotNil(t, first.Active)
	assert.True(t, *first.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, first.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodPassword, first.AuthMethod)
	assert.NotNil(t, first.CreatedAt)
	assert.NotNil(t, first.LastLogin)

	second := records[1]
	assert.Equal(t, "developer-user", second.FullName)
	require.NotNil(t, second.Active)
	assert.True(t, *second.Active)
	assert.Equal(t, coredata.MFAStatusEnabled, second.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodUnknown, second.AuthMethod)

	third := records[2]
	assert.Equal(t, "blocked@example.com", third.FullName)
	require.NotNil(t, third.Active)
	assert.False(t, *third.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, third.MFAStatus)
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
