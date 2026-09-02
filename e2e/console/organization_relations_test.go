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
	orgID                  string
	ownerProfileID         string
	contextProduct         string
	frameworkID            string
	controlID              string
	measureID              string
	taskID                 string
	soaID                  string
	assetID                string
	datumID                string
	documentID             string
	auditID                string
	findingID              string
	processingActivityID   string
	compliancePortalID     string
	thirdPartyID           string
	riskID                 string
	riskAnalysisID         string
	riskScenarioID         string
	treatmentPlanID        string
	obligationID           string
	aiSystemID             string
	businessFunctionID     string
	deviceID               string
	cookieBannerID         string
	accessReviewSourceID   string
	accessReviewCampaignID string
	logoFileID             string
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

	g.logoFileID = createOrganizationRelationLogo(t, owner)
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
	approveTestDocument(t, owner, g.documentID)
	requestDocumentSignature(
		t,
		owner,
		latestDocumentVersionID(t, owner, g.documentID),
		g.ownerProfileID,
	)
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
		factory.Attrs{
			"name":             marker + " third party",
			"administratorIds": []string{g.ownerProfileID},
		},
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
	factory.LinkRiskAnalysisScenarioRisk(owner, g.riskScenarioID, g.riskID)
	g.treatmentPlanID = factory.CreateTreatmentPlan(
		owner,
		g.riskID,
		g.riskAnalysisID,
		factory.Attrs{"ownerId": g.ownerProfileID},
	)
	g.obligationID = createOrganizationRelationObligation(
		t,
		owner,
		g.ownerProfileID,
		marker+" obligation",
	)
	g.aiSystemID = createOrganizationRelationAiSystem(
		t,
		owner,
		g.ownerProfileID,
		marker+" ai system",
	)
	g.businessFunctionID = createOrganizationRelationBusinessFunction(
		t,
		owner,
		g.ownerProfileID,
		marker+" business function",
	)
	g.deviceID = createOrganizationRelationDevice(t, owner, g.ownerProfileID)
	g.cookieBannerID = factory.NewCookieBanner(owner).
		WithName(marker + " banner").
		Create()
	g.accessReviewSourceID = factory.NewAccessReviewSource(owner, g.orgID).
		WithName(marker + " source").
		Create()
	g.accessReviewCampaignID = factory.NewAccessReviewCampaign(owner, g.orgID).
		WithName(marker + " campaign").
		WithAccessReviewSourceIDs([]string{g.accessReviewSourceID}).
		Create()

	return g
}

func createOrganizationRelationLogo(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var result struct {
		UpdateOrganization struct {
			Organization struct {
				Logo *struct {
					ID string `json:"id"`
				} `json:"logo"`
			} `json:"organization"`
		} `json:"updateOrganization"`
	}

	err := owner.ExecuteConnectWithFile(
		`
			mutation($input: UpdateOrganizationInput!) {
				updateOrganization(input: $input) {
					organization { logo { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"logoFile":       nil,
			},
		},
		"input.logoFile",
		testutil.UploadFile{
			Filename:    "org-relations-logo.png",
			ContentType: "image/png",
			Content:     pngContent,
		},
		&result,
	)
	require.NoError(t, err)
	require.NotNil(t, result.UpdateOrganization.Organization.Logo)

	return result.UpdateOrganization.Organization.Logo.ID
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

func createOrganizationRelationObligation(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID, requirement string,
) string {
	t.Helper()

	var result struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"requirement":    requirement,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
			"ownerId":        ownerProfileID,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateObligation.ObligationEdge.Node.ID
}

func createOrganizationRelationAiSystem(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID, name string,
) string {
	t.Helper()

	var result struct {
		CreateAiSystem struct {
			AiSystemEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"aiSystemEdge"`
		} `json:"createAiSystem"`
	}

	err := owner.Execute(`
		mutation($input: CreateAiSystemInput!) {
			createAiSystem(input: $input) {
				aiSystemEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"name":               name,
			"status":             "ACTIVE",
			"riskClassification": "MINIMAL",
			"ownerId":            ownerProfileID,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateAiSystem.AiSystemEdge.Node.ID
}

func createOrganizationRelationBusinessFunction(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID, name string,
) string {
	t.Helper()

	var result struct {
		CreateBusinessFunction struct {
			BusinessFunctionEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"businessFunctionEdge"`
		} `json:"createBusinessFunction"`
	}

	err := owner.Execute(`
		mutation($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           name,
			"classification": "CRITICAL",
			"mtdMinutes":     240,
			"rtoMinutes":     120,
			"rpoMinutes":     60,
			"ownerId":        ownerProfileID,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateBusinessFunction.BusinessFunctionEdge.Node.ID
}

func createOrganizationRelationDevice(
	t *testing.T,
	owner *testutil.Client,
	ownerProfileID string,
) string {
	t.Helper()

	var result struct {
		CreateDevice struct {
			Device struct {
				ID string `json:"id"`
			} `json:"device"`
		} `json:"createDevice"`
	}

	err := owner.Execute(`
		mutation($input: CreateDeviceInput!) {
			createDevice(input: $input) {
				device { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"ownerId":        ownerProfileID,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateDevice.Device.ID
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
		{name: "OwnedPrivacyAndReviews", run: orgRelationsOwnedPrivacyAndReviews},
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
			Logo *struct {
				ID string `json:"id"`
			} `json:"logo"`
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
					logo { id }
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
	require.NotNil(t, result.Node.Logo)
	assert.Equal(t, g.logoFileID, result.Node.Logo.ID)
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

func orgRelationsOwnedPrivacyAndReviews(
	t *testing.T,
	owner *testutil.Client,
	g organizationRelationGraph,
) {
	var result struct {
		Node struct {
			Obligations struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"obligations"`
			BusinessFunctions struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"businessFunctions"`
			AiSystems struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"aiSystems"`
			Devices struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"devices"`
			TreatmentPlans struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"treatmentPlans"`
			CookieBanners struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"cookieBanners"`
			AccessReviewSources struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accessReviewSources"`
			AccessReviewCampaigns struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accessReviewCampaigns"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					obligations(first: 50) {
						totalCount
						edges { node { id } }
					}
					businessFunctions(first: 50) {
						totalCount
						edges { node { id } }
					}
					aiSystems(first: 50) {
						totalCount
						edges { node { id } }
					}
					devices(first: 50) {
						totalCount
						edges { node { id } }
					}
					treatmentPlans(first: 50) {
						totalCount
						edges { node { id } }
					}
					cookieBanners(first: 50) {
						totalCount
						edges { node { id } }
					}
					accessReviewSources(first: 50) {
						totalCount
						edges { node { id } }
					}
					accessReviewCampaigns(first: 50) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": g.orgID}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Obligations.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Obligations.Edges)[g.obligationID])
	assert.GreaterOrEqual(t, result.Node.BusinessFunctions.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.BusinessFunctions.Edges)[g.businessFunctionID])
	assert.GreaterOrEqual(t, result.Node.AiSystems.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.AiSystems.Edges)[g.aiSystemID])
	assert.GreaterOrEqual(t, result.Node.Devices.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.Devices.Edges)[g.deviceID])
	assert.GreaterOrEqual(t, result.Node.TreatmentPlans.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.TreatmentPlans.Edges)[g.treatmentPlanID])
	assert.GreaterOrEqual(t, result.Node.CookieBanners.TotalCount, 1)
	assert.True(t, collectRelationNodeIDs(result.Node.CookieBanners.Edges)[g.cookieBannerID])
	assert.GreaterOrEqual(t, result.Node.AccessReviewSources.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.AccessReviewSources.Edges)[g.accessReviewSourceID],
	)
	assert.GreaterOrEqual(t, result.Node.AccessReviewCampaigns.TotalCount, 1)
	assert.True(
		t,
		collectRelationNodeIDs(result.Node.AccessReviewCampaigns.Edges)[g.accessReviewCampaignID],
	)
}
