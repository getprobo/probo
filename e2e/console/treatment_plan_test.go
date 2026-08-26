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

func TestTreatmentPlan_Create(t *testing.T) {
	t.Parallel()

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)

		var result struct {
			CreateTreatmentPlan struct {
				TreatmentPlanEdge struct {
					Node struct {
						ID                 string `json:"id"`
						Treatment          string `json:"treatment"`
						InherentLikelihood int    `json:"inherentLikelihood"`
						InherentImpact     int    `json:"inherentImpact"`
						InherentRiskScore  int    `json:"inherentRiskScore"`
						ResidualLikelihood int    `json:"residualLikelihood"`
						ResidualImpact     int    `json:"residualImpact"`
						ResidualRiskScore  int    `json:"residualRiskScore"`
					} `json:"node"`
				} `json:"treatmentPlanEdge"`
			} `json:"createTreatmentPlan"`
		}

		err := owner.Execute(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge {
						node {
							id
							treatment
							inherentLikelihood
							inherentImpact
							inherentRiskScore
							residualLikelihood
							residualImpact
							residualRiskScore
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 3,
				"inherentImpact":     4,
			},
		}, &result)

		require.NoError(t, err)

		node := result.CreateTreatmentPlan.TreatmentPlanEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "MITIGATED", node.Treatment)
		assert.Equal(t, 3, node.InherentLikelihood)
		assert.Equal(t, 4, node.InherentImpact)
		assert.Equal(t, 12, node.InherentRiskScore)
		assert.Equal(t, 3, node.ResidualLikelihood)
		assert.Equal(t, 4, node.ResidualImpact)
		assert.Equal(t, 12, node.ResidualRiskScore)
	})

	t.Run("with owner and residual scores", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)
		profileID := owner.GetProfileID().String()

		var result struct {
			CreateTreatmentPlan struct {
				TreatmentPlanEdge struct {
					Node struct {
						ID                 string `json:"id"`
						ResidualLikelihood int    `json:"residualLikelihood"`
						ResidualImpact     int    `json:"residualImpact"`
						ResidualRiskScore  int    `json:"residualRiskScore"`
						Owner              struct {
							ID string `json:"id"`
						} `json:"owner"`
					} `json:"node"`
				} `json:"treatmentPlanEdge"`
			} `json:"createTreatmentPlan"`
		}

		err := owner.Execute(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge {
						node {
							id
							residualLikelihood
							residualImpact
							residualRiskScore
							owner { id }
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            profileID,
				"inherentLikelihood": 4,
				"inherentImpact":     4,
				"residualLikelihood": 2,
				"residualImpact":     3,
			},
		}, &result)

		require.NoError(t, err)

		node := result.CreateTreatmentPlan.TreatmentPlanEdge.Node
		assert.Equal(t, profileID, node.Owner.ID)
		assert.Equal(t, 2, node.ResidualLikelihood)
		assert.Equal(t, 3, node.ResidualImpact)
		assert.Equal(t, 6, node.ResidualRiskScore)
	})

	t.Run("rejects unpaired residual score", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)

		_, err := owner.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 3,
				"inherentImpact":     3,
				"residualLikelihood": 1,
			},
		})
		testutil.RequireErrorCode(t, err, "INVALID", "residual fields must be set together")
	})

	t.Run("rejects risk not linked to a scenario", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)

		_, err := owner.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
		})
		testutil.RequireErrorCode(t, err, "INVALID", "risk must be linked to a scenario")
	})

	t.Run("rejects duplicate risk and analysis", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)
		factory.CreateTreatmentPlan(owner, riskID, analysisID)

		_, err := owner.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "AVOIDED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
		})
		testutil.RequireConflictError(t, err, "duplicate treatment plan for risk and analysis")
	})

	t.Run("rejects scores outside analysis matrix", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{"matrixRows": 3, "matrixCols": 3})
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)

		_, err := owner.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 4,
				"inherentImpact":     2,
			},
		})
		testutil.RequireErrorCode(t, err, "INVALID", "score exceeds analysis matrix")
	})

	t.Run("rejects accepted residual that differs from inherent", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)

		_, err := owner.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "ACCEPTED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 4,
				"inherentImpact":     4,
				"residualLikelihood": 1,
				"residualImpact":     2,
			},
		})
		testutil.RequireErrorCode(t, err, "INVALID", "accepted residual must match inherent")
	})
}

func TestTreatmentPlan_Update(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID)

	var result struct {
		UpdateTreatmentPlan struct {
			TreatmentPlan struct {
				ID        string `json:"id"`
				Treatment string `json:"treatment"`
			} `json:"treatmentPlan"`
		} `json:"updateTreatmentPlan"`
	}

	err := owner.Execute(`
		mutation($input: UpdateTreatmentPlanInput!) {
			updateTreatmentPlan(input: $input) {
				treatmentPlan { id treatment }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":        tpID,
			"treatment": "TRANSFERRED",
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "TRANSFERRED", result.UpdateTreatmentPlan.TreatmentPlan.Treatment)

	var ownerResult struct {
		UpdateTreatmentPlan struct {
			TreatmentPlan struct {
				Owner struct {
					ID string `json:"id"`
				} `json:"owner"`
			} `json:"treatmentPlan"`
		} `json:"updateTreatmentPlan"`
	}

	err = owner.Execute(`
		mutation($input: UpdateTreatmentPlanInput!) {
			updateTreatmentPlan(input: $input) {
				treatmentPlan { owner { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      tpID,
			"ownerId": owner.GetProfileID().String(),
		},
	}, &ownerResult)
	require.NoError(t, err)
	assert.Equal(t, owner.GetProfileID().String(), ownerResult.UpdateTreatmentPlan.TreatmentPlan.Owner.ID)

	var clearResult struct {
		UpdateTreatmentPlan struct {
			TreatmentPlan struct {
				Owner *struct {
					ID string `json:"id"`
				} `json:"owner"`
			} `json:"treatmentPlan"`
		} `json:"updateTreatmentPlan"`
	}

	err = owner.Execute(`
		mutation($input: UpdateTreatmentPlanInput!) {
			updateTreatmentPlan(input: $input) {
				treatmentPlan { owner { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      tpID,
			"ownerId": nil,
		},
	}, &clearResult)
	testutil.RequireErrorCode(t, err, "INVALID", "cannot clear treatment plan owner")
}

func TestTreatmentPlan_Update_AcceptedRejectsMismatchedResidual(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID, factory.Attrs{
		"inherentLikelihood": 4,
		"inherentImpact":     4,
		"residualLikelihood": 1,
		"residualImpact":     2,
	})

	_, err := owner.Do(`
		mutation($input: UpdateTreatmentPlanInput!) {
			updateTreatmentPlan(input: $input) {
				treatmentPlan { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":        tpID,
			"treatment": "ACCEPTED",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "accepted residual must match inherent")
}

func TestTreatmentPlan_Delete(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID)

	var result struct {
		DeleteTreatmentPlan struct {
			DeletedTreatmentPlanID string `json:"deletedTreatmentPlanId"`
		} `json:"deleteTreatmentPlan"`
	}

	err := owner.Execute(`
		mutation($input: DeleteTreatmentPlanInput!) {
			deleteTreatmentPlan(input: $input) {
				deletedTreatmentPlanId
			}
		}
	`, map[string]any{
		"input": map[string]any{"treatmentPlanId": tpID},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, tpID, result.DeleteTreatmentPlan.DeletedTreatmentPlanID)
}

func TestTreatmentPlan_List(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID)

	var result struct {
		Node struct {
			TreatmentPlans struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"treatmentPlans"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					treatmentPlans(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.TreatmentPlans.TotalCount, 1)

	found := false

	for _, edge := range result.Node.TreatmentPlans.Edges {
		if edge.Node.ID == tpID {
			found = true
			break
		}
	}

	assert.True(t, found)

	t.Run("on risk", func(t *testing.T) {
		var riskResult struct {
			Node struct {
				TreatmentPlans struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"treatmentPlans"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Risk {
						treatmentPlans(first: 10) {
							totalCount
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": riskID}, &riskResult)
		require.NoError(t, err)
		assert.Equal(t, 1, riskResult.Node.TreatmentPlans.TotalCount)
		require.Len(t, riskResult.Node.TreatmentPlans.Edges, 1)
		assert.Equal(t, tpID, riskResult.Node.TreatmentPlans.Edges[0].Node.ID)
	})

	t.Run("on risk analysis", func(t *testing.T) {
		var analysisResult struct {
			Node struct {
				TreatmentPlans struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"treatmentPlans"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on RiskAnalysis {
						treatmentPlans(first: 10) {
							totalCount
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": analysisID}, &analysisResult)
		require.NoError(t, err)
		assert.Equal(t, 1, analysisResult.Node.TreatmentPlans.TotalCount)
		require.Len(t, analysisResult.Node.TreatmentPlans.Edges, 1)
		assert.Equal(t, tpID, analysisResult.Node.TreatmentPlans.Edges[0].Node.ID)
	})

	t.Run("scenario risks exclude planned", func(t *testing.T) {
		plannedRiskID := factory.CreateRisk(owner)
		unplannedRiskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, plannedRiskID, analysisID)
		factory.LinkRiskToAnalysis(owner, unplannedRiskID, analysisID)
		factory.CreateTreatmentPlan(owner, plannedRiskID, analysisID)

		var result struct {
			Node struct {
				ScenarioRisks struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"scenarioRisks"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on RiskAnalysis {
						scenarioRisks(first: 50) {
							totalCount
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": analysisID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.ScenarioRisks.TotalCount)
		require.Len(t, result.Node.ScenarioRisks.Edges, 1)
		assert.Equal(t, unplannedRiskID, result.Node.ScenarioRisks.Edges[0].Node.ID)
	})

	t.Run("scenario risks filter by query", func(t *testing.T) {
		matchedRiskID := factory.CreateRisk(owner, factory.Attrs{"name": "FilterableScenarioRiskAlpha"})
		otherRiskID := factory.CreateRisk(owner, factory.Attrs{"name": "UnrelatedScenarioRiskBeta"})
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, matchedRiskID, analysisID)
		factory.LinkRiskToAnalysis(owner, otherRiskID, analysisID)

		var result struct {
			Node struct {
				ScenarioRisks struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"scenarioRisks"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!, $filter: RiskFilter) {
				node(id: $id) {
					... on RiskAnalysis {
						scenarioRisks(first: 50, filter: $filter) {
							totalCount
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{
			"id":     analysisID,
			"filter": map[string]any{"query": "FilterableScenarioRiskAlpha"},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.ScenarioRisks.TotalCount)
		require.Len(t, result.Node.ScenarioRisks.Edges, 1)
		assert.Equal(t, matchedRiskID, result.Node.ScenarioRisks.Edges[0].Node.ID)
	})
}

func TestTreatmentPlan_LinkMeasure(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID)
	measureID := factory.CreateMeasure(owner, factory.Attrs{"name": "Access reviews"})

	factory.LinkTreatmentPlanMeasure(owner, tpID, measureID)

	var listResult struct {
		Node struct {
			Measures struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"measures"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on TreatmentPlan {
					measures(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": tpID}, &listResult)
	require.NoError(t, err)
	assert.Equal(t, 1, listResult.Node.Measures.TotalCount)
	require.Len(t, listResult.Node.Measures.Edges, 1)
	assert.Equal(t, measureID, listResult.Node.Measures.Edges[0].Node.ID)
}

func TestTreatmentPlan_NetFromMeasureProgress(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID, factory.Attrs{
		"treatment":          "MITIGATED",
		"inherentLikelihood": 4,
		"inherentImpact":     4,
		"residualLikelihood": 1,
		"residualImpact":     2,
	})

	assertTreatmentPlanNet(t, owner, tpID, 4, 4, 16)

	measureIDs := []string{
		factory.CreateMeasure(owner, factory.Attrs{"name": "First mitigation measure"}),
		factory.CreateMeasure(owner, factory.Attrs{"name": "Second mitigation measure"}),
	}
	factory.LinkTreatmentPlanMeasure(owner, tpID, measureIDs[0])
	factory.LinkTreatmentPlanMeasure(owner, tpID, measureIDs[1])

	assertTreatmentPlanNet(t, owner, tpID, 4, 4, 16)

	updateMeasureState(t, owner, measureIDs[0], "IN_PROGRESS")
	assertTreatmentPlanNet(t, owner, tpID, 4, 4, 16)

	updateMeasureState(t, owner, measureIDs[0], "IMPLEMENTED")
	assertTreatmentPlanNet(t, owner, tpID, 4, 4, 16)

	updateMeasureState(t, owner, measureIDs[1], "IMPLEMENTED")
	assertTreatmentPlanNet(t, owner, tpID, 1, 2, 2)
}

func TestTreatmentPlan_AcceptedEmptyNet(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID, factory.Attrs{
		"treatment":          "ACCEPTED",
		"inherentLikelihood": 4,
		"inherentImpact":     4,
		"residualLikelihood": 4,
		"residualImpact":     4,
	})

	assertTreatmentPlanNet(t, owner, tpID, 4, 4, 16)
}

func TestTreatmentPlan_AcceptedRejectsMeasureLink(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)
	tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID, factory.Attrs{
		"treatment": "ACCEPTED",
	})
	measureID := factory.CreateMeasure(owner)

	_, err := owner.Do(`
		mutation($input: CreateTreatmentPlanMeasureMappingInput!) {
			createTreatmentPlanMeasureMapping(input: $input) {
				measureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"treatmentPlanId": tpID,
			"measureId":       measureID,
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "accepted treatment plan cannot link measures")
}

type treatmentPlanNet struct {
	NetLikelihood int `json:"netLikelihood"`
	NetImpact     int `json:"netImpact"`
	NetRiskScore  int `json:"netRiskScore"`
}

func assertTreatmentPlanNet(
	t *testing.T,
	owner *testutil.Client,
	tpID string,
	likelihood, impact, score int,
) {
	t.Helper()

	var result struct {
		Node treatmentPlanNet `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on TreatmentPlan {
					netLikelihood
					netImpact
					netRiskScore
				}
			}
		}
	`, map[string]any{"id": tpID}, &result)
	require.NoError(t, err)
	assert.Equal(t, likelihood, result.Node.NetLikelihood)
	assert.Equal(t, impact, result.Node.NetImpact)
	assert.Equal(t, score, result.Node.NetRiskScore)
}

func updateMeasureState(t *testing.T, owner *testutil.Client, measureID, state string) {
	t.Helper()

	var result struct {
		UpdateMeasure struct {
			Measure struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"measure"`
		} `json:"updateMeasure"`
	}

	err := owner.Execute(`
		mutation($input: UpdateMeasureInput!) {
			updateMeasure(input: $input) {
				measure { id state }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":    measureID,
			"state": state,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, state, result.UpdateMeasure.Measure.State)
}

func TestTreatmentPlan_FilterByMatrixCell(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	highRiskID := factory.CreateRisk(owner)
	lowRiskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, highRiskID, analysisID)
	factory.LinkRiskToAnalysis(owner, lowRiskID, analysisID)

	highPlanID := factory.CreateTreatmentPlan(owner, highRiskID, analysisID, factory.Attrs{
		"treatment":          "MITIGATED",
		"inherentLikelihood": 4,
		"inherentImpact":     4,
		"residualLikelihood": 1,
		"residualImpact":     2,
	})
	lowPlanID := factory.CreateTreatmentPlan(owner, lowRiskID, analysisID, factory.Attrs{
		"treatment":          "MITIGATED",
		"inherentLikelihood": 2,
		"inherentImpact":     3,
		"residualLikelihood": 2,
		"residualImpact":     3,
	})

	assertFilteredPlanIDs(t, owner, analysisID, "INHERENT", 4, 4, highPlanID)
	assertFilteredPlanIDs(t, owner, analysisID, "RESIDUAL", 1, 2, highPlanID)
	assertFilteredPlanIDs(t, owner, analysisID, "NET", 4, 4, highPlanID)
	assertFilteredPlanIDs(t, owner, analysisID, "INHERENT", 2, 3, lowPlanID)
	assertFilteredPlanIDs(t, owner, analysisID, "NET", 1, 2)

	measureID := factory.CreateMeasure(owner, factory.Attrs{"name": "Close residual gap"})
	factory.LinkTreatmentPlanMeasure(owner, highPlanID, measureID)
	updateMeasureState(t, owner, measureID, "IMPLEMENTED")

	assertFilteredPlanIDs(t, owner, analysisID, "NET", 1, 2, highPlanID)
	assertFilteredPlanIDs(t, owner, analysisID, "NET", 4, 4)
	assertFilteredPlanIDs(t, owner, analysisID, "INHERENT", 4, 4, highPlanID)
}

func TestTreatmentPlan_RejectIncompleteFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	analysisID := factory.CreateRiskAnalysis(owner)

	query := `
		query($id: ID!, $filter: TreatmentPlanFilter) {
			node(id: $id) {
				... on RiskAnalysis {
					treatmentPlans(first: 50, filter: $filter) {
						totalCount
					}
				}
			}
		}
	`

	_, err := owner.Do(query, map[string]any{
		"id": analysisID,
		"filter": map[string]any{
			"scoreType": "NET",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")

	_, err = owner.Do(query, map[string]any{
		"id": analysisID,
		"filter": map[string]any{
			"scoreType":  "INHERENT",
			"likelihood": 4,
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")

	var emptyResult struct {
		Node struct {
			TreatmentPlans struct {
				TotalCount int `json:"totalCount"`
			} `json:"treatmentPlans"`
		} `json:"node"`
	}

	err = owner.Execute(query, map[string]any{
		"id":     analysisID,
		"filter": map[string]any{},
	}, &emptyResult)
	require.NoError(t, err)
	assert.Equal(t, 0, emptyResult.Node.TreatmentPlans.TotalCount)
}

func TestTreatmentPlan_OrderByCategory(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	zetaRiskID := factory.CreateRisk(owner, factory.Attrs{"category": "ZETA"})
	alphaRiskID := factory.CreateRisk(owner, factory.Attrs{"category": "ALPHA"})
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, zetaRiskID, analysisID)
	factory.LinkRiskToAnalysis(owner, alphaRiskID, analysisID)

	zetaPlanID := factory.CreateTreatmentPlan(owner, zetaRiskID, analysisID)
	alphaPlanID := factory.CreateTreatmentPlan(owner, alphaRiskID, analysisID)

	var result struct {
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

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysis {
					treatmentPlans(first: 50, orderBy: { field: CATEGORY, direction: ASC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": analysisID}, &result)
	require.NoError(t, err)
	require.Len(t, result.Node.TreatmentPlans.Edges, 2)
	assert.Equal(t, alphaPlanID, result.Node.TreatmentPlans.Edges[0].Node.ID)
	assert.Equal(t, zetaPlanID, result.Node.TreatmentPlans.Edges[1].Node.ID)
}

func TestTreatmentPlan_MatrixCounts(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	highRiskID := factory.CreateRisk(owner)
	lowRiskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, highRiskID, analysisID)
	factory.LinkRiskToAnalysis(owner, lowRiskID, analysisID)

	highPlanID := factory.CreateTreatmentPlan(owner, highRiskID, analysisID, factory.Attrs{
		"treatment":          "MITIGATED",
		"inherentLikelihood": 4,
		"inherentImpact":     4,
		"residualLikelihood": 1,
		"residualImpact":     2,
	})
	factory.CreateTreatmentPlan(owner, lowRiskID, analysisID, factory.Attrs{
		"treatment":          "MITIGATED",
		"inherentLikelihood": 2,
		"inherentImpact":     3,
		"residualLikelihood": 2,
		"residualImpact":     3,
	})

	counts := queryMatrixCells(t, owner, analysisID)
	assert.Equal(t, 1, matrixCellCount(counts, "INHERENT", 4, 4))
	assert.Equal(t, 1, matrixCellCount(counts, "INHERENT", 2, 3))
	assert.Equal(t, 1, matrixCellCount(counts, "RESIDUAL", 1, 2))
	assert.Equal(t, 1, matrixCellCount(counts, "RESIDUAL", 2, 3))
	assert.Equal(t, 1, matrixCellCount(counts, "NET", 4, 4))
	assert.Equal(t, 1, matrixCellCount(counts, "NET", 2, 3))
	assert.Equal(t, 0, matrixCellCount(counts, "NET", 1, 2))

	measureID := factory.CreateMeasure(owner, factory.Attrs{"name": "Finish mitigation"})
	factory.LinkTreatmentPlanMeasure(owner, highPlanID, measureID)
	updateMeasureState(t, owner, measureID, "IMPLEMENTED")

	counts = queryMatrixCells(t, owner, analysisID)
	assert.Equal(t, 1, matrixCellCount(counts, "INHERENT", 4, 4))
	assert.Equal(t, 1, matrixCellCount(counts, "RESIDUAL", 1, 2))
	assert.Equal(t, 1, matrixCellCount(counts, "NET", 1, 2))
	assert.Equal(t, 1, matrixCellCount(counts, "NET", 2, 3))
	assert.Equal(t, 0, matrixCellCount(counts, "NET", 4, 4))
}

type matrixCellCountRow struct {
	Type       string `json:"type"`
	Likelihood int    `json:"likelihood"`
	Impact     int    `json:"impact"`
	Count      int    `json:"count"`
}

func queryMatrixCells(
	t *testing.T,
	owner *testutil.Client,
	analysisID string,
) []matrixCellCountRow {
	t.Helper()

	var result struct {
		Node struct {
			MatrixCells []matrixCellCountRow `json:"matrixCells"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on RiskAnalysis {
					matrixCells {
						type
						likelihood
						impact
						count
					}
				}
			}
		}
	`, map[string]any{"id": analysisID}, &result)
	require.NoError(t, err)

	return result.Node.MatrixCells
}

func matrixCellCount(
	counts []matrixCellCountRow,
	scoreType string,
	likelihood, impact int,
) int {
	for _, cell := range counts {
		if cell.Type == scoreType && cell.Likelihood == likelihood && cell.Impact == impact {
			return cell.Count
		}
	}

	return 0
}

func assertFilteredPlanIDs(
	t *testing.T,
	owner *testutil.Client,
	analysisID, scoreType string,
	likelihood, impact int,
	wantIDs ...string,
) {
	t.Helper()

	var result struct {
		Node struct {
			TreatmentPlans struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"treatmentPlans"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!, $filter: TreatmentPlanFilter) {
			node(id: $id) {
				... on RiskAnalysis {
					treatmentPlans(first: 50, filter: $filter) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{
		"id": analysisID,
		"filter": map[string]any{
			"scoreType":  scoreType,
			"likelihood": likelihood,
			"impact":     impact,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, len(wantIDs), result.Node.TreatmentPlans.TotalCount)

	gotIDs := make([]string, 0, len(result.Node.TreatmentPlans.Edges))
	for _, edge := range result.Node.TreatmentPlans.Edges {
		gotIDs = append(gotIDs, edge.Node.ID)
	}

	assert.ElementsMatch(t, wantIDs, gotIDs)
}
