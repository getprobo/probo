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
