// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type thirdPartyContact struct {
	ID       string  `json:"id"`
	FullName string  `json:"full_name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Role     *string `json:"role"`
}

func mcpThirdPartyContactInput(thirdPartyID, fullName, email string) map[string]any {
	return map[string]any{
		"third_party_id": thirdPartyID,
		"full_name":      fullName,
		"email":          email,
		"phone":          "+1-555-0100",
		"role":           "Account Manager",
	}
}

func TestMCP_AddThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	var result struct {
		ThirdPartyContact thirdPartyContact `json:"third_party_contact"`
	}
	mc.CallToolInto("addThirdPartyContact", mcpThirdPartyContactInput(
		thirdPartyID,
		"Alice Smith",
		"alice@example.com",
	), &result)

	assert.NotEmpty(t, result.ThirdPartyContact.ID)
	assert.Equal(t, "Alice Smith", result.ThirdPartyContact.FullName)
}

func TestMCP_UpdateThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyContact thirdPartyContact `json:"third_party_contact"`
	}
	mc.CallToolInto("addThirdPartyContact", mcpThirdPartyContactInput(
		thirdPartyID,
		"Bob Jones",
		"bob@example.com",
	), &addResult)
	require.NotEmpty(t, addResult.ThirdPartyContact.ID)

	// Update
	var updateResult struct {
		ThirdPartyContact thirdPartyContact `json:"third_party_contact"`
	}
	mc.CallToolInto("updateThirdPartyContact", map[string]any{
		"id":        addResult.ThirdPartyContact.ID,
		"full_name": "Robert Jones",
		"role":      "CTO",
	}, &updateResult)

	assert.Equal(t, addResult.ThirdPartyContact.ID, updateResult.ThirdPartyContact.ID)
	assert.Equal(t, "Robert Jones", updateResult.ThirdPartyContact.FullName)
}

func TestMCP_DeleteThirdPartyContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyContact thirdPartyContact `json:"third_party_contact"`
	}
	mc.CallToolInto("addThirdPartyContact", mcpThirdPartyContactInput(
		thirdPartyID,
		"Contact to delete",
		factory.SafeEmail(),
	), &addResult)
	require.NotEmpty(t, addResult.ThirdPartyContact.ID)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyContactID string `json:"deleted_third_party_contact_id"`
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
			ThirdPartyContact thirdPartyContact `json:"third_party_contact"`
		}
		mc.CallToolInto("addThirdPartyContact", mcpThirdPartyContactInput(
			thirdPartyID,
			factory.SafeName("Contact"),
			factory.SafeEmail(),
		), &result)
		require.NotEmpty(t, result.ThirdPartyContact.ID)

		_ = i
	}

	// List
	var listResult struct {
		ThirdPartyContacts []thirdPartyContact `json:"third_party_contacts"`
	}
	mc.CallToolInto("listThirdPartyContacts", map[string]any{
		"third_party_id": thirdPartyID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.ThirdPartyContacts), 3)
}
