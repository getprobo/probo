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

func TestTallyDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/tally", "TALLY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("TALLY_TOKEN")))

	orgID := os.Getenv("TALLY_ORG_ID")
	if orgID == "" {
		orgID = "wvBzxD"
	}

	driver := NewTallyDriver(client, orgID, "https://api.tally.so")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
}

// TestTallyCurrentUser covers the /users/me fetch shared by the
// create-connector validation (organization id derivation) and the name
// resolver (source label). Both calls replay from the same cassette.
func TestTallyCurrentUser(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/tally_user", "TALLY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("TALLY_TOKEN")))

	user, err := GetTallyCurrentUser(context.Background(), client, "https://api.tally.so")
	require.NoError(t, err)
	assert.NotEmpty(t, user.OrganizationID)

	resolver := NewTallyNameResolver(client, "https://api.tally.so")
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}
