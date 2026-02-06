// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package console_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAsset_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
			"ownerId":         profileID.String(),
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
	assert.Equal(t, profileID.String(), asset.Owner.ID)
}

func TestAsset_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
			"ownerId":         profileID.String(),
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

	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
			"ownerId":         profileID.String(),
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

	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	// Create multiple assets
	for i := 0; i < 3; i++ {
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
				"ownerId":         profileID.String(),
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

	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
					"ownerId":         profileID.String(),
					"assetType":       assetType,
					"dataTypesStored": "Test data",
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, assetType, result.CreateAsset.AssetEdge.Node.AssetType)
		})
	}
}
