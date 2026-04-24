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

func TestMCP_ThirdParty_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"thirdParty"`
	}
	name := factory.SafeName("ThirdParty")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organizationId": orgID,
		"name":           name,
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdParty.ID)
	assert.Equal(t, name, addResult.ThirdParty.Name)

	// Update
	var updateResult struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"thirdParty"`
	}
	mc.CallToolInto("updateThirdParty", map[string]any{
		"id":   addResult.ThirdParty.ID,
		"name": "Updated ThirdParty",
	}, &updateResult)
	assert.Equal(t, "Updated ThirdParty", updateResult.ThirdParty.Name)

	// List
	var listResult struct {
		ThirdParties []struct {
			ID string `json:"id"`
		} `json:"thirdParties"`
	}
	mc.CallToolInto("listThirdParties", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.ThirdParties)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyID string `json:"deletedThirdPartyId"`
	}
	mc.CallToolInto("deleteThirdParty", map[string]any{
		"id": addResult.ThirdParty.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.ThirdParty.ID, deleteResult.DeletedThirdPartyID)
}
