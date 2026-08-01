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
)

func TestGoogleAnalyticsDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/google_analytics", "GOOGLE_ANALYTICS_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GOOGLE_ANALYTICS_TOKEN")))

	driver := NewGoogleAnalyticsDriver(client, "123456", "https://analyticsadmin.googleapis.com/v1alpha")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

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

	// dave holds a binding only on properties/99999, a subproperty parented to
	// another property rather than to the account. He is reachable only through
	// the ancestor filter, so his presence is what proves subproperties are
	// walked. A third property (55555) returns 403 and is skipped rather than
	// failing the whole account — otherwise these four records would be zero.
	dave := records[3]
	assert.Equal(t, "dave@example.com", dave.Email)
	assert.False(t, dave.IsAdmin)
	assert.Equal(t, []string{"analyst"}, dave.Roles)
}
