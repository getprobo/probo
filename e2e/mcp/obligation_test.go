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

func TestMCP_Obligation_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Obligation struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"obligation"`
	}
	mc.CallToolInto("addObligation", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Obligation"),
		"description":    "Test obligation",
	}, &addResult)
	require.NotEmpty(t, addResult.Obligation.ID)

	// Get
	var getResult struct {
		Obligation struct {
			ID string `json:"id"`
		} `json:"obligation"`
	}
	mc.CallToolInto("getObligation", map[string]any{
		"id": addResult.Obligation.ID,
	}, &getResult)
	assert.Equal(t, addResult.Obligation.ID, getResult.Obligation.ID)

	// Update
	var updateResult struct {
		Obligation struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"obligation"`
	}
	mc.CallToolInto("updateObligation", map[string]any{
		"id":   addResult.Obligation.ID,
		"name": "Updated Obligation",
	}, &updateResult)
	assert.Equal(t, "Updated Obligation", updateResult.Obligation.Name)

	// List
	var listResult struct {
		Obligations []struct {
			ID string `json:"id"`
		} `json:"obligations"`
	}
	mc.CallToolInto("listObligations", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Obligations)

	// Delete
	var deleteResult struct {
		DeletedObligationID string `json:"deletedObligationId"`
	}
	mc.CallToolInto("deleteObligation", map[string]any{
		"id": addResult.Obligation.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Obligation.ID, deleteResult.DeletedObligationID)
}
