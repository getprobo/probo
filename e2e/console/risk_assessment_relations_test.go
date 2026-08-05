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

type riskAssessmentRelationGraph struct {
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

func populateRiskAssessmentRelationGraph(
	t *testing.T,
	owner *testutil.Client,
) riskAssessmentRelationGraph {
	t.Helper()

	g := riskAssessmentRelationGraph{
		orgID: owner.GetOrganizationID().String(),
	}
	g.assessmentID = factory.CreateRiskAssessment(
		owner,
		factory.Attrs{"name": factory.SafeName("Relations assessment")},
	)
	g.scopeID = factory.CreateRiskAssessmentScope(
		owner,
		g.assessmentID,
		factory.Attrs{"name": factory.SafeName("Relations scope")},
	)
	g.boundaryID = factory.CreateRiskAssessmentBoundary(
		owner,
		g.scopeID,
		factory.Attrs{"name": factory.SafeName("Relations boundary")},
	)
	g.sourceNodeID = factory.CreateRiskAssessmentNode(
		owner,
		g.scopeID,
		factory.Attrs{"nodeType": "ENTITY", "name": factory.SafeName("Relations source")},
	)
	g.targetNodeID = factory.CreateRiskAssessmentNode(
		owner,
		g.scopeID,
		factory.Attrs{"nodeType": "ASSET", "name": factory.SafeName("Relations target")},
	)
	g.boundedNodeID = factory.CreateRiskAssessmentNode(
		owner,
		g.scopeID,
		factory.Attrs{
			"nodeType":   "DATA",
			"name":       factory.SafeName("Relations bounded"),
			"boundaryId": g.boundaryID,
		},
	)
	g.processID = factory.CreateRiskAssessmentProcess(
		owner,
		g.scopeID,
		g.sourceNodeID,
		g.targetNodeID,
		factory.Attrs{"name": factory.SafeName("Relations process")},
	)
	g.threatID = factory.CreateRiskAssessmentThreat(
		owner,
		g.scopeID,
		g.processID,
		factory.Attrs{
			"name":     factory.SafeName("Relations threat"),
			"category": "Confidentiality",
		},
	)
	g.scenarioID = factory.CreateRiskAssessmentScenario(
		owner,
		g.scopeID,
		factory.Attrs{"name": factory.SafeName("Relations scenario")},
	)
	g.riskID = factory.CreateRisk(owner)
	factory.LinkRiskAssessmentScenarioThreat(owner, g.scenarioID, g.threatID)
	factory.LinkRiskAssessmentScenarioRisk(owner, g.scenarioID, g.riskID)

	return g
}

func TestRiskAssessment_Relations(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	graph := populateRiskAssessmentRelationGraph(t, owner)

	cases := []struct {
		name string
		run  func(*testing.T, *testutil.Client, riskAssessmentRelationGraph)
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
	g riskAssessmentRelationGraph,
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
			} `json:"scopes"`
			CanUpdate bool `json:"canUpdate"`
			CanDelete bool `json:"canDelete"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessment {
					id
					organization { id }
					scopes(first: 10) {
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
					canUpdate: permission(action: "core:risk-assessment:update")
					canDelete: permission(action: "core:risk-assessment:delete")
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
	g riskAssessmentRelationGraph,
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
				... on RiskAssessmentScope {
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
	g riskAssessmentRelationGraph,
) {
	var result struct {
		Node struct {
			RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
			Scope                 struct {
				ID string `json:"id"`
			} `json:"scope"`
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
				... on RiskAssessmentScenario {
					riskAssessmentScopeId
					scope { id }
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

	assert.Equal(t, g.scopeID, result.Node.RiskAssessmentScopeID)
	assert.Equal(t, g.scopeID, result.Node.Scope.ID)
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
	g riskAssessmentRelationGraph,
) {
	var boundedResult struct {
		Node struct {
			RiskAssessmentScopeID string  `json:"riskAssessmentScopeId"`
			BoundaryID            *string `json:"boundaryId"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentNode {
					riskAssessmentScopeId
					boundaryId
				}
			}
		}
	`, map[string]any{"id": g.boundedNodeID}, &boundedResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, boundedResult.Node.RiskAssessmentScopeID)
	require.NotNil(t, boundedResult.Node.BoundaryID)
	assert.Equal(t, g.boundaryID, *boundedResult.Node.BoundaryID)

	var processResult struct {
		Node struct {
			RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
			SourceNodeID          string `json:"sourceNodeId"`
			TargetNodeID          string `json:"targetNodeId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentProcess {
					riskAssessmentScopeId
					sourceNodeId
					targetNodeId
				}
			}
		}
	`, map[string]any{"id": g.processID}, &processResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, processResult.Node.RiskAssessmentScopeID)
	assert.Equal(t, g.sourceNodeID, processResult.Node.SourceNodeID)
	assert.Equal(t, g.targetNodeID, processResult.Node.TargetNodeID)

	var threatResult struct {
		Node struct {
			RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
			ProcessID             string `json:"processId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentThreat {
					riskAssessmentScopeId
					processId
				}
			}
		}
	`, map[string]any{"id": g.threatID}, &threatResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, threatResult.Node.RiskAssessmentScopeID)
	assert.Equal(t, g.processID, threatResult.Node.ProcessID)

	var boundaryResult struct {
		Node struct {
			RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAssessmentBoundary {
					riskAssessmentScopeId
				}
			}
		}
	`, map[string]any{"id": g.boundaryID}, &boundaryResult)
	require.NoError(t, err)
	assert.Equal(t, g.scopeID, boundaryResult.Node.RiskAssessmentScopeID)
}

func riskRelationsOrganizationReverse(
	t *testing.T,
	owner *testutil.Client,
	g riskAssessmentRelationGraph,
) {
	var result struct {
		Node struct {
			RiskAssessments struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"riskAssessments"`
			RiskAssessmentScenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID                    string `json:"id"`
						RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"riskAssessmentScenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					riskAssessments(first: 50) {
						totalCount
						edges { node { id } }
						pageInfo { hasNextPage }
					}
					riskAssessmentScenarios(first: 50) {
						totalCount
						edges {
							node {
								id
								riskAssessmentScopeId
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.RiskAssessments.TotalCount, 1)
	assessmentIDs := collectRelationNodeIDs(result.Node.RiskAssessments.Edges)
	assert.True(t, assessmentIDs[g.assessmentID], "organization riskAssessments must include created assessment")

	assert.GreaterOrEqual(t, result.Node.RiskAssessmentScenarios.TotalCount, 1)

	var scenarioScopeID string

	for _, edge := range result.Node.RiskAssessmentScenarios.Edges {
		if edge.Node.ID != g.scenarioID {
			continue
		}

		scenarioScopeID = edge.Node.RiskAssessmentScopeID

		break
	}

	assert.Equal(t, g.scopeID, scenarioScopeID)
}

func riskRelationsRiskScenariosReverse(
	t *testing.T,
	owner *testutil.Client,
	g riskAssessmentRelationGraph,
) {
	var result struct {
		Node struct {
			Scenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Cursor string `json:"cursor"`
					Node   struct {
						ID                    string `json:"id"`
						RiskAssessmentScopeID string `json:"riskAssessmentScopeId"`
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
								riskAssessmentScopeId
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
	assert.Equal(t, g.scopeID, result.Node.Scenarios.Edges[0].Node.RiskAssessmentScopeID)
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
