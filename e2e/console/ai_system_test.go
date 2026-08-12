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
	"maps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAiSystem_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	const query = `
		mutation CreateAiSystem($input: CreateAiSystemInput!) {
			createAiSystem(input: $input) {
				aiSystemEdge {
					node {
						id
						name
						version
						companyRoles
						status
						riskClassification
						source
						purpose
						owner { id }
					}
				}
			}
		}
	`

	var result struct {
		CreateAiSystem struct {
			AiSystemEdge struct {
				Node struct {
					ID                 string   `json:"id"`
					Name               string   `json:"name"`
					Version            *string  `json:"version"`
					CompanyRoles       []string `json:"companyRoles"`
					Status             string   `json:"status"`
					RiskClassification string   `json:"riskClassification"`
					Source             *string  `json:"source"`
					Purpose            *string  `json:"purpose"`
					Owner              *struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"aiSystemEdge"`
		} `json:"createAiSystem"`
	}

	name := factory.SafeName("Customer Support Bot")
	version := "2.1"
	source := "Internal ML team"
	purpose := "Automate tier-1 support responses"

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               name,
			"version":            version,
			"companyRoles":       []string{"PROVIDER", "DEPLOYER"},
			"status":             "ACTIVE",
			"ownerId":            profileID,
			"source":             source,
			"purpose":            purpose,
			"riskClassification": "LIMITED",
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateAiSystem.AiSystemEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, name, node.Name)
	require.NotNil(t, node.Version)
	assert.Equal(t, version, *node.Version)
	assert.Equal(t, []string{"PROVIDER", "DEPLOYER"}, node.CompanyRoles)
	assert.Equal(t, "ACTIVE", node.Status)
	assert.Equal(t, "LIMITED", node.RiskClassification)
	require.NotNil(t, node.Source)
	assert.Equal(t, source, *node.Source)
	require.NotNil(t, node.Purpose)
	assert.Equal(t, purpose, *node.Purpose)
	require.NotNil(t, node.Owner)
	assert.Equal(t, profileID, node.Owner.ID)
}

func TestAiSystem_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	aiSystemID := createAiSystem(t, owner, map[string]any{
		"name":               "Original AI System",
		"status":             "IN_DEVELOPMENT",
		"riskClassification": "MINIMAL",
	})

	const query = `
		mutation UpdateAiSystem($input: UpdateAiSystemInput!) {
			updateAiSystem(input: $input) {
				aiSystem {
					id
					name
					status
					riskClassification
					notes
				}
			}
		}
	`

	var result struct {
		UpdateAiSystem struct {
			AiSystem struct {
				ID                 string  `json:"id"`
				Name               string  `json:"name"`
				Status             string  `json:"status"`
				RiskClassification string  `json:"riskClassification"`
				Notes              *string `json:"notes"`
			} `json:"aiSystem"`
		} `json:"updateAiSystem"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":                 aiSystemID,
			"name":               "Updated AI System",
			"status":             "ACTIVE",
			"riskClassification": "HIGH_RISK",
			"notes":              "Requires quarterly review",
		},
	}, &result)
	require.NoError(t, err)

	aiSystem := result.UpdateAiSystem.AiSystem
	assert.Equal(t, aiSystemID, aiSystem.ID)
	assert.Equal(t, "Updated AI System", aiSystem.Name)
	assert.Equal(t, "ACTIVE", aiSystem.Status)
	assert.Equal(t, "HIGH_RISK", aiSystem.RiskClassification)
	require.NotNil(t, aiSystem.Notes)
	assert.Equal(t, "Requires quarterly review", *aiSystem.Notes)
}

func TestAiSystem_Delete(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	aiSystemID := createAiSystem(t, owner, map[string]any{
		"name":               "Delete Me",
		"status":             "DECOMMISSIONED",
		"riskClassification": "MINIMAL",
	})

	const query = `
		mutation DeleteAiSystem($input: DeleteAiSystemInput!) {
			deleteAiSystem(input: $input) {
				deletedAiSystemId
			}
		}
	`

	var result struct {
		DeleteAiSystem struct {
			DeletedAiSystemID string `json:"deletedAiSystemId"`
		} `json:"deleteAiSystem"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"aiSystemId": aiSystemID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, aiSystemID, result.DeleteAiSystem.DeletedAiSystemID)
}

func TestAiSystem_List(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	for i := range 3 {
		createAiSystem(t, owner, map[string]any{
			"name":               fmt.Sprintf("List AI System %d", i),
			"status":             "ACTIVE",
			"riskClassification": "MINIMAL",
		})
	}

	const query = `
		query GetAiSystems($id: ID!) {
			node(id: $id) {
				... on Organization {
					aiSystems(first: 10) {
						edges {
							node {
								id
								name
								status
								riskClassification
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
			AiSystems struct {
				Edges []struct {
					Node struct {
						ID                 string `json:"id"`
						Name               string `json:"name"`
						Status             string `json:"status"`
						RiskClassification string `json:"riskClassification"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"aiSystems"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AiSystems.TotalCount, 3)
}

func TestAiSystem_ListWithFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	createAiSystem(t, owner, map[string]any{
		"name":               "High Risk System",
		"status":             "ACTIVE",
		"riskClassification": "HIGH_RISK",
	})
	createAiSystem(t, owner, map[string]any{
		"name":               "Minimal Risk System",
		"status":             "ACTIVE",
		"riskClassification": "MINIMAL",
	})

	const query = `
		query GetAiSystems($id: ID!, $filter: AiSystemFilter) {
			node(id: $id) {
				... on Organization {
					aiSystems(first: 10, filter: $filter) {
						edges {
							node {
								id
								riskClassification
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
			AiSystems struct {
				Edges []struct {
					Node struct {
						ID                 string `json:"id"`
						RiskClassification string `json:"riskClassification"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"aiSystems"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id":     owner.GetOrganizationID().String(),
		"filter": map[string]any{"riskClassification": "HIGH_RISK"},
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AiSystems.TotalCount, 1)

	for _, edge := range result.Node.AiSystems.Edges {
		assert.Equal(t, "HIGH_RISK", edge.Node.RiskClassification)
	}
}

func TestAiSystem_GetByID(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const (
		name               = "Get By ID AI System"
		status             = "ACTIVE"
		riskClassification = "GPAI"
	)

	aiSystemID := createAiSystem(t, owner, map[string]any{
		"name":               name,
		"status":             status,
		"riskClassification": riskClassification,
	})

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AiSystem {
					id
					name
					status
					riskClassification
					createdAt
					updatedAt
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID                 string    `json:"id"`
			Name               string    `json:"name"`
			Status             string    `json:"status"`
			RiskClassification string    `json:"riskClassification"`
			CreatedAt          time.Time `json:"createdAt"`
			UpdatedAt          time.Time `json:"updatedAt"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": aiSystemID}, &result)
	require.NoError(t, err)

	node := result.Node
	assert.Equal(t, aiSystemID, node.ID)
	assert.Equal(t, name, node.Name)
	assert.Equal(t, status, node.Status)
	assert.Equal(t, riskClassification, node.RiskClassification)
	assert.False(t, node.CreatedAt.IsZero())
	assert.False(t, node.UpdatedAt.IsZero())
}

func TestAiSystem_Create_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             map[string]any
		wantErrorContains string
	}{
		{
			name: "empty name",
			input: map[string]any{
				"name":               "",
				"status":             "ACTIVE",
				"riskClassification": "MINIMAL",
			},
			wantErrorContains: "name",
		},
		{
			name: "missing risk classification",
			input: map[string]any{
				"name":   factory.SafeName("Validation Test"),
				"status": "ACTIVE",
			},
			wantErrorContains: "riskClassification",
		},
		{
			name: "invalid status enum",
			input: map[string]any{
				"name":               factory.SafeName("Validation Test"),
				"status":             "INVALID",
				"riskClassification": "MINIMAL",
			},
			wantErrorContains: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			const query = `
				mutation CreateAiSystem($input: CreateAiSystemInput!) {
					createAiSystem(input: $input) {
						aiSystemEdge {
							node { id }
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
			}
			maps.Copy(input, tt.input)

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestAiSystem_RBAC(t *testing.T) {
	t.Parallel()

	baseCreateInput := func(client *testutil.Client) map[string]any {
		return map[string]any{
			"organizationId":     client.GetOrganizationID().String(),
			"name":               factory.SafeName("RBAC Test AI System"),
			"status":             "ACTIVE",
			"riskClassification": "MINIMAL",
		}
	}

	t.Run("create", func(t *testing.T) {
		t.Parallel()

		t.Run("owner can create", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			_, err := owner.Do(`
				mutation CreateAiSystem($input: CreateAiSystemInput!) {
					createAiSystem(input: $input) {
						aiSystemEdge { node { id } }
					}
				}
			`, map[string]any{"input": baseCreateInput(owner)})
			require.NoError(t, err, "owner should be able to create ai system")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			_, err := viewer.Do(`
				mutation CreateAiSystem($input: CreateAiSystemInput!) {
					createAiSystem(input: $input) {
						aiSystemEdge { node { id } }
					}
				}
			`, map[string]any{"input": baseCreateInput(viewer)})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create ai system")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		aiSystemID := createAiSystem(t, owner, map[string]any{
			"name":               "RBAC Read Test",
			"status":             "ACTIVE",
			"riskClassification": "MINIMAL",
		})

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on AiSystem { id name }
				}
			}
		`, map[string]any{"id": aiSystemID}, &result)
		require.NoError(t, err, "viewer should be able to read ai system")
		require.NotNil(t, result.Node, "viewer should receive ai system data")
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		t.Run("owner can update", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			aiSystemID := createAiSystem(t, owner, map[string]any{
				"name":               "RBAC Update Test",
				"status":             "ACTIVE",
				"riskClassification": "MINIMAL",
			})

			_, err := owner.Do(`
				mutation UpdateAiSystem($input: UpdateAiSystemInput!) {
					updateAiSystem(input: $input) {
						aiSystem { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   aiSystemID,
					"name": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update ai system")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			aiSystemID := createAiSystem(t, owner, map[string]any{
				"name":               "RBAC Update Test",
				"status":             "ACTIVE",
				"riskClassification": "MINIMAL",
			})

			_, err := viewer.Do(`
				mutation UpdateAiSystem($input: UpdateAiSystemInput!) {
					updateAiSystem(input: $input) {
						aiSystem { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   aiSystemID,
					"name": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update ai system")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		t.Run("owner can delete", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			aiSystemID := createAiSystem(t, owner, map[string]any{
				"name":               "RBAC Delete Test",
				"status":             "ACTIVE",
				"riskClassification": "MINIMAL",
			})

			_, err := owner.Do(`
				mutation DeleteAiSystem($input: DeleteAiSystemInput!) {
					deleteAiSystem(input: $input) {
						deletedAiSystemId
					}
				}
			`, map[string]any{
				"input": map[string]any{"aiSystemId": aiSystemID},
			})
			require.NoError(t, err, "owner should be able to delete ai system")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			aiSystemID := createAiSystem(t, owner, map[string]any{
				"name":               "RBAC Delete Test",
				"status":             "ACTIVE",
				"riskClassification": "MINIMAL",
			})

			_, err := viewer.Do(`
				mutation DeleteAiSystem($input: DeleteAiSystemInput!) {
					deleteAiSystem(input: $input) {
						deletedAiSystemId
					}
				}
			`, map[string]any{
				"input": map[string]any{"aiSystemId": aiSystemID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete ai system")
		})
	})
}

func TestAiSystem_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ProfileID := factory.CreateUser(org1Owner)
	org2ProfileID := factory.CreateUser(org2Owner)

	aiSystemID := createAiSystem(t, org1Owner, map[string]any{
		"name":               "Org1 AI System",
		"status":             "ACTIVE",
		"riskClassification": "MINIMAL",
		"ownerId":            org1ProfileID,
	})

	const createQuery = `
		mutation CreateAiSystem($input: CreateAiSystemInput!) {
			createAiSystem(input: $input) {
				aiSystemEdge { node { id } }
			}
		}
	`

	t.Run(
		"cannot read ai system from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				query($id: ID!) {
					node(id: $id) {
						... on AiSystem {
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

			err := org2Owner.Execute(query, map[string]any{"id": aiSystemID}, &result)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "ai system")
		},
	)

	t.Run(
		"cannot update ai system from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation UpdateAiSystem($input: UpdateAiSystemInput!) {
					updateAiSystem(input: $input) {
						aiSystem { id }
					}
				}
			`

			_, err := org2Owner.Do(
				query,
				map[string]any{
					"input": map[string]any{
						"id":   aiSystemID,
						"name": "Updated by Org2",
					},
				},
			)
			require.Error(t, err, "must not update ai system from another organization")
		},
	)

	t.Run(
		"cannot delete ai system from another organization",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation DeleteAiSystem($input: DeleteAiSystemInput!) {
					deleteAiSystem(input: $input) {
						deletedAiSystemId
					}
				}
			`

			_, err := org2Owner.Do(
				query,
				map[string]any{
					"input": map[string]any{"aiSystemId": aiSystemID},
				},
			)
			require.Error(t, err, "must not delete ai system from another organization")
		},
	)

	t.Run(
		"cannot create ai system referencing owner from another organization",
		func(t *testing.T) {
			t.Parallel()

			_, err := org1Owner.Do(
				createQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId":     org1Owner.GetOrganizationID().String(),
						"name":               factory.SafeName("Cross-tenant Owner Link"),
						"status":             "ACTIVE",
						"riskClassification": "MINIMAL",
						"ownerId":            org2ProfileID,
					},
				},
			)
			require.Error(t, err, "must not accept ownerId belonging to another organization")
		},
	)
}

func createAiSystem(t *testing.T, client *testutil.Client, attrs map[string]any) string {
	t.Helper()

	const query = `
		mutation CreateAiSystem($input: CreateAiSystemInput!) {
			createAiSystem(input: $input) {
				aiSystemEdge {
					node { id }
				}
			}
		}
	`

	input := map[string]any{
		"organizationId": client.GetOrganizationID().String(),
		"name":           attrs["name"],
		"status":         attrs["status"],
	}

	if riskClassification, ok := attrs["riskClassification"]; ok {
		input["riskClassification"] = riskClassification
	} else {
		input["riskClassification"] = "MINIMAL"
	}

	if ownerID, ok := attrs["ownerId"]; ok {
		input["ownerId"] = ownerID
	}

	var result struct {
		CreateAiSystem struct {
			AiSystemEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"aiSystemEdge"`
		} `json:"createAiSystem"`
	}

	err := client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.CreateAiSystem.AiSystemEdge.Node.ID
}
