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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestBusinessFunction_CreateCritical(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	const query = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge {
					node {
						id
						referenceId
						name
						classification
						mtdMinutes
						rtoMinutes
						rpoMinutes
						impactTolerance
						notes
						owner { id }
					}
				}
			}
		}
	`

	var result struct {
		CreateBusinessFunction struct {
			BusinessFunctionEdge struct {
				Node struct {
					ID              string  `json:"id"`
					ReferenceID     string  `json:"referenceId"`
					Name            string  `json:"name"`
					Classification  string  `json:"classification"`
					MTDMinutes      int     `json:"mtdMinutes"`
					RTOMinutes      int     `json:"rtoMinutes"`
					RPOMinutes      int     `json:"rpoMinutes"`
					ImpactTolerance *string `json:"impactTolerance"`
					Notes           *string `json:"notes"`
					Owner           *struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"businessFunctionEdge"`
		} `json:"createBusinessFunction"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":  owner.GetOrganizationID().String(),
			"referenceId":     "BF-CRIT-001",
			"name":            factory.SafeName("Critical Function"),
			"classification":  "CRITICAL",
			"mtdMinutes":      240,
			"rtoMinutes":      120,
			"rpoMinutes":      60,
			"impactTolerance": "Severe financial and reputational impact",
			"notes":           "Primary payment processing",
			"ownerId":         profileID,
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateBusinessFunction.BusinessFunctionEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "BF-CRIT-001", node.ReferenceID)
	assert.Equal(t, "CRITICAL", node.Classification)
	assert.Equal(t, 240, node.MTDMinutes)
	assert.Equal(t, 120, node.RTOMinutes)
	assert.Equal(t, 60, node.RPOMinutes)
	require.NotNil(t, node.ImpactTolerance)
	assert.Equal(t, "Severe financial and reputational impact", *node.ImpactTolerance)
	require.NotNil(t, node.Notes)
	assert.Equal(t, "Primary payment processing", *node.Notes)
	require.NotNil(t, node.Owner)
	assert.Equal(t, profileID, node.Owner.ID)
}

func TestBusinessFunction_CreateImportant(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	const query = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge {
					node {
						id
						referenceId
						name
						classification
						mtdMinutes
						rtoMinutes
						rpoMinutes
						owner { id }
					}
				}
			}
		}
	`

	var result struct {
		CreateBusinessFunction struct {
			BusinessFunctionEdge struct {
				Node struct {
					ID             string `json:"id"`
					ReferenceID    string `json:"referenceId"`
					Name           string `json:"name"`
					Classification string `json:"classification"`
					MTDMinutes     int    `json:"mtdMinutes"`
					RTOMinutes     int    `json:"rtoMinutes"`
					RPOMinutes     int    `json:"rpoMinutes"`
					Owner          *struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"businessFunctionEdge"`
		} `json:"createBusinessFunction"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    "BF-IMP-001",
			"name":           factory.SafeName("Important Function"),
			"classification": "IMPORTANT",
			"mtdMinutes":     480,
			"rtoMinutes":     240,
			"rpoMinutes":     120,
			"ownerId":        profileID,
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateBusinessFunction.BusinessFunctionEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "BF-IMP-001", node.ReferenceID)
	assert.Equal(t, "IMPORTANT", node.Classification)
	assert.Equal(t, 480, node.MTDMinutes)
	assert.Equal(t, 240, node.RTOMinutes)
	assert.Equal(t, 120, node.RPOMinutes)
	require.NotNil(t, node.Owner)
	assert.Equal(t, profileID, node.Owner.ID)
}

func TestBusinessFunction_CreateWithAssetAndThirdPartyLinks(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	assetID := createAssetForBusinessFunction(t, owner, profileID)
	thirdPartyID := factory.CreateThirdParty(owner, factory.Attrs{"name": factory.SafeName("Vendor")})

	const query = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge {
					node {
						id
						assets(first: 10) {
							edges { node { id } }
							totalCount
						}
						thirdParties(first: 10) {
							edges { node { id } }
							totalCount
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateBusinessFunction struct {
			BusinessFunctionEdge struct {
				Node struct {
					ID     string `json:"id"`
					Assets struct {
						Edges []struct {
							Node struct {
								ID string `json:"id"`
							} `json:"node"`
						} `json:"edges"`
						TotalCount int `json:"totalCount"`
					} `json:"assets"`
					ThirdParties struct {
						Edges []struct {
							Node struct {
								ID string `json:"id"`
							} `json:"node"`
						} `json:"edges"`
						TotalCount int `json:"totalCount"`
					} `json:"thirdParties"`
				} `json:"node"`
			} `json:"businessFunctionEdge"`
		} `json:"createBusinessFunction"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    "BF-LINK-001",
			"name":           factory.SafeName("Linked Function"),
			"classification": "CRITICAL",
			"mtdMinutes":     60,
			"rtoMinutes":     30,
			"rpoMinutes":     15,
			"ownerId":        profileID,
			"assetIds":       []string{assetID},
			"thirdPartyIds":  []string{thirdPartyID},
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateBusinessFunction.BusinessFunctionEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, 1, node.Assets.TotalCount)
	require.Len(t, node.Assets.Edges, 1)
	assert.Equal(t, assetID, node.Assets.Edges[0].Node.ID)
	assert.Equal(t, 1, node.ThirdParties.TotalCount)
	require.Len(t, node.ThirdParties.Edges, 1)
	assert.Equal(t, thirdPartyID, node.ThirdParties.Edges[0].Node.ID)
}

func TestBusinessFunction_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	businessFunctionID := createBusinessFunction(t, owner, map[string]any{
		"referenceId":    "BF-UPD-001",
		"name":           "Original Function",
		"classification": "STANDARD",
		"mtdMinutes":     60,
		"rtoMinutes":     30,
		"rpoMinutes":     15,
		"ownerId":        profileID,
	})

	const query = `
		mutation UpdateBusinessFunction($input: UpdateBusinessFunctionInput!) {
			updateBusinessFunction(input: $input) {
				businessFunction {
					id
					referenceId
					name
					classification
					mtdMinutes
					rtoMinutes
					rpoMinutes
					impactTolerance
					notes
				}
			}
		}
	`

	var result struct {
		UpdateBusinessFunction struct {
			BusinessFunction struct {
				ID              string  `json:"id"`
				ReferenceID     string  `json:"referenceId"`
				Name            string  `json:"name"`
				Classification  string  `json:"classification"`
				MTDMinutes      int     `json:"mtdMinutes"`
				RTOMinutes      int     `json:"rtoMinutes"`
				RPOMinutes      int     `json:"rpoMinutes"`
				ImpactTolerance *string `json:"impactTolerance"`
				Notes           *string `json:"notes"`
			} `json:"businessFunction"`
		} `json:"updateBusinessFunction"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":              businessFunctionID,
			"referenceId":     "BF-UPD-002",
			"name":            "Updated Function",
			"classification":  "IMPORTANT",
			"mtdMinutes":      120,
			"rtoMinutes":      60,
			"rpoMinutes":      30,
			"impactTolerance": "Moderate customer impact",
			"notes":           "Updated recovery objectives",
		},
	}, &result)
	require.NoError(t, err)

	bf := result.UpdateBusinessFunction.BusinessFunction
	assert.Equal(t, businessFunctionID, bf.ID)
	assert.Equal(t, "BF-UPD-002", bf.ReferenceID)
	assert.Equal(t, "Updated Function", bf.Name)
	assert.Equal(t, "IMPORTANT", bf.Classification)
	assert.Equal(t, 120, bf.MTDMinutes)
	assert.Equal(t, 60, bf.RTOMinutes)
	assert.Equal(t, 30, bf.RPOMinutes)
	require.NotNil(t, bf.ImpactTolerance)
	assert.Equal(t, "Moderate customer impact", *bf.ImpactTolerance)
	require.NotNil(t, bf.Notes)
	assert.Equal(t, "Updated recovery objectives", *bf.Notes)
}

func TestBusinessFunction_Delete(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	businessFunctionID := createBusinessFunction(t, owner, map[string]any{
		"referenceId":    "BF-DEL-001",
		"name":           "Delete Me",
		"classification": "SECONDARY",
		"mtdMinutes":     60,
		"rtoMinutes":     30,
		"rpoMinutes":     15,
	})

	const query = `
		mutation DeleteBusinessFunction($input: DeleteBusinessFunctionInput!) {
			deleteBusinessFunction(input: $input) {
				deletedBusinessFunctionId
			}
		}
	`

	var result struct {
		DeleteBusinessFunction struct {
			DeletedBusinessFunctionID string `json:"deletedBusinessFunctionId"`
		} `json:"deleteBusinessFunction"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"businessFunctionId": businessFunctionID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, businessFunctionID, result.DeleteBusinessFunction.DeletedBusinessFunctionID)
}

func TestBusinessFunction_List(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	for i := range 3 {
		createBusinessFunction(t, owner, map[string]any{
			"referenceId":    fmt.Sprintf("BF-LIST-%03d", i),
			"name":           fmt.Sprintf("List Function %d", i),
			"classification": "STANDARD",
			"mtdMinutes":     60,
			"rtoMinutes":     30,
			"rpoMinutes":     15,
		})
	}

	const query = `
		query GetBusinessFunctions($id: ID!) {
			node(id: $id) {
				... on Organization {
					businessFunctions(first: 10) {
						edges {
							node {
								id
								referenceId
								name
								classification
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
			BusinessFunctions struct {
				Edges []struct {
					Node struct {
						ID             string `json:"id"`
						ReferenceID    string `json:"referenceId"`
						Name           string `json:"name"`
						Classification string `json:"classification"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"businessFunctions"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.BusinessFunctions.TotalCount, 3)
}

func TestBusinessFunction_ListWithClassificationFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	classifications := []string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"}
	for i, classification := range classifications {
		createBusinessFunction(t, owner, map[string]any{
			"referenceId":    fmt.Sprintf("BF-CLS-%03d", i),
			"name":           fmt.Sprintf("Classification %s", classification),
			"classification": classification,
			"mtdMinutes":     60,
			"rtoMinutes":     30,
			"rpoMinutes":     15,
		})
	}

	const query = `
		query GetBusinessFunctions($id: ID!, $filter: BusinessFunctionFilter) {
			node(id: $id) {
				... on Organization {
					businessFunctions(first: 10, filter: $filter) {
						edges {
							node {
								id
								classification
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
			BusinessFunctions struct {
				Edges []struct {
					Node struct {
						ID             string `json:"id"`
						Classification string `json:"classification"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"businessFunctions"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id":     owner.GetOrganizationID().String(),
		"filter": map[string]any{"classification": "CRITICAL"},
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.BusinessFunctions.TotalCount, 1)

	for _, edge := range result.Node.BusinessFunctions.Edges {
		assert.Equal(t, "CRITICAL", edge.Node.Classification)
	}
}

func TestBusinessFunction_ListWithCifOnlyFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	classifications := []string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"}
	for i, classification := range classifications {
		createBusinessFunction(t, owner, map[string]any{
			"referenceId":    fmt.Sprintf("BF-CIF-%03d", i),
			"name":           fmt.Sprintf("CIF Filter %s", classification),
			"classification": classification,
			"mtdMinutes":     60,
			"rtoMinutes":     30,
			"rpoMinutes":     15,
		})
	}

	const query = `
		query GetBusinessFunctions($id: ID!, $filter: BusinessFunctionFilter) {
			node(id: $id) {
				... on Organization {
					businessFunctions(first: 10, filter: $filter) {
						edges {
							node {
								id
								classification
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
			BusinessFunctions struct {
				Edges []struct {
					Node struct {
						ID             string `json:"id"`
						Classification string `json:"classification"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"businessFunctions"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id":     owner.GetOrganizationID().String(),
		"filter": map[string]any{"cifOnly": true},
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.BusinessFunctions.TotalCount, 2)

	for _, edge := range result.Node.BusinessFunctions.Edges {
		assert.Contains(t, []string{"CRITICAL", "IMPORTANT"}, edge.Node.Classification)
	}
}

func TestBusinessFunction_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner)
	org2AssetID := createAssetForBusinessFunction(t, org2Owner, org2ProfileID)
	org2ThirdPartyID := factory.CreateThirdParty(org2Owner, factory.Attrs{"name": "Org2 Vendor"})

	businessFunctionID := createBusinessFunction(t, org1Owner, map[string]any{
		"referenceId":    "BF-ISO-001",
		"name":           "Org1 Function",
		"classification": "CRITICAL",
		"mtdMinutes":     60,
		"rtoMinutes":     30,
		"rpoMinutes":     15,
		"ownerId":        org1ProfileID,
	})

	const createQuery = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge { node { id } }
			}
		}
	`

	t.Run(
		"cannot read business function from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				query($id: ID!) {
					node(id: $id) {
						... on BusinessFunction {
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

			err := org2Owner.Execute(query, map[string]any{"id": businessFunctionID}, &result)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "business function")
		},
	)

	t.Run(
		"cannot update business function from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation UpdateBusinessFunction($input: UpdateBusinessFunctionInput!) {
					updateBusinessFunction(input: $input) {
						businessFunction { id }
					}
				}
			`

			_, err := org2Owner.Do(
				query,
				map[string]any{
					"input": map[string]any{
						"id":   businessFunctionID,
						"name": "Hijacked Function",
					},
				},
			)
			require.Error(t, err, "must not update business function from another organization")
		},
	)

	t.Run(
		"cannot delete business function from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation DeleteBusinessFunction($input: DeleteBusinessFunctionInput!) {
					deleteBusinessFunction(input: $input) {
						deletedBusinessFunctionId
					}
				}
			`

			_, err := org2Owner.Do(
				query,
				map[string]any{
					"input": map[string]any{
						"businessFunctionId": businessFunctionID,
					},
				},
			)
			require.Error(t, err, "must not delete business function from another organization")
		},
	)

	t.Run(
		"cannot create business function referencing asset from another organization",
		func(t *testing.T) {
			t.Parallel()

			_, err := org1Owner.Do(
				createQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId": org1Owner.GetOrganizationID().String(),
						"referenceId":    "BF-ISO-ASSET",
						"name":           "Cross-tenant Asset Link",
						"classification": "CRITICAL",
						"mtdMinutes":     60,
						"rtoMinutes":     30,
						"rpoMinutes":     15,
						"assetIds":       []string{org2AssetID},
					},
				},
			)
			require.Error(t, err, "must not accept assetIds belonging to another organization")
		},
	)

	t.Run(
		"cannot create business function referencing third party from another organization",
		func(t *testing.T) {
			t.Parallel()

			_, err := org1Owner.Do(
				createQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId": org1Owner.GetOrganizationID().String(),
						"referenceId":    "BF-ISO-TP",
						"name":           "Cross-tenant Third Party Link",
						"classification": "CRITICAL",
						"mtdMinutes":     60,
						"rtoMinutes":     30,
						"rpoMinutes":     15,
						"thirdPartyIds":  []string{org2ThirdPartyID},
					},
				},
			)
			require.Error(t, err, "must not accept thirdPartyIds belonging to another organization")
		},
	)

	t.Run(
		"cannot create business function referencing owner from another organization",
		func(t *testing.T) {
			t.Parallel()

			_, err := org1Owner.Do(
				createQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId": org1Owner.GetOrganizationID().String(),
						"referenceId":    "BF-ISO-OWNER",
						"name":           "Cross-tenant Owner Link",
						"classification": "CRITICAL",
						"mtdMinutes":     60,
						"rtoMinutes":     30,
						"rpoMinutes":     15,
						"ownerId":        org2ProfileID,
					},
				},
			)
			require.Error(t, err, "must not accept ownerId belonging to another organization")
		},
	)

	t.Run(
		"cannot update business function to reference asset from another organization",
		func(t *testing.T) {
			t.Parallel()

			bfID := createBusinessFunction(
				t,
				org1Owner,
				map[string]any{
					"referenceId":    "BF-ISO-UPD-ASSET",
					"name":           "Update Asset Isolation",
					"classification": "STANDARD",
					"mtdMinutes":     60,
					"rtoMinutes":     30,
					"rpoMinutes":     15,
				},
			)

			const query = `
				mutation UpdateBusinessFunction($input: UpdateBusinessFunctionInput!) {
					updateBusinessFunction(input: $input) {
						businessFunction { id }
					}
				}
			`

			_, err := org1Owner.Do(
				query,
				map[string]any{
					"input": map[string]any{
						"id":       bfID,
						"assetIds": []string{org2AssetID},
					},
				},
			)
			require.Error(t, err, "must not accept assetIds belonging to another organization")
		},
	)
}

func createBusinessFunction(t *testing.T, client *testutil.Client, attrs map[string]any) string {
	t.Helper()

	const query = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": client.GetOrganizationID().String(),
		"referenceId":    attrs["referenceId"],
		"name":           attrs["name"],
		"classification": attrs["classification"],
		"mtdMinutes":     attrs["mtdMinutes"],
		"rtoMinutes":     attrs["rtoMinutes"],
		"rpoMinutes":     attrs["rpoMinutes"],
	}

	if ownerID, ok := attrs["ownerId"]; ok {
		input["ownerId"] = ownerID
	}

	if assetIDs, ok := attrs["assetIds"]; ok {
		input["assetIds"] = assetIDs
	}

	if thirdPartyIDs, ok := attrs["thirdPartyIds"]; ok {
		input["thirdPartyIds"] = thirdPartyIDs
	}

	var result struct {
		CreateBusinessFunction struct {
			BusinessFunctionEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"businessFunctionEdge"`
		} `json:"createBusinessFunction"`
	}

	err := client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.CreateBusinessFunction.BusinessFunctionEdge.Node.ID
}

func createAssetForBusinessFunction(t *testing.T, client *testutil.Client, ownerID string) string {
	t.Helper()

	const query = `
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge { node { id } }
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

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":  client.GetOrganizationID().String(),
			"name":            factory.SafeName("Asset"),
			"amount":          1,
			"ownerId":         ownerID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Operational data",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateAsset.AssetEdge.Node.ID
}
