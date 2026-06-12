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

func TestMetabaseDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/metabase", "METABASE_API_KEY")
	client := newVCRClientWithHeader(rec, "x-api-key", os.Getenv("METABASE_API_KEY"))

	instanceURL := os.Getenv("METABASE_INSTANCE_URL")
	if instanceURL == "" {
		instanceURL = "https://k7.metabaseapp.com"
	}

	driver := NewMetabaseDriver(client, instanceURL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice A.", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.True(t, records[0].IsAdmin)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	assert.Equal(t, "1", records[0].ExternalID)
	require.NotNil(t, records[0].LastLogin)
	require.NotNil(t, records[0].CreatedAt)

	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, "Bob Builder", records[1].FullName)
	assert.Equal(t, []string{"User"}, records[1].Roles)
	assert.False(t, records[1].IsAdmin)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
	assert.Equal(t, "2", records[1].ExternalID)
	assert.Nil(t, records[1].LastLogin)
	require.NotNil(t, records[1].CreatedAt)
}
