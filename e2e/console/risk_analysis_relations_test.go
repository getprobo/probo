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

type riskAnalysisRelationGraph struct {
	orgID         string
	assessmentID  string
	scopeID       string
	boundaryID    string
	sourceNodeID  string
	targetNodeID  string
	boundedNodeID string
	processID     string
	threatID      string
	scenarioID    string
	riskID        string
}

func populateRiskAnalysisRelationGraph(
	t *testing.T,
	owner *testutil.Client,
) riskAnalysisRelationGraph {
	t.Helper()

	g := riskAnalysisRelationGraph{
		orgID: owner.GetOrganizationID().String(),
	}
	g.assessmentID = factory.CreateRiskAnalysis(
		owner,
		factory.Attrs{"name": factory.SafeName("Relations assessment")},
	)
	g.scopeID = factory.CreateRiskAnalysisDiagram(
		owner,
		g.assessmentID,
		factory.Attrs{"name": factory.SafeName("Relations scope")},
	)
	g.boundaryID = factory.CreateRiskAnalysisBoundary(
		owner,
		g.scopeID,
		factory.Attrs{"name": factory.SafeName("Relations boundary")},
	)
	g.sourceNodeID = factory.CreateRiskAnalysisNode(
		owner,
		g.scopeID,
		factory.Attrs{"nodeType": "ENTITY", "name": factory.SafeName("Relations source")},
	)
	g.targetNodeID = factory.CreateRiskAnalysisNode(
		owner,
		g.scopeID,
		factory.Attrs{"nodeType": "ASSET", "name": factory.SafeName("Relations target")},
	)
	g.boundedNodeID = factory.CreateRiskAnalysisNode(
		owner,
		g.scopeID,
		factory.Attrs{
			"nodeType":   "DATA",
			"name":       factory.SafeName("Relations bounded"),
			"boundaryId": g.boundaryID,
		},
	)
	g.processID = factory.CreateRiskAnalysisProcess(
		owner,
		g.scopeID,
		g.sourceNodeID,
		g.targetNodeID,
		factory.Attrs{"name": factory.SafeName("Relations process")},
	)
	g.threatID = factory.CreateRiskAnalysisThreat(
		owner,
		g.scopeID,
		g.processID,
		factory.Attrs{
			"name":     factory.SafeName("Relations threat"),
			"category": "Confidentiality",
		},
	)
	g.scenarioID = factory.CreateRiskAnalysisScenario(
		owner,
		g.scopeID,
		factory.Attrs{"name": factory.SafeName("Relations scenario")},
	)
	g.riskID = factory.CreateRisk(owner)
	factory.LinkRiskAnalysisScenarioThreat(owner, g.scenarioID, g.threatID)
	factory.LinkRiskAnalysisScenarioRisk(owner, g.scenarioID, g.riskID)

	return g
}

func TestRiskAnalysis_Relations(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	graph := populateRiskAnalysisRelationGraph(t, owner)

	cases := []struct {
		name string
		run  func(*testing.T, *testutil.Client, riskAnalysisRelationGraph)
	}{
		{name: "AssessmentOrganizationScopesPermission", run: riskRelationsAssessmentOrganizationScopesPermission},
		{name: "ScopeNodesBoundariesProcessesThreatsScenarios", run: riskRelationsScopeChildren},
		{name: "ScenarioScopeThreatsRisks", run: riskRelationsScenarioLinks},
		{name: "NodeBoundaryAndProcessEndpoints", run: riskRelationsNodeAndProcessFields},
		{name: "OrganizationReverseConnections", run: riskRelationsOrganizationReverse},
		{name: "RiskScenariosReverse", run: riskRelationsRiskScenariosReverse},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				tc.run(t, owner.ForTest(t), graph)
			},
		)
	}
}

func riskRelationsAssessmentOrganizationScopesPermission(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var result struct {
		Node struct {
			ID           string `json:"id"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			Scopes struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Cursor string `json:"cursor"`
					Node   struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage     bool `json:"hasNextPage"`
					HasPreviousPage bool `json:"hasPreviousPage"`
				} `json:"pageInfo"`
			} `json:"diagrams"`
			CanUpdate bool `json:"canUpdate"`
			CanDelete bool `json:"canDelete"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysis {
					id
					organization { id }
					diagrams(first: 10) {
						totalCount
						edges {
							cursor
							node { id name }
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
						}
					}
					canUpdate: permission(action: "risk-management:risk-analysis:update")
					canDelete: permission(action: "risk-management:risk-analysis:delete")
				}
			}
		}
	`, map[string]any{"id": g.assessmentID}, &result)
	require.NoError(t, err)

	assert.Equal(t, g.assessmentID, result.Node.ID)
	assert.Equal(t, g.orgID, result.Node.Organization.ID)
	assert.Equal(t, 1, result.Node.Scopes.TotalCount)
	require.Len(t, result.Node.Scopes.Edges, 1)
	assert.Equal(t, g.scopeID, result.Node.Scopes.Edges[0].Node.ID)
	assert.NotEmpty(t, result.Node.Scopes.Edges[0].Cursor)
	assert.False(t, result.Node.Scopes.PageInfo.HasNextPage)
	assert.False(t, result.Node.Scopes.PageInfo.HasPreviousPage)
	assert.True(t, result.Node.CanUpdate)
	assert.True(t, result.Node.CanDelete)
}

func riskRelationsScopeChildren(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var result struct {
		Node struct {
			Nodes struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"nodes"`
			Boundaries struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"boundaries"`
			Processes struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"processes"`
			Threats struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"threats"`
			Scenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"scenarios"`
			MermaidChart string `json:"mermaidChart"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisDiagram {
					nodes(first: 10) {
						totalCount
						edges { node { id } }
						pageInfo { hasNextPage }
					}
					boundaries(first: 10) {
						totalCount
						edges { node { id } }
					}
					processes(first: 10) {
						totalCount
						edges { node { id } }
					}
					threats(first: 10) {
						totalCount
						edges { node { id } }
					}
					scenarios(first: 10) {
						totalCount
						edges { node { id } }
					}
					mermaidChart
				}
			}
		}
	`, map[string]any{"id": g.scopeID}, &result)
	require.NoError(t, err)

	assert.Equal(t, 3, result.Node.Nodes.TotalCount)
	assert.False(t, result.Node.Nodes.PageInfo.HasNextPage)
	nodeIDs := collectRelationNodeIDs(result.Node.Nodes.Edges)
	assert.Equal(t, map[string]bool{
		g.sourceNodeID:  true,
		g.targetNodeID:  true,
		g.boundedNodeID: true,
	}, nodeIDs)

	assert.Equal(t, 1, result.Node.Boundaries.TotalCount)
	require.Len(t, result.Node.Boundaries.Edges, 1)
	assert.Equal(t, g.boundaryID, result.Node.Boundaries.Edges[0].Node.ID)

	assert.Equal(t, 1, result.Node.Processes.TotalCount)
	require.Len(t, result.Node.Processes.Edges, 1)
	assert.Equal(t, g.processID, result.Node.Processes.Edges[0].Node.ID)

	assert.Equal(t, 1, result.Node.Threats.TotalCount)
	require.Len(t, result.Node.Threats.Edges, 1)
	assert.Equal(t, g.threatID, result.Node.Threats.Edges[0].Node.ID)

	assert.Equal(t, 1, result.Node.Scenarios.TotalCount)
	require.Len(t, result.Node.Scenarios.Edges, 1)
	assert.Equal(t, g.scenarioID, result.Node.Scenarios.Edges[0].Node.ID)

	assert.NotEmpty(t, result.Node.MermaidChart)
	assert.Contains(t, result.Node.MermaidChart, "flowchart")
}

func riskRelationsScenarioLinks(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var result struct {
		Node struct {
			RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
			Diagram               struct {
				ID string `json:"id"`
			} `json:"diagram"`
			Threats struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID        string `json:"id"`
						ProcessID string `json:"processId"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasPreviousPage bool `json:"hasPreviousPage"`
				} `json:"pageInfo"`
			} `json:"threats"`
			Risks struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"risks"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisScenario {
					riskAnalysisDiagramId
					diagram { id }
					threats(first: 10) {
						totalCount
						edges { node { id processId } }
						pageInfo { hasPreviousPage }
					}
					risks(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.scenarioID}, &result)
	require.NoError(t, err)

	assert.Equal(t, g.scopeID, result.Node.RiskAnalysisDiagramID)
	assert.Equal(t, g.scopeID, result.Node.Diagram.ID)
	assert.Equal(t, 1, result.Node.Threats.TotalCount)
	require.Len(t, result.Node.Threats.Edges, 1)
	assert.Equal(t, g.threatID, result.Node.Threats.Edges[0].Node.ID)
	assert.Equal(t, g.processID, result.Node.Threats.Edges[0].Node.ProcessID)
	assert.False(t, result.Node.Threats.PageInfo.HasPreviousPage)
	assert.Equal(t, 1, result.Node.Risks.TotalCount)
	require.Len(t, result.Node.Risks.Edges, 1)
	assert.Equal(t, g.riskID, result.Node.Risks.Edges[0].Node.ID)
}

func riskRelationsNodeAndProcessFields(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var boundedResult struct {
		Node struct {
			RiskAnalysisDiagramID string  `json:"riskAnalysisDiagramId"`
			BoundaryID            *string `json:"boundaryId"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisNode {
					riskAnalysisDiagramId
					boundaryId
				}
			}
		}
	`, map[string]any{"id": g.boundedNodeID}, &boundedResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, boundedResult.Node.RiskAnalysisDiagramID)
	require.NotNil(t, boundedResult.Node.BoundaryID)
	assert.Equal(t, g.boundaryID, *boundedResult.Node.BoundaryID)

	var processResult struct {
		Node struct {
			RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
			SourceNodeID          string `json:"sourceNodeId"`
			TargetNodeID          string `json:"targetNodeId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisProcess {
					riskAnalysisDiagramId
					sourceNodeId
					targetNodeId
				}
			}
		}
	`, map[string]any{"id": g.processID}, &processResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, processResult.Node.RiskAnalysisDiagramID)
	assert.Equal(t, g.sourceNodeID, processResult.Node.SourceNodeID)
	assert.Equal(t, g.targetNodeID, processResult.Node.TargetNodeID)

	var threatResult struct {
		Node struct {
			RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
			ProcessID             string `json:"processId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisThreat {
					riskAnalysisDiagramId
					processId
				}
			}
		}
	`, map[string]any{"id": g.threatID}, &threatResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, threatResult.Node.RiskAnalysisDiagramID)
	assert.Equal(t, g.processID, threatResult.Node.ProcessID)

	var boundaryResult struct {
		Node struct {
			RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysisBoundary {
					riskAnalysisDiagramId
				}
			}
		}
	`, map[string]any{"id": g.boundaryID}, &boundaryResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, boundaryResult.Node.RiskAnalysisDiagramID)
}

func riskRelationsOrganizationReverse(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var result struct {
		Node struct {
			RiskAnalyses struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"riskAnalyses"`
			RiskAnalysisScenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID                    string `json:"id"`
						RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"riskAnalysisScenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					riskAnalyses(first: 50) {
						totalCount
						edges { node { id } }
						pageInfo { hasNextPage }
					}
					riskAnalysisScenarios(first: 50) {
						totalCount
						edges {
							node {
								id
								riskAnalysisDiagramId
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.RiskAnalyses.TotalCount, 1)
	assessmentIDs := collectRelationNodeIDs(result.Node.RiskAnalyses.Edges)
	assert.True(t, assessmentIDs[g.assessmentID], "organization riskAnalyses must include created assessment")

	assert.GreaterOrEqual(t, result.Node.RiskAnalysisScenarios.TotalCount, 1)

	var scenarioScopeID string

	for _, edge := range result.Node.RiskAnalysisScenarios.Edges {
		if edge.Node.ID != g.scenarioID {
			continue
		}

		scenarioScopeID = edge.Node.RiskAnalysisDiagramID

		break
	}

	assert.Equal(t, g.scopeID, scenarioScopeID)
}

func riskRelationsRiskScenariosReverse(
	t *testing.T,
	owner *testutil.Client,
	g riskAnalysisRelationGraph,
) {
	var result struct {
		Node struct {
			Scenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Cursor string `json:"cursor"`
					Node   struct {
						ID                    string `json:"id"`
						RiskAnalysisDiagramID string `json:"riskAnalysisDiagramId"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"scenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					scenarios(first: 10) {
						totalCount
						edges {
							cursor
							node {
								id
								riskAnalysisDiagramId
							}
						}
						pageInfo { hasNextPage }
					}
				}
			}
		}
	`, map[string]any{"id": g.riskID}, &result)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Node.Scenarios.TotalCount)
	require.Len(t, result.Node.Scenarios.Edges, 1)
	assert.Equal(t, g.scenarioID, result.Node.Scenarios.Edges[0].Node.ID)
	assert.Equal(t, g.scopeID, result.Node.Scenarios.Edges[0].Node.RiskAnalysisDiagramID)
	assert.NotEmpty(t, result.Node.Scenarios.Edges[0].Cursor)
	assert.False(t, result.Node.Scenarios.PageInfo.HasNextPage)
}

func collectRelationNodeIDs(
	edges []struct {
		Node struct {
			ID string `json:"id"`
		} `json:"node"`
	},
) map[string]bool {
	ids := make(map[string]bool, len(edges))
	for _, edge := range edges {
		ids[edge.Node.ID] = true
	}

	return ids
}
