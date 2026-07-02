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

func TestOktaDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/okta", "OKTA_API_TOKEN")

	authValue := ""
	if token := os.Getenv("OKTA_API_TOKEN"); token != "" {
		authValue = "SSWS " + token
	}

	client := newVCRClient(rec, authValue)

	domain := os.Getenv("OKTA_DOMAIN")
	if domain == "" {
		domain = "acme.okta.com"
	}

	driver := NewOktaDriver(client, domain)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// Two pages followed via the Link header; the third page-1 user has no
	// email and is dropped, so three records survive.
	require.Len(t, records, 3)

	// Alice: active, displayName preferred, title + timestamps populated.
	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "Alice Active", records[0].FullName)
	assert.Equal(t, "Security Engineer", records[0].JobTitle)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	assert.Equal(t, "00u1aaaaaaaaaaaaa0h7", records[0].ExternalID)
	require.NotNil(t, records[0].CreatedAt)
	require.NotNil(t, records[0].LastLogin)

	// Bob: SUSPENDED → inactive, no displayName (falls back to first+last),
	// null lastLogin stays nil.
	assert.Equal(t, "bob@example.com", records[1].Email)
	assert.Equal(t, "Bob Suspended", records[1].FullName)
	assert.Empty(t, records[1].JobTitle)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
	assert.Equal(t, "00u2bbbbbbbbbbbbb1h7", records[1].ExternalID)
	assert.Nil(t, records[1].LastLogin)
	require.NotNil(t, records[1].CreatedAt)

	// Carol: page 2, DEPROVISIONED → inactive.
	assert.Equal(t, "carol@example.com", records[2].Email)
	assert.Equal(t, "Carol Gone", records[2].FullName)
	assert.Equal(t, "Contractor", records[2].JobTitle)
	require.NotNil(t, records[2].Active)
	assert.False(t, *records[2].Active)
	assert.Equal(t, "00u4ddddddddddddd4h7", records[2].ExternalID)
}
