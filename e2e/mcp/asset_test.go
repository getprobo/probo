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

func TestMCP_Asset_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	// Create
	var addResult struct {
		Asset struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			AssetType string `json:"assetType"`
		} `json:"asset"`
	}
	mc.CallToolInto("addAsset", map[string]any{
		"organizationId":  orgID,
		"name":            factory.SafeName("Asset"),
		"amount":          5,
		"ownerId":         profileID,
		"assetType":       "VIRTUAL",
		"dataTypesStored": "PII",
	}, &addResult)
	require.NotEmpty(t, addResult.Asset.ID)
	assert.Equal(t, "VIRTUAL", addResult.Asset.AssetType)

	// Get
	var getResult struct {
		Asset struct {
			ID string `json:"id"`
		} `json:"asset"`
	}
	mc.CallToolInto("getAsset", map[string]any{
		"id": addResult.Asset.ID,
	}, &getResult)
	assert.Equal(t, addResult.Asset.ID, getResult.Asset.ID)

	// Update
	var updateResult struct {
		Asset struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"asset"`
	}
	mc.CallToolInto("updateAsset", map[string]any{
		"id":   addResult.Asset.ID,
		"name": "Updated Asset",
	}, &updateResult)
	assert.Equal(t, "Updated Asset", updateResult.Asset.Name)

	// List
	var listResult struct {
		Assets []struct {
			ID string `json:"id"`
		} `json:"assets"`
	}
	mc.CallToolInto("listAssets", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Assets)

	// Delete
	var deleteResult struct {
		DeletedAssetID string `json:"deletedAssetId"`
	}
	mc.CallToolInto("deleteAsset", map[string]any{
		"id": addResult.Asset.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Asset.ID, deleteResult.DeletedAssetID)

	// Get deleted asset returns sanitized not-found error
	msg := mc.CallToolExpectError("getAsset", map[string]any{
		"id": addResult.Asset.ID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Asset_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectError("addAsset", map[string]any{
		"organizationId":  orgID,
		"name":            factory.SafeName("Asset"),
		"amount":          1,
		"assetType":       "VIRTUAL",
		"dataTypesStored": "PII",
	})
	assert.Contains(t, msg, "permission denied")
}
