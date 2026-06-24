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

func TestOpenRouterDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/openrouter", "OPENROUTER_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("OPENROUTER_API_KEY")))

	driver := NewOpenRouterDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	// Cassette recorded live against an OpenRouter organization (single admin
	// member), then anonymized.
	admin := records[0]
	assert.Equal(t, "user_000000000000000000000admin", admin.ExternalID)
	assert.Equal(t, "ada.admin@example.com", admin.Email)
	assert.Equal(t, "Ada Admin", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.True(t, admin.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
	// The members endpoint carries no account-status field, so Active stays nil.
	assert.Nil(t, admin.Active)
}

func TestOpenRouterRoles(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"Admin"}, openRouterRoles("org:admin"))
	assert.Equal(t, []string{"Member"}, openRouterRoles("org:member"))
	// An unknown future role is preserved verbatim; an empty role yields none.
	assert.Equal(t, []string{"org:billing"}, openRouterRoles("org:billing"))
	assert.Equal(t, []string{}, openRouterRoles(""))
}

func TestOpenRouterFullName(t *testing.T) {
	t.Parallel()

	first, last := "Bob", "Member"

	// first + last.
	assert.Equal(t, "Bob Member", openRouterFullName(openRouterMember{FirstName: &first, LastName: &last}, "bob@example.com"))
	// last_name null → first name alone.
	assert.Equal(t, "Bob", openRouterFullName(openRouterMember{FirstName: &first}, "bob@example.com"))
	// first_name null → last name alone.
	assert.Equal(t, "Member", openRouterFullName(openRouterMember{LastName: &last}, "bob@example.com"))
	// both null → email fallback.
	assert.Equal(t, "carol@example.com", openRouterFullName(openRouterMember{}, "carol@example.com"))
}
