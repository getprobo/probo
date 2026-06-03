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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyRelation_AddAndList(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Parent Corp").Create()
	childID := factory.NewThirdParty(owner).WithName("Child Corp").Create()

	addRelation(t, owner, parentID, childID)

	t.Run("list child third parties", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						childThirdParties(first: 10) {
							totalCount
							edges {
								node {
									id
									name
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ChildThirdParties struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"childThirdParties"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": parentID}, &result)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.ChildThirdParties.TotalCount)
		require.Len(t, result.Node.ChildThirdParties.Edges, 1)
		assert.Equal(t, childID, result.Node.ChildThirdParties.Edges[0].Node.ID)
	})
}

func TestThirdPartyRelation_Remove(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Parent Remove").Create()
	childID := factory.NewThirdParty(owner).WithName("Child Remove").Create()

	addRelation(t, owner, parentID, childID)

	const removeQuery = `
		mutation($input: DeleteThirdPartyThirdPartyMappingInput!) {
			deleteThirdPartyThirdPartyMapping(input: $input) {
				removedThirdPartyId
			}
		}
	`

	var result struct {
		DeleteThirdPartyThirdPartyMapping struct {
			RemovedThirdPartyID string `json:"removedThirdPartyId"`
		} `json:"deleteThirdPartyThirdPartyMapping"`
	}

	err := owner.Execute(removeQuery, map[string]any{
		"input": map[string]any{
			"parentThirdPartyId": parentID,
			"childThirdPartyId":  childID,
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, childID, result.DeleteThirdPartyThirdPartyMapping.RemovedThirdPartyID)

	count := countChildThirdParties(t, owner, parentID)
	assert.Equal(t, 0, count)
}

func TestThirdPartyRelation_Bidirectional(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	aID := factory.NewThirdParty(owner).WithName("Company A").Create()
	bID := factory.NewThirdParty(owner).WithName("Company B").Create()

	addRelation(t, owner, aID, bID)
	addRelation(t, owner, bID, aID)

	t.Run("A has B as child", func(t *testing.T) {
		t.Parallel()
		count := countChildThirdParties(t, owner, aID)
		assert.Equal(t, 1, count)
	})

	t.Run("B has A as child", func(t *testing.T) {
		t.Parallel()
		count := countChildThirdParties(t, owner, bID)
		assert.Equal(t, 1, count)
	})
}

func TestThirdPartyRelation_Idempotent(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Idempotent Parent").Create()
	childID := factory.NewThirdParty(owner).WithName("Idempotent Child").Create()

	addRelation(t, owner, parentID, childID)
	addRelation(t, owner, parentID, childID)

	count := countChildThirdParties(t, owner, parentID)
	assert.Equal(t, 1, count)
}

func TestThirdPartyRelation_CascadeOnDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Cascade Parent").Create()
	childID := factory.NewThirdParty(owner).WithName("Cascade Child").Create()

	addRelation(t, owner, parentID, childID)

	const deleteQuery = `
		mutation($input: DeleteThirdPartyInput!) {
			deleteThirdParty(input: $input) {
				deletedThirdPartyId
			}
		}
	`

	var result struct {
		DeleteThirdParty struct {
			DeletedThirdPartyID string `json:"deletedThirdPartyId"`
		} `json:"deleteThirdParty"`
	}

	err := owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"thirdPartyId": childID,
		},
	}, &result)
	require.NoError(t, err)

	count := countChildThirdParties(t, owner, parentID)
	assert.Equal(t, 0, count)
}

func TestThirdPartyRelation_Authorization(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	parentID := factory.NewThirdParty(owner).WithName("Auth Parent").Create()
	childID := factory.NewThirdParty(owner).WithName("Auth Child").Create()

	t.Run("viewer cannot add relation", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateThirdPartyThirdPartyMappingInput!) {
				createThirdPartyThirdPartyMapping(input: $input) {
					thirdPartyEdge {
						node { id }
					}
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"parentThirdPartyId": parentID,
				"childThirdPartyId":  childID,
			},
		})
		testutil.RequireForbiddenError(t, err)
	})

	t.Run("viewer cannot remove relation", func(t *testing.T) {
		t.Parallel()

		addRelation(t, owner, parentID, childID)

		const query = `
			mutation($input: DeleteThirdPartyThirdPartyMappingInput!) {
				deleteThirdPartyThirdPartyMapping(input: $input) {
					removedThirdPartyId
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"parentThirdPartyId": parentID,
				"childThirdPartyId":  childID,
			},
		})
		testutil.RequireForbiddenError(t, err)
	})

	t.Run("viewer can list child third parties", func(t *testing.T) {
		t.Parallel()

		count := countChildThirdParties(t, viewer, parentID)
		assert.GreaterOrEqual(t, count, 0)
	})
}

func TestThirdPartyRelation_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(org1Owner).WithName("Org1 Parent").Create()
	childID := factory.NewThirdParty(org1Owner).WithName("Org1 Child").Create()
	org2ChildID := factory.NewThirdParty(org2Owner).WithName("Org2 Child").Create()

	addRelation(t, org1Owner, parentID, childID)

	t.Run("cannot add cross-org relation", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateThirdPartyThirdPartyMappingInput!) {
				createThirdPartyThirdPartyMapping(input: $input) {
					thirdPartyEdge {
						node { id }
					}
				}
			}
		`

		_, err := org1Owner.Do(query, map[string]any{
			"input": map[string]any{
				"parentThirdPartyId": parentID,
				"childThirdPartyId":  org2ChildID,
			},
		})
		require.Error(t, err)
	})

	t.Run("cannot list children of other org third party", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						childThirdParties(first: 10) {
							totalCount
						}
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ChildThirdParties *struct {
					TotalCount int `json:"totalCount"`
				} `json:"childThirdParties"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": parentID}, &result)

		nodeInaccessible := err != nil || result.Node == nil || result.Node.ChildThirdParties == nil
		emptyResult := result.Node != nil && result.Node.ChildThirdParties != nil && result.Node.ChildThirdParties.TotalCount == 0
		assert.True(t, nodeInaccessible || emptyResult, "expected either inaccessible node or zero children")
	})
}

func TestThirdParty_DirectFilter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	factory.NewThirdParty(owner).WithName("Direct TP").Create()

	const createNonDirect = `
		mutation($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
						firstLevel
					}
				}
			}
		}
	`

	var createResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID         string `json:"id"`
					FirstLevel bool   `json:"firstLevel"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := owner.Execute(createNonDirect, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           factory.SafeName("NonDirect TP"),
			"firstLevel":     false,
		},
	}, &createResult)
	require.NoError(t, err)
	assert.False(t, createResult.CreateThirdParty.ThirdPartyEdge.Node.FirstLevel)

	t.Run("filter firstLevel only", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($orgId: ID!) {
				node(id: $orgId) {
					... on Organization {
						thirdParties(first: 100, filter: { firstLevel: true }) {
							edges {
								node {
									id
									firstLevel
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ThirdParties struct {
					Edges []struct {
						Node struct {
							ID         string `json:"id"`
							FirstLevel bool   `json:"firstLevel"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"orgId": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		for _, edge := range result.Node.ThirdParties.Edges {
			assert.True(t, edge.Node.FirstLevel, "expected all third parties to be firstLevel when filtering direct=true")
		}
	})

	t.Run("filter all", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($orgId: ID!) {
				node(id: $orgId) {
					... on Organization {
						thirdParties(first: 100) {
							edges {
								node {
									id
									firstLevel
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ThirdParties struct {
					Edges []struct {
						Node struct {
							ID         string `json:"id"`
							FirstLevel bool   `json:"firstLevel"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"orgId": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		hasFirstLevel := false
		hasNonFirstLevel := false

		for _, edge := range result.Node.ThirdParties.Edges {
			if edge.Node.FirstLevel {
				hasFirstLevel = true
			} else {
				hasNonFirstLevel = true
			}
		}

		assert.True(t, hasFirstLevel, "expected at least one first-level third party")
		assert.True(t, hasNonFirstLevel, "expected at least one non-first-level third party")
	})
}

func addRelation(t *testing.T, c *testutil.Client, parentID, childID string) {
	t.Helper()

	const query = `
		mutation($input: CreateThirdPartyThirdPartyMappingInput!) {
			createThirdPartyThirdPartyMapping(input: $input) {
				thirdPartyEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		CreateThirdPartyThirdPartyMapping struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdPartyThirdPartyMapping"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"parentThirdPartyId": parentID,
			"childThirdPartyId":  childID,
		},
	}, &result)
	require.NoError(t, err)
}

func countChildThirdParties(t *testing.T, c *testutil.Client, parentID string) int {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					childThirdParties(first: 1) {
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ChildThirdParties struct {
				TotalCount int `json:"totalCount"`
			} `json:"childThirdParties"`
		} `json:"node"`
	}

	err := c.Execute(query, map[string]any{"id": parentID}, &result)
	require.NoError(t, err)

	return result.Node.ChildThirdParties.TotalCount
}
