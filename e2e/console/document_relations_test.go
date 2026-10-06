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

func TestDocument_Relations(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	cases := []struct {
		name string
		run  func(*testing.T, *testutil.Client)
	}{
		{name: "MappedControlsInternalControlsRisks", run: documentRelationsMappedControlsInternalControlsRisks},
		{name: "OrganizationDefaultApproversPermission", run: documentRelationsOrganizationDefaultApproversPermission},
		{name: "VersionSignatureAndApprovalCounts", run: documentRelationsVersionSignatureAndApprovalCounts},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				tc.run(t, owner.ForTest(t))
			},
		)
	}
}

func linkControlToDocument(
	t *testing.T,
	owner *testutil.Client,
	controlID, documentID string,
) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: CreateControlDocumentMappingInput!) {
			createControlDocumentMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId":  controlID,
			"documentId": documentID,
		},
	})
	require.NoError(t, err)
}

func linkInternalControlToDocument(
	t *testing.T,
	owner *testutil.Client,
	internalControlID, documentID string,
) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: CreateInternalControlDocumentMappingInput!) {
			createInternalControlDocumentMapping(input: $input) {
				internalControlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"internalControlId": internalControlID,
			"documentId":        documentID,
		},
	})
	require.NoError(t, err)
}

func linkRiskToInternalControl(
	t *testing.T,
	owner *testutil.Client,
	riskID, internalControlID string,
) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: CreateRiskInternalControlMappingInput!) {
			createRiskInternalControlMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":            riskID,
			"internalControlId": internalControlID,
		},
	})
	require.NoError(t, err)
}

func linkRiskToDocument(
	t *testing.T,
	owner *testutil.Client,
	riskID, documentID string,
) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": documentID,
		},
	})
	require.NoError(t, err)
}

func documentRelationsMappedControlsInternalControlsRisks(t *testing.T, owner *testutil.Client) {
	frameworkID := factory.NewFramework(owner).
		WithName(factory.SafeName("Relations framework")).
		Create()
	controlID := factory.NewControl(owner, frameworkID).
		WithName(factory.SafeName("Relations control")).
		WithDescription("Control for document relations").
		WithSectionTitle("Section").
		WithBestPractice(true).
		WithMaturityLevel("INITIAL").
		Create()
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("Relations internal control")).
		WithCategory("POLICY").
		Create()
	riskID := factory.NewRisk(owner).
		WithName(factory.SafeName("Relations risk")).
		WithCategory("OPERATIONAL").
		Create()

	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Relations document")).
		Create()

	linkControlToDocument(t, owner, controlID, documentID)
	linkInternalControlToDocument(t, owner, internalControlID, documentID)
	linkRiskToDocument(t, owner, riskID, documentID)

	var documentResult struct {
		Node struct {
			Controls struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"controls"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					controls(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": documentID}, &documentResult)
	require.NoError(t, err)
	assert.Equal(t, 1, documentResult.Node.Controls.TotalCount)
	require.Len(t, documentResult.Node.Controls.Edges, 1)
	assert.Equal(t, controlID, documentResult.Node.Controls.Edges[0].Node.ID)

	var internalControlResult struct {
		Node struct {
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

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on InternalControl {
					documents(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": internalControlID}, &internalControlResult)
	require.NoError(t, err)
	assert.Equal(t, 1, internalControlResult.Node.Documents.TotalCount)
	require.Len(t, internalControlResult.Node.Documents.Edges, 1)
	assert.Equal(t, documentID, internalControlResult.Node.Documents.Edges[0].Node.ID)

	var riskResult struct {
		Node struct {
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

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Risk {
					documents(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": riskID}, &riskResult)
	require.NoError(t, err)
	assert.Equal(t, 1, riskResult.Node.Documents.TotalCount)
	require.Len(t, riskResult.Node.Documents.Edges, 1)
	assert.Equal(t, documentID, riskResult.Node.Documents.Edges[0].Node.ID)
}

func documentRelationsOrganizationDefaultApproversPermission(t *testing.T, owner *testutil.Client) {
	approverID := getOwnerProfileID(t, owner)
	orgID := owner.GetOrganizationID().String()

	documentID := createTestDocumentWithApprovers(t, owner, []string{approverID})

	var result struct {
		Node struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			DefaultApprovers []struct {
				ID string `json:"id"`
			} `json:"defaultApprovers"`
			CanUpdate  bool `json:"canUpdate"`
			CanArchive bool `json:"canArchive"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					organization { id }
					defaultApprovers { id }
					canUpdate: permission(action: "core:document:update")
					canArchive: permission(action: "core:document:archive")
				}
			}
		}
	`, map[string]any{"id": documentID}, &result)
	require.NoError(t, err)

	assert.Equal(t, orgID, result.Node.Organization.ID)
	require.Len(t, result.Node.DefaultApprovers, 1)
	assert.Equal(t, approverID, result.Node.DefaultApprovers[0].ID)
	assert.True(t, result.Node.CanUpdate)
	assert.True(t, result.Node.CanArchive)
}

func documentRelationsVersionSignatureAndApprovalCounts(t *testing.T, owner *testutil.Client) {
	docID, _ := createTestDocument(t, owner)
	approverID := getOwnerProfileID(t, owner)

	requestDocumentApproval(t, owner, docID, []string{approverID})
	pendingVersionID := latestDocumentVersionID(t, owner, docID)

	var pendingResult struct {
		Node struct {
			ApprovalQuorums struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						Decisions struct {
							TotalCount int `json:"totalCount"`
						} `json:"decisions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"approvalQuorums"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					approvalQuorums(first: 5) {
						totalCount
						edges {
							node {
								decisions(first: 10) {
									totalCount
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": pendingVersionID}, &pendingResult)
	require.NoError(t, err)
	assert.Equal(t, 1, pendingResult.Node.ApprovalQuorums.TotalCount)
	require.NotEmpty(t, pendingResult.Node.ApprovalQuorums.Edges)
	assert.Equal(t, 1, pendingResult.Node.ApprovalQuorums.Edges[0].Node.Decisions.TotalCount)

	approveLatestDocumentVersion(t, owner, docID)
	publishedVersionID := latestDocumentVersionID(t, owner, docID)

	requestDocumentSignature(t, owner, publishedVersionID, approverID)

	var signatureResult struct {
		Node struct {
			Signatures struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						State string `json:"state"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"signatures"`
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					document { id }
					signatures(first: 5) {
						totalCount
						edges { node { state } }
					}
				}
			}
		}
	`, map[string]any{"id": publishedVersionID}, &signatureResult)
	require.NoError(t, err)
	assert.Equal(t, docID, signatureResult.Node.Document.ID)
	assert.Equal(t, 1, signatureResult.Node.Signatures.TotalCount)
	require.Len(t, signatureResult.Node.Signatures.Edges, 1)
	assert.Equal(t, "REQUESTED", signatureResult.Node.Signatures.Edges[0].Node.State)
}
