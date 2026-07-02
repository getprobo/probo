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

func TestMondayDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/monday", "MONDAY_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("MONDAY_TOKEN")))

	driver := NewMondayDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.FullName)
	assert.True(t, r.IsAdmin)
	assert.NotEmpty(t, r.JobTitle)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	require.NotNil(t, r.LastLogin)

	// Pending users should surface as Active=false even when enabled.
	require.Len(t, records, 2)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
}
