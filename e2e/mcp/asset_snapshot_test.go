// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestMCP_AssetSnapshot_PublishListAndGet(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)
	assetName := factory.SafeName("SnapshotAsset")

	var created struct {
		Asset struct {
			ID string `json:"id"`
		} `json:"asset"`
	}
	mc.CallToolInto("addAsset", map[string]any{
		"organization_id":   orgID,
		"name":              assetName,
		"amount":            3,
		"owner_id":          profileID,
		"asset_type":        "VIRTUAL",
		"data_types_stored": "PII",
	}, &created)
	require.NotEmpty(t, created.Asset.ID)

	var published struct {
		DocumentID string `json:"document_id"`
	}
	mc.CallToolInto("publishAssetList", map[string]any{
		"organization_id": orgID,
		"minor":           false,
		"name":            "Quarter",
		"note":            "mcp",
	}, &published)
	require.NotEmpty(t, published.DocumentID)

	var listed struct {
		AssetSnapshots []struct {
			ID   string  `json:"id"`
			Name *string `json:"name"`
		} `json:"asset_snapshots"`
	}
	mc.CallToolInto("listAssetSnapshots", map[string]any{
		"organization_id": orgID,
	}, &listed)
	require.Len(t, listed.AssetSnapshots, 1)
	require.NotNil(t, listed.AssetSnapshots[0].Name)
	assert.Equal(t, "Quarter", *listed.AssetSnapshots[0].Name)

	var got struct {
		AssetSnapshot struct {
			Note *string `json:"note"`
		} `json:"asset_snapshot"`
		Assets []struct {
			Name    string `json:"name"`
			Amount  int    `json:"amount"`
			AssetID string `json:"asset_id"`
		} `json:"assets"`
	}
	mc.CallToolInto("getAssetSnapshot", map[string]any{
		"id": listed.AssetSnapshots[0].ID,
	}, &got)
	require.NotNil(t, got.AssetSnapshot.Note)
	assert.Equal(t, "mcp", *got.AssetSnapshot.Note)
	require.NotEmpty(t, got.Assets)
	assert.Equal(t, assetName, got.Assets[0].Name)
	assert.Equal(t, 3, got.Assets[0].Amount)
	assert.Equal(t, created.Asset.ID, got.Assets[0].AssetID)
}
