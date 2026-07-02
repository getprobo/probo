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

func TestIncidentIODriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/incidentio", "INCIDENT_IO_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("INCIDENT_IO_API_KEY")))

	driver := NewIncidentIODriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	// Three records spread across two pages: the cassette's first page is
	// short (2 < page_size) but carries a non-empty `after`, so getting all
	// three proves the driver follows the cursor instead of stopping early.
	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "01ABCOWNER", owner.ExternalID)
	assert.Equal(t, "lisa@example.com", owner.Email)
	assert.Equal(t, "Lisa Curtis", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	// /v2/users carries no account-status field, so Active stays nil.
	assert.Nil(t, owner.Active)

	// Live base_role + custom_roles are unioned; a responder is not an admin.
	responder := records[1]
	assert.Equal(t, []string{"Responder", "On-call Lead"}, responder.Roles)
	assert.False(t, responder.IsAdmin)

	// base_role null → falls back to the deprecated role enum; empty name →
	// the display name falls back to the email.
	legacy := records[2]
	assert.Equal(t, "legacy-admin@example.com", legacy.FullName)
	assert.Equal(t, []string{"Administrator"}, legacy.Roles)
	assert.True(t, legacy.IsAdmin)
}

func TestIncidentIORoles(t *testing.T) {
	t.Parallel()

	// base_role + custom_roles are unioned, base first.
	full := incidentIOUser{
		BaseRole:    &incidentIORole{Name: "Owner", Slug: "owner"},
		CustomRoles: []incidentIORole{{Name: "On-call Lead", Slug: "on-call-lead"}},
		Role:        "viewer",
	}
	assert.Equal(t, []string{"Owner", "On-call Lead"}, incidentIORoles(full))

	// No live RBAC role → falls back to the deprecated enum.
	assert.Equal(t, []string{"Responder"}, incidentIORoles(incidentIOUser{Role: "responder"}))

	// No role at all → empty slice.
	assert.Equal(t, []string{}, incidentIORoles(incidentIOUser{Role: "unset"}))
}

func TestIncidentIOIsAdmin(t *testing.T) {
	t.Parallel()

	// The live base_role slug wins, including the "administrator" slug that
	// the cassette does not exercise.
	assert.True(t, incidentIOIsAdmin(incidentIOUser{BaseRole: &incidentIORole{Slug: "owner"}}))
	assert.True(t, incidentIOIsAdmin(incidentIOUser{BaseRole: &incidentIORole{Slug: "administrator"}}))
	assert.False(t, incidentIOIsAdmin(incidentIOUser{BaseRole: &incidentIORole{Slug: "responder"}}))
	// A non-admin base_role is NOT overridden by an admin deprecated role.
	assert.False(t, incidentIOIsAdmin(incidentIOUser{BaseRole: &incidentIORole{Slug: "viewer"}, Role: "administrator"}))
	// No base_role → falls back to the deprecated role enum.
	assert.True(t, incidentIOIsAdmin(incidentIOUser{Role: "administrator"}))
	assert.True(t, incidentIOIsAdmin(incidentIOUser{Role: "owner"}))
	assert.False(t, incidentIOIsAdmin(incidentIOUser{Role: "viewer"}))
}

func TestIncidentIODeprecatedRoleName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Owner", incidentIODeprecatedRoleName("owner"))
	assert.Equal(t, "Administrator", incidentIODeprecatedRoleName("administrator"))
	assert.Equal(t, "Responder", incidentIODeprecatedRoleName("responder"))
	assert.Equal(t, "Viewer", incidentIODeprecatedRoleName("viewer"))
	// "unset" and an absent value map to no role.
	assert.Equal(t, "", incidentIODeprecatedRoleName("unset"))
	assert.Equal(t, "", incidentIODeprecatedRoleName(""))
}
