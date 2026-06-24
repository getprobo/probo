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
