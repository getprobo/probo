// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmployeeDocument_NodeAccess(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	docID, _ := createTestDocument(t, owner)

	t.Run("employee cannot list documents", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query ListDocuments($orgId: ID!) {
				node(id: $orgId) {
					... on Organization {
						documents(first: 10) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()})
		testutil.RequireForbiddenError(t, err, "employee should not list documents")
	})

	t.Run("employee cannot access document via node query", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query GetDocument($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						title
					}
				}
			}
		`, map[string]any{"id": docID})
		testutil.RequireForbiddenError(t, err, "employee should not access document via node")
	})

	t.Run("owner can access document via node query", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			query GetDocument($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						title
					}
				}
			}
		`, map[string]any{"id": docID})
		require.NoError(t, err, "owner should access document via node")
	})

	t.Run("employee cannot access signableDocument they are not signer of", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query($docId: ID!) {
				viewer {
					signableDocument(id: $docId) {
						id
					}
				}
			}
		`, map[string]any{"docId": docID})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "employee should not access signableDocument for doc they are not signer of")
	})

	t.Run("employee cannot access approvableDocument they are not approver of", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query($docId: ID!) {
				viewer {
					approvableDocument(id: $docId) {
						id
					}
				}
			}
		`, map[string]any{"docId": docID})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "employee should not access approvableDocument for doc they are not approver of")
	})
}

func TestEmployeeDocument_ExportPDF(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	_, docVersionID := createTestDocument(t, owner)

	t.Run("employee cannot use exportDocumentVersionPDF", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			mutation ExportPDF($input: ExportDocumentVersionPDFInput!) {
				exportDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": docVersionID,
				"withWatermark":     false,
				"withSignatures":    false,
			},
		})
		testutil.RequireForbiddenError(t, err, "employee should not use exportDocumentVersionPDF")
	})

	t.Run("owner can use exportDocumentVersionPDF", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation ExportPDF($input: ExportDocumentVersionPDFInput!) {
				exportDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": docVersionID,
				"withWatermark":     false,
				"withSignatures":    false,
			},
		})
		require.NoError(t, err, "owner should use exportDocumentVersionPDF")
	})

	t.Run("employee cannot use exportEmployeeDocumentVersionPDF without being signer", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			mutation ExportEmployeePDF($input: ExportEmployeeDocumentVersionPDFInput!) {
				exportEmployeeDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": docVersionID,
			},
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "employee should not export PDF for document they are not signer of")
	})
}

func TestEmployeeDocument_SignableDocuments(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	t.Run("employee can list signable documents", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				SignableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"signableDocuments"`
			} `json:"viewer"`
		}

		err := employee.Execute(`
			query($orgId: ID!) {
				viewer {
					signableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()}, &result)
		require.NoError(t, err, "employee should list signable documents")
	})

	t.Run("employee can list approvable documents", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				ApprovableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"approvableDocuments"`
			} `json:"viewer"`
		}

		err := employee.Execute(`
			query($orgId: ID!) {
				viewer {
					approvableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()}, &result)
		require.NoError(t, err, "employee should list approvable documents")
	})

	t.Run("employee does not see document they are not signer of in signableDocuments", func(t *testing.T) {
		t.Parallel()

		// Create a document with signature requested from owner only
		docID, _ := createTestDocument(t, owner)
		approveTestDocument(t, owner, docID)

		var versionResult struct {
			Node struct {
				Versions struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"versions"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": docID}, &versionResult)
		require.NoError(t, err)
		require.NotEmpty(t, versionResult.Node.Versions.Edges)

		ownerProfileID := owner.GetProfileID().String()
		_, err = owner.Do(`
			mutation($input: RequestSignatureInput!) {
				requestSignature(input: $input) {
					documentVersionSignatureEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": versionResult.Node.Versions.Edges[0].Node.ID,
				"signatoryId":       ownerProfileID,
			},
		})
		require.NoError(t, err)

		// Employee should not see this document
		var listResult struct {
			Viewer struct {
				SignableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"signableDocuments"`
			} `json:"viewer"`
		}

		err = employee.Execute(`
			query($orgId: ID!) {
				viewer {
					signableDocuments(organizationId: $orgId, first: 100) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()}, &listResult)
		require.NoError(t, err)

		for _, edge := range listResult.Viewer.SignableDocuments.Edges {
			assert.NotEqual(t, docID, edge.Node.ID, "employee should not see document they are not signer of")
		}
	})

	t.Run("employee does not see document they are not approver of in approvableDocuments", func(t *testing.T) {
		t.Parallel()

		// Create a document with approval requested from owner only
		docID, _ := createTestDocument(t, owner)
		ownerProfileID := owner.GetProfileID().String()

		_, err := owner.Do(`
			mutation($input: RequestDocumentVersionApprovalInput!) {
				requestDocumentVersionApproval(input: $input) {
					approvalQuorum { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId":  docID,
				"approverIds": []string{ownerProfileID},
				"changelog":   "Test changelog",
			},
		})
		require.NoError(t, err)

		// Employee should not see this document
		var listResult struct {
			Viewer struct {
				ApprovableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"approvableDocuments"`
			} `json:"viewer"`
		}

		err = employee.Execute(`
			query($orgId: ID!) {
				viewer {
					approvableDocuments(organizationId: $orgId, first: 100) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()}, &listResult)
		require.NoError(t, err)

		for _, edge := range listResult.Viewer.ApprovableDocuments.Edges {
			assert.NotEqual(t, docID, edge.Node.ID, "employee should not see document they are not approver of")
		}
	})
}

func TestEmployeeDocument_FilterModeIsolation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	// Create a document and have the admin approve it (so owner is NOT an approver)
	docID, _ := createTestDocument(t, owner)

	adminProfileID := admin.GetProfileID().String()

	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{adminProfileID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	// Get the version ID created by the approval request
	var approveVersionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &approveVersionResult)
	require.NoError(t, err)
	require.NotEmpty(t, approveVersionResult.Node.Versions.Edges)

	pendingVersionID := approveVersionResult.Node.Versions.Edges[0].Node.ID

	_, err = admin.Do(`
		mutation($input: ApproveDocumentVersionInput!) {
			approveDocumentVersion(input: $input) {
				approvalDecision { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": pendingVersionID,
		},
	})
	require.NoError(t, err)

	// Get the published version ID
	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	publishedVersionID := versionResult.Node.Versions.Edges[0].Node.ID
	ownerProfileID := owner.GetProfileID().String()

	// Request signature from the owner only (owner is a signer but NOT an approver)
	_, err = owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge {
					node { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
			"signatoryId":       ownerProfileID,
		},
	})
	require.NoError(t, err)

	t.Run("signer sees document in signableDocuments list", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				SignableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"signableDocuments"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($orgId: ID!) {
				viewer {
					signableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": owner.GetOrganizationID().String()}, &result)
		require.NoError(t, err)

		var found bool
		for _, edge := range result.Viewer.SignableDocuments.Edges {
			if edge.Node.ID == docID {
				found = true
				break
			}
		}
		assert.True(t, found, "signer should see document in signableDocuments list")
	})

	t.Run("signer does not see document in approvableDocuments list", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				ApprovableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"approvableDocuments"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($orgId: ID!) {
				viewer {
					approvableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": owner.GetOrganizationID().String()}, &result)
		require.NoError(t, err)

		for _, edge := range result.Viewer.ApprovableDocuments.Edges {
			assert.NotEqual(t, docID, edge.Node.ID, "signer-only document should not appear in approvableDocuments list")
		}
	})

	t.Run("signer cannot access document via approvableDocument", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			query($docId: ID!) {
				viewer {
					approvableDocument(id: $docId) {
						id
					}
				}
			}
		`, map[string]any{"docId": docID})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "signer-only should not access approvableDocument")
	})
}

func TestEmployeeDocument_ApproverFilterModeIsolation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a document and request approval (owner is approver but NOT a signer)
	docID, _ := createTestDocument(t, owner)
	ownerProfileID := owner.GetProfileID().String()

	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{ownerProfileID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	t.Run("approver sees document in approvableDocuments list", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				ApprovableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"approvableDocuments"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($orgId: ID!) {
				viewer {
					approvableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": owner.GetOrganizationID().String()}, &result)
		require.NoError(t, err)

		var found bool
		for _, edge := range result.Viewer.ApprovableDocuments.Edges {
			if edge.Node.ID == docID {
				found = true
				break
			}
		}
		assert.True(t, found, "approver should see document in approvableDocuments list")
	})

	t.Run("approver does not see document in signableDocuments list", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				SignableDocuments struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"signableDocuments"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($orgId: ID!) {
				viewer {
					signableDocuments(organizationId: $orgId, first: 10) {
						edges { node { id } }
					}
				}
			}
		`, map[string]any{"orgId": owner.GetOrganizationID().String()}, &result)
		require.NoError(t, err)

		for _, edge := range result.Viewer.SignableDocuments.Edges {
			assert.NotEqual(t, docID, edge.Node.ID, "approver-only document should not appear in signableDocuments list")
		}
	})

	t.Run("approver cannot access document via signableDocument", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			query($docId: ID!) {
				viewer {
					signableDocument(id: $docId) {
						id
					}
				}
			}
		`, map[string]any{"docId": docID})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "approver-only should not access signableDocument")
	})
}

func TestEmployeeDocument_UnsignedDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a document, approve and publish it, then request signature from owner
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	publishedVersionID := versionResult.Node.Versions.Edges[0].Node.ID
	ownerProfileID := owner.GetProfileID().String()

	_, err = owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
			"signatoryId":       ownerProfileID,
		},
	})
	require.NoError(t, err)

	t.Run("unsigned document shows signed=false on version", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				SignableDocument struct {
					ID       string `json:"id"`
					Signed   *bool  `json:"signed"`
					Versions struct {
						Edges []struct {
							Node struct {
								ID     string `json:"id"`
								Signed bool   `json:"signed"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"versions"`
				} `json:"signableDocument"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($docId: ID!) {
				viewer {
					signableDocument(id: $docId) {
						id
						signed
						versions(first: 10) {
							edges {
								node {
									id
									signed
								}
							}
						}
					}
				}
			}
		`, map[string]any{"docId": docID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Viewer.SignableDocument.Signed)
		assert.False(t, *result.Viewer.SignableDocument.Signed, "document should not be signed yet")
		require.NotEmpty(t, result.Viewer.SignableDocument.Versions.Edges)
		assert.False(t, result.Viewer.SignableDocument.Versions.Edges[0].Node.Signed, "version should not be signed yet")
	})

	t.Run("unsigned document still allows exportEmployeeDocumentVersionPDF", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
				exportEmployeeDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": publishedVersionID,
			},
		})
		require.NoError(t, err, "signer should export PDF even before signing")
	})
}

func TestEmployeeDocument_UnapprovedDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a document and request approval from owner (don't approve yet)
	docID, _ := createTestDocument(t, owner)
	ownerProfileID := owner.GetProfileID().String()

	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{ownerProfileID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	t.Run("unapproved document shows PENDING state", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				ApprovableDocument struct {
					ID            string  `json:"id"`
					ApprovalState *string `json:"approvalState"`
					Versions      struct {
						Edges []struct {
							Node struct {
								ID               string `json:"id"`
								ApprovalDecision *struct {
									ID    string `json:"id"`
									State string `json:"state"`
								} `json:"approvalDecision"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"versions"`
				} `json:"approvableDocument"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($docId: ID!) {
				viewer {
					approvableDocument(id: $docId) {
						id
						approvalState
						versions(first: 10) {
							edges {
								node {
									id
									approvalDecision {
										id
										state
									}
								}
							}
						}
					}
				}
			}
		`, map[string]any{"docId": docID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Viewer.ApprovableDocument.ApprovalState)
		assert.Equal(t, "PENDING", *result.Viewer.ApprovableDocument.ApprovalState, "approval state should be PENDING")
		require.NotEmpty(t, result.Viewer.ApprovableDocument.Versions.Edges)
		require.NotNil(t, result.Viewer.ApprovableDocument.Versions.Edges[0].Node.ApprovalDecision)
		assert.Equal(t, "PENDING", result.Viewer.ApprovableDocument.Versions.Edges[0].Node.ApprovalDecision.State, "decision should be PENDING")
	})

	t.Run("unapproved document still allows exportEmployeeDocumentVersionPDF", func(t *testing.T) {
		t.Parallel()

		var versionResult struct {
			Node struct {
				Versions struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"versions"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": docID}, &versionResult)
		require.NoError(t, err)
		require.NotEmpty(t, versionResult.Node.Versions.Edges)

		_, err = owner.Do(`
			mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
				exportEmployeeDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": versionResult.Node.Versions.Edges[0].Node.ID,
			},
		})
		require.NoError(t, err, "approver should export PDF even before approving")
	})
}

func TestEmployeeDocument_SignableDocumentNestedFields(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a document, approve and publish it
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Get the published version ID
	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	publishedVersionID := versionResult.Node.Versions.Edges[0].Node.ID

	// Request signature from the owner
	ownerProfileID := owner.GetProfileID().String()

	_, err = owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge {
					node { id state }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
			"signatoryId":       ownerProfileID,
		},
	})
	require.NoError(t, err)

	t.Run("owner can access signableDocument with nested fields", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				SignableDocument struct {
					ID       string `json:"id"`
					Title    string `json:"title"`
					Signed   *bool  `json:"signed"`
					Versions struct {
						Edges []struct {
							Node struct {
								ID     string `json:"id"`
								Signed bool   `json:"signed"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"versions"`
				} `json:"signableDocument"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($docId: ID!) {
				viewer {
					signableDocument(id: $docId) {
						id
						title
						signed
						versions(first: 10) {
							edges {
								node {
									id
									signed
								}
							}
						}
					}
				}
			}
		`, map[string]any{"docId": docID}, &result)
		require.NoError(t, err, "owner should access signableDocument with nested fields")
		assert.NotEmpty(t, result.Viewer.SignableDocument.ID)
		assert.NotNil(t, result.Viewer.SignableDocument.Signed)
		assert.NotEmpty(t, result.Viewer.SignableDocument.Versions.Edges)
	})

	t.Run("owner can use exportEmployeeDocumentVersionPDF as signer", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
				exportEmployeeDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": publishedVersionID,
			},
		})
		require.NoError(t, err, "owner signer should use exportEmployeeDocumentVersionPDF")
	})
}

func TestEmployeeDocument_ApprovableDocumentNestedFields(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a document and request approval with the owner as approver
	docID, _ := createTestDocument(t, owner)
	ownerProfileID := owner.GetProfileID().String()

	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{ownerProfileID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	t.Run("owner can access approvableDocument with nested fields", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Viewer struct {
				ApprovableDocument struct {
					ID            string  `json:"id"`
					Title         string  `json:"title"`
					ApprovalState *string `json:"approvalState"`
					Versions      struct {
						Edges []struct {
							Node struct {
								ID               string `json:"id"`
								ApprovalDecision *struct {
									ID    string `json:"id"`
									State string `json:"state"`
								} `json:"approvalDecision"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"versions"`
				} `json:"approvableDocument"`
			} `json:"viewer"`
		}

		err := owner.Execute(`
			query($docId: ID!) {
				viewer {
					approvableDocument(id: $docId) {
						id
						title
						approvalState
						versions(first: 10) {
							edges {
								node {
									id
									approvalDecision {
										id
										state
									}
								}
							}
						}
					}
				}
			}
		`, map[string]any{"docId": docID}, &result)
		require.NoError(t, err, "owner should access approvableDocument with nested fields")
		assert.NotEmpty(t, result.Viewer.ApprovableDocument.ID)
		assert.NotEmpty(t, result.Viewer.ApprovableDocument.Versions.Edges)
	})

	t.Run("owner can use exportEmployeeDocumentVersionPDF as approver", func(t *testing.T) {
		t.Parallel()

		// Get the version ID
		var versionResult struct {
			Node struct {
				Versions struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"versions"`
			} `json:"node"`
		}

		err := owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"id": docID}, &versionResult)
		require.NoError(t, err)
		require.NotEmpty(t, versionResult.Node.Versions.Edges)

		versionID := versionResult.Node.Versions.Edges[0].Node.ID

		_, err = owner.Do(`
			mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
				exportEmployeeDocumentVersionPDF(input: $input) {
					data
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentVersionId": versionID,
			},
		})
		require.NoError(t, err, "owner approver should use exportEmployeeDocumentVersionPDF")
	})
}
