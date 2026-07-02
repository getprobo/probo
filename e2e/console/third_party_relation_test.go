// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func TestThirdPartyRelation_CreateChildAndList(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Parent Corp").Create()
	childID := createChildThirdParty(t, owner, parentID, "Child Corp")

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

	t.Run("child has parent reference", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						parentThirdParty {
							id
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ParentThirdParty *struct {
					ID string `json:"id"`
				} `json:"parentThirdParty"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": childID}, &result)

		require.NoError(t, err)
		require.NotNil(t, result.Node.ParentThirdParty)
		assert.Equal(t, parentID, result.Node.ParentThirdParty.ID)
	})
}

func TestThirdPartyRelation_Ancestors(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	rootID := factory.NewThirdParty(owner).WithName("Ancestor Root").Create()
	midID := createChildThirdParty(t, owner, rootID, "Ancestor Mid")
	leafID := createChildThirdParty(t, owner, midID, "Ancestor Leaf")

	t.Run("root has no ancestors", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						ancestors {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Ancestors []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"ancestors"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": rootID}, &result)
		require.NoError(t, err)
		assert.Empty(t, result.Node.Ancestors)
	})

	t.Run("mid has one ancestor", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						ancestors {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Ancestors []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"ancestors"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": midID}, &result)
		require.NoError(t, err)
		require.Len(t, result.Node.Ancestors, 1)
		assert.Equal(t, rootID, result.Node.Ancestors[0].ID)
	})

	t.Run("leaf returns ancestors root-first", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						ancestors {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Ancestors []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"ancestors"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": leafID}, &result)
		require.NoError(t, err)
		require.Len(t, result.Node.Ancestors, 2)
		assert.Equal(t, rootID, result.Node.Ancestors[0].ID, "root should be first")
		assert.Equal(t, midID, result.Node.Ancestors[1].ID, "mid should be second")
	})
}

func TestThirdPartyRelation_DeleteChild(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(owner).WithName("Parent Remove").Create()
	childID := createChildThirdParty(t, owner, parentID, "Child Remove")

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
	assert.Equal(t, childID, result.DeleteThirdParty.DeletedThirdPartyID)

	count := countChildThirdParties(t, owner, parentID)
	assert.Equal(t, 0, count)
}

func TestThirdPartyRelation_Authorization(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	parentID := factory.NewThirdParty(owner).WithName("Auth Parent").Create()

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
	createChildThirdParty(t, org1Owner, parentID, "Org1 Child")

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

func TestThirdParty_LevelFilter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	factory.NewThirdParty(owner).WithName("Direct TP").Create()

	const createThirdParty = `
		mutation($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
						level
					}
				}
			}
		}
	`

	type createThirdPartyResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID    string `json:"id"`
					Level int    `json:"level"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	// A level-2 third party can only exist as the child of a level-1 parent;
	// the level is derived from the parent rather than supplied by the client.
	var parentResult createThirdPartyResult

	err := owner.Execute(createThirdParty, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           factory.SafeName("Parent TP"),
		},
	}, &parentResult)
	require.NoError(t, err)
	require.Equal(t, 1, parentResult.CreateThirdParty.ThirdPartyEdge.Node.Level)

	var createResult createThirdPartyResult

	err = owner.Execute(createThirdParty, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               factory.SafeName("NonDirect TP"),
			"parentThirdPartyId": parentResult.CreateThirdParty.ThirdPartyEdge.Node.ID,
		},
	}, &createResult)
	require.NoError(t, err)
	assert.Equal(t, 2, createResult.CreateThirdParty.ThirdPartyEdge.Node.Level)

	t.Run("filter level 1 only", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($orgId: ID!) {
				node(id: $orgId) {
					... on Organization {
						thirdParties(first: 100, filter: { level: 1 }) {
							edges {
								node {
									id
									level
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
							ID    string `json:"id"`
							Level int    `json:"level"`
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
			assert.Equal(t, 1, edge.Node.Level, "expected all third parties to be level 1 when filtering level=1")
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
									level
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
							ID    string `json:"id"`
							Level int    `json:"level"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"orgId": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		hasLevel1 := false
		hasLevel2 := false

		for _, edge := range result.Node.ThirdParties.Edges {
			if edge.Node.Level == 1 {
				hasLevel1 = true
			} else if edge.Node.Level >= 2 {
				hasLevel2 = true
			}
		}

		assert.True(t, hasLevel1, "expected at least one level-1 third party")
		assert.True(t, hasLevel2, "expected at least one level-2+ third party")
	})
}

func createChildThirdParty(t *testing.T, c *testutil.Client, parentID, name string) string {
	t.Helper()

	const query = `
		mutation($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":     c.GetOrganizationID().String(),
			"name":               factory.SafeName(name),
			"parentThirdPartyId": parentID,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateThirdParty.ThirdPartyEdge.Node.ID
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
