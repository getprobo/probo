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
