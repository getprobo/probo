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

package console_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAsset_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	const query = `
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge {
					node {
						id
						name
						amount
						assetType
						dataTypesStored
						owner {
							id
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID              string `json:"id"`
					Name            string `json:"name"`
					Amount          int    `json:"amount"`
					AssetType       string `json:"assetType"`
					DataTypesStored string `json:"dataTypesStored"`
					Owner           struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":  owner.GetOrganizationID().String(),
			"name":            "Production Database Server",
			"amount":          5,
			"ownerId":         profileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Customer PII, Financial Records",
		},
	}, &result)
	require.NoError(t, err)

	asset := result.CreateAsset.AssetEdge.Node
	assert.NotEmpty(t, asset.ID)
	assert.Equal(t, "Production Database Server", asset.Name)
	assert.Equal(t, 5, asset.Amount)
	assert.Equal(t, "VIRTUAL", asset.AssetType)
	assert.Equal(t, "Customer PII, Financial Records", asset.DataTypesStored)
	assert.Equal(t, profileID, asset.Owner.ID)
}

func TestAsset_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	const createQuery = `
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":  owner.GetOrganizationID().String(),
			"name":            "Test Asset",
			"amount":          10,
			"ownerId":         profileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Test data",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateAsset.AssetEdge.Node.ID

	const query = `
		mutation($input: UpdateAssetInput!) {
			updateAsset(input: $input) {
				asset {
					id
					name
					amount
					dataTypesStored
				}
			}
		}
	`

	var result struct {
		UpdateAsset struct {
			Asset struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				Amount          int    `json:"amount"`
				DataTypesStored string `json:"dataTypesStored"`
			} `json:"asset"`
		} `json:"updateAsset"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":              assetID,
			"name":            "Updated Asset Name",
			"amount":          20,
			"dataTypesStored": "Updated data types",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, assetID, result.UpdateAsset.Asset.ID)
	assert.Equal(t, "Updated Asset Name", result.UpdateAsset.Asset.Name)
	assert.Equal(t, 20, result.UpdateAsset.Asset.Amount)
}

func TestAsset_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	const createQuery = `
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":  owner.GetOrganizationID().String(),
			"name":            "Asset to delete",
			"amount":          1,
			"ownerId":         profileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "None",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateAsset.AssetEdge.Node.ID

	const query = `
		mutation($input: DeleteAssetInput!) {
			deleteAsset(input: $input) {
				deletedAssetId
			}
		}
	`

	var result struct {
		DeleteAsset struct {
			DeletedAssetID string `json:"deletedAssetId"`
		} `json:"deleteAsset"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"assetId": assetID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, assetID, result.DeleteAsset.DeletedAssetID)
}

func TestAsset_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	// Create multiple assets
	for i := range 3 {
		const query = `
			mutation($input: CreateAssetInput!) {
				createAsset(input: $input) {
					assetEdge {
						node {
							id
						}
					}
				}
			}
		`

		var result struct {
			CreateAsset struct {
				AssetEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"assetEdge"`
			} `json:"createAsset"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":  owner.GetOrganizationID().String(),
				"name":            fmt.Sprintf("Asset %c", 'A'+i),
				"amount":          i + 1,
				"ownerId":         profileID,
				"assetType":       "VIRTUAL",
				"dataTypesStored": "Test data",
			},
		}, &result)
		require.NoError(t, err)
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					assets(first: 10) {
						edges {
							node {
								id
								name
								amount
								assetType
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Assets struct {
				Edges []struct {
					Node struct {
						ID        string `json:"id"`
						Name      string `json:"name"`
						Amount    int    `json:"amount"`
						AssetType string `json:"assetType"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"assets"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Assets.TotalCount, 3)
}

func TestAsset_Types(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	assetTypes := []string{"PHYSICAL", "VIRTUAL"}

	for _, assetType := range assetTypes {
		t.Run(assetType, func(t *testing.T) {
			const query = `
				mutation($input: CreateAssetInput!) {
					createAsset(input: $input) {
						assetEdge {
							node {
								id
								assetType
							}
						}
					}
				}
			`

			var result struct {
				CreateAsset struct {
					AssetEdge struct {
						Node struct {
							ID        string `json:"id"`
							AssetType string `json:"assetType"`
						} `json:"node"`
					} `json:"assetEdge"`
				} `json:"createAsset"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId":  owner.GetOrganizationID().String(),
					"name":            "Asset " + assetType,
					"amount":          1,
					"ownerId":         profileID,
					"assetType":       assetType,
					"dataTypesStored": "Test data",
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, assetType, result.CreateAsset.AssetEdge.Node.AssetType)
		})
	}
}
