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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
)

// injectCrossTenantFK bypasses the application entirely and writes a foreign
// key directly via SQL against the same Postgres database the e2e probod
// instance is running against. This simulates a cross-tenant reference that
// exists in the row despite the application's own write-time validation --
// e.g. a future regression in that specific check, a data migration bug, or
// direct DB access -- so that the read-path fix (GHSA-c74x-79w6-63jh pattern
// 2: authorizing the actual child id, not the parent's) can be verified on
// its own, independently of whether the write-time check is still in place.
func injectCrossTenantFK(t *testing.T, table, column, rowID, foreignID string) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `UPDATE `+table+` SET `+column+` = $1 WHERE id = $2`, foreignID, rowID)
		return err
	})
	require.NoError(t, err, "test setup: cannot inject cross-tenant FK into %s.%s", table, column)
}

// TestSecurity_ReadGap_FindingRisk independently verifies the fix for
// GHSA-c74x-79w6-63jh's confirmed exploit #1. findingResolver.Risk now
// authorizes obj.Risk.ID instead of obj.ID; this test proves that holds even
// when findings.risk_id points at another organization's risk by means other
// than FindingService.Create/Update (which is separately validated and
// covered by TestFinding_TenantIsolation).
func TestSecurity_ReadGap_FindingRisk(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2RiskID := factory.CreateRisk(org2Owner, factory.Attrs{"name": "Org2 Secret Risk (read-gap probe)"})

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org1Owner.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)

	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	injectCrossTenantFK(t, "findings", "risk_id", findingID, org2RiskID)

	var readResult struct {
		Node struct {
			Risk *struct {
				ID string `json:"id"`
			} `json:"risk"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Finding {
					risk { id }
				}
			}
		}
	`, map[string]any{"id": findingID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Risk == nil, "cross-tenant risk via finding.risk")
}

// TestSecurity_ReadGap_ProcessingActivityDataProtectionOfficer independently
// verifies the fix for GHSA-c74x-79w6-63jh's confirmed exploit #2 (PII
// disclosure). processingActivityResolver.DataProtectionOfficer now
// authorizes obj.DataProtectionOfficer.ID instead of obj.ID.
func TestSecurity_ReadGap_ProcessingActivityDataProtectionOfficer(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret DPO (read-gap probe)"})

	paID := factory.NewProcessingActivity(org1Owner).WithName("Org1 PA for read-gap probe").Create()

	injectCrossTenantFK(t, "processing_activities", "dpo_profile_id", paID, org2ProfileID)

	var readResult struct {
		Node struct {
			DataProtectionOfficer *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"dataProtectionOfficer"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on ProcessingActivity {
					dataProtectionOfficer { id fullName }
				}
			}
		}
	`, map[string]any{"id": paID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.DataProtectionOfficer == nil, "cross-tenant profile PII via processingActivity.dataProtectionOfficer")
}

// TestSecurity_ReadGap_AuditFramework independently verifies the
// defense-in-depth fix for auditResolver.Framework, which now authorizes
// obj.Framework.ID instead of obj.ID.
func TestSecurity_ReadGap_AuditFramework(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1FrameworkID := factory.CreateFramework(org1Owner)
	auditID := factory.CreateAudit(org1Owner, org1FrameworkID)

	org2FrameworkID := factory.CreateFramework(org2Owner, factory.Attrs{"name": "Org2 Secret Framework (read-gap probe)"})

	injectCrossTenantFK(t, "audits", "framework_id", auditID, org2FrameworkID)

	var readResult struct {
		Node struct {
			Framework *struct {
				ID string `json:"id"`
			} `json:"framework"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Audit {
					framework { id }
				}
			}
		}
	`, map[string]any{"id": auditID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.Framework == nil, "cross-tenant framework via audit.framework")
}

// TestSecurity_ReadGap_ApplicabilityStatementControl independently verifies
// the defense-in-depth fix for applicabilityStatementResolver.Control, which
// now authorizes obj.Control.ID instead of obj.ID.
func TestSecurity_ReadGap_ApplicabilityStatementControl(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	soaID := factory.NewStatementOfApplicability(org1Owner).Create()
	org1FrameworkID := factory.CreateFramework(org1Owner)
	org1ControlID := factory.CreateControl(org1Owner, org1FrameworkID)
	asID := factory.CreateApplicabilityStatement(org1Owner, soaID, org1ControlID, true, nil)

	org2FrameworkID := factory.CreateFramework(org2Owner)
	org2ControlID := factory.CreateControl(org2Owner, org2FrameworkID, factory.Attrs{"name": "Org2 Secret Control (read-gap probe)"})

	injectCrossTenantFK(t, "applicability_statements", "control_id", asID, org2ControlID)

	// ApplicabilityStatement is not reachable via node(id:...) directly (a
	// separate, unrelated gap: coredata.ApplicabilityStatementEntityType has
	// no case in queryResolver.Node's dispatch switch, so it always denies
	// with an empty action regardless of tenant). Reach the same
	// applicabilityStatementResolver.Control sub-resolver through
	// StatementOfApplicability.applicabilityStatements instead, which does
	// implement Node.
	var readResult struct {
		Node struct {
			ApplicabilityStatements struct {
				Edges []struct {
					Node struct {
						ID      string `json:"id"`
						Control *struct {
							ID string `json:"id"`
						} `json:"control"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"applicabilityStatements"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on StatementOfApplicability {
					applicabilityStatements(first: 10) {
						edges {
							node {
								id
								control { id }
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": soaID}, &readResult)

	// control: Control! is non-nullable, so a denial on that field nulls the
	// whole ancestor chain up to the query's nullable "node" root -- err != nil
	// with no edges IS the expected block, same as AssertNodeNotAccessible's
	// contract elsewhere in this suite.
	if err == nil {
		var got *struct {
			ID string `json:"id"`
		}
		for _, edge := range readResult.Node.ApplicabilityStatements.Edges {
			if edge.Node.ID == asID {
				got = edge.Node.Control
			}
		}

		require.Nil(t, got, "must not be able to read a cross-tenant control via applicabilityStatement.control")
	}
}

// TestSecurity_ReadGap_DocumentVersionSignatureSignedBy independently
// verifies the defense-in-depth fix for documentVersionSignatureResolver.SignedBy,
// which now authorizes obj.SignedBy.ID instead of obj.ID.
func TestSecurity_ReadGap_DocumentVersionSignatureSignedBy(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, org1Owner)
	approveTestDocument(t, org1Owner, docID)
	versionID := latestDocumentVersionID(t, org1Owner, docID)
	org1ProfileID := getOwnerProfileID(t, org1Owner)

	var sigResult struct {
		RequestSignature struct {
			DocumentVersionSignatureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentVersionSignatureEdge"`
		} `json:"requestSignature"`
	}

	err := org1Owner.Execute(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
			"signatoryId":       org1ProfileID,
		},
	}, &sigResult)
	require.NoError(t, err)

	signatureID := sigResult.RequestSignature.DocumentVersionSignatureEdge.Node.ID

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Signatory (read-gap probe)"})

	injectCrossTenantFK(t, "document_version_signatures", "signed_by_profile_id", signatureID, org2ProfileID)

	var readResult struct {
		Node struct {
			SignedBy *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"signedBy"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersionSignature {
					signedBy { id fullName }
				}
			}
		}
	`, map[string]any{"id": signatureID}, &readResult)

	testutil.AssertNodeNotAccessible(t, err, readResult.Node.SignedBy == nil, "cross-tenant profile PII via documentVersionSignature.signedBy")
}

// TestSecurity_ReadGap_DocumentVersionApprovalDecisionApprover independently
// verifies a read-gap found while auditing GHSA-c74x-79w6-63jh's blast
// radius: documentVersionApprovalDecisionResolver.Approver authorized
// obj.ID (the decision, caller's own org) instead of obj.Approver.ID, and
// the profile loader it called into (iam.OrganizationService.GetProfile)
// derives its scope from coredata.NewScopeFromObjectID(profileID) -- the
// exact same scope-from-attacker-key mechanism as the two confirmed GHSA
// exploits. Not independently exploitable today because the write path
// (DocumentApprovalService.createDecisions, reached via
// validateApproverProfileIDs in document_service.go/generated_document_service.go)
// already rejects a cross-tenant approverId, but this proves the read-side
// fix (authorizing obj.Approver.ID) holds on its own.
func TestSecurity_ReadGap_DocumentVersionApprovalDecisionApprover(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, org1Owner)
	org1ProfileID := getOwnerProfileID(t, org1Owner)
	requestDocumentApproval(t, org1Owner, docID, []string{org1ProfileID})
	versionID := latestDocumentVersionID(t, org1Owner, docID)

	var quorumResult struct {
		Node struct {
			ApprovalQuorums struct {
				Edges []struct {
					Node struct {
						ID        string `json:"id"`
						Decisions struct {
							Edges []struct {
								Node struct {
									ID string `json:"id"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"decisions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"approvalQuorums"`
		} `json:"node"`
	}

	err := org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					approvalQuorums(first: 10) {
						edges {
							node {
								id
								decisions(first: 10) {
									edges { node { id } }
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": versionID}, &quorumResult)
	require.NoError(t, err)
	require.NotEmpty(t, quorumResult.Node.ApprovalQuorums.Edges)
	require.NotEmpty(t, quorumResult.Node.ApprovalQuorums.Edges[0].Node.Decisions.Edges)

	decisionID := quorumResult.Node.ApprovalQuorums.Edges[0].Node.Decisions.Edges[0].Node.ID

	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Secret Approver Profile"})

	injectCrossTenantFK(t, "document_version_approval_decisions", "approver_id", decisionID, org2ProfileID)

	var readResult struct {
		Node struct {
			ApprovalQuorums struct {
				Edges []struct {
					Node struct {
						Decisions struct {
							Edges []struct {
								Node struct {
									ID       string `json:"id"`
									Approver *struct {
										ID       string `json:"id"`
										FullName string `json:"fullName"`
									} `json:"approver"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"decisions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"approvalQuorums"`
		} `json:"node"`
	}

	err = org1Owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					approvalQuorums(first: 10) {
						edges {
							node {
								decisions(first: 10) {
									edges { node { id approver { id fullName } } }
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": versionID}, &readResult)

	// approver: Profile! is non-nullable, so a denial on that field nulls the
	// whole ancestor chain up to the query's nullable "node" root -- err !=
	// nil with no edges IS the expected block.
	if err == nil {
		var got *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		}

		for _, qEdge := range readResult.Node.ApprovalQuorums.Edges {
			for _, dEdge := range qEdge.Node.Decisions.Edges {
				if dEdge.Node.ID == decisionID {
					got = dEdge.Node.Approver
				}
			}
		}

		require.Nil(t, got, "must not be able to read a cross-tenant profile via documentVersionApprovalDecision.approver")
	}
}
