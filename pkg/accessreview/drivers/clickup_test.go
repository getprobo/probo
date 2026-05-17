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
)

func TestClickUpDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/clickup", "CLICKUP_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("CLICKUP_TOKEN")))

	teamID := os.Getenv("CLICKUP_TEAM_ID")
	if teamID == "" {
		teamID = "9999999"
	}

	driver := NewClickUpDriver(client, teamID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	r := records[0]
	assert.Equal(t, "111", r.ExternalID)
	assert.Equal(t, "jane@example.com", r.Email)
	assert.Equal(t, "jane.doe", r.FullName)
	assert.Equal(t, "owner", r.Role)
	assert.True(t, r.IsAdmin)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	require.NotNil(t, r.LastLogin)

	// Pending invite -> Active=false.
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
	assert.Equal(t, "member", records[1].Role)
	assert.False(t, records[1].IsAdmin)
}
