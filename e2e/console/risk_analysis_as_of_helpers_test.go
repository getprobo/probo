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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	asOfMeasureNodeSelection = `
		id
		name
		category
		state
		asOf
		treatmentPlans(asOf: $asOf) {
			totalCount
			edges { node { id asOf } }
		}
	`

	asOfPlanNodeSelection = `
		id
		asOf
		category
		inherentLikelihood
		inherentImpact
		netLikelihood
		netImpact
		residualLikelihood
		residualImpact
		canUpdate: permission(action: "risk-management:treatment-plan:update")
		canDelete: permission(action: "risk-management:treatment-plan:delete")
		progress { done inProgress notImplemented total }
		owner { fullName }
		risk { id name category }
		measures(first: 10, asOf: $asOf) {
			asOf
			edges { node { id name category state } }
		}
	`

	riskAnalysisPlansAsOfQuery = `
		query($id: ID!, $asOf: Datetime, $first: Int, $orderBy: TreatmentPlanOrder) {
			node(id: $id) {
				... on RiskAnalysis {
					treatmentPlans(asOf: $asOf, first: $first, orderBy: $orderBy) {
						totalCount
						edges {
							node {
								` + asOfPlanNodeSelection + `
							}
						}
					}
					matrixCells(asOf: $asOf) {
						type
						likelihood
						impact
						count
					}
				}
			}
		}
	`

	treatmentPlanMeasuresAsOfQuery = `
		query($id: ID!, $asOf: Datetime, $first: Int, $after: CursorKey, $filter: MeasureFilter) {
			node(id: $id) {
				... on TreatmentPlan {
					__typename
					asOf
					measures(first: $first, after: $after, asOf: $asOf, filter: $filter) {
						asOf
						edges {
							node {
								` + asOfMeasureNodeSelection + `
							}
						}
						pageInfo { hasNextPage endCursor }
					}
				}
			}
		}
	`

	measureTreatmentPlansAsOfQuery = `
		query($id: ID!, $asOf: Datetime) {
			node(id: $id) {
				... on Measure {
					treatmentPlans(asOf: $asOf) {
						totalCount
						edges { node { id asOf } }
					}
				}
			}
		}
	`
)

type (
	asOfMeasureNode struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		Category       string  `json:"category"`
		State          string  `json:"state"`
		AsOf           *string `json:"asOf"`
		TreatmentPlans struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID   string  `json:"id"`
					AsOf *string `json:"asOf"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"treatmentPlans"`
	}

	asOfMeasureConnection struct {
		AsOf  *string `json:"asOf"`
		Edges []struct {
			Node asOfMeasureNode `json:"node"`
		} `json:"edges"`
		PageInfo struct {
			HasNextPage bool    `json:"hasNextPage"`
			EndCursor   *string `json:"endCursor"`
		} `json:"pageInfo"`
	}

	asOfPlanNode struct {
		ID                 string  `json:"id"`
		AsOf               *string `json:"asOf"`
		Category           string  `json:"category"`
		InherentLikelihood int     `json:"inherentLikelihood"`
		InherentImpact     int     `json:"inherentImpact"`
		NetLikelihood      int     `json:"netLikelihood"`
		NetImpact          int     `json:"netImpact"`
		ResidualLikelihood int     `json:"residualLikelihood"`
		ResidualImpact     int     `json:"residualImpact"`
		CanUpdate          bool    `json:"canUpdate"`
		CanDelete          bool    `json:"canDelete"`
		Progress           struct {
			Done           int `json:"done"`
			InProgress     int `json:"inProgress"`
			NotImplemented int `json:"notImplemented"`
			Total          int `json:"total"`
		} `json:"progress"`
		Owner *struct {
			FullName string `json:"fullName"`
		} `json:"owner"`
		Risk struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"risk"`
		Measures asOfMeasureConnection `json:"measures"`
	}

	asOfMatrixCell struct {
		Type       string `json:"type"`
		Likelihood int    `json:"likelihood"`
		Impact     int    `json:"impact"`
		Count      int    `json:"count"`
	}

	asOfPlansResult struct {
		Node struct {
			TreatmentPlans struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node asOfPlanNode `json:"node"`
				} `json:"edges"`
			} `json:"treatmentPlans"`
			MatrixCells []asOfMatrixCell `json:"matrixCells"`
		} `json:"node"`
	}

	asOfTreatmentPlanMeasuresResult struct {
		Typename string
		AsOf     *string
		Measures asOfMeasureConnection
	}

	asOfPlanRefConnection struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node struct {
				ID   string  `json:"id"`
				AsOf *string `json:"asOf"`
			} `json:"node"`
		} `json:"edges"`
	}
)

func queryRiskAnalysisPlansAsOf(
	t *testing.T,
	client *testutil.Client,
	analysisID string,
	asOf *string,
	first int,
	orderBy any,
) asOfPlansResult {
	t.Helper()

	var result asOfPlansResult

	err := client.Execute(
		riskAnalysisPlansAsOfQuery,
		map[string]any{
			"id":      analysisID,
			"asOf":    asOf,
			"first":   first,
			"orderBy": orderBy,
		},
		&result,
	)
	require.NoError(t, err)

	return result
}

func queryTreatmentPlanMeasuresAsOf(
	t *testing.T,
	client *testutil.Client,
	planID string,
	asOf *string,
	first int,
	after *string,
	filter map[string]any,
) asOfTreatmentPlanMeasuresResult {
	t.Helper()

	var result struct {
		Node struct {
			Typename string                `json:"__typename"`
			AsOf     *string               `json:"asOf"`
			Measures asOfMeasureConnection `json:"measures"`
		} `json:"node"`
	}

	err := client.Execute(
		treatmentPlanMeasuresAsOfQuery,
		map[string]any{
			"id":     planID,
			"asOf":   asOf,
			"first":  first,
			"after":  after,
			"filter": filter,
		},
		&result,
	)
	require.NoError(t, err)

	return asOfTreatmentPlanMeasuresResult{
		Typename: result.Node.Typename,
		AsOf:     result.Node.AsOf,
		Measures: result.Node.Measures,
	}
}

func queryMeasureTreatmentPlansAsOf(
	t *testing.T,
	client *testutil.Client,
	measureID string,
	asOf *string,
) asOfPlanRefConnection {
	t.Helper()

	var result struct {
		Node struct {
			TreatmentPlans asOfPlanRefConnection `json:"treatmentPlans"`
		} `json:"node"`
	}

	err := client.Execute(
		measureTreatmentPlansAsOfQuery,
		map[string]any{"id": measureID, "asOf": asOf},
		&result,
	)
	require.NoError(t, err)

	return result.Node.TreatmentPlans
}

func asOfMatrixCellCount(
	cells []asOfMatrixCell,
	scoreType string,
	likelihood, impact int,
) int {
	for _, cell := range cells {
		if cell.Type == scoreType && cell.Likelihood == likelihood && cell.Impact == impact {
			return cell.Count
		}
	}

	return 0
}

func asOfPlanNodes(result asOfPlansResult) []asOfPlanNode {
	plans := make([]asOfPlanNode, 0, len(result.Node.TreatmentPlans.Edges))
	for _, edge := range result.Node.TreatmentPlans.Edges {
		plans = append(plans, edge.Node)
	}

	return plans
}

func asOfMeasureNodes(conn asOfMeasureConnection) []asOfMeasureNode {
	measures := make([]asOfMeasureNode, 0, len(conn.Edges))
	for _, edge := range conn.Edges {
		measures = append(measures, edge.Node)
	}

	return measures
}

func asOfPlanIDs(plans []asOfPlanNode) []string {
	ids := make([]string, 0, len(plans))
	for _, plan := range plans {
		ids = append(ids, plan.ID)
	}

	return ids
}

func asOfPlanByID(t *testing.T, plans []asOfPlanNode, id string) asOfPlanNode {
	t.Helper()

	for _, plan := range plans {
		if plan.ID == id {
			return plan
		}
	}

	require.FailNow(t, "treatment plan not found", id)

	return asOfPlanNode{}
}
