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

func TestCursorDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/cursor", "CURSOR_ADMIN_TOKEN")
	// Cursor authenticates via HTTP Basic auth (the admin key as the
	// username). The cassette matcher ignores Authorization, so replay
	// needs no auth; the value matters only when re-recording.
	client := newVCRClient(rec, basicAuth(os.Getenv("CURSOR_ADMIN_TOKEN")))

	driver := NewCursorDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	member := records[0]
	assert.Equal(t, "jane@example.com", member.Email)
	assert.Equal(t, "Jane Doe", member.FullName)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	// The Cursor Admin API returns the member id as a string; it is used
	// verbatim as the stable ExternalID.
	assert.Equal(t, "10000001", member.ExternalID)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	owner := records[1]
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)

	// A removed member (role "removed", isRemoved true) is still returned,
	// flagged inactive rather than dropped, per the AccountRecord contract.
	removed := records[2]
	assert.Equal(t, []string{"Removed"}, removed.Roles)
	assert.False(t, removed.IsAdmin)
	require.NotNil(t, removed.Active)
	assert.False(t, *removed.Active)

	// Cursor's two removal signals are not always consistent: a member can
	// carry role "removed" while isRemoved is still false. The role alone
	// must mark the account inactive.
	removedByRole := records[3]
	assert.Equal(t, []string{"Removed"}, removedByRole.Roles)
	assert.False(t, removedByRole.IsAdmin)
	require.NotNil(t, removedByRole.Active)
	assert.False(t, *removedByRole.Active)
}

func TestCursorRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    []string
		isAdmin bool
	}{
		{"owner", []string{"Owner"}, true},
		{"free-owner", []string{"Owner"}, true},
		{"member", []string{"Member"}, false},
		{"removed", []string{"Removed"}, false},
		{"unknown_future_role", []string{"unknown_future_role"}, false},
		{"", []string{}, false},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, cursorRoles(c.in))
			assert.Equal(t, c.isAdmin, cursorIsAdmin(c.in))
		})
	}
}
