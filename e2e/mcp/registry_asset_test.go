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

func TestMCP_Asset_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	// Create
	var addResult struct {
		RegistryAsset struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			RegistryAssetType string `json:"registry_asset_type"`
		} `json:"registry_asset"`
	}
	mc.CallToolInto("addRegistryAsset", map[string]any{
		"organization_id":     orgID,
		"name":                factory.SafeName("RegistryAsset"),
		"amount":              5,
		"owner_id":            profileID,
		"registry_asset_type": "VIRTUAL",
		"data_types_stored":   "PII",
	}, &addResult)
	require.NotEmpty(t, addResult.RegistryAsset.ID)
	assert.Equal(t, "VIRTUAL", addResult.RegistryAsset.RegistryAssetType)

	// Get
	var getResult struct {
		RegistryAsset struct {
			ID string `json:"id"`
		} `json:"registry_asset"`
	}
	mc.CallToolInto("getRegistryAsset", map[string]any{
		"id": addResult.RegistryAsset.ID,
	}, &getResult)
	assert.Equal(t, addResult.RegistryAsset.ID, getResult.RegistryAsset.ID)

	// Update
	var updateResult struct {
		RegistryAsset struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"registry_asset"`
	}
	mc.CallToolInto("updateRegistryAsset", map[string]any{
		"id":   addResult.RegistryAsset.ID,
		"name": "Updated RegistryAsset",
	}, &updateResult)
	assert.Equal(t, "Updated RegistryAsset", updateResult.RegistryAsset.Name)

	// List
	var listResult struct {
		RegistryAssets []struct {
			ID string `json:"id"`
		} `json:"registry_assets"`
	}
	mc.CallToolInto("listRegistryAssets", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.RegistryAssets)

	// Delete
	var deleteResult struct {
		DeletedAssetID string `json:"deleted_registry_asset_id"`
	}
	mc.CallToolInto("deleteRegistryAsset", map[string]any{
		"id": addResult.RegistryAsset.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.RegistryAsset.ID, deleteResult.DeletedAssetID)

	// Get deleted asset returns sanitized not-found error
	msg := mc.CallToolExpectToolError("getRegistryAsset", map[string]any{
		"id": addResult.RegistryAsset.ID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Asset_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("addRegistryAsset", map[string]any{
		"organization_id":     orgID,
		"name":                factory.SafeName("RegistryAsset"),
		"amount":              1,
		"owner_id":            factory.CreateUser(owner),
		"registry_asset_type": "VIRTUAL",
		"data_types_stored":   "PII",
	})
	assert.Contains(t, msg, "permission denied")
}
