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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestRiskAnalysis_Create(t *testing.T) {
	t.Parallel()

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		var result struct {
			CreateRiskAnalysis struct {
				RiskAnalysisEdge struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"riskAnalysisEdge"`
			} `json:"createRiskAnalysis"`
		}

		err := owner.Execute(`
			mutation($input: CreateRiskAnalysisInput!) {
				createRiskAnalysis(input: $input) {
					riskAnalysisEdge { node { id name } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Platform Threat Model",
			},
		}, &result)

		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateRiskAnalysis.RiskAnalysisEdge.Node.ID)
		assert.Equal(t, "Platform Threat Model", result.CreateRiskAnalysis.RiskAnalysisEdge.Node.Name)
	})
}

func TestRiskAnalysis_Delete(t *testing.T) {
	t.Parallel()

	t.Run("cascades to scopes", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		raID := factory.CreateRiskAnalysis(owner)
		scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)

		_, err := owner.Do(`
			mutation($input: DeleteRiskAnalysisInput!) {
				deleteRiskAnalysis(input: $input) { deletedRiskAnalysisId }
			}
		`, map[string]any{"input": map[string]any{"riskAnalysisId": raID}})
		require.NoError(t, err)

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err = owner.Execute(`query($id: ID!) { node(id: $id) { ... on RiskAnalysisDiagram { id } } }`,
			map[string]any{"id": scopeID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "RiskAnalysisDiagram")
	})
}

func TestRiskAnalysisDiagram_CRUD(t *testing.T) {
	t.Parallel()

	t.Run("create and list via assessment", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		raID := factory.CreateRiskAnalysis(owner)
		factory.CreateRiskAnalysisDiagram(owner, raID, factory.Attrs{"name": "API scope"})
		factory.CreateRiskAnalysisDiagram(owner, raID, factory.Attrs{"name": "Infra diagram"})

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
				} `json:"diagrams"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on RiskAnalysis {
						diagrams(first: 10) {
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

func TestRiskAnalysisNode_Create(t *testing.T) {
	t.Parallel()

	for _, nodeType := range []string{"ENTITY", "ASSET", "DATA"} {
		t.Run("nodeType="+nodeType, func(t *testing.T) {
			t.Parallel()
			owner := testutil.NewClient(t, testutil.RoleOwner)
			raID := factory.CreateRiskAnalysis(owner)
			scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)

			var result struct {
				CreateRiskAnalysisNode struct {
					RiskAnalysisNodeEdge struct {
						Node struct {
							ID       string `json:"id"`
							NodeType string `json:"nodeType"`
						} `json:"node"`
					} `json:"riskAnalysisNodeEdge"`
				} `json:"createRiskAnalysisNode"`
			}

			err := owner.Execute(`
				mutation($input: CreateRiskAnalysisNodeInput!) {
					createRiskAnalysisNode(input: $input) {
						riskAnalysisNodeEdge { node { id nodeType } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"riskAnalysisDiagramId": scopeID,
					"nodeType":              nodeType,
					"name":                  "Node-" + nodeType,
				},
			}, &result)

			require.NoError(t, err)
			assert.Equal(t, nodeType, result.CreateRiskAnalysisNode.RiskAnalysisNodeEdge.Node.NodeType)
		})
	}
}

func TestRiskAnalysisProcess_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	src := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY"})
	dst := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"nodeType": "ASSET"})

	var result struct {
		CreateRiskAnalysisProcess struct {
			RiskAnalysisProcessEdge struct {
				Node struct {
					ID           string `json:"id"`
					SourceNodeID string `json:"sourceNodeId"`
					TargetNodeID string `json:"targetNodeId"`
					Name         string `json:"name"`
				} `json:"node"`
			} `json:"riskAnalysisProcessEdge"`
		} `json:"createRiskAnalysisProcess"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAnalysisProcessInput!) {
			createRiskAnalysisProcess(input: $input) {
				riskAnalysisProcessEdge { node { id sourceNodeId targetNodeId name } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisDiagramId": scopeID,
			"sourceNodeId":          src,
			"targetNodeId":          dst,
			"name":                  "User → API",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, src, result.CreateRiskAnalysisProcess.RiskAnalysisProcessEdge.Node.SourceNodeID)
	assert.Equal(t, dst, result.CreateRiskAnalysisProcess.RiskAnalysisProcessEdge.Node.TargetNodeID)
}

func TestRiskAnalysisThreat_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	src := factory.CreateRiskAnalysisNode(owner, scopeID)
	dst := factory.CreateRiskAnalysisNode(owner, scopeID)
	processID := factory.CreateRiskAnalysisProcess(owner, scopeID, src, dst)

	var result struct {
		CreateRiskAnalysisThreat struct {
			RiskAnalysisThreatEdge struct {
				Node struct {
					ID        string `json:"id"`
					ProcessID string `json:"processId"`
					Category  string `json:"category"`
				} `json:"node"`
			} `json:"riskAnalysisThreatEdge"`
		} `json:"createRiskAnalysisThreat"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAnalysisThreatInput!) {
			createRiskAnalysisThreat(input: $input) {
				riskAnalysisThreatEdge { node { id processId category } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisDiagramId": scopeID,
			"processId":             processID,
			"name":                  "SQL injection",
			"category":              "Confidentiality",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, processID, result.CreateRiskAnalysisThreat.RiskAnalysisThreatEdge.Node.ProcessID)
	assert.Equal(t, "Confidentiality", result.CreateRiskAnalysisThreat.RiskAnalysisThreatEdge.Node.Category)
}

func TestRiskAnalysisScenario_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)

	var result struct {
		CreateRiskAnalysisScenario struct {
			RiskAnalysisScenarioEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"riskAnalysisScenarioEdge"`
		} `json:"createRiskAnalysisScenario"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAnalysisScenarioInput!) {
			createRiskAnalysisScenario(input: $input) {
				riskAnalysisScenarioEdge { node { id name } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisDiagramId": scopeID,
			"name":                  "SQL injection impacts data breach risk",
		},
	}, &result)

	require.NoError(t, err)
	assert.NotEmpty(t, result.CreateRiskAnalysisScenario.RiskAnalysisScenarioEdge.Node.ID)
	assert.Equal(t, "SQL injection impacts data breach risk", result.CreateRiskAnalysisScenario.RiskAnalysisScenarioEdge.Node.Name)
}

func TestRiskAnalysisScenario_ListViaRisk(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	riskID := factory.CreateRisk(owner)
	s1 := factory.CreateRiskAnalysisScenario(owner, scopeID, factory.Attrs{"name": "S1"})
	s2 := factory.CreateRiskAnalysisScenario(owner, scopeID, factory.Attrs{"name": "S2"})

	factory.LinkRiskAnalysisScenarioRisk(owner, s1, riskID)
	factory.LinkRiskAnalysisScenarioRisk(owner, s2, riskID)

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

func TestRiskAnalysisScenario_ListViaScope(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	factory.CreateRiskAnalysisScenario(owner, scopeID, factory.Attrs{"name": "Scenario A"})
	factory.CreateRiskAnalysisScenario(owner, scopeID, factory.Attrs{"name": "Scenario B"})

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
				... on RiskAnalysisDiagram {
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

func TestRiskAnalysis_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAnalysis struct {
			RiskAnalysis struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
			} `json:"riskAnalysis"`
		} `json:"updateRiskAnalysis"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisInput!) {
			updateRiskAnalysis(input: $input) {
				riskAnalysis { id name description }
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
	assert.Equal(t, "Updated", result.UpdateRiskAnalysis.RiskAnalysis.Name)
	require.NotNil(t, result.UpdateRiskAnalysis.RiskAnalysis.Description)
	assert.Equal(t, "New description", *result.UpdateRiskAnalysis.RiskAnalysis.Description)
}

func TestRiskAnalysisDiagram_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAnalysisDiagram struct {
			RiskAnalysisDiagram struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"riskAnalysisDiagram"`
		} `json:"updateRiskAnalysisDiagram"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisDiagramInput!) {
			updateRiskAnalysisDiagram(input: $input) {
				riskAnalysisDiagram { id name }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":   scopeID,
			"name": "Updated scope",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated scope", result.UpdateRiskAnalysisDiagram.RiskAnalysisDiagram.Name)
}

func TestRiskAnalysisNode_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	nodeID := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY", "name": "Original"})

	var result struct {
		UpdateRiskAnalysisNode struct {
			RiskAnalysisNode struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				NodeType string `json:"nodeType"`
			} `json:"riskAnalysisNode"`
		} `json:"updateRiskAnalysisNode"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisNodeInput!) {
			updateRiskAnalysisNode(input: $input) {
				riskAnalysisNode { id name nodeType }
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
	assert.Equal(t, "Updated node", result.UpdateRiskAnalysisNode.RiskAnalysisNode.Name)
	assert.Equal(t, "ASSET", result.UpdateRiskAnalysisNode.RiskAnalysisNode.NodeType)
}

func TestRiskAnalysisProcess_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	src := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"nodeType": "ENTITY"})
	dst := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"nodeType": "ASSET"})
	processID := factory.CreateRiskAnalysisProcess(owner, scopeID, src, dst)

	var result struct {
		UpdateRiskAnalysisProcess struct {
			RiskAnalysisProcess struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"riskAnalysisProcess"`
		} `json:"updateRiskAnalysisProcess"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisProcessInput!) {
			updateRiskAnalysisProcess(input: $input) {
				riskAnalysisProcess { id name }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":   processID,
			"name": "Updated process",
		},
	}, &result)

	require.NoError(t, err)
	assert.Equal(t, "Updated process", result.UpdateRiskAnalysisProcess.RiskAnalysisProcess.Name)
}

func TestRiskAnalysisThreat_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	src := factory.CreateRiskAnalysisNode(owner, scopeID)
	dst := factory.CreateRiskAnalysisNode(owner, scopeID)
	processID := factory.CreateRiskAnalysisProcess(owner, scopeID, src, dst)
	threatID := factory.CreateRiskAnalysisThreat(owner, scopeID, processID, factory.Attrs{"name": "Original", "category": "Confidentiality"})

	var result struct {
		UpdateRiskAnalysisThreat struct {
			RiskAnalysisThreat struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Category string `json:"category"`
			} `json:"riskAnalysisThreat"`
		} `json:"updateRiskAnalysisThreat"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisThreatInput!) {
			updateRiskAnalysisThreat(input: $input) {
				riskAnalysisThreat { id name category }
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
	assert.Equal(t, "Updated threat", result.UpdateRiskAnalysisThreat.RiskAnalysisThreat.Name)
	assert.Equal(t, "Integrity", result.UpdateRiskAnalysisThreat.RiskAnalysisThreat.Category)
}

func TestRiskAnalysisScenario_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	scenarioID := factory.CreateRiskAnalysisScenario(owner, scopeID, factory.Attrs{"name": "Original"})

	var result struct {
		UpdateRiskAnalysisScenario struct {
			RiskAnalysisScenario struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
			} `json:"riskAnalysisScenario"`
		} `json:"updateRiskAnalysisScenario"`
	}

	err := owner.Execute(`
		mutation($input: UpdateRiskAnalysisScenarioInput!) {
			updateRiskAnalysisScenario(input: $input) {
				riskAnalysisScenario { id name description }
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
	assert.Equal(t, "Updated scenario", result.UpdateRiskAnalysisScenario.RiskAnalysisScenario.Name)
	require.NotNil(t, result.UpdateRiskAnalysisScenario.RiskAnalysisScenario.Description)
	assert.Equal(t, "Scenario desc", *result.UpdateRiskAnalysisScenario.RiskAnalysisScenario.Description)
}

func TestRiskAnalysisScenario_LinkUnlinkThreat(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	src := factory.CreateRiskAnalysisNode(owner, scopeID)
	dst := factory.CreateRiskAnalysisNode(owner, scopeID)
	processID := factory.CreateRiskAnalysisProcess(owner, scopeID, src, dst)
	threatID := factory.CreateRiskAnalysisThreat(owner, scopeID, processID)
	scenarioID := factory.CreateRiskAnalysisScenario(owner, scopeID)

	factory.LinkRiskAnalysisScenarioThreat(owner, scenarioID, threatID)

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
				... on RiskAnalysisScenario {
					threats(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Node.Threats.TotalCount)

	_, err = owner.Do(`
		mutation($input: UnlinkRiskAnalysisScenarioThreatInput!) {
			unlinkRiskAnalysisScenarioThreat(input: $input) { riskAnalysisScenario { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisScenarioId": scenarioID,
			"threatId":               threatID,
		},
	})
	require.NoError(t, err)

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisScenario {
					threats(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Node.Threats.TotalCount)
}

func TestRiskAnalysisScenario_LinkUnlinkRisk(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	riskID := factory.CreateRisk(owner)
	scenarioID := factory.CreateRiskAnalysisScenario(owner, scopeID)

	factory.LinkRiskAnalysisScenarioRisk(owner, scenarioID, riskID)

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
				... on RiskAnalysisScenario {
					risks(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Node.Risks.TotalCount)

	_, err = owner.Do(`
		mutation($input: UnlinkRiskAnalysisScenarioRiskInput!) {
			unlinkRiskAnalysisScenarioRisk(input: $input) { riskAnalysisScenario { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisScenarioId": scenarioID,
			"riskId":                 riskID,
		},
	})
	require.NoError(t, err)

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisScenario {
					risks(first: 10) { totalCount }
				}
			}
		}
	`, map[string]any{"id": scenarioID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Node.Risks.TotalCount)
}

func TestRiskAnalysisBoundary_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	parentID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "External Zone"})

	var result struct {
		CreateRiskAnalysisBoundary struct {
			RiskAnalysisBoundaryEdge struct {
				Node struct {
					ID               string  `json:"id"`
					Name             string  `json:"name"`
					ParentBoundaryID *string `json:"parentBoundaryId"`
				} `json:"node"`
			} `json:"riskAnalysisBoundaryEdge"`
		} `json:"createRiskAnalysisBoundary"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAnalysisBoundaryInput!) {
			createRiskAnalysisBoundary(input: $input) {
				riskAnalysisBoundaryEdge { node { id name parentBoundaryId } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisDiagramId": scopeID,
			"parentBoundaryId":      parentID,
			"name":                  "Internal Network",
		},
	}, &result)

	require.NoError(t, err)

	node := result.CreateRiskAnalysisBoundary.RiskAnalysisBoundaryEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "Internal Network", node.Name)
	require.NotNil(t, node.ParentBoundaryID)
	assert.Equal(t, parentID, *node.ParentBoundaryID)
}

func TestRiskAnalysisBoundary_ListViaScope(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Zone A"})
	factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Zone B"})

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
				... on RiskAnalysisDiagram {
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

func TestRiskAnalysisBoundary_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	parentID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Parent"})
	boundaryID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Original"})

	const mutation = `
		mutation($input: UpdateRiskAnalysisBoundaryInput!) {
			updateRiskAnalysisBoundary(input: $input) {
				riskAnalysisBoundary { id name parentBoundaryId }
			}
		}
	`

	var result struct {
		UpdateRiskAnalysisBoundary struct {
			RiskAnalysisBoundary struct {
				ID               string  `json:"id"`
				Name             string  `json:"name"`
				ParentBoundaryID *string `json:"parentBoundaryId"`
			} `json:"riskAnalysisBoundary"`
		} `json:"updateRiskAnalysisBoundary"`
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
	assert.Equal(t, "Renamed", result.UpdateRiskAnalysisBoundary.RiskAnalysisBoundary.Name)
	require.NotNil(t, result.UpdateRiskAnalysisBoundary.RiskAnalysisBoundary.ParentBoundaryID)
	assert.Equal(t, parentID, *result.UpdateRiskAnalysisBoundary.RiskAnalysisBoundary.ParentBoundaryID)

	// Clear the parent (move back to the top level).
	err = owner.Execute(mutation, map[string]any{
		"input": map[string]any{
			"id":               boundaryID,
			"parentBoundaryId": nil,
		},
	}, &result)
	require.NoError(t, err)
	assert.Nil(t, result.UpdateRiskAnalysisBoundary.RiskAnalysisBoundary.ParentBoundaryID)
}

func TestRiskAnalysisBoundary_PreventCycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	a := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "A"})
	b := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "B", "parentBoundaryId": a})

	// B is nested under A, so nesting A under B would create a cycle.
	_, err := owner.Do(`
		mutation($input: UpdateRiskAnalysisBoundaryInput!) {
			updateRiskAnalysisBoundary(input: $input) {
				riskAnalysisBoundary { id }
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

func TestRiskAnalysisBoundary_Delete(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	parentID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Parent"})
	childID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Child", "parentBoundaryId": parentID})
	nodeID := factory.CreateRiskAnalysisNode(owner, scopeID, factory.Attrs{"name": "Member", "boundaryId": parentID})

	_, err := owner.Do(`
		mutation($input: DeleteRiskAnalysisBoundaryInput!) {
			deleteRiskAnalysisBoundary(input: $input) { deletedRiskAnalysisBoundaryId }
		}
	`, map[string]any{"input": map[string]any{"riskAnalysisBoundaryId": parentID}})
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
			child: node(id: $child) { ... on RiskAnalysisBoundary { parentBoundaryId } }
			member: node(id: $member) { ... on RiskAnalysisNode { boundaryId } }
		}
	`, map[string]any{"child": childID, "member": nodeID}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Child)
	assert.Nil(t, result.Child.ParentBoundaryID)
	require.NotNil(t, result.Member)
	assert.Nil(t, result.Member.BoundaryID)
}

func TestRiskAnalysisNode_WithBoundary(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner)
	scopeID := factory.CreateRiskAnalysisDiagram(owner, raID)
	boundaryID := factory.CreateRiskAnalysisBoundary(owner, scopeID)

	var createResult struct {
		CreateRiskAnalysisNode struct {
			RiskAnalysisNodeEdge struct {
				Node struct {
					ID         string  `json:"id"`
					BoundaryID *string `json:"boundaryId"`
				} `json:"node"`
			} `json:"riskAnalysisNodeEdge"`
		} `json:"createRiskAnalysisNode"`
	}

	err := owner.Execute(`
		mutation($input: CreateRiskAnalysisNodeInput!) {
			createRiskAnalysisNode(input: $input) {
				riskAnalysisNodeEdge { node { id boundaryId } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskAnalysisDiagramId": scopeID,
			"nodeType":              "ASSET",
			"name":                  "Member",
			"boundaryId":            boundaryID,
		},
	}, &createResult)
	require.NoError(t, err)

	created := createResult.CreateRiskAnalysisNode.RiskAnalysisNodeEdge.Node
	require.NotNil(t, created.BoundaryID)
	assert.Equal(t, boundaryID, *created.BoundaryID)

	// Clearing boundaryId moves the node back to the top level.
	var updateResult struct {
		UpdateRiskAnalysisNode struct {
			RiskAnalysisNode struct {
				BoundaryID *string `json:"boundaryId"`
			} `json:"riskAnalysisNode"`
		} `json:"updateRiskAnalysisNode"`
	}

	err = owner.Execute(`
		mutation($input: UpdateRiskAnalysisNodeInput!) {
			updateRiskAnalysisNode(input: $input) {
				riskAnalysisNode { boundaryId }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":         created.ID,
			"boundaryId": nil,
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Nil(t, updateResult.UpdateRiskAnalysisNode.RiskAnalysisNode.BoundaryID)
}
