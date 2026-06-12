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

func TestDeepgramDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/deepgram", "DEEPGRAM_API_KEY")
	// Deepgram authenticates with the `Token` scheme. The matcher ignores
	// Authorization, so replay needs no auth; the value matters only when
	// re-recording.
	auth := ""
	if token := os.Getenv("DEEPGRAM_API_KEY"); token != "" {
		auth = "Token " + token
	}

	client := newVCRClient(rec, auth)

	driver := NewDeepgramDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	// owner@example.com appears in both projects and must be deduped.
	require.Len(t, records, 3)

	// owner@example.com is first seen with only ["member"] in project-1 and
	// gains ["owner"] in project-2. The Owner role / admin flag therefore
	// depend on the cross-project scope union actually taking effect.
	owner := records[0]
	assert.Equal(t, "m-0000-0000-0001", owner.ExternalID)
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "Olivia Owner", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)

	member := records[1]
	assert.Equal(t, "member@example.com", member.Email)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)

	dev := records[2]
	assert.Equal(t, "dev@example.com", dev.Email)
	assert.Equal(t, []string{"Member"}, dev.Roles)
}

func TestDeepgramRoles(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"Owner"}, deepgramRoles([]string{"owner"}))
	assert.Equal(t, []string{"Admin"}, deepgramRoles([]string{"admin", "read:transcripts"}))
	assert.Equal(t, []string{"Member"}, deepgramRoles([]string{"read:transcripts"}))
	assert.True(t, deepgramIsAdmin([]string{"owner"}))
	assert.True(t, deepgramIsAdmin([]string{"admin"}))
	assert.False(t, deepgramIsAdmin([]string{"member"}))
}

func TestDeepgramUnionScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"member", "owner"}, deepgramUnionScopes([]string{"member"}, []string{"owner"}))
	assert.Equal(t, []string{"a", "b"}, deepgramUnionScopes([]string{"a", "b"}, []string{"a"}))
	assert.Empty(t, deepgramUnionScopes(nil, nil))
}
