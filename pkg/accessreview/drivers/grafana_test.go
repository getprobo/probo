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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrafanaDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/grafana", "GRAFANA_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GRAFANA_TOKEN")))

	baseURL := os.Getenv("GRAFANA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://grafana.example.com"
	}

	driver := NewGrafanaDriver(client, baseURL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, strconv.Itoa(1), records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].LastLogin)

	assert.Equal(t, "viewer@example.com", records[1].Email)
	assert.Equal(t, "Viewer User", records[1].FullName)
	assert.Equal(t, []string{"Viewer"}, records[1].Roles)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, strconv.Itoa(2), records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	resolver := NewGrafanaNameResolver(client, baseURL)
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Acme Grafana", name)
}
