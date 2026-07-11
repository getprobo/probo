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

func TestGoogleAnalyticsDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/google_analytics", "GOOGLE_ANALYTICS_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GOOGLE_ANALYTICS_TOKEN")))

	driver := NewGoogleAnalyticsDriver(client, "123456")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// alice holds an account-level admin binding AND a property-level viewer
	// binding: roles are merged, deduplicated, sorted, prefix-stripped, and the
	// admin role sets IsAdmin. GA4 has no active signal → Active is nil.
	alice := records[0]
	assert.Equal(t, "alice@example.com", alice.Email)
	assert.Equal(t, "alice@example.com", alice.ExternalID)
	assert.Equal(t, "alice@example.com", alice.FullName)
	assert.True(t, alice.IsAdmin)
	assert.Equal(t, []string{"admin", "viewer"}, alice.Roles)
	assert.Nil(t, alice.Active)

	// bob: account-level analyst only.
	bob := records[1]
	assert.Equal(t, "bob@example.com", bob.Email)
	assert.False(t, bob.IsAdmin)
	assert.Equal(t, []string{"analyst"}, bob.Roles)

	// carol: property-level analyst only (never appears at account level).
	carol := records[2]
	assert.Equal(t, "carol@example.com", carol.Email)
	assert.False(t, carol.IsAdmin)
	assert.Equal(t, []string{"analyst"}, carol.Roles)
}
