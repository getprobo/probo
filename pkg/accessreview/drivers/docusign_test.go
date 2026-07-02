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
)

func TestDocuSignDriver(t *testing.T) {
	t.Parallel()

	// Account UUID matches the userinfo cassette: the driver resolves the
	// selected account's data-center base URI before listing its users.
	const accountID = "a1a1a1a1-1111-4111-8111-111111111111"

	rec := newRecorder(t, "testdata/docusign", "DOCUSIGN_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("DOCUSIGN_TOKEN")))
	driver := NewDocuSignDriver(client, accountID)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Len(t, records, 2)

	r := records[0]
	assert.Equal(t, "jane.doe@example.com", r.Email)
	assert.Equal(t, "Jane Doe", r.FullName)
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", r.ExternalID)
	assert.Equal(t, []string{"Account Administrator"}, r.Roles)
	assert.Equal(t, "CTO", r.JobTitle)
	assert.True(t, r.IsAdmin)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)

	assert.False(t, records[1].IsAdmin)
}
