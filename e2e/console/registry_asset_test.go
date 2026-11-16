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
		mutation($input: CreateRegistryAssetInput!) {
			createRegistryAsset(input: $input) {
				registryAssetEdge {
					node {
						id
						name
						amount
						registryAssetType
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
		CreateRegistryAsset struct {
			RegistryAssetEdge struct {
				Node struct {
					ID                string `json:"id"`
					Name              string `json:"name"`
					Amount            int    `json:"amount"`
					RegistryAssetType string `json:"registryAssetType"`
					DataTypesStored   string `json:"dataTypesStored"`
					Owner             struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"registryAssetEdge"`
		} `json:"createRegistryAsset"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"name":              "Production Database Server",
			"amount":            5,
			"ownerId":           profileID,
			"registryAssetType": "VIRTUAL",
			"dataTypesStored":   "Customer PII, Financial Records",
		},
	}, &result)
	require.NoError(t, err)

	asset := result.CreateRegistryAsset.RegistryAssetEdge.Node
	assert.NotEmpty(t, asset.ID)
	assert.Equal(t, "Production Database Server", asset.Name)
	assert.Equal(t, 5, asset.Amount)
	assert.Equal(t, "VIRTUAL", asset.RegistryAssetType)
	assert.Equal(t, "Customer PII, Financial Records", asset.DataTypesStored)
	assert.Equal(t, profileID, asset.Owner.ID)
}

func TestAsset_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	const createQuery = `
		mutation($input: CreateRegistryAssetInput!) {
			createRegistryAsset(input: $input) {
				registryAssetEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateRegistryAsset struct {
			RegistryAssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"registryAssetEdge"`
		} `json:"createRegistryAsset"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"name":              "Test RegistryAsset",
			"amount":            10,
			"ownerId":           profileID,
			"registryAssetType": "VIRTUAL",
			"dataTypesStored":   "Test data",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateRegistryAsset.RegistryAssetEdge.Node.ID

	const query = `
		mutation($input: UpdateRegistryAssetInput!) {
			updateRegistryAsset(input: $input) {
				registryAsset {
					id
					name
					amount
					dataTypesStored
				}
			}
		}
	`

	var result struct {
		UpdateRegistryAsset struct {
			RegistryAsset struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				Amount          int    `json:"amount"`
				DataTypesStored string `json:"dataTypesStored"`
			} `json:"registryAsset"`
		} `json:"updateRegistryAsset"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":              assetID,
			"name":            "Updated RegistryAsset Name",
			"amount":          20,
			"dataTypesStored": "Updated data types",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, assetID, result.UpdateRegistryAsset.RegistryAsset.ID)
	assert.Equal(t, "Updated RegistryAsset Name", result.UpdateRegistryAsset.RegistryAsset.Name)
	assert.Equal(t, 20, result.UpdateRegistryAsset.RegistryAsset.Amount)
}

func TestAsset_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	const createQuery = `
		mutation($input: CreateRegistryAssetInput!) {
			createRegistryAsset(input: $input) {
				registryAssetEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateRegistryAsset struct {
			RegistryAssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"registryAssetEdge"`
		} `json:"createRegistryAsset"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"name":              "RegistryAsset to delete",
			"amount":            1,
			"ownerId":           profileID,
			"registryAssetType": "VIRTUAL",
			"dataTypesStored":   "None",
		},
	}, &createResult)
	require.NoError(t, err)

	assetID := createResult.CreateRegistryAsset.RegistryAssetEdge.Node.ID

	const query = `
		mutation($input: DeleteRegistryAssetInput!) {
			deleteRegistryAsset(input: $input) {
				deletedRegistryAssetId
			}
		}
	`

	var result struct {
		DeleteRegistryAsset struct {
			DeletedAssetID string `json:"deletedRegistryAssetId"`
		} `json:"deleteRegistryAsset"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"registryAssetId": assetID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, assetID, result.DeleteRegistryAsset.DeletedAssetID)
}

func TestAsset_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	// Create multiple assets
	for i := range 3 {
		const query = `
			mutation($input: CreateRegistryAssetInput!) {
				createRegistryAsset(input: $input) {
					registryAssetEdge {
						node {
							id
						}
					}
				}
			}
		`

		var result struct {
			CreateRegistryAsset struct {
				RegistryAssetEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"registryAssetEdge"`
			} `json:"createRegistryAsset"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"name":              fmt.Sprintf("RegistryAsset %c", 'A'+i),
				"amount":            i + 1,
				"ownerId":           profileID,
				"registryAssetType": "VIRTUAL",
				"dataTypesStored":   "Test data",
			},
		}, &result)
		require.NoError(t, err)
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					registryAssets(first: 10) {
						edges {
							node {
								id
								name
								amount
								registryAssetType
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
			RegistryAssets struct {
				Edges []struct {
					Node struct {
						ID                string `json:"id"`
						Name              string `json:"name"`
						Amount            int    `json:"amount"`
						RegistryAssetType string `json:"registryAssetType"`
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
	assert.GreaterOrEqual(t, result.Node.RegistryAssets.TotalCount, 3)
}

func TestAsset_Types(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(owner)

	registryAssetTypes := []string{"PHYSICAL", "VIRTUAL"}

	for _, registryAssetType := range registryAssetTypes {
		t.Run(registryAssetType, func(t *testing.T) {
			const query = `
				mutation($input: CreateRegistryAssetInput!) {
					createRegistryAsset(input: $input) {
						registryAssetEdge {
							node {
								id
								registryAssetType
							}
						}
					}
				}
			`

			var result struct {
				CreateRegistryAsset struct {
					RegistryAssetEdge struct {
						Node struct {
							ID                string `json:"id"`
							RegistryAssetType string `json:"registryAssetType"`
						} `json:"node"`
					} `json:"registryAssetEdge"`
				} `json:"createRegistryAsset"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId":    owner.GetOrganizationID().String(),
					"name":              "RegistryAsset " + registryAssetType,
					"amount":            1,
					"ownerId":           profileID,
					"registryAssetType": registryAssetType,
					"dataTypesStored":   "Test data",
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, registryAssetType, result.CreateRegistryAsset.RegistryAssetEdge.Node.RegistryAssetType)
		})
	}
}
