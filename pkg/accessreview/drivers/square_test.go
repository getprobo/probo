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

func TestSquareDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/square", "SQUARE_ACCESS_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SQUARE_ACCESS_TOKEN")))

	driver := NewSquareDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Non-owner, ACTIVE → Member, active, not admin.
	member := records[0]
	assert.Equal(t, "-3oZQKPKVk6gUXU_V5Qa", member.ExternalID)
	assert.Equal(t, "sherlock.holmes@example.com", member.Email)
	assert.Equal(t, "Sherlock Holmes", member.FullName)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, []string{"Member"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	// Owner → is_owner drives IsAdmin directly.
	owner := records[1]
	assert.Equal(t, "Pw67AzUomYUdF04AN17i", owner.ExternalID)
	assert.Equal(t, "john.watson@example.com", owner.Email)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)

	// INACTIVE member → active=false.
	inactive := records[2]
	assert.Equal(t, "martha.hudson@example.com", inactive.Email)
	assert.False(t, inactive.IsAdmin)
	require.NotNil(t, inactive.Active)
	assert.False(t, *inactive.Active)
}
