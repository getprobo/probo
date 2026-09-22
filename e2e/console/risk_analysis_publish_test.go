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

func TestRiskAnalysis_PublishRiskAnalysis(t *testing.T) {
	t.Parallel()

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			riskAnalysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{
				"name":        "Export Analysis",
				"description": "Export analysis description",
			})
			riskID := factory.CreateRisk(owner, factory.Attrs{"name": "Export Risk"})
			factory.LinkRiskToAnalysis(owner, riskID, riskAnalysisID)
			planID := factory.CreateTreatmentPlan(owner, riskID, riskAnalysisID)
			measureID := factory.CreateMeasure(owner, factory.Attrs{"name": "Export Measure"})
			factory.LinkTreatmentPlanMeasure(owner, planID, measureID)
			diagramID := factory.CreateRiskAnalysisDiagram(owner, riskAnalysisID, factory.Attrs{"name": "Export Diagram"})
			sourceNodeID := factory.CreateRiskAnalysisNode(owner, diagramID, factory.Attrs{"name": "Export Node"})
			targetNodeID := factory.CreateRiskAnalysisNode(owner, diagramID, factory.Attrs{"name": "Export Target"})
			processID := factory.CreateRiskAnalysisProcess(
				owner,
				diagramID,
				sourceNodeID,
				targetNodeID,
				factory.Attrs{"name": "Export Process"},
			)
			threatID := factory.CreateRiskAnalysisThreat(
				owner,
				diagramID,
				processID,
				factory.Attrs{"name": "Export Threat", "category": "Confidentiality"},
			)
			scenarioID := factory.CreateRiskAnalysisScenario(
				owner,
				diagramID,
				factory.Attrs{"name": "Export Scenario"},
			)
			factory.LinkRiskAnalysisScenarioRisk(owner, scenarioID, riskID)
			factory.LinkRiskAnalysisScenarioThreat(owner, scenarioID, threatID)

			const query = `
				mutation($input: PublishRiskAnalysisInput!) {
					publishRiskAnalysis(input: $input) {
						documentEdge {
							node {
								id
								writeMode
								status
							}
						}
						documentVersionEdge {
							node {
								id
								title
								documentType
								status
								major
								minor
								content
							}
						}
					}
				}
			`

			var result struct {
				PublishRiskAnalysis struct {
					DocumentEdge struct {
						Node struct {
							ID        string `json:"id"`
							WriteMode string `json:"writeMode"`
							Status    string `json:"status"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID           string `json:"id"`
							Title        string `json:"title"`
							DocumentType string `json:"documentType"`
							Status       string `json:"status"`
							Major        int    `json:"major"`
							Minor        int    `json:"minor"`
							Content      string `json:"content"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishRiskAnalysis"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"riskAnalysisId": riskAnalysisID,
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishRiskAnalysis.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishRiskAnalysis.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "Export Analysis", ver.Title)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
			assert.Contains(t, ver.Content, "Risk Matrix")
			assert.Contains(t, ver.Content, "Initial")
			assert.Contains(t, ver.Content, "Net")
			assert.Contains(t, ver.Content, "Residual")
			assert.Contains(t, ver.Content, "Impact")
			assert.Contains(t, ver.Content, "Likelihood")
			assert.Contains(t, ver.Content, "5×5")
			assert.Contains(t, ver.Content, "Export Analysis")
			assert.Contains(t, ver.Content, "Export analysis description")
			assert.Contains(t, ver.Content, "Export Risk")
			assert.Contains(t, ver.Content, "RSK-")
			assert.Contains(t, ver.Content, "Export Measure")
			assert.Contains(t, ver.Content, "Export Diagram")
			assert.Contains(t, ver.Content, "Export Node")
			assert.Contains(t, ver.Content, "Export Scenario")
			assert.Contains(t, ver.Content, "Export Threat")
		},
	)

	t.Run(
		"publish with approvers creates draft pending approval",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			riskAnalysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Approval Analysis"})

			const query = `
				mutation($input: PublishRiskAnalysisInput!) {
					publishRiskAnalysis(input: $input) {
						documentEdge {
							node { id writeMode }
						}
						documentVersionEdge {
							node { id status major }
						}
					}
				}
			`

			var result struct {
				PublishRiskAnalysis struct {
					DocumentEdge struct {
						Node struct {
							ID        string `json:"id"`
							WriteMode string `json:"writeMode"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
							Major  int    `json:"major"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishRiskAnalysis"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"riskAnalysisId": riskAnalysisID,
						"approverIds":    []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishRiskAnalysis.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)

			ver := result.PublishRiskAnalysis.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"second publish reuses document and bumps major version",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			riskAnalysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Reuse Analysis"})

			const query = `
				mutation($input: PublishRiskAnalysisInput!) {
					publishRiskAnalysis(input: $input) {
						documentEdge { node { id } }
						documentVersionEdge { node { id major } }
					}
				}
			`

			var r1, r2 struct {
				PublishRiskAnalysis struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID    string `json:"id"`
							Major int    `json:"major"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishRiskAnalysis"`
			}

			input := map[string]any{
				"input": map[string]any{
					"minor":          false,
					"riskAnalysisId": riskAnalysisID,
				},
			}

			err := owner.Execute(query, input, &r1)
			require.NoError(t, err)

			err = owner.Execute(query, input, &r2)
			require.NoError(t, err)

			assert.Equal(t,
				r1.PublishRiskAnalysis.DocumentEdge.Node.ID,
				r2.PublishRiskAnalysis.DocumentEdge.Node.ID,
				"should reuse same document",
			)
			assert.Equal(t, 1, r1.PublishRiskAnalysis.DocumentVersionEdge.Node.Major)
			assert.Equal(t, 2, r2.PublishRiskAnalysis.DocumentVersionEdge.Node.Major)
		},
	)

	t.Run(
		"risk analysis document links to published document",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			riskAnalysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Linked Analysis"})

			const publishQuery = `
				mutation($input: PublishRiskAnalysisInput!) {
					publishRiskAnalysis(input: $input) {
						documentEdge { node { id } }
						documentVersionEdge { node { id } }
					}
				}
			`

			var publishResult struct {
				PublishRiskAnalysis struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishRiskAnalysis"`
			}

			err := owner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"riskAnalysisId": riskAnalysisID,
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishRiskAnalysis.DocumentEdge.Node.ID

			const nodeQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on RiskAnalysis {
							id
							document { id }
						}
					}
				}
			`

			var nodeResult struct {
				Node struct {
					ID       string `json:"id"`
					Document *struct {
						ID string `json:"id"`
					} `json:"document"`
				} `json:"node"`
			}

			err = owner.Execute(
				nodeQuery,
				map[string]any{"id": riskAnalysisID},
				&nodeResult,
			)
			require.NoError(t, err)
			require.NotNil(t, nodeResult.Node.Document)
			assert.Equal(t, docID, nodeResult.Node.Document.ID)
		},
	)
}
