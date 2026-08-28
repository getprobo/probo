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
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestRisk_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with full details", func(t *testing.T) {
		query := `
			mutation CreateRisk($input: CreateRiskInput!) {
				createRisk(input: $input) {
					riskEdge {
						node {
							id
							name
							category
							description
						}
					}
				}
			}
		`

		var result struct {
			CreateRisk struct {
				RiskEdge struct {
					Node struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Category    string  `json:"category"`
						Description *string `json:"description"`
					} `json:"node"`
				} `json:"riskEdge"`
			} `json:"createRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Data Breach Risk",
				"description":    "Risk of unauthorized data access",
				"category":       "SECURITY",
			},
		}, &result)
		require.NoError(t, err)

		risk := result.CreateRisk.RiskEdge.Node
		assert.NotEmpty(t, risk.ID)
		assert.Equal(t, "Data Breach Risk", risk.Name)
		assert.Equal(t, "SECURITY", risk.Category)
		require.NotNil(t, risk.Description)
		assert.Equal(t, "Risk of unauthorized data access", *risk.Description)
	})

	t.Run("with required fields only", func(t *testing.T) {
		query := `
			mutation CreateRisk($input: CreateRiskInput!) {
				createRisk(input: $input) {
					riskEdge {
						node {
							id
							name
							category
						}
					}
				}
			}
		`

		var result struct {
			CreateRisk struct {
				RiskEdge struct {
					Node struct {
						ID       string `json:"id"`
						Name     string `json:"name"`
						Category string `json:"category"`
					} `json:"node"`
				} `json:"riskEdge"`
			} `json:"createRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Catalog Risk",
				"category":       "SECURITY",
			},
		}, &result)
		require.NoError(t, err)

		risk := result.CreateRisk.RiskEdge.Node
		assert.NotEmpty(t, risk.ID)
		assert.Equal(t, "Catalog Risk", risk.Name)
		assert.Equal(t, "SECURITY", risk.Category)
	})
}

func TestRisk_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.NewRisk(owner).
		WithName("Risk to Update").
		WithDescription("Original description").
		Create()

	query := `
		mutation UpdateRisk($input: UpdateRiskInput!) {
			updateRisk(input: $input) {
				risk {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateRisk struct {
			Risk struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"risk"`
		} `json:"updateRisk"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":          riskID,
			"name":        "Updated Risk Name",
			"description": "Updated by owner",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, riskID, result.UpdateRisk.Risk.ID)
	assert.Equal(t, "Updated Risk Name", result.UpdateRisk.Risk.Name)
}

func TestRisk_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.NewRisk(owner).
		WithName("Risk to Delete").
		Create()

	query := `
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}
	`

	var result struct {
		DeleteRisk struct {
			DeletedRiskID string `json:"deletedRiskId"`
		} `json:"deleteRisk"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"riskId": riskID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, riskID, result.DeleteRisk.DeletedRiskID)
}

func TestRisk_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple risks
	categories := []string{"SECURITY", "OPERATIONAL", "COMPLIANCE"}
	for _, category := range categories {
		factory.NewRisk(owner).
			WithName(category + " Risk").
			WithCategory(category).
			Create()
	}

	query := `
		query ListRisks($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					risks(first: 10) {
						edges {
							node {
								id
								name
								category
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
			Risks struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Name     string `json:"name"`
						Category string `json:"category"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"risks"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Risks.TotalCount, 3)
}

func TestRisk_RequiredFields(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":     "Test Risk",
				"category": "SECURITY",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing name",
			input: map[string]any{
				"category": "SECURITY",
			},
			wantErrorContains: "name",
		},
		{
			name: "missing category",
			input: map[string]any{
				"name": "Test Risk",
			},
			wantErrorContains: "category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateRisk($input: CreateRiskInput!) {
					createRisk(input: $input) {
						riskEdge {
							node {
								id
							}
						}
					}
				}
			`

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}

			maps.Copy(input, tt.input)

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestRisk_CategoryEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	categories := []string{
		"SECURITY",
		"OPERATIONAL",
		"COMPLIANCE",
		"FINANCIAL",
		"REPUTATIONAL",
		"STRATEGIC",
	}

	for _, category := range categories {
		t.Run("create with category "+category, func(t *testing.T) {
			riskID := factory.NewRisk(owner).
				WithName("Category Test " + category).
				WithCategory(category).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on Risk {
							id
							category
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID       string `json:"id"`
					Category string `json:"category"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": riskID}, &result)
			require.NoError(t, err)
			assert.Equal(t, category, result.Node.Category)
		})
	}
}

func TestRisk_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.NewRisk(owner).
		WithName("SubResolver Test Risk").
		Create()

	t.Run("risk node query", func(t *testing.T) {
		query := `
			query GetRisk($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						name
						description
						category
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Category    string  `json:"category"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.Equal(t, riskID, result.Node.ID)
		assert.Equal(t, "SubResolver Test Risk", result.Node.Name)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						organization {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})

	t.Run("measures sub-resolver (empty)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						measures(first: 10) {
							edges {
								node {
									id
									name
								}
							}
							pageInfo {
								hasNextPage
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID       string `json:"id"`
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"measures"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Node.Measures.Edges)
	})

	t.Run("documents sub-resolver (empty)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						documents(first: 10) {
							edges {
								node {
									id
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID        string `json:"id"`
				Documents struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Node.Documents.Edges)
	})
}

func TestRisk_InvalidID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update with invalid ID", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   "invalid-id-format",
				"name": "Test",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		query := `
			mutation DeleteRisk($input: DeleteRiskInput!) {
				deleteRisk(input: $input) {
					deletedRiskId
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"riskId": "invalid-id-format",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("query with non-existent ID", func(t *testing.T) {
		query := `
			query GetRisk($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo",
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestRisk_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.NewRisk(owner).
		WithName("Description Test Risk").
		WithDescription("Initial description").
		Create()

	t.Run("set description", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          riskID,
				"description": "Updated description",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateRisk.Risk.Description)
		assert.Equal(t, "Updated description", *result.UpdateRisk.Risk.Description)
	})

	t.Run("clear description with null", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          riskID,
				"description": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateRisk.Risk.Description)
	})

	t.Run("update without description preserves value", func(t *testing.T) {
		// First set a description
		setQuery := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
					}
				}
			}
		`

		err := owner.Execute(setQuery, map[string]any{
			"input": map[string]any{
				"id":          riskID,
				"description": "Should persist",
			},
		}, nil)
		require.NoError(t, err)

		// Update only name
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						name
						description
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":   riskID,
				"name": "Updated Name",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateRisk.Risk.Description)
		assert.Equal(t, "Should persist", *result.UpdateRisk.Risk.Description)
	})
}
