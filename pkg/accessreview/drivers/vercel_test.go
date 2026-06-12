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

func TestVercelDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/vercel", "VERCEL_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("VERCEL_TOKEN")))

	teamID := os.Getenv("VERCEL_TEAM_ID")
	if teamID == "" {
		teamID = "team_acme"
	}

	driver := NewVercelDriver(client, teamID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.Roles)
	assert.True(t, r.IsAdmin)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)

	// Unconfirmed members must surface as Active=false.
	require.Len(t, records, 2)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
}
