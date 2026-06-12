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

func TestLangfuseDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/langfuse", "LANGFUSE_API_KEY")
	// Langfuse authenticates with HTTP Basic (publicKey:secretKey). The
	// matcher ignores Authorization, so replay needs no auth.
	client := newVCRClient(rec, basicAuthUserPass(os.Getenv("LANGFUSE_API_KEY")))

	driver := NewLangfuseDriver(client, "https://cloud.langfuse.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	owner := records[0]
	assert.Equal(t, "lf-user-1", owner.ExternalID)
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "Olivia Owner", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)

	admin := records[1]
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)

	member := records[2]
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)

	viewer := records[3]
	assert.Equal(t, []string{"Viewer"}, viewer.Roles)
	assert.False(t, viewer.IsAdmin)
	// name is empty in the payload, so the email is used as the display name.
	assert.Equal(t, "viewer@example.com", viewer.FullName)
}

func TestLangfuseRoles(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"Owner"}, langfuseRoles("OWNER"))
	assert.Equal(t, []string{"Admin"}, langfuseRoles("ADMIN"))
	assert.Equal(t, []string{"Member"}, langfuseRoles("MEMBER"))
	assert.Equal(t, []string{"Viewer"}, langfuseRoles("VIEWER"))
	assert.Equal(t, []string{"None"}, langfuseRoles("NONE"))
	assert.True(t, langfuseIsAdmin("OWNER"))
	assert.True(t, langfuseIsAdmin("ADMIN"))
	assert.False(t, langfuseIsAdmin("MEMBER"))
}
