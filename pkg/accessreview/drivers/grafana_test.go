// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	assert.Equal(t, "Admin", records[0].Role)
	assert.True(t, records[0].IsAdmin)
	assert.Equal(t, strconv.Itoa(1), records[0].ExternalID)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].LastLogin)

	assert.Equal(t, "viewer@example.com", records[1].Email)
	assert.Equal(t, "Viewer User", records[1].FullName)
	assert.Equal(t, "Viewer", records[1].Role)
	assert.False(t, records[1].IsAdmin)
	assert.Equal(t, strconv.Itoa(2), records[1].ExternalID)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	resolver := NewGrafanaNameResolver(client, baseURL)
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Acme Grafana", name)
}
