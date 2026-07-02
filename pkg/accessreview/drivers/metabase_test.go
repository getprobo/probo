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
