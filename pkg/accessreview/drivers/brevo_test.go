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
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBrevoDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/brevo", "BREVO_API_KEY")
	// Brevo authenticates via the api-key header, not Authorization.
	client := newVCRClientWithHeader(rec, "api-key", os.Getenv("BREVO_API_KEY"))

	driver := NewBrevoDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Cassette recorded live (api-key header), then anonymized: the owner has
	// every feature at "owner"; the two members have crm/transactional "full"
	// and the rest "none".
	owner := records[0]
	assert.Equal(t, "000000000000000000000001", owner.ExternalID)
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "owner@example.com", owner.FullName)
	assert.True(t, owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	assert.Equal(t, []string{"owner"}, owner.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)

	// Non-owner; "none" levels filtered, the remaining "full" de-duplicated.
	member := records[1]
	assert.Equal(t, "000000000000000000000002", member.ExternalID)
	assert.False(t, member.IsAdmin)
	assert.Equal(t, []string{"full"}, member.Roles)
	require.NotNil(t, member.Active)
	assert.True(t, *member.Active)

	// The third record is asserted too, so a swap or corruption is caught.
	viewer := records[2]
	assert.Equal(t, "000000000000000000000003", viewer.ExternalID)
	assert.Equal(t, "viewer@example.com", viewer.Email)
	assert.False(t, viewer.IsAdmin)
	assert.Equal(t, []string{"full"}, viewer.Roles)
}

func TestBrevoExternalID(t *testing.T) {
	t.Parallel()

	// The stable id is preferred when present.
	assert.Equal(t, "abc123", brevoExternalID(brevoInvitedUser{ID: "abc123"}, "x@example.com"))
	// With no id, the email is the fallback.
	assert.Equal(t, "x@example.com", brevoExternalID(brevoInvitedUser{}, "x@example.com"))
	assert.Equal(t, "x@example.com", brevoExternalID(brevoInvitedUser{ID: "  "}, "x@example.com"))
}

func TestBrevoIsOwner(t *testing.T) {
	t.Parallel()

	// The live API returns a JSON boolean; older docs/SDK show a string.
	// Both must be tolerated.
	assert.True(t, brevoIsOwner(json.RawMessage(`true`)))
	assert.False(t, brevoIsOwner(json.RawMessage(`false`)))
	assert.True(t, brevoIsOwner(json.RawMessage(`"true"`)))
	assert.False(t, brevoIsOwner(json.RawMessage(`"false"`)))
	assert.False(t, brevoIsOwner(nil))
}

func TestBrevoRoles(t *testing.T) {
	t.Parallel()

	// Distinct non-"none" levels, sorted; "none" is filtered out.
	roles := brevoRoles(map[string]json.RawMessage{
		"marketing":     json.RawMessage(`"owner"`),
		"conversations": json.RawMessage(`"owner"`),
		"crm":           json.RawMessage(`"none"`),
	})
	assert.Equal(t, []string{"owner"}, roles)

	// All "none" → no roles.
	assert.Empty(t, brevoRoles(map[string]json.RawMessage{"crm": json.RawMessage(`"none"`)}))

	// A non-string shape is ignored rather than failing.
	assert.Empty(t, brevoRoles(map[string]json.RawMessage{"crm": json.RawMessage(`{"x":1}`)}))
}
