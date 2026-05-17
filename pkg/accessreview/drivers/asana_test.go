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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsanaDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/asana", "ASANA_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("ASANA_TOKEN")))

	workspaceGID := os.Getenv("ASANA_WORKSPACE_GID")
	if workspaceGID == "" {
		workspaceGID = "9999999"
	}

	driver := NewAsanaDriver(client, workspaceGID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.Email)

	// Records without an email should be marked Active=false.
	require.Len(t, records, 2)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)
}
