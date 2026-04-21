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

func TestMCP_Finding_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Finding struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"finding"`
	}
	mc.CallToolInto("addFinding", map[string]any{
		"organizationId": orgID,
		"title":          factory.SafeName("Finding"),
	}, &addResult)
	require.NotEmpty(t, addResult.Finding.ID)

	// Get
	var getResult struct {
		Finding struct {
			ID string `json:"id"`
		} `json:"finding"`
	}
	mc.CallToolInto("getFinding", map[string]any{
		"id": addResult.Finding.ID,
	}, &getResult)
	assert.Equal(t, addResult.Finding.ID, getResult.Finding.ID)

	// Update
	var updateResult struct {
		Finding struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"finding"`
	}
	mc.CallToolInto("updateFinding", map[string]any{
		"id":    addResult.Finding.ID,
		"title": "Updated Finding",
	}, &updateResult)
	assert.Equal(t, "Updated Finding", updateResult.Finding.Title)

	// List
	var listResult struct {
		Findings []struct {
			ID string `json:"id"`
		} `json:"findings"`
	}
	mc.CallToolInto("listFindings", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Findings)

	// Delete
	var deleteResult struct {
		DeletedFindingID string `json:"deletedFindingId"`
	}
	mc.CallToolInto("deleteFinding", map[string]any{
		"id": addResult.Finding.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Finding.ID, deleteResult.DeletedFindingID)
}
