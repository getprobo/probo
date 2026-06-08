// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

func TestRiskAssessment_Create(t *testing.T) {
	t.Parallel()

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		var result struct {
			CreateRiskAssessment struct {
				RiskAssessmentEdge struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"riskAssessmentEdge"`
			} `json:"createRiskAssessment"`
		}

		err := owner.Execute(`
			mutation($input: CreateRiskAssessmentInput!) {
				createRiskAssessment(input: $input) {
					riskAssessmentEdge { node { id name } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Platform Threat Model",
			},
		}, &result)

		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateRiskAssessment.RiskAssessmentEdge.Node.ID)
		assert.Equal(t, "Platform Threat Model", result.CreateRiskAssessment.RiskAssessmentEdge.Node.Name)
	})
}

func TestRiskAssessment_Delete(t *testing.T) {
	t.Parallel()

	t.Run("cascades to scopes", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		raID := factory.CreateRiskAssessment(owner)
		scopeID := factory.CreateRiskAssessmentScope(owner, raID)

		_, err := owner.Do(`
			mutation($input: DeleteRiskAssessmentInput!) {
				deleteRiskAssessment(input: $input) { deletedRiskAssessmentId }
			}
		`, map[string]any{"input": map[string]any{"riskAssessmentId": raID}})
		require.NoError(t, err)

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err = owner.Execute(`query($id: ID!) { node(id: $id) { ... on RiskAssessmentScope { id } } }`,
			map[string]any{"id": scopeID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "RiskAssessmentScope")
	})
}

func TestRiskAssessmentScope_CRUD(t *testing.T) {
	t.Parallel()

	t.Run("create and list via assessment", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		raID := factory.CreateRiskAssessment(owner)
		factory.CreateRiskAssessmentScope(owner, raID, factory.Attrs{"name": "API scope"})
		factory.CreateRiskAssessmentScope(owner, raID, factory.Attrs{"name": "Infra scope"})

		var result struct {
			Node struct {
				Scopes struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"scopes"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on RiskAssessment {
						scopes(first: 10) {
							totalCount
							edges { node { id name } }
						}
					}
				}
			}
		`, map[string]any{"id": raID}, &result)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Node.Scopes.TotalCount)
		assert.Len(t, result.Node.Scopes.Edges, 2)
	})
}

func TestRiskAssessmentNode_Create(t *testing.T) {
	t.Parallel()

	for _, nodeType := range []string{"ENTITY", "ASSET", "DATA"} {
		t.Run("nodeType="+nodeType, func(t *testing.T) {
			t.Parallel()
			owner := testutil.NewClient(t, testutil.RoleOwner)
			raID := factory.CreateRiskAssessment(owner)
			scopeID := factory.CreateRiskAssessmentScope(owner, raID)

			var result struct {
				CreateRiskAssessmentNode struct {
					RiskAssessmentNodeEdge struct {
						Node struct {
							ID       string `json:"id"`
							NodeType string `json:"nodeType"`
						} `json:"node"`
					} `json:"riskAssessmentNodeEdge"`
				} `json:"createRiskAssessmentNode"`
			}

			err := owner.Execute(`
				mutation($input: CreateRiskAssessmentNodeInput!) {
					createRiskAssessmentNode(input: $input) {
						riskAssessmentNodeEdge { node { id nodeType } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"riskAssessmentScopeId": scopeID,
					"nodeType":              nodeType,
					"name":                  "Node-" + nodeType,
				},
			}, &result)

			require.NoError(t, err)
			assert.Equal(t, nodeType, result.CreateRiskAssessmentNode.RiskAssessmentNodeEdge.Node.NodeType)
		})
	}
}

func TestRiskAssessmentProcess_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	src := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY"})
	dst := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"nodeType": "ASSET"})

	var result struct {
		CreateRiskAssessmentProcess struct {
			RiskAssessmentProcessEdge struct {
				Node struct {
					ID           string `json:"id"`
					SourceNodeID string `json:"sourceNodeId"`
					TargetNodeID string `json:"targetNodeId"`
					Name         string `json:"name"`
				} `json:"node"`
			} `json:"riskAssessmentProcessEdge"`
		} `json:"createRiskAssessmentProcess"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAssessmentProcessInput!) {
			createRiskAssessmentProcess(input: $input) {
				riskAssessmentProcessEdge { node { id sourceNodeId targetNodeId name } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScopeId": scopeID,
			"sourceNodeId":          src,
			"targetNodeId":          dst,
			"name":                  "User → API",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, src, result.CreateRiskAssessmentProcess.RiskAssessmentProcessEdge.Node.SourceNodeID)
	assert.Equal(t, dst, result.CreateRiskAssessmentProcess.RiskAssessmentProcessEdge.Node.TargetNodeID)
}

func TestRiskAssessmentThreat_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	src := factory.CreateRiskAssessmentNode(owner, scopeID)
	dst := factory.CreateRiskAssessmentNode(owner, scopeID)
	processID := factory.CreateRiskAssessmentProcess(owner, scopeID, src, dst)

	var result struct {
		CreateRiskAssessmentThreat struct {
			RiskAssessmentThreatEdge struct {
				Node struct {
					ID        string `json:"id"`
					ProcessID string `json:"processId"`
					Category  string `json:"category"`
				} `json:"node"`
			} `json:"riskAssessmentThreatEdge"`
		} `json:"createRiskAssessmentThreat"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAssessmentThreatInput!) {
			createRiskAssessmentThreat(input: $input) {
				riskAssessmentThreatEdge { node { id processId category } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScopeId": scopeID,
			"processId":             processID,
			"name":                  "SQL injection",
			"category":              "Confidentiality",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, processID, result.CreateRiskAssessmentThreat.RiskAssessmentThreatEdge.Node.ProcessID)
	assert.Equal(t, "Confidentiality", result.CreateRiskAssessmentThreat.RiskAssessmentThreatEdge.Node.Category)
}

func TestRiskAssessmentScenario_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)

	var result struct {
		CreateRiskAssessmentScenario struct {
			RiskAssessmentScenarioEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"riskAssessmentScenarioEdge"`
		} `json:"createRiskAssessmentScenario"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAssessmentScenarioInput!) {
			createRiskAssessmentScenario(input: $input) {
				riskAssessmentScenarioEdge { node { id name } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScopeId": scopeID,
			"name":                  "SQL injection impacts data breach risk",
		},
	}, &result)

	require.NoError(t, err)
	assert.NotEmpty(t, result.CreateRiskAssessmentScenario.RiskAssessmentScenarioEdge.Node.ID)
	assert.Equal(t, "SQL injection impacts data breach risk", result.CreateRiskAssessmentScenario.RiskAssessmentScenarioEdge.Node.Name)
}

func TestRiskAssessmentScenario_ListViaRisk(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	riskID := factory.CreateRisk(owner)
	s1 := factory.CreateRiskAssessmentScenario(owner, scopeID, factory.Attrs{"name": "S1"})
	s2 := factory.CreateRiskAssessmentScenario(owner, scopeID, factory.Attrs{"name": "S2"})

	factory.LinkRiskAssessmentScenarioRisk(owner, s1, riskID)
	factory.LinkRiskAssessmentScenarioRisk(owner, s2, riskID)

	var result struct {
		Node struct {
			Scenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"scenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					scenarios(first: 10) {
						totalCount
						edges { node { id name } }
					}
				}
			}
		}
	`, map[string]any{"id": riskID}, &result)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Node.Scenarios.TotalCount)
	assert.Len(t, result.Node.Scenarios.Edges, 2)
}

func TestRiskAssessmentScenario_ListViaScope(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	factory.CreateRiskAssessmentScenario(owner, scopeID, factory.Attrs{"name": "Scenario A"})
	factory.CreateRiskAssessmentScenario(owner, scopeID, factory.Attrs{"name": "Scenario B"})

	var result struct {
		Node struct {
			Scenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"scenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScope {
					scenarios(first: 10) {
						totalCount
						edges { node { id name } }
					}
				}
			}
		}
	`, map[string]any{"id": scopeID}, &result)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Node.Scenarios.TotalCount)
	assert.Len(t, result.Node.Scenarios.Edges, 2)
}

func TestRiskAssessment_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAssessment struct {
			RiskAssessment struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
			} `json:"riskAssessment"`
		} `json:"updateRiskAssessment"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentInput!) {
			updateRiskAssessment(input: $input) {
				riskAssessment { id name description }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":          raID,
			"name":        "Updated",
			"description": "New description",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated", result.UpdateRiskAssessment.RiskAssessment.Name)
	require.NotNil(t, result.UpdateRiskAssessment.RiskAssessment.Description)
	assert.Equal(t, "New description", *result.UpdateRiskAssessment.RiskAssessment.Description)
}

func TestRiskAssessmentScope_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAssessmentScope struct {
			RiskAssessmentScope struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"riskAssessmentScope"`
		} `json:"updateRiskAssessmentScope"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentScopeInput!) {
			updateRiskAssessmentScope(input: $input) {
				riskAssessmentScope { id name }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":   scopeID,
			"name": "Updated scope",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated scope", result.UpdateRiskAssessmentScope.RiskAssessmentScope.Name)
}

func TestRiskAssessmentNode_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	nodeID := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY", "name": "Original"})

	var result struct {
		UpdateRiskAssessmentNode struct {
			RiskAssessmentNode struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				NodeType string `json:"nodeType"`
			} `json:"riskAssessmentNode"`
		} `json:"updateRiskAssessmentNode"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentNodeInput!) {
			updateRiskAssessmentNode(input: $input) {
				riskAssessmentNode { id name nodeType }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":       nodeID,
			"name":     "Updated node",
			"nodeType": "ASSET",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated node", result.UpdateRiskAssessmentNode.RiskAssessmentNode.Name)
	assert.Equal(t, "ASSET", result.UpdateRiskAssessmentNode.RiskAssessmentNode.NodeType)
}

func TestRiskAssessmentProcess_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	src := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY"})
	dst := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"nodeType": "ASSET"})
	processID := factory.CreateRiskAssessmentProcess(owner, scopeID, src, dst)

	var result struct {
		UpdateRiskAssessmentProcess struct {
			RiskAssessmentProcess struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"riskAssessmentProcess"`
		} `json:"updateRiskAssessmentProcess"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentProcessInput!) {
			updateRiskAssessmentProcess(input: $input) {
				riskAssessmentProcess { id name }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":   processID,
			"name": "Updated process",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated process", result.UpdateRiskAssessmentProcess.RiskAssessmentProcess.Name)
}

func TestRiskAssessmentThreat_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	src := factory.CreateRiskAssessmentNode(owner, scopeID)
	dst := factory.CreateRiskAssessmentNode(owner, scopeID)
	processID := factory.CreateRiskAssessmentProcess(owner, scopeID, src, dst)
	threatID := factory.CreateRiskAssessmentThreat(owner, scopeID, processID, factory.Attrs{"name": "Original", "category": "Confidentiality"})

	var result struct {
		UpdateRiskAssessmentThreat struct {
			RiskAssessmentThreat struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Category string `json:"category"`
			} `json:"riskAssessmentThreat"`
		} `json:"updateRiskAssessmentThreat"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentThreatInput!) {
			updateRiskAssessmentThreat(input: $input) {
				riskAssessmentThreat { id name category }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":       threatID,
			"name":     "Updated threat",
			"category": "Integrity",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated threat", result.UpdateRiskAssessmentThreat.RiskAssessmentThreat.Name)
	assert.Equal(t, "Integrity", result.UpdateRiskAssessmentThreat.RiskAssessmentThreat.Category)
}

func TestRiskAssessmentScenario_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	scenarioID := factory.CreateRiskAssessmentScenario(owner, scopeID, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAssessmentScenario struct {
			RiskAssessmentScenario struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
			} `json:"riskAssessmentScenario"`
		} `json:"updateRiskAssessmentScenario"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAssessmentScenarioInput!) {
			updateRiskAssessmentScenario(input: $input) {
				riskAssessmentScenario { id name description }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":          scenarioID,
			"name":        "Updated scenario",
			"description": "Scenario desc",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated scenario", result.UpdateRiskAssessmentScenario.RiskAssessmentScenario.Name)
	require.NotNil(t, result.UpdateRiskAssessmentScenario.RiskAssessmentScenario.Description)
	assert.Equal(t, "Scenario desc", *result.UpdateRiskAssessmentScenario.RiskAssessmentScenario.Description)
}

func TestRiskAssessmentScenario_LinkUnlinkThreat(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	src := factory.CreateRiskAssessmentNode(owner, scopeID)
	dst := factory.CreateRiskAssessmentNode(owner, scopeID)
	processID := factory.CreateRiskAssessmentProcess(owner, scopeID, src, dst)
	threatID := factory.CreateRiskAssessmentThreat(owner, scopeID, processID)
	scenarioID := factory.CreateRiskAssessmentScenario(owner, scopeID)

	factory.LinkRiskAssessmentScenarioThreat(owner, scenarioID, threatID)

	var result struct {
		Node struct {
			Threats struct {
				TotalCount int `json:"totalCount"`
			} `json:"threats"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScenario {
					threats(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Node.Threats.TotalCount)

	_, err = owner.Do(`
		mutation($input: UnlinkRiskAssessmentScenarioThreatInput!) {
			unlinkRiskAssessmentScenarioThreat(input: $input) { riskAssessmentScenario { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScenarioId": scenarioID,
			"threatId":                 threatID,
		},
	})
	require.NoError(t, err)

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScenario {
					threats(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Node.Threats.TotalCount)
}

func TestRiskAssessmentScenario_LinkUnlinkRisk(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	riskID := factory.CreateRisk(owner)
	scenarioID := factory.CreateRiskAssessmentScenario(owner, scopeID)

	factory.LinkRiskAssessmentScenarioRisk(owner, scenarioID, riskID)

	var result struct {
		Node struct {
			Risks struct {
				TotalCount int `json:"totalCount"`
			} `json:"risks"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScenario {
					risks(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Node.Risks.TotalCount)

	_, err = owner.Do(`
		mutation($input: UnlinkRiskAssessmentScenarioRiskInput!) {
			unlinkRiskAssessmentScenarioRisk(input: $input) { riskAssessmentScenario { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScenarioId": scenarioID,
			"riskId":                   riskID,
		},
	})
	require.NoError(t, err)

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScenario {
					risks(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Node.Risks.TotalCount)
}

func TestRiskAssessment_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(`
			mutation($input: CreateRiskAssessmentInput!) {
				createRiskAssessment(input: $input) { riskAssessmentEdge { node { id } } }
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"name":           "test",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create risk assessment")
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAssessment(owner, factory.Attrs{"name": "Visible"})

		var result struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) { node(id: $id) { ... on RiskAssessment { id name } } }
		`, map[string]any{"id": raID}, &result)
		require.NoError(t, err)
		assert.Equal(t, "Visible", result.Node.Name)
	})
}

func TestRiskAssessment_TenantIsolation(t *testing.T) {
	t.Parallel()

	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner1)

	var result struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err := owner2.Execute(`
		query($id: ID!) { node(id: $id) { ... on RiskAssessment { id } } }
	`, map[string]any{"id": raID}, &result)
	testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "RiskAssessment")
}

func TestRiskAssessmentBoundary_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	parentID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "External Zone"})

	var result struct {
		CreateRiskAssessmentBoundary struct {
			RiskAssessmentBoundaryEdge struct {
				Node struct {
					ID               string  `json:"id"`
					Name             string  `json:"name"`
					ParentBoundaryID *string `json:"parentBoundaryId"`
				} `json:"node"`
			} `json:"riskAssessmentBoundaryEdge"`
		} `json:"createRiskAssessmentBoundary"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAssessmentBoundaryInput!) {
			createRiskAssessmentBoundary(input: $input) {
				riskAssessmentBoundaryEdge { node { id name parentBoundaryId } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScopeId": scopeID,
			"parentBoundaryId":      parentID,
			"name":                  "Internal Network",
		},
	}, &result)

	require.NoError(t, err)

	node := result.CreateRiskAssessmentBoundary.RiskAssessmentBoundaryEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "Internal Network", node.Name)
	require.NotNil(t, node.ParentBoundaryID)
	assert.Equal(t, parentID, *node.ParentBoundaryID)
}

func TestRiskAssessmentBoundary_ListViaScope(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Zone A"})
	factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Zone B"})

	var result struct {
		Node struct {
			Boundaries struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"boundaries"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentScope {
					boundaries(first: 10) {
						totalCount
						edges { node { id name } }
					}
				}
			}
		}
	`, map[string]any{"id": scopeID}, &result)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Node.Boundaries.TotalCount)
	assert.Len(t, result.Node.Boundaries.Edges, 2)
}

func TestRiskAssessmentBoundary_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	parentID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Parent"})
	boundaryID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Original"})

	const mutation = `
		mutation($input: UpdateRiskAssessmentBoundaryInput!) {
			updateRiskAssessmentBoundary(input: $input) {
				riskAssessmentBoundary { id name parentBoundaryId }
			}
		}
	`

	var result struct {
		UpdateRiskAssessmentBoundary struct {
			RiskAssessmentBoundary struct {
				ID               string  `json:"id"`
				Name             string  `json:"name"`
				ParentBoundaryID *string `json:"parentBoundaryId"`
			} `json:"riskAssessmentBoundary"`
		} `json:"updateRiskAssessmentBoundary"`
	}

	// Rename and assign a parent.
	err := owner.Execute(mutation, map[string]any{
		"input": map[string]any{
			"id":               boundaryID,
			"name":             "Renamed",
			"parentBoundaryId": parentID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", result.UpdateRiskAssessmentBoundary.RiskAssessmentBoundary.Name)
	require.NotNil(t, result.UpdateRiskAssessmentBoundary.RiskAssessmentBoundary.ParentBoundaryID)
	assert.Equal(t, parentID, *result.UpdateRiskAssessmentBoundary.RiskAssessmentBoundary.ParentBoundaryID)

	// Clear the parent (move back to the top level).
	err = owner.Execute(mutation, map[string]any{
		"input": map[string]any{
			"id":               boundaryID,
			"parentBoundaryId": nil,
		},
	}, &result)
	require.NoError(t, err)
	assert.Nil(t, result.UpdateRiskAssessmentBoundary.RiskAssessmentBoundary.ParentBoundaryID)
}

func TestRiskAssessmentBoundary_PreventCycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	a := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "A"})
	b := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "B", "parentBoundaryId": a})

	// B is nested under A, so nesting A under B would create a cycle.
	_, err := owner.Do(`
		mutation($input: UpdateRiskAssessmentBoundaryInput!) {
			updateRiskAssessmentBoundary(input: $input) {
				riskAssessmentBoundary { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":               a,
			"parentBoundaryId": b,
		},
	})
	require.Error(t, err, "nesting a boundary under its own descendant should be rejected")
}

func TestRiskAssessmentBoundary_Delete(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	parentID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Parent"})
	childID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Child", "parentBoundaryId": parentID})
	nodeID := factory.CreateRiskAssessmentNode(owner, scopeID, factory.Attrs{"name": "Member", "boundaryId": parentID})

	_, err := owner.Do(`
		mutation($input: DeleteRiskAssessmentBoundaryInput!) {
			deleteRiskAssessmentBoundary(input: $input) { deletedRiskAssessmentBoundaryId }
		}
	`, map[string]any{"input": map[string]any{"riskAssessmentBoundaryId": parentID}})
	require.NoError(t, err)

	// Deleting a parent moves its nested boundary and member node to the top
	// level instead of cascading the delete (ON DELETE SET NULL).
	var result struct {
		Child *struct {
			ParentBoundaryID *string `json:"parentBoundaryId"`
		} `json:"child"`
		Member *struct {
			BoundaryID *string `json:"boundaryId"`
		} `json:"member"`
	}

	err = owner.Execute(`
		query($child: ID!, $member: ID!) {
			child: node(id: $child) { ... on RiskAssessmentBoundary { parentBoundaryId } }
			member: node(id: $member) { ... on RiskAssessmentNode { boundaryId } }
		}
	`, map[string]any{"child": childID, "member": nodeID}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Child)
	assert.Nil(t, result.Child.ParentBoundaryID)
	require.NotNil(t, result.Member)
	assert.Nil(t, result.Member.BoundaryID)
}

func TestRiskAssessmentNode_WithBoundary(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAssessment(owner)
	scopeID := factory.CreateRiskAssessmentScope(owner, raID)
	boundaryID := factory.CreateRiskAssessmentBoundary(owner, scopeID)

	var createResult struct {
		CreateRiskAssessmentNode struct {
			RiskAssessmentNodeEdge struct {
				Node struct {
					ID         string  `json:"id"`
					BoundaryID *string `json:"boundaryId"`
				} `json:"node"`
			} `json:"riskAssessmentNodeEdge"`
		} `json:"createRiskAssessmentNode"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAssessmentNodeInput!) {
			createRiskAssessmentNode(input: $input) {
				riskAssessmentNodeEdge { node { id boundaryId } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAssessmentScopeId": scopeID,
			"nodeType":              "ASSET",
			"name":                  "Member",
			"boundaryId":            boundaryID,
		},
	}, &createResult)
	require.NoError(t, err)

	created := createResult.CreateRiskAssessmentNode.RiskAssessmentNodeEdge.Node
	require.NotNil(t, created.BoundaryID)
	assert.Equal(t, boundaryID, *created.BoundaryID)

	// Clearing boundaryId moves the node back to the top level.
	var updateResult struct {
		UpdateRiskAssessmentNode struct {
			RiskAssessmentNode struct {
				BoundaryID *string `json:"boundaryId"`
			} `json:"riskAssessmentNode"`
		} `json:"updateRiskAssessmentNode"`
	}

	err = owner.Execute(`
		mutation($input: UpdateRiskAssessmentNodeInput!) {
			updateRiskAssessmentNode(input: $input) {
				riskAssessmentNode { boundaryId }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":         created.ID,
			"boundaryId": nil,
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Nil(t, updateResult.UpdateRiskAssessmentNode.RiskAssessmentNode.BoundaryID)
}

func TestRiskAssessmentBoundary_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAssessment(owner)
		scopeID := factory.CreateRiskAssessmentScope(owner, raID)

		_, err := viewer.Do(`
			mutation($input: CreateRiskAssessmentBoundaryInput!) {
				createRiskAssessmentBoundary(input: $input) {
					riskAssessmentBoundaryEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskAssessmentScopeId": scopeID,
				"name":                  "Nope",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create risk assessment boundary")
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAssessment(owner)
		scopeID := factory.CreateRiskAssessmentScope(owner, raID)
		boundaryID := factory.CreateRiskAssessmentBoundary(owner, scopeID, factory.Attrs{"name": "Visible"})

		var result struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) { node(id: $id) { ... on RiskAssessmentBoundary { id name } } }
		`, map[string]any{"id": boundaryID}, &result)
		require.NoError(t, err)
		assert.Equal(t, "Visible", result.Node.Name)
	})
}
