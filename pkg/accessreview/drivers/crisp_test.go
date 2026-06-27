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

	member := records[1]
	assert.Equal(t, "9a1f3c2e-6b4d-4f8a-bc11-7d2e9f0a1b22", member.ExternalID)
	assert.Equal(t, "jordan@example.com", member.Email)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, "Support Agent", member.JobTitle)
}
