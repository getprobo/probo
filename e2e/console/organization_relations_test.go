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

// Organization list-document resolvers (assetListDocument, dataListDocument,
// findingsDocument, processingActivitiesDocument, risksDocument,
// thirdPartiesDocument, and similar) require publish-* mutations and are not
// covered here.

type organizationRelationGraph struct {
	orgID                string
	ownerProfileID       string
	contextProduct       string
	frameworkID          string
	controlID            string
	measureID            string
	taskID               string
	soaID                string
	assetID              string
	datumID              string
	documentID           string
	auditID              string
	findingID            string
	processingActivityID string
	compliancePortalID   string
	thirdPartyID         string
	riskID               string
	riskAnalysisID       string
	riskScenarioID       string
}

func populateOrganizationRelationGraph(
	t *testing.T,
	owner *testutil.Client,
) organizationRelationGraph {
	t.Helper()

	marker := factory.SafeName("Org relations")
	g := organizationRelationGraph{
		orgID:          owner.GetOrganizationID().String(),
		ownerProfileID: getOwnerProfileID(t, owner),
		contextProduct: marker + " product context",
	}

	err := owner.Execute(`
		mutation($input: UpdateOrganizationContextInput!) {
			updateOrganizationContext(input: $input) {
				context { organizationId product }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": g.orgID,
			"product":        g.contextProduct,
			"team":           marker + " team",
		},
	}, &struct{}{})
	require.NoError(t, err)

	g.frameworkID = factory.NewFramework(owner).
		WithName(marker + " framework").
		Create()
	g.controlID = factory.NewControl(owner, g.frameworkID).
		WithName(marker + " control").
		WithDescription("Organization relations control").
		WithSectionTitle("Section").
		WithBestPractice(true).
		WithMaturityLevel("INITIAL").
		Create()
	g.measureID = factory.NewMeasure(owner).
		WithName(marker + " measure").
		WithCategory("POLICY").
		Create()
	g.taskID = factory.NewTask(owner, g.measureID).
		WithName(marker + " task").
		Create()
	g.soaID = factory.NewStatementOfApplicability(owner).
		WithName(marker + " soa").
		Create()

	g.assetID = createOrganizationRelationAsset(t, owner, g.ownerProfileID, marker+" asset")
	g.datumID = factory.NewDatum(owner, g.ownerProfileID).
		WithName(marker + " datum").
		Create()
	g.documentID = factory.NewDocument(owner).
		WithTitle(marker + " document").
		Create()
	g.auditID = factory.NewAudit(owner, g.frameworkID).
		WithName(marker + " audit").
		Create()
	g.findingID = createOrganizationRelationFinding(t, owner, g.ownerProfileID, marker)
	g.processingActivityID = factory.NewProcessingActivity(owner).
		WithName(marker + " processing").
		Create()
	g.compliancePortalID = factory.CreateCompliancePortal(
		owner,
		factory.Attrs{"entityName": marker + " portal"},
	)
	g.thirdPartyID = factory.CreateThirdParty(
		owner,
		factory.Attrs{"name": marker + " third party"},
	)
	g.riskID = factory.NewRisk(owner).
		WithName(marker + " risk").
		WithCategory("OPERATIONAL").
		Create()
	g.riskAnalysisID = factory.CreateRiskAnalysis(
		owner,
		factory.Attrs{"name": marker + " assessment"},
	)
	scopeID := factory.CreateRiskAnalysisDiagram(
		owner,
		g.riskAnalysisID,
		factory.Attrs{"name": marker + " scope"},
	)
	g.riskScenarioID = factory.CreateRiskAnalysisScenario(
		owner,
		scopeID,
		factory.Attrs{"name": marker + " scenario"},
	)

	return g
}

func createOrganizationRelationAsset(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID, name string,
) string {
	t.Helper()

	var result struct {
		CreateAsset struct {
			AssetEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"assetEdge"`
		} `json:"createAsset"`
	}

	err := owner.Execute(`
		mutation($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":  owner.GetOrganizationID().String(),
			"name":            name,
			"amount":          1,
			"ownerId":         ownerProfileID,
			"assetType":       "VIRTUAL",
			"dataTypesStored": "Organization relations fixture",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateAsset.AssetEdge.Node.ID
}

func createOrganizationRelationFinding(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID, marker string,
) string {
	t.Helper()

	var result struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(`
		mutation($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":   owner.GetOrganizationID().String(),
			"kind":             "MINOR_NONCONFORMITY",
			"description":      marker + " finding",
			"rootCause":        "Fixture root cause",
			"correctiveAction": "Fixture corrective action",
			"ownerId":          ownerProfileID,
			"status":           "OPEN",
			"priority":         "MEDIUM",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateFinding.FindingEdge.Node.ID
}

func TestOrganization_Relations(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	graph := populateOrganizationRelationGraph(t, owner)

	cases := []struct {
		name string
		run  func(*testing.T, *testutil.Client, organizationRelationGraph)
	}{
		{name: "ContextProfilesPermission", run: orgRelationsContextProfilesPermission},
		{name: "GovernanceCollections", run: orgRelationsGovernanceCollections},
		{name: "AssetsDataDocuments", run: orgRelationsAssetsDataDocuments},
		{name: "AuditsFindings", run: orgRelationsAuditsFindings},
		{name: "PrivacyPortalThirdParties", run: orgRelationsPrivacyPortalThirdParties},
		{name: "RiskPortfolio", run: orgRelationsRiskPortfolio},
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

func orgRelationsContextProfilesPermission(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			ID        string `json:"id"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
			Context   struct {
				OrganizationID string  `json:"organizationId"`
				Product        *string `json:"product"`
				Team           *string `json:"team"`
			} `json:"context"`
			Profiles struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
			CanCreateFramework bool `json:"canCreateFramework"`
			CanListRisks       bool `json:"canListRisks"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					createdAt
					updatedAt
					context {
						organizationId
						product
						team
					}
					profiles(first: 50) {
						totalCount
						edges { node { id } }
					}
					canCreateFramework: permission(action: "core:framework:create")
					canListRisks: permission(action: "risk-management:risk:list")
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.Equal(t, g.orgID, result.Node.ID)
	assert.NotEmpty(t, result.Node.CreatedAt)
	assert.NotEmpty(t, result.Node.UpdatedAt)
	assert.Equal(t, g.orgID, result.Node.Context.OrganizationID)
	require.NotNil(t, result.Node.Context.Product)
	assert.Equal(t, g.contextProduct, *result.Node.Context.Product)
	require.NotNil(t, result.Node.Context.Team)
	assert.Contains(t, *result.Node.Context.Team, "Org relations")
	assert.GreaterOrEqual(t, result.Node.Profiles.TotalCount, 1)

	profileIDs := collectRelationNodeIDs(result.Node.Profiles.Edges)
	assert.True(t, profileIDs[g.ownerProfileID], "profiles must include owner membership")
	assert.True(t, result.Node.CanCreateFramework)
	assert.True(t, result.Node.CanListRisks)
}

func orgRelationsGovernanceCollections(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			MeasureCategories []string `json:"measureCategories"`
			Frameworks        struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"frameworks"`
			Controls struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"controls"`
			Measures struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"measures"`
			StatementsOfApplicability struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"statementsOfApplicability"`
			Tasks struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"tasks"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					measureCategories
					frameworks(first: 50) {
						totalCount
						edges { node { id } }
					}
					controls(first: 50) {
						totalCount
						edges { node { id } }
					}
					measures(first: 50) {
						totalCount
						edges { node { id } }
					}
					statementsOfApplicability(first: 50) {
						totalCount
						edges { node { id } }
					}
					tasks(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.Contains(t, result.Node.MeasureCategories, "POLICY")
	assert.GreaterOrEqual(t, result.Node.Frameworks.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Frameworks.Edges)[g.frameworkID])
	assert.GreaterOrEqual(t, result.Node.Controls.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Controls.Edges)[g.controlID])
	assert.GreaterOrEqual(t, result.Node.Measures.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Measures.Edges)[g.measureID])
	assert.GreaterOrEqual(t, result.Node.StatementsOfApplicability.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.StatementsOfApplicability.Edges)[g.soaID],
	)
	assert.GreaterOrEqual(t, result.Node.Tasks.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Tasks.Edges)[g.taskID])

	var controlResult struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			CanUpdate bool `json:"canUpdate"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Control {
					organization { id }
					canUpdate: permission(action: "core:control:update")
				}
			}
		}
	`, map[string]any{"id": g.controlID}, &controlResult)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, controlResult.Node.Organization.ID)
	assert.True(t, controlResult.Node.CanUpdate)
}

func orgRelationsAssetsDataDocuments(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			Assets struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"assets"`
			Data struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"data"`
			Documents struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"documents"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					assets(first: 50) {
						totalCount
						edges { node { id } }
					}
					data(first: 50) {
						totalCount
						edges { node { id } }
					}
					documents(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Assets.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Assets.Edges)[g.assetID])
	assert.GreaterOrEqual(t, result.Node.Data.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Data.Edges)[g.datumID])
	assert.GreaterOrEqual(t, result.Node.Documents.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Documents.Edges)[g.documentID])

	type reverseNode struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	for _, tc := range []struct {
		id    string
		graph string
	}{
		{id: g.assetID, graph: "Asset"},
		{id: g.datumID, graph: "Datum"},
		{id: g.documentID, graph: "Document"},
	} {
		var reverse reverseNode

		err = owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on `+tc.graph+` {
						organization { id }
					}
				}
			}
		`, map[string]any{"id": tc.id}, &reverse)
		require.NoError(t, err)
		assert.Equal(t, g.orgID, reverse.Node.Organization.ID)
	}
}

func orgRelationsAuditsFindings(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			Audits struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"audits"`
			Findings struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"findings"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					audits(first: 50) {
						totalCount
						edges { node { id } }
					}
					findings(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Audits.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Audits.Edges)[g.auditID])
	assert.GreaterOrEqual(t, result.Node.Findings.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Findings.Edges)[g.findingID])

	var auditReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Audit {
					organization { id }
				}
			}
		}
	`, map[string]any{"id": g.auditID}, &auditReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, auditReverse.Node.Organization.ID)

	var findingReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			CanUpdate bool `json:"canUpdate"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Finding {
					organization { id }
					canUpdate: permission(action: "core:finding:update")
				}
			}
		}
	`, map[string]any{"id": g.findingID}, &findingReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, findingReverse.Node.Organization.ID)
	assert.True(t, findingReverse.Node.CanUpdate)
}

func orgRelationsPrivacyPortalThirdParties(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			ProcessingActivities struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"processingActivities"`
			CompliancePortals struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"compliancePortals"`
			ThirdParties struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"thirdParties"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					processingActivities(first: 50) {
						totalCount
						edges { node { id } }
					}
					compliancePortals(first: 50) {
						totalCount
						edges { node { id } }
					}
					thirdParties(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.ProcessingActivities.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.ProcessingActivities.Edges)[g.processingActivityID],
	)
	assert.GreaterOrEqual(t, result.Node.CompliancePortals.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.CompliancePortals.Edges)[g.compliancePortalID],
	)
	assert.GreaterOrEqual(t, result.Node.ThirdParties.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.ThirdParties.Edges)[g.thirdPartyID])

	var paReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on ProcessingActivity {
					organization { id }
				}
			}
		}
	`, map[string]any{"id": g.processingActivityID}, &paReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, paReverse.Node.Organization.ID)

	var portalReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			CanUpdate bool `json:"canUpdate"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					organization { id }
					canUpdate: permission(action: "compliance-portal:portal:update")
				}
			}
		}
	`, map[string]any{"id": g.compliancePortalID}, &portalReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, portalReverse.Node.Organization.ID)
	assert.True(t, portalReverse.Node.CanUpdate)

	var tpReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					organization { id }
				}
			}
		}
	`, map[string]any{"id": g.thirdPartyID}, &tpReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, tpReverse.Node.Organization.ID)
}

func orgRelationsRiskPortfolio(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			Risks struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"risks"`
			RiskAnalyses struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"riskAnalyses"`
			RiskAnalysisScenarios struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"riskAnalysisScenarios"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					risks(first: 50) {
						totalCount
						edges { node { id } }
					}
					riskAnalyses(first: 50) {
						totalCount
						edges { node { id } }
					}
					riskAnalysisScenarios(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Risks.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Risks.Edges)[g.riskID])
	assert.GreaterOrEqual(t, result.Node.RiskAnalyses.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.RiskAnalyses.Edges)[g.riskAnalysisID],
	)
	assert.GreaterOrEqual(t, result.Node.RiskAnalysisScenarios.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.RiskAnalysisScenarios.Edges)[g.riskScenarioID],
	)

	var riskReverse struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			CanUpdate bool `json:"canUpdate"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					organization { id }
					canUpdate: permission(action: "risk-management:risk:update")
				}
			}
		}
	`, map[string]any{"id": g.riskID}, &riskReverse)
	require.NoError(t, err)
	assert.Equal(t, g.orgID, riskReverse.Node.Organization.ID)
	assert.True(t, riskReverse.Node.CanUpdate)
}
