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

func TestSegmentDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/segment", "SEGMENT_API_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SEGMENT_API_TOKEN")))

	driver := NewSegmentDriver(client, "https://api.segmentapis.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// First user: empty name falls back to email; Workspace Owner ⇒ admin.
	owner := records[0]
	assert.Equal(t, "sgJDWk3K21k6LE3tLU9nRK", owner.ExternalID)
	assert.Equal(t, "papi@example.com", owner.Email)
	assert.Equal(t, "papi@example.com", owner.FullName)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, []string{"Workspace Owner"}, owner.Roles)
	// Segment exposes no active/suspended status, so confirmed members carry
	// no Active signal (nil), unlike pending invites below.
	assert.Nil(t, owner.Active)

	// Second user: named; resource-scoped read-only role ⇒ not admin.
	member := records[1]
	assert.Equal(t, "i2VTJURQprNfqdwjLFPWYx", member.ExternalID)
	assert.Equal(t, "Sloth", member.FullName)
	assert.Equal(t, new(false), member.IsAdmin)
	assert.Equal(t, []string{"Source Read-only"}, member.Roles)
	assert.Nil(t, member.Active)

	// Pending invite: inactive, no roles, keyed by email.
	invite := records[2]
	assert.Equal(t, "foo@example.com", invite.Email)
	assert.Equal(t, "foo@example.com", invite.ExternalID)
	assert.Nil(t, invite.IsAdmin)
	assert.Empty(t, invite.Roles)
	require.NotNil(t, invite.Active)
	assert.False(t, *invite.Active)
}
