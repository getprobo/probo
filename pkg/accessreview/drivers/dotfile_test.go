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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotfileDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/dotfile", "DOTFILE_API_KEY")
	// Dotfile authenticates via the X-DOTFILE-API-KEY header, not Authorization.
	client := newVCRClientWithHeader(rec, "X-DOTFILE-API-KEY", os.Getenv("DOTFILE_API_KEY"))

	driver := NewDotfileDriver(client, "https://api.dotfile.com/v1")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Owner: active, admin, id-first ExternalID, first+last name joined.
	owner := records[0]
	assert.Equal(t, "a1b2c3d4-0000-4000-8000-000000000001", owner.ExternalID)
	assert.Equal(t, "alice.martin@example.com", owner.Email)
	assert.Equal(t, "Alice Martin", owner.FullName)
	assert.Equal(t, new(true), owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	assert.Equal(t, []string{"owner"}, owner.Roles)
	// parseRFC3339Ptr nils on any format mismatch, so pin the parsed instant:
	// a silently dropped created_at would otherwise look like a clean pass.
	require.NotNil(t, owner.CreatedAt)
	assert.Equal(t, "2023-01-31T13:30:00Z", owner.CreatedAt.UTC().Format(time.RFC3339))

	// Admin: also administrative.
	admin := records[1]
	assert.Equal(t, "bob.durand@example.com", admin.Email)
	assert.Equal(t, new(true), admin.IsAdmin)
	assert.Equal(t, []string{"admin"}, admin.Roles)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)

	// Member: suspended (suspended_at set) → inactive, not admin.
	member := records[2]
	assert.Equal(t, "carla.petit@example.com", member.Email)
	assert.Equal(t, new(false), member.IsAdmin)
	assert.Equal(t, []string{"member"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.False(t, *member.Active)
}
