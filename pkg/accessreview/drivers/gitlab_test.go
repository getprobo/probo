// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/coredata"
)

func TestGitLabDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/gitlab", "GITLAB_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GITLAB_TOKEN")))

	groupID := os.Getenv("GITLAB_GROUP_ID")
	if groupID == "" {
		groupID = "12345"
	}

	driver := NewGitLabDriver(client, groupID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.Roles)
	assert.Equal(t, coredata.MFAStatusUnknown, r.MFAStatus)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)
	assert.True(t, r.IsAdmin) // first record is access_level=50 (Owner)
}
