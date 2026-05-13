// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type thirdPartyContact struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
	Role  *string `json:"role"`
}

func TestMCP_AddThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	var result struct {
		ThirdPartyContact thirdPartyContact `json:"thirdPartyContact"`
	}
	mc.CallToolInto("addThirdPartyContact", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Alice Smith",
		"email":        "alice@example.com",
		"phone":        "+1-555-0100",
		"role":         "Account Manager",
	}, &result)

	assert.NotEmpty(t, result.ThirdPartyContact.ID)
	assert.Equal(t, "Alice Smith", result.ThirdPartyContact.Name)
}

func TestMCP_UpdateThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyContact thirdPartyContact `json:"thirdPartyContact"`
	}
	mc.CallToolInto("addThirdPartyContact", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Bob Jones",
		"email":        "bob@example.com",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdPartyContact.ID)

	// Update
	var updateResult struct {
		ThirdPartyContact thirdPartyContact `json:"thirdPartyContact"`
	}
	mc.CallToolInto("updateThirdPartyContact", map[string]any{
		"id":   addResult.ThirdPartyContact.ID,
		"name": "Robert Jones",
		"role": "CTO",
	}, &updateResult)

	assert.Equal(t, addResult.ThirdPartyContact.ID, updateResult.ThirdPartyContact.ID)
	assert.Equal(t, "Robert Jones", updateResult.ThirdPartyContact.Name)
}

func TestMCP_DeleteThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyContact thirdPartyContact `json:"thirdPartyContact"`
	}
	mc.CallToolInto("addThirdPartyContact", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Contact to delete",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdPartyContact.ID)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyContactID string `json:"deletedThirdPartyContactId"`
	}
	mc.CallToolInto("deleteThirdPartyContact", map[string]any{
		"id": addResult.ThirdPartyContact.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.ThirdPartyContact.ID, deleteResult.DeletedThirdPartyContactID)
}

func TestMCP_ListThirdPartyContacts(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create contacts
	for i := range 3 {
		var result struct {
			ThirdPartyContact thirdPartyContact `json:"thirdPartyContact"`
		}
		mc.CallToolInto("addThirdPartyContact", map[string]any{
			"thirdPartyId": thirdPartyID,
			"name":         factory.SafeName("Contact"),
			"email":        factory.SafeEmail(),
		}, &result)
		require.NotEmpty(t, result.ThirdPartyContact.ID)
		_ = i
	}

	// List
	var listResult struct {
		ThirdPartyContacts []thirdPartyContact `json:"thirdPartyContacts"`
	}
	mc.CallToolInto("listThirdPartyContacts", map[string]any{
		"thirdPartyId": thirdPartyID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.ThirdPartyContacts), 3)
}
