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

func TestMCP_Datum_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	// Create
	var addResult struct {
		Datum struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"datum"`
	}
	mc.CallToolInto("addDatum", map[string]any{
		"organizationId":     orgID,
		"name":               factory.SafeName("Datum"),
		"ownerId":            profileID,
		"dataClassification": "PUBLIC",
	}, &addResult)
	require.NotEmpty(t, addResult.Datum.ID)

	// Get
	var getResult struct {
		Datum struct {
			ID string `json:"id"`
		} `json:"datum"`
	}
	mc.CallToolInto("getDatum", map[string]any{
		"id": addResult.Datum.ID,
	}, &getResult)
	assert.Equal(t, addResult.Datum.ID, getResult.Datum.ID)

	// Update
	var updateResult struct {
		Datum struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"datum"`
	}
	mc.CallToolInto("updateDatum", map[string]any{
		"id":   addResult.Datum.ID,
		"name": "Updated Datum",
	}, &updateResult)
	assert.Equal(t, "Updated Datum", updateResult.Datum.Name)

	// List
	var listResult struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	mc.CallToolInto("listData", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Data)

	// Delete
	var deleteResult struct {
		DeletedDatumID string `json:"deletedDatumId"`
	}
	mc.CallToolInto("deleteDatum", map[string]any{
		"id": addResult.Datum.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Datum.ID, deleteResult.DeletedDatumID)
}
