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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const forkRiskAnalysisMutation = `
	mutation($input: ForkRiskAnalysisInput!) {
		forkRiskAnalysis(input: $input) {
			riskAnalysisEdge {
				node {
					id
					name
					description
					period { start end }
					matrixSize { rows cols }
				}
			}
		}
	}
`

type forkRiskAnalysisResult struct {
	ForkRiskAnalysis struct {
		RiskAnalysisEdge struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Period      *struct {
					Start *string `json:"start"`
					End   *string `json:"end"`
				} `json:"period"`
				MatrixSize struct {
					Rows int `json:"rows"`
					Cols int `json:"cols"`
				} `json:"matrixSize"`
			} `json:"node"`
		} `json:"riskAnalysisEdge"`
	} `json:"forkRiskAnalysis"`
}

func parseDatetime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339Nano, value)
	require.NoError(t, err)

	return parsed
}

func TestRiskAnalysis_Fork(t *testing.T) {
	t.Parallel()

	t.Run("copies diagrams treatment plans and relations", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		graph := populateRiskAnalysisRelationGraph(t, owner)
		childBoundaryID := factory.CreateRiskAnalysisBoundary(
			owner,
			graph.scopeID,
			factory.Attrs{
				"name":             factory.SafeName("Child boundary"),
				"parentBoundaryId": graph.boundaryID,
			},
		)
		measureID := factory.CreateMeasure(owner)
		tpID := factory.CreateTreatmentPlan(
			owner,
			graph.riskID,
			graph.assessmentID,
			factory.Attrs{"inherentLikelihood": 3, "inherentImpact": 4},
		)
		factory.LinkTreatmentPlanMeasure(owner, tpID, measureID)

		var sourcePlan struct {
			Node struct {
				CreatedAt string `json:"createdAt"`
			} `json:"node"`
		}

		err := owner.Execute(
			`query($id: ID!) { node(id: $id) { ... on TreatmentPlan { createdAt } } }`,
			map[string]any{"id": tpID},
			&sourcePlan,
		)
		require.NoError(t, err)

		sourceCreatedAt := parseDatetime(t, sourcePlan.Node.CreatedAt)

		var result forkRiskAnalysisResult

		err = owner.Execute(
			forkRiskAnalysisMutation,
			map[string]any{
				"input": map[string]any{
					"riskAnalysisId": graph.assessmentID,
					"name":           "Forked analysis",
					"description":    "Forked description",
				},
			},
			&result,
		)
		require.NoError(t, err)

		forkedID := result.ForkRiskAnalysis.RiskAnalysisEdge.Node.ID
		require.NotEmpty(t, forkedID)
		assert.NotEqual(t, graph.assessmentID, forkedID)
		assert.Equal(t, "Forked analysis", result.ForkRiskAnalysis.RiskAnalysisEdge.Node.Name)
		require.NotNil(t, result.ForkRiskAnalysis.RiskAnalysisEdge.Node.Description)
		assert.Equal(t, "Forked description", *result.ForkRiskAnalysis.RiskAnalysisEdge.Node.Description)
		assert.Nil(t, result.ForkRiskAnalysis.RiskAnalysisEdge.Node.Period)
		assert.Equal(t, 5, result.ForkRiskAnalysis.RiskAnalysisEdge.Node.MatrixSize.Rows)
		assert.Equal(t, 5, result.ForkRiskAnalysis.RiskAnalysisEdge.Node.MatrixSize.Cols)

		var copied struct {
			Node struct {
				Diagrams struct {
					Edges []struct {
						Node struct {
							ID         string `json:"id"`
							Name       string `json:"name"`
							Boundaries struct {
								Edges []struct {
									Node struct {
										ID               string  `json:"id"`
										Name             string  `json:"name"`
										ParentBoundaryID *string `json:"parentBoundaryId"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"boundaries"`
							Nodes struct {
								Edges []struct {
									Node struct {
										ID         string  `json:"id"`
										Name       string  `json:"name"`
										BoundaryID *string `json:"boundaryId"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"nodes"`
							Processes struct {
								Edges []struct {
									Node struct {
										ID           string `json:"id"`
										Name         string `json:"name"`
										SourceNodeID string `json:"sourceNodeId"`
										TargetNodeID string `json:"targetNodeId"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"processes"`
							Threats struct {
								Edges []struct {
									Node struct {
										ID        string `json:"id"`
										Name      string `json:"name"`
										ProcessID string `json:"processId"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"threats"`
							Scenarios struct {
								Edges []struct {
									Node struct {
										ID      string `json:"id"`
										Name    string `json:"name"`
										Threats struct {
											Edges []struct {
												Node struct {
													ID string `json:"id"`
												} `json:"node"`
											} `json:"edges"`
										} `json:"threats"`
										Risks struct {
											Edges []struct {
												Node struct {
													ID string `json:"id"`
												} `json:"node"`
											} `json:"edges"`
										} `json:"risks"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"scenarios"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"diagrams"`
				TreatmentPlans struct {
					Edges []struct {
						Node struct {
							ID                 string `json:"id"`
							CreatedAt          string `json:"createdAt"`
							InherentLikelihood int    `json:"inherentLikelihood"`
							InherentImpact     int    `json:"inherentImpact"`
							Risk               struct {
								ID string `json:"id"`
							} `json:"risk"`
							Measures struct {
								Edges []struct {
									Node struct {
										ID string `json:"id"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"measures"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"treatmentPlans"`
				AsOfPlans struct {
					Edges []struct {
						Node struct {
							ID        string `json:"id"`
							CreatedAt string `json:"createdAt"`
							Measures  struct {
								Edges []struct {
									Node struct {
										ID string `json:"id"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"measures"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"asOfPlans"`
			} `json:"node"`
		}

		err = owner.Execute(`
			query($id: ID!, $asOf: Datetime) {
				node(id: $id) {
					... on RiskAnalysis {
						diagrams(first: 10) {
							edges {
								node {
									id
									name
									boundaries(first: 10) {
										edges { node { id name parentBoundaryId } }
									}
									nodes(first: 10) {
										edges { node { id name boundaryId } }
									}
									processes(first: 10) {
										edges { node { id name sourceNodeId targetNodeId } }
									}
									threats(first: 10) {
										edges { node { id name processId } }
									}
									scenarios(first: 10) {
										edges {
											node {
												id
												name
												threats(first: 10) { edges { node { id } } }
												risks(first: 10) { edges { node { id } } }
											}
										}
									}
								}
							}
						}
						treatmentPlans(first: 10) {
							edges {
								node {
									id
									createdAt
									inherentLikelihood
									inherentImpact
									risk { id }
									measures(first: 10) { edges { node { id } } }
								}
							}
						}
						asOfPlans: treatmentPlans(first: 10, asOf: $asOf) {
							edges {
								node {
									id
									createdAt
									measures(first: 10) { edges { node { id } } }
								}
							}
						}
					}
				}
			}
		`, map[string]any{
			"id":   forkedID,
			"asOf": time.Now().UTC().Format(time.RFC3339Nano),
		}, &copied)
		require.NoError(t, err)

		require.Len(t, copied.Node.Diagrams.Edges, 1)
		diagram := copied.Node.Diagrams.Edges[0].Node
		assert.NotEqual(t, graph.scopeID, diagram.ID)

		require.Len(t, diagram.Boundaries.Edges, 2)

		boundariesByName := map[string]struct {
			ID               string
			ParentBoundaryID *string
		}{}

		for _, edge := range diagram.Boundaries.Edges {
			assert.NotEqual(t, graph.boundaryID, edge.Node.ID)
			assert.NotEqual(t, childBoundaryID, edge.Node.ID)

			boundariesByName[edge.Node.Name] = struct {
				ID               string
				ParentBoundaryID *string
			}{ID: edge.Node.ID, ParentBoundaryID: edge.Node.ParentBoundaryID}
		}

		var parentName, childName string

		for name, boundary := range boundariesByName {
			if boundary.ParentBoundaryID == nil {
				parentName = name
			} else {
				childName = name
			}
		}

		require.NotEmpty(t, parentName)
		require.NotEmpty(t, childName)
		require.NotNil(t, boundariesByName[childName].ParentBoundaryID)
		assert.Equal(t, boundariesByName[parentName].ID, *boundariesByName[childName].ParentBoundaryID)

		require.Len(t, diagram.Nodes.Edges, 3)

		nodesByID := map[string]string{}

		for _, edge := range diagram.Nodes.Edges {
			assert.NotEqual(t, graph.sourceNodeID, edge.Node.ID)
			assert.NotEqual(t, graph.targetNodeID, edge.Node.ID)
			assert.NotEqual(t, graph.boundedNodeID, edge.Node.ID)

			nodesByID[edge.Node.ID] = edge.Node.Name

			if edge.Node.BoundaryID != nil {
				assert.Equal(t, boundariesByName[parentName].ID, *edge.Node.BoundaryID)
			}
		}

		require.Len(t, diagram.Processes.Edges, 1)
		process := diagram.Processes.Edges[0].Node
		assert.NotEqual(t, graph.processID, process.ID)
		_, ok := nodesByID[process.SourceNodeID]
		assert.True(t, ok)
		_, ok = nodesByID[process.TargetNodeID]
		assert.True(t, ok)

		require.Len(t, diagram.Threats.Edges, 1)
		threat := diagram.Threats.Edges[0].Node
		assert.NotEqual(t, graph.threatID, threat.ID)
		assert.Equal(t, process.ID, threat.ProcessID)

		require.Len(t, diagram.Scenarios.Edges, 1)
		scenario := diagram.Scenarios.Edges[0].Node
		assert.NotEqual(t, graph.scenarioID, scenario.ID)
		require.Len(t, scenario.Threats.Edges, 1)
		assert.Equal(t, threat.ID, scenario.Threats.Edges[0].Node.ID)
		require.Len(t, scenario.Risks.Edges, 1)
		assert.Equal(t, graph.riskID, scenario.Risks.Edges[0].Node.ID)

		require.Len(t, copied.Node.TreatmentPlans.Edges, 1)
		copiedTP := copied.Node.TreatmentPlans.Edges[0].Node
		assert.NotEqual(t, tpID, copiedTP.ID)
		assert.Equal(t, graph.riskID, copiedTP.Risk.ID)
		assert.Equal(t, 3, copiedTP.InherentLikelihood)
		assert.Equal(t, 4, copiedTP.InherentImpact)

		copiedCreatedAt := parseDatetime(t, copiedTP.CreatedAt)
		assert.False(t, copiedCreatedAt.Before(sourceCreatedAt))
		require.Len(t, copiedTP.Measures.Edges, 1)
		assert.Equal(t, measureID, copiedTP.Measures.Edges[0].Node.ID)
		require.Len(t, copied.Node.AsOfPlans.Edges, 1)
		assert.Equal(t, copiedTP.CreatedAt, copied.Node.AsOfPlans.Edges[0].Node.CreatedAt)
		require.Len(t, copied.Node.AsOfPlans.Edges[0].Node.Measures.Edges, 1)
		assert.Equal(t, measureID, copied.Node.AsOfPlans.Edges[0].Node.Measures.Edges[0].Node.ID)

		var beforeFork struct {
			Node struct {
				TreatmentPlans struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"treatmentPlans"`
			} `json:"node"`
		}

		err = owner.Execute(`
			query($id: ID!, $asOf: Datetime) {
				node(id: $id) {
					... on RiskAnalysis {
						treatmentPlans(first: 10, asOf: $asOf) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{
			"id":   forkedID,
			"asOf": sourceCreatedAt.Add(-time.Second).UTC().Format(time.RFC3339Nano),
		}, &beforeFork)
		require.NoError(t, err)
		assert.Empty(t, beforeFork.Node.TreatmentPlans.Edges)
	})

	t.Run("copies scenario threats linked across diagrams", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		sourceID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Cross-diagram source"})

		scenarioDiagramID := factory.CreateRiskAnalysisDiagram(
			owner,
			sourceID,
			factory.Attrs{"name": "Scenario diagram"},
		)
		threatDiagramID := factory.CreateRiskAnalysisDiagram(
			owner,
			sourceID,
			factory.Attrs{"name": "Threat diagram"},
		)
		sourceNodeID := factory.CreateRiskAnalysisNode(
			owner,
			threatDiagramID,
			factory.Attrs{"nodeType": "ENTITY", "name": "Source node"},
		)
		targetNodeID := factory.CreateRiskAnalysisNode(
			owner,
			threatDiagramID,
			factory.Attrs{"nodeType": "ASSET", "name": "Target node"},
		)
		processID := factory.CreateRiskAnalysisProcess(
			owner,
			threatDiagramID,
			sourceNodeID,
			targetNodeID,
			factory.Attrs{"name": "Process"},
		)
		threatID := factory.CreateRiskAnalysisThreat(
			owner,
			threatDiagramID,
			processID,
			factory.Attrs{"name": "Later-diagram threat", "category": "Confidentiality"},
		)
		scenarioID := factory.CreateRiskAnalysisScenario(
			owner,
			scenarioDiagramID,
			factory.Attrs{"name": "Earlier-diagram scenario"},
		)
		factory.LinkRiskAnalysisScenarioThreat(owner, scenarioID, threatID)

		var result forkRiskAnalysisResult

		err := owner.Execute(
			forkRiskAnalysisMutation,
			map[string]any{
				"input": map[string]any{
					"riskAnalysisId": sourceID,
					"name":           "Cross-diagram fork",
				},
			},
			&result,
		)
		require.NoError(t, err)

		forkedID := result.ForkRiskAnalysis.RiskAnalysisEdge.Node.ID
		require.NotEmpty(t, forkedID)

		var copied struct {
			Node struct {
				Diagrams struct {
					Edges []struct {
						Node struct {
							ID      string `json:"id"`
							Name    string `json:"name"`
							Threats struct {
								Edges []struct {
									Node struct {
										ID string `json:"id"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"threats"`
							Scenarios struct {
								Edges []struct {
									Node struct {
										ID      string `json:"id"`
										Name    string `json:"name"`
										Threats struct {
											Edges []struct {
												Node struct {
													ID string `json:"id"`
												} `json:"node"`
											} `json:"edges"`
										} `json:"threats"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"scenarios"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"diagrams"`
			} `json:"node"`
		}

		err = owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on RiskAnalysis {
						diagrams(first: 10) {
							edges {
								node {
									id
									name
									threats(first: 10) { edges { node { id } } }
									scenarios(first: 10) {
										edges {
											node {
												id
												name
												threats(first: 10) { edges { node { id } } }
											}
										}
									}
								}
							}
						}
					}
				}
			}
		`, map[string]any{"id": forkedID}, &copied)
		require.NoError(t, err)
		require.Len(t, copied.Node.Diagrams.Edges, 2)

		diagramsByName := map[string]int{}

		for i, edge := range copied.Node.Diagrams.Edges {
			assert.NotEqual(t, scenarioDiagramID, edge.Node.ID)
			assert.NotEqual(t, threatDiagramID, edge.Node.ID)
			diagramsByName[edge.Node.Name] = i
		}

		scenarioIdx, ok := diagramsByName["Scenario diagram"]
		require.True(t, ok)
		threatIdx, ok := diagramsByName["Threat diagram"]
		require.True(t, ok)

		scenarioDiagram := copied.Node.Diagrams.Edges[scenarioIdx].Node
		threatDiagram := copied.Node.Diagrams.Edges[threatIdx].Node

		require.Len(t, threatDiagram.Threats.Edges, 1)
		require.Len(t, scenarioDiagram.Scenarios.Edges, 1)
		assert.Equal(t, "Earlier-diagram scenario", scenarioDiagram.Scenarios.Edges[0].Node.Name)
		require.Len(t, scenarioDiagram.Scenarios.Edges[0].Node.Threats.Edges, 1)
		assert.Equal(
			t,
			threatDiagram.Threats.Edges[0].Node.ID,
			scenarioDiagram.Scenarios.Edges[0].Node.Threats.Edges[0].Node.ID,
		)
	})

	t.Run("copies matrix size and uses supplied period", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		var created struct {
			CreateRiskAnalysis struct {
				RiskAnalysisEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"riskAnalysisEdge"`
			} `json:"createRiskAnalysis"`
		}

		err := owner.Execute(`
			mutation($input: CreateRiskAnalysisInput!) {
				createRiskAnalysis(input: $input) {
					riskAnalysisEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Source with period",
				"matrixSize":     map[string]any{"rows": 3, "cols": 3},
				"period": map[string]any{
					"start": "2025-01-01T00:00:00Z",
					"end":   "2025-12-31T00:00:00Z",
				},
			},
		}, &created)
		require.NoError(t, err)

		sourceID := created.CreateRiskAnalysis.RiskAnalysisEdge.Node.ID

		var result forkRiskAnalysisResult

		err = owner.Execute(
			forkRiskAnalysisMutation,
			map[string]any{
				"input": map[string]any{
					"riskAnalysisId": sourceID,
					"name":           "Forked with new period",
					"period": map[string]any{
						"start": "2026-01-01T00:00:00Z",
						"end":   "2026-12-31T00:00:00Z",
					},
				},
			},
			&result,
		)
		require.NoError(t, err)

		node := result.ForkRiskAnalysis.RiskAnalysisEdge.Node
		assert.Equal(t, 3, node.MatrixSize.Rows)
		assert.Equal(t, 3, node.MatrixSize.Cols)
		require.NotNil(t, node.Period)
		require.NotNil(t, node.Period.Start)
		require.NotNil(t, node.Period.End)
		assert.Contains(t, *node.Period.Start, "2026-01-01")
		assert.Contains(t, *node.Period.End, "2026-12-31")
		assert.NotContains(t, *node.Period.Start, "2025")
	})

	t.Run("forks empty analysis", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		sourceID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Empty source"})

		var result forkRiskAnalysisResult

		err := owner.Execute(
			forkRiskAnalysisMutation,
			map[string]any{
				"input": map[string]any{
					"riskAnalysisId": sourceID,
					"name":           "Empty fork",
				},
			},
			&result,
		)
		require.NoError(t, err)
		assert.NotEqual(t, sourceID, result.ForkRiskAnalysis.RiskAnalysisEdge.Node.ID)
		assert.Equal(t, "Empty fork", result.ForkRiskAnalysis.RiskAnalysisEdge.Node.Name)
	})
}

func TestRiskAnalysis_Fork_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	sourceID := factory.CreateRiskAnalysis(owner)

	_, err := viewer.Do(
		forkRiskAnalysisMutation,
		map[string]any{
			"input": map[string]any{
				"riskAnalysisId": sourceID,
				"name":           "Viewer fork",
			},
		},
	)
	testutil.RequireForbiddenError(t, err, "viewer cannot fork risk analysis")
}

func TestRiskAnalysis_Fork_TenantIsolation(t *testing.T) {
	t.Parallel()

	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)
	sourceID := factory.CreateRiskAnalysis(owner1)

	_, err := owner2.Do(
		forkRiskAnalysisMutation,
		map[string]any{
			"input": map[string]any{
				"riskAnalysisId": sourceID,
				"name":           "Cross tenant fork",
			},
		},
	)
	testutil.RequireErrorCode(t, err, "NOT_FOUND", "cannot fork a risk analysis from another tenant")
}
