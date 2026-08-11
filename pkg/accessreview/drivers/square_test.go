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

func TestSquareDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/square", "SQUARE_ACCESS_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SQUARE_ACCESS_TOKEN")))

	driver := NewSquareDriver(client, "https://connect.squareup.com/v2")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Non-owner, ACTIVE → Member, active, not admin.
	member := records[0]
	assert.Equal(t, "-3oZQKPKVk6gUXU_V5Qa", member.ExternalID)
	assert.Equal(t, "sherlock.holmes@example.com", member.Email)
	assert.Equal(t, "Sherlock Holmes", member.FullName)
	assert.Equal(t, new(false), member.IsAdmin)
	assert.Equal(t, []string{"Member"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)
	// parseRFC3339Ptr nils on any format mismatch, so pin the parsed instant:
	// a silently dropped created_at would otherwise look like a clean pass.
	require.NotNil(t, member.CreatedAt)
	assert.Equal(t, "2020-03-24T18:14:26Z", member.CreatedAt.UTC().Format(time.RFC3339))

	// Owner → is_owner drives IsAdmin directly.
	owner := records[1]
	assert.Equal(t, "Pw67AzUomYUdF04AN17i", owner.ExternalID)
	assert.Equal(t, "john.watson@example.com", owner.Email)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)

	// INACTIVE member → active=false.
	inactive := records[2]
	assert.Equal(t, "martha.hudson@example.com", inactive.Email)
	assert.Equal(t, new(false), inactive.IsAdmin)
	require.NotNil(t, inactive.Active)
	assert.False(t, *inactive.Active)
}
