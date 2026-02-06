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
							treatment
							inherentLikelihood
							inherentImpact
						}
					}
				}
			}
		`

		var result struct {
			CreateRisk struct {
				RiskEdge struct {
					Node struct {
						ID                 string `json:"id"`
						Name               string `json:"name"`
						Category           string `json:"category"`
						Treatment          string `json:"treatment"`
						InherentLikelihood int    `json:"inherentLikelihood"`
						InherentImpact     int    `json:"inherentImpact"`
					} `json:"node"`
				} `json:"riskEdge"`
			} `json:"createRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"name":               "Data Breach Risk",
				"description":        "Risk of unauthorized data access",
				"category":           "SECURITY",
				"treatment":          "MITIGATED",
				"inherentLikelihood": 2,
				"inherentImpact":     3,
			},
		}, &result)
		require.NoError(t, err)

		risk := result.CreateRisk.RiskEdge.Node
		assert.NotEmpty(t, risk.ID)
		assert.Equal(t, "Data Breach Risk", risk.Name)
		assert.Equal(t, "SECURITY", risk.Category)
		assert.Equal(t, "MITIGATED", risk.Treatment)
		assert.Equal(t, 2, risk.InherentLikelihood)
		assert.Equal(t, 3, risk.InherentImpact)
	})

	t.Run("with different treatments", func(t *testing.T) {
		treatments := []struct {
			name      string
			treatment string
		}{
			{"Mitigated", "MITIGATED"},
			{"Transferred", "TRANSFERRED"},
			{"Accepted", "ACCEPTED"},
			{"Avoided", "AVOIDED"},
		}

		for _, tt := range treatments {
			t.Run(tt.name, func(t *testing.T) {
				riskID := factory.NewRisk(owner).
					WithName("Risk with " + tt.treatment).
					WithTreatment(tt.treatment).
					Create()

				query := `
					query GetRisk($id: ID!) {
						node(id: $id) {
							... on Risk {
								id
								treatment
							}
						}
					}
				`

				var result struct {
					Node struct {
						ID        string `json:"id"`
						Treatment string `json:"treatment"`
					} `json:"node"`
				}

				err := owner.Execute(query, map[string]any{"id": riskID}, &result)
				require.NoError(t, err)
				assert.Equal(t, tt.treatment, result.Node.Treatment)
			})
		}
	})

	t.Run("with different likelihood and impact", func(t *testing.T) {
		levelNames := map[int]string{1: "Low", 2: "Medium", 3: "High", 4: "Critical"}

		for likelihood := 1; likelihood <= 4; likelihood++ {
			for impact := 1; impact <= 4; impact++ {
				testName := levelNames[likelihood] + "_" + levelNames[impact]
				t.Run(testName, func(t *testing.T) {
					riskID := factory.NewRisk(owner).
						WithName("Risk " + testName).
						WithLikelihood(likelihood).
						WithImpact(impact).
						Create()

					query := `
						query GetRisk($id: ID!) {
							node(id: $id) {
								... on Risk {
									id
									inherentLikelihood
									inherentImpact
								}
							}
						}
					`

					var result struct {
						Node struct {
							ID                 string `json:"id"`
							InherentLikelihood int    `json:"inherentLikelihood"`
							InherentImpact     int    `json:"inherentImpact"`
						} `json:"node"`
					}

					err := owner.Execute(query, map[string]any{"id": riskID}, &result)
					require.NoError(t, err)
					assert.Equal(t, likelihood, result.Node.InherentLikelihood)
					assert.Equal(t, impact, result.Node.InherentImpact)
				})
			}
		}
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
					treatment
				}
			}
		}
	`

	var result struct {
		UpdateRisk struct {
			Risk struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Treatment string `json:"treatment"`
			} `json:"risk"`
		} `json:"updateRisk"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":                 riskID,
			"name":               "Updated Risk Name",
			"description":        "Updated by owner",
			"treatment":          "TRANSFERRED",
			"inherentLikelihood": 3,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, riskID, result.UpdateRisk.Risk.ID)
	assert.Equal(t, "Updated Risk Name", result.UpdateRisk.Risk.Name)
	assert.Equal(t, "TRANSFERRED", result.UpdateRisk.Risk.Treatment)
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
				"name":               "Test Risk",
				"category":           "SECURITY",
				"treatment":          "MITIGATED",
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing name",
			input: map[string]any{
				"category":           "SECURITY",
				"treatment":          "MITIGATED",
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
			wantErrorContains: "name",
		},
		{
			name: "missing category",
			input: map[string]any{
				"name":               "Test Risk",
				"treatment":          "MITIGATED",
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
			wantErrorContains: "category",
		},
		{
			name: "missing treatment",
			input: map[string]any{
				"name":               "Test Risk",
				"category":           "SECURITY",
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
			wantErrorContains: "treatment",
		},
		{
			name: "missing inherentLikelihood",
			input: map[string]any{
				"name":           "Test Risk",
				"category":       "SECURITY",
				"treatment":      "MITIGATED",
				"inherentImpact": 2,
			},
			wantErrorContains: "inherentLikelihood",
		},
		{
			name: "missing inherentImpact",
			input: map[string]any{
				"name":               "Test Risk",
				"category":           "SECURITY",
				"treatment":          "MITIGATED",
				"inherentLikelihood": 2,
			},
			wantErrorContains: "inherentImpact",
		},
		{
			name: "invalid treatment enum",
			input: map[string]any{
				"name":               "Test Risk",
				"category":           "SECURITY",
				"treatment":          "INVALID_TREATMENT",
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
			wantErrorContains: "treatment",
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
			for k, v := range tt.input {
				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestRisk_TreatmentEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	treatments := []string{
		"MITIGATED",
		"ACCEPTED",
		"TRANSFERRED",
		"AVOIDED",
	}

	for _, treatment := range treatments {
		t.Run("create with treatment "+treatment, func(t *testing.T) {
			riskID := factory.NewRisk(owner).
				WithName("Treatment Test " + treatment).
				WithTreatment(treatment).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on Risk {
							id
							treatment
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID        string `json:"id"`
					Treatment string `json:"treatment"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": riskID}, &result)
			require.NoError(t, err)
			assert.Equal(t, treatment, result.Node.Treatment)
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
						treatment
						inherentLikelihood
						inherentImpact
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID                 string  `json:"id"`
				Name               string  `json:"name"`
				Description        *string `json:"description"`
				Category           string  `json:"category"`
				Treatment          string  `json:"treatment"`
				InherentLikelihood int     `json:"inherentLikelihood"`
				InherentImpact     int     `json:"inherentImpact"`
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
									title
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
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Node.Documents.Edges)
	})

	t.Run("owner sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
						id
						owner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID    string `json:"id"`
				Owner *struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"owner"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": riskID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.Owner)
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

func TestRisk_OmittableOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a people for owner assignment
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	riskID := factory.NewRisk(owner).WithName("Owner Test Risk").Create()

	t.Run("set owner", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						owner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID    string `json:"id"`
					Owner struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"owner"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":      riskID,
				"ownerId": profileID.String(),
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID.String(), result.UpdateRisk.Risk.Owner.ID)
	})

	t.Run("clear owner with null", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						owner {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID    string `json:"id"`
					Owner *struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":      riskID,
				"ownerId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateRisk.Risk.Owner)
	})
}

func TestRisk_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.NewRisk(org1Owner).WithName("Org1 Risk").Create()

	t.Run("cannot read risk from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
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

		err := org2Owner.Execute(query, map[string]any{"id": riskID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "risk")
	})

	t.Run("cannot update risk from another organization", func(t *testing.T) {
		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   riskID,
				"name": "Hijacked Risk",
			},
		})
		require.Error(t, err, "Should not be able to update risk from another org")
	})

	t.Run("cannot delete risk from another organization", func(t *testing.T) {
		query := `
			mutation DeleteRisk($input: DeleteRiskInput!) {
				deleteRisk(input: $input) {
					deletedRiskId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"riskId": riskID,
			},
		})
		require.Error(t, err, "Should not be able to delete risk from another org")
	})
}

func TestRisk_LikelihoodImpactValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("create with various likelihood and impact values", func(t *testing.T) {
		// Test valid values 1-5
		for i := 1; i <= 5; i++ {
			riskID := factory.NewRisk(owner).
				WithName("Likelihood Impact Test").
				WithLikelihood(i).
				WithImpact(i).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on Risk {
							id
							inherentLikelihood
							inherentImpact
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID                 string `json:"id"`
					InherentLikelihood int    `json:"inherentLikelihood"`
					InherentImpact     int    `json:"inherentImpact"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": riskID}, &result)
			require.NoError(t, err)
			assert.Equal(t, i, result.Node.InherentLikelihood)
			assert.Equal(t, i, result.Node.InherentImpact)
		}
	})

	t.Run("update likelihood and impact", func(t *testing.T) {
		riskID := factory.NewRisk(owner).
			WithName("Update Likelihood Impact Test").
			WithLikelihood(1).
			WithImpact(1).
			Create()

		query := `
			mutation UpdateRisk($input: UpdateRiskInput!) {
				updateRisk(input: $input) {
					risk {
						id
						inherentLikelihood
						inherentImpact
					}
				}
			}
		`

		var result struct {
			UpdateRisk struct {
				Risk struct {
					ID                 string `json:"id"`
					InherentLikelihood int    `json:"inherentLikelihood"`
					InherentImpact     int    `json:"inherentImpact"`
				} `json:"risk"`
			} `json:"updateRisk"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":                 riskID,
				"inherentLikelihood": 5,
				"inherentImpact":     4,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, 5, result.UpdateRisk.Risk.InherentLikelihood)
		assert.Equal(t, 4, result.UpdateRisk.Risk.InherentImpact)
	})
}
