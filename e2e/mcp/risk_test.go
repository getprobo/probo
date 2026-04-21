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

func TestMCP_Risk_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("addRisk", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Risk"),
	}, &addResult)
	require.NotEmpty(t, addResult.Risk.ID)

	// Get
	var getResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("getRisk", map[string]any{
		"id": addResult.Risk.ID,
	}, &getResult)
	assert.Equal(t, addResult.Risk.ID, getResult.Risk.ID)

	// Update
	var updateResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("updateRisk", map[string]any{
		"id":   addResult.Risk.ID,
		"name": "Updated Risk",
	}, &updateResult)
	assert.Equal(t, "Updated Risk", updateResult.Risk.Name)

	// List
	var listResult struct {
		Risks []struct {
			ID string `json:"id"`
		} `json:"risks"`
	}
	mc.CallToolInto("listRisks", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Risks)

	// Delete
	var deleteResult struct {
		DeletedRiskID string `json:"deletedRiskId"`
	}
	mc.CallToolInto("deleteRisk", map[string]any{
		"id": addResult.Risk.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Risk.ID, deleteResult.DeletedRiskID)
}
