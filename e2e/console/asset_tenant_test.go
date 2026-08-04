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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAsset_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(org1Owner)

	var createResult struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":  org1Owner.GetOrganizationID().String(),
			"name":            "Org1 Asset",
			"amount":          1,
			"ownerId":         profileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Test data",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateAsset.AssetEdge.Node.ID

	t.Run("cannot read asset from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Asset {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": assetID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "asset")
	})

	t.Run("cannot update asset from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: UpdateAssetInput!) {
				updateAsset(input: $input) {
					asset { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":   assetID,
				"name": "Hijacked Asset",
			},
		})
		require.Error(t, err, "Should not be able to update asset from another org")
	})

	t.Run("cannot delete asset from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteAssetInput!) {
				deleteAsset(input: $input) {
					deletedAssetId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"assetId": assetID,
			},
		})
		require.Error(t, err, "Should not be able to delete asset from another org")
	})

	t.Run("cannot create asset referencing an owner from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)

		_, err := org1Owner.Do(`
			mutation($input: CreateAssetInput!) {
				createAsset(input: $input) {
					assetEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":  org1Owner.GetOrganizationID().String(),
				"name":            factory.SafeName("Asset"),
				"amount":          1,
				"ownerId":         org2ProfileID,
				"assetType":       "VIRTUAL",
				"dataTypesStored": "Test data",
			},
		})
		require.Error(t, err, "must not accept an ownerId belonging to another organization")
	})

	t.Run("cannot create asset referencing a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty").Create()

		_, err := org1Owner.Do(`
			mutation($input: CreateAssetInput!) {
				createAsset(input: $input) {
					assetEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":  org1Owner.GetOrganizationID().String(),
				"name":            factory.SafeName("Asset"),
				"amount":          1,
				"ownerId":         profileID,
				"assetType":       "VIRTUAL",
				"dataTypesStored": "Test data",
				"thirdPartyIds":   []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})

	t.Run("cannot update asset to reference a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty for Update").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateAssetInput!) {
				updateAsset(input: $input) {
					asset { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":            assetID,
				"thirdPartyIds": []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
