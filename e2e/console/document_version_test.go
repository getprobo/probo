// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func getOwnerProfileID(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	return owner.GetProfileID().String()
}

// createTestDocument creates a document and returns its ID and the document version ID
func createTestDocument(t *testing.T, owner *testutil.Client) (docID string, docVersionID string) {
	t.Helper()

	query := `
		mutation CreateDocument($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node {
						id
						versions(first: 1) {
							edges {
								node {
									id
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID       string `json:"id"`
					Versions struct {
						Edges []struct {
							Node struct {
								ID string `json:"id"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"versions"`
				} `json:"node"`
			} `json:"documentEdge"`
		} `json:"createDocument"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"title":          "Test Document",
			"content":        testutil.ProseMirrorTextDoc("Initial content"),
			"documentType":   "POLICY",
			"classification": "INTERNAL",
		},
	}, &result)
	require.NoError(t, err)

	docID = result.CreateDocument.DocumentEdge.Node.ID
	if len(result.CreateDocument.DocumentEdge.Node.Versions.Edges) > 0 {
		docVersionID = result.CreateDocument.DocumentEdge.Node.Versions.Edges[0].Node.ID
	}
	return docID, docVersionID
}

// approveTestDocument requests approval and approves the document so it can be published.
func approveTestDocument(t *testing.T, owner *testutil.Client, docID string) {
	t.Helper()

	requestQuery := `
		mutation RequestApproval($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum {
					id
				}
			}
		}
	`

	// Use the owner's profile as the approver
	approverID := getOwnerProfileID(t, owner)

	_, err := owner.Do(requestQuery, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	// Approve for each approver
	approveQuery := `
		mutation ApproveDocumentVersion($input: ApproveDocumentVersionInput!) {
			approveDocumentVersion(input: $input) {
				approvalDecision {
					id
					state
				}
			}
		}
	`

	// Get the latest version ID
	versionQuery := `
		query GetVersions($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`

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

	err = owner.Execute(versionQuery, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	versionID := versionResult.Node.Versions.Edges[0].Node.ID

	_, err = owner.Do(approveQuery, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
		},
	})
	require.NoError(t, err)
}

func TestDocumentVersion_PublishVersion(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// After approval, the version is auto-published.
	// Verify the version status by querying.
	query := `
		query GetDocument($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges {
							node {
								id
								status
								major
								minor
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Status string `json:"status"`
						Major  int    `json:"major"`
						Minor  int    `json:"minor"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": docID}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.Versions.Edges)

	assert.Equal(t, "PUBLISHED", result.Node.Versions.Edges[0].Node.Status)
	assert.Equal(t, 1, result.Node.Versions.Edges[0].Node.Major)
	assert.Equal(t, 0, result.Node.Versions.Edges[0].Node.Minor)
}

func TestDocumentVersion_AutoCreateDraft(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Updating content should auto-create a draft
	query := `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					id
				}
				documentVersion {
					id
					status
					content
				}
			}
		}
	`

	var result struct {
		UpdateDocument struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
			DocumentVersion *struct {
				ID      string `json:"id"`
				Status  string `json:"status"`
				Content string `json:"content"`
			} `json:"documentVersion"`
		} `json:"updateDocument"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Updated content"),
		},
	}, &result)
	require.NoError(t, err)

	require.NotNil(t, result.UpdateDocument.DocumentVersion)
	assert.Equal(t, "DRAFT", result.UpdateDocument.DocumentVersion.Status)
}

func TestDocumentVersion_AutoDeleteDraft(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// First update to create a draft
	query := `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					id
				}
				documentVersion {
					id
					status
				}
			}
		}
	`

	var createResult struct {
		UpdateDocument struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
			DocumentVersion *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentVersion"`
		} `json:"updateDocument"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Updated content"),
		},
	}, &createResult)
	require.NoError(t, err)
	require.NotNil(t, createResult.UpdateDocument.DocumentVersion)

	// Now revert content to match the published version — draft should be auto-deleted
	var revertResult struct {
		UpdateDocument struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
			DocumentVersion *struct {
				ID string `json:"id"`
			} `json:"documentVersion"`
		} `json:"updateDocument"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Initial content"),
		},
	}, &revertResult)
	require.NoError(t, err)

	assert.Nil(t, revertResult.UpdateDocument.DocumentVersion)
}

func TestDocumentVersion_RequestSignature(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Get the published version ID
	versionQuery := `
		query GetVersions($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`

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

	err := owner.Execute(versionQuery, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	publishedVersionID := versionResult.Node.Versions.Edges[0].Node.ID

	// Create a person to sign
	signerProfileID := factory.CreateUser(owner)

	query := `
		mutation RequestSignature($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge {
					node {
						id
						state
						signedBy {
							id
							fullName
						}
					}
				}
			}
		}
	`

	var result struct {
		RequestSignature struct {
			DocumentVersionSignatureEdge struct {
				Node struct {
					ID       string `json:"id"`
					State    string `json:"state"`
					SignedBy struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"signedBy"`
				} `json:"node"`
			} `json:"documentVersionSignatureEdge"`
		} `json:"requestSignature"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
			"signatoryId":       signerProfileID,
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.RequestSignature.DocumentVersionSignatureEdge.Node.ID)
	assert.Equal(t, "REQUESTED", result.RequestSignature.DocumentVersionSignatureEdge.Node.State)
	assert.Equal(t, signerProfileID, result.RequestSignature.DocumentVersionSignatureEdge.Node.SignedBy.ID)
}

func createTestDocumentWithApprovers(t *testing.T, owner *testutil.Client, approverIDs []string) (docID string) {
	t.Helper()

	var result struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentEdge"`
		} `json:"createDocument"`
	}

	err := owner.Execute(`
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId":     owner.GetOrganizationID().String(),
			"title":              "Test Document With Approvers",
			"content":            testutil.ProseMirrorTextDoc("Initial content"),
			"documentType":       "POLICY",
			"classification":     "INTERNAL",
			"defaultApproverIds": approverIDs,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateDocument.DocumentEdge.Node.ID
}

func TestDocumentVersion_BulkPublish(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple draft documents (no default approvers — should publish directly)
	docID1, _ := createTestDocument(t, owner)
	docID2, _ := createTestDocument(t, owner)

	query := `
		mutation BulkPublishMajorDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishMajorDocumentVersions(input: $input) {
				documentVersions {
					id
					status
				}
			}
		}
	`

	var result struct {
		BulkPublishMajorDocumentVersions struct {
			DocumentVersions []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentVersions"`
		} `json:"bulkPublishMajorDocumentVersions"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID1, docID2},
			"changelog":   "Bulk publish release",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.BulkPublishMajorDocumentVersions.DocumentVersions))
	for _, dv := range result.BulkPublishMajorDocumentVersions.DocumentVersions {
		assert.Equal(t, "PUBLISHED", dv.Status)
	}
}

func TestDocumentVersion_BulkPublishRequestsApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	approverID := getOwnerProfileID(t, owner)

	// Create a document with default approvers
	docID := createTestDocumentWithApprovers(t, owner, []string{approverID})

	query := `
		mutation BulkPublishMajorDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishMajorDocumentVersions(input: $input) {
				documentVersions {
					id
					status
					major
					minor
				}
			}
		}
	`

	var result struct {
		BulkPublishMajorDocumentVersions struct {
			DocumentVersions []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersions"`
		} `json:"bulkPublishMajorDocumentVersions"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID},
			"changelog":   "Needs approval",
		},
	}, &result)
	require.NoError(t, err)

	require.Len(t, result.BulkPublishMajorDocumentVersions.DocumentVersions, 1)
	dv := result.BulkPublishMajorDocumentVersions.DocumentVersions[0]
	assert.Equal(t, "PENDING_APPROVAL", dv.Status)
	assert.Equal(t, 1, dv.Major)
	assert.Equal(t, 0, dv.Minor)
}

func TestDocumentVersion_BulkPublishSkipsPendingApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	approverID := getOwnerProfileID(t, owner)

	// Create a document with default approvers and bulk publish it (puts it in PENDING_APPROVAL)
	docID := createTestDocumentWithApprovers(t, owner, []string{approverID})

	_, err := owner.Do(`
		mutation($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishMajorDocumentVersions(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID},
			"changelog":   "First approval request",
		},
	})
	require.NoError(t, err)

	// Bulk publish again — should skip the pending document and return empty
	var result struct {
		BulkPublishMajorDocumentVersions struct {
			DocumentVersions []struct {
				ID string `json:"id"`
			} `json:"documentVersions"`
		} `json:"bulkPublishMajorDocumentVersions"`
	}

	err = owner.Execute(`
		mutation($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishMajorDocumentVersions(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID},
			"changelog":   "Second attempt",
		},
	}, &result)
	require.NoError(t, err)

	assert.Empty(t, result.BulkPublishMajorDocumentVersions.DocumentVersions)
}

func TestDocumentVersion_BulkPublishMinorSkipsPendingApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and publish a document first (need a published version for minor publish)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Create a draft by updating content (auto-creates draft)
	_, err := owner.Do(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				documentVersion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Updated content to create a draft"),
		},
	})
	require.NoError(t, err)

	// Request approval to put it in PENDING_APPROVAL
	approverID := getOwnerProfileID(t, owner)
	_, err = owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Approval request",
		},
	})
	require.NoError(t, err)

	// Bulk publish minor — should skip the pending document
	var result struct {
		BulkPublishMinorDocumentVersions struct {
			DocumentVersions []struct {
				ID string `json:"id"`
			} `json:"documentVersions"`
		} `json:"bulkPublishMinorDocumentVersions"`
	}

	err = owner.Execute(`
		mutation($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishMinorDocumentVersions(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID},
			"changelog":   "Minor publish attempt",
		},
	}, &result)
	require.NoError(t, err)

	assert.Empty(t, result.BulkPublishMinorDocumentVersions.DocumentVersions)
}

func TestDocumentVersion_BulkRequestSignatures(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Create multiple signers
	signer1ProfileID := factory.CreateUser(owner)
	signer2ProfileID := factory.CreateUser(owner)

	query := `
		mutation BulkRequestSignatures($input: BulkRequestSignaturesInput!) {
			bulkRequestSignatures(input: $input) {
				documentVersionSignatureEdges {
					node {
						id
						state
					}
				}
			}
		}
	`

	var result struct {
		BulkRequestSignatures struct {
			DocumentVersionSignatureEdges []struct {
				Node struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"node"`
			} `json:"documentVersionSignatureEdges"`
		} `json:"bulkRequestSignatures"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds":  []string{docID},
			"signatoryIds": []string{signer1ProfileID, signer2ProfileID},
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.BulkRequestSignatures.DocumentVersionSignatureEdges))
	for _, edge := range result.BulkRequestSignatures.DocumentVersionSignatureEdges {
		assert.Equal(t, "REQUESTED", edge.Node.State)
	}
}

func TestDocumentVersion_AutoCreateDraftOnClassificationOrTypeUpdate(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("documentType update creates draft", func(t *testing.T) {
		t.Parallel()

		// Create and approve a document (auto-publishes on approval)
		docID, _ := createTestDocument(t, owner)
		approveTestDocument(t, owner, docID)

		// Updating documentType should auto-create a draft
		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
					}
					documentVersion {
						id
						status
					}
				}
			}
		`

		var result struct {
			UpdateDocument struct {
				Document struct {
					ID string `json:"id"`
				} `json:"document"`
				DocumentVersion *struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"documentVersion"`
			} `json:"updateDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":           docID,
				"documentType": "PROCEDURE",
			},
		}, &result)
		require.NoError(t, err)

		require.NotNil(t, result.UpdateDocument.DocumentVersion)
		assert.Equal(t, "DRAFT", result.UpdateDocument.DocumentVersion.Status)
	})

	t.Run("classification update creates draft", func(t *testing.T) {
		t.Parallel()

		// Create and approve a document (auto-publishes on approval)
		docID, _ := createTestDocument(t, owner)
		approveTestDocument(t, owner, docID)

		// Updating classification should auto-create a draft
		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
					}
					documentVersion {
						id
						status
					}
				}
			}
		`

		var result struct {
			UpdateDocument struct {
				Document struct {
					ID string `json:"id"`
				} `json:"document"`
				DocumentVersion *struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"documentVersion"`
			} `json:"updateDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":             docID,
				"classification": "CONFIDENTIAL",
			},
		}, &result)
		require.NoError(t, err)

		require.NotNil(t, result.UpdateDocument.DocumentVersion)
		assert.Equal(t, "DRAFT", result.UpdateDocument.DocumentVersion.Status)
	})
}

func TestDocumentVersion_ViewerCannotUpdateDocument(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Viewer attempts to update content on the published document
	_, err := viewer.Do(`
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Viewer updated content"),
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer should not be able to update document")
}

func TestDocumentVersion_BulkDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple documents to delete
	docID1, _ := createTestDocument(t, owner)
	docID2, _ := createTestDocument(t, owner)

	query := `
		mutation BulkDeleteDocuments($input: BulkDeleteDocumentsInput!) {
			bulkDeleteDocuments(input: $input) {
				deletedDocumentIds
			}
		}
	`

	var result struct {
		BulkDeleteDocuments struct {
			DeletedDocumentIds []string `json:"deletedDocumentIds"`
		} `json:"bulkDeleteDocuments"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID1, docID2},
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.BulkDeleteDocuments.DeletedDocumentIds))
	assert.Contains(t, result.BulkDeleteDocuments.DeletedDocumentIds, docID1)
	assert.Contains(t, result.BulkDeleteDocuments.DeletedDocumentIds, docID2)
}

func TestDocumentVersion_VoidApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approverID := getOwnerProfileID(t, owner)

	// Request approval
	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	// Get version ID and verify version bumped to 1.0
	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Status string `json:"status"`
						Major  int    `json:"major"`
						Minor  int    `json:"minor"`
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
						edges { node { id status major minor } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)
	assert.Equal(t, "PENDING_APPROVAL", versionResult.Node.Versions.Edges[0].Node.Status)
	assert.Equal(t, 1, versionResult.Node.Versions.Edges[0].Node.Major)
	assert.Equal(t, 0, versionResult.Node.Versions.Edges[0].Node.Minor)

	versionID := versionResult.Node.Versions.Edges[0].Node.ID

	// Void approval — version should revert to 0.1
	var voidResult struct {
		VoidDocumentVersionApproval struct {
			ApprovalQuorum struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"approvalQuorum"`
			DocumentVersion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersion"`
		} `json:"voidDocumentVersionApproval"`
	}

	err = owner.Execute(`
		mutation($input: VoidDocumentVersionApprovalInput!) {
			voidDocumentVersionApproval(input: $input) {
				approvalQuorum { id status }
				documentVersion { id status major minor }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
		},
	}, &voidResult)
	require.NoError(t, err)

	assert.Equal(t, "VOIDED", voidResult.VoidDocumentVersionApproval.ApprovalQuorum.Status)
	assert.Equal(t, "DRAFT", voidResult.VoidDocumentVersionApproval.DocumentVersion.Status)
	assert.Equal(t, 0, voidResult.VoidDocumentVersionApproval.DocumentVersion.Major)
	assert.Equal(t, 1, voidResult.VoidDocumentVersionApproval.DocumentVersion.Minor)

	// Verify decisions are VOIDED after voiding
	var quorumResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ApprovalQuorums struct {
							Edges []struct {
								Node struct {
									Decisions struct {
										Edges []struct {
											Node struct {
												State string `json:"state"`
											} `json:"node"`
										} `json:"edges"`
									} `json:"decisions"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"approvalQuorums"`
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
						edges {
							node {
								approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
									edges {
										node {
											decisions(first: 100) {
												edges { node { state } }
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &quorumResult)
	require.NoError(t, err)
	require.NotEmpty(t, quorumResult.Node.Versions.Edges)
	require.NotEmpty(t, quorumResult.Node.Versions.Edges[0].Node.ApprovalQuorums.Edges)
	decisions := quorumResult.Node.Versions.Edges[0].Node.ApprovalQuorums.Edges[0].Node.Decisions.Edges
	require.NotEmpty(t, decisions)
	for _, d := range decisions {
		assert.Equal(t, "VOIDED", d.Node.State, "decisions should be VOIDED after voiding")
	}
}

func TestDocumentVersion_RejectApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approverID := getOwnerProfileID(t, owner)

	// Request approval — version should bump to 1.0
	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	// Get version ID
	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Status string `json:"status"`
						Major  int    `json:"major"`
						Minor  int    `json:"minor"`
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
						edges { node { id status major minor } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)
	assert.Equal(t, "PENDING_APPROVAL", versionResult.Node.Versions.Edges[0].Node.Status)
	assert.Equal(t, 1, versionResult.Node.Versions.Edges[0].Node.Major)
	assert.Equal(t, 0, versionResult.Node.Versions.Edges[0].Node.Minor)

	versionID := versionResult.Node.Versions.Edges[0].Node.ID

	// Reject approval
	_, err = owner.Do(`
		mutation($input: RejectDocumentVersionInput!) {
			rejectDocumentVersion(input: $input) {
				approvalDecision { id state }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
			"comment":           "Needs rework",
		},
	})
	require.NoError(t, err)

	// Verify version reverted to 0.1 DRAFT
	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id status major minor } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)
	assert.Equal(t, "DRAFT", versionResult.Node.Versions.Edges[0].Node.Status)
	assert.Equal(t, 0, versionResult.Node.Versions.Edges[0].Node.Major)
	assert.Equal(t, 1, versionResult.Node.Versions.Edges[0].Node.Minor)

	// Verify decisions are VOIDED after reject
	var quorumResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ApprovalQuorums struct {
							Edges []struct {
								Node struct {
									Decisions struct {
										Edges []struct {
											Node struct {
												State string `json:"state"`
											} `json:"node"`
										} `json:"edges"`
									} `json:"decisions"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"approvalQuorums"`
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
						edges {
							node {
								approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
									edges {
										node {
											decisions(first: 100) {
												edges { node { state } }
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &quorumResult)
	require.NoError(t, err)
	require.NotEmpty(t, quorumResult.Node.Versions.Edges)
	require.NotEmpty(t, quorumResult.Node.Versions.Edges[0].Node.ApprovalQuorums.Edges)
	decisions := quorumResult.Node.Versions.Edges[0].Node.ApprovalQuorums.Edges[0].Node.Decisions.Edges
	require.Len(t, decisions, 1)
	assert.Equal(t, "REJECTED", decisions[0].Node.State, "rejecting approver's decision should be REJECTED")
}

func TestDocumentVersion_PublishBlockedWhenPendingApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approverID := getOwnerProfileID(t, owner)

	// Request approval (puts version in PENDING_APPROVAL)
	_, err := owner.Do(`
		mutation($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	t.Run("publish major blocked", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: PublishMajorDocumentVersionInput!) {
				publishMajorDocumentVersion(input: $input) {
					documentVersion { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId": docID,
				"changelog":  "Major release",
			},
		})
		require.Error(t, err)
	})

	t.Run("publish minor blocked", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: PublishMinorDocumentVersionInput!) {
				publishMinorDocumentVersion(input: $input) {
					documentVersion { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId": docID,
				"changelog":  "Minor release",
			},
		})
		require.Error(t, err)
	})
}

func TestDocument_DefaultApprovers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	approverID := getOwnerProfileID(t, owner)

	t.Run("create document with default approvers", func(t *testing.T) {
		t.Parallel()

		var result struct {
			CreateDocument struct {
				DocumentEdge struct {
					Node struct {
						ID               string `json:"id"`
						DefaultApprovers []struct {
							ID string `json:"id"`
						} `json:"defaultApprovers"`
					} `json:"node"`
				} `json:"documentEdge"`
			} `json:"createDocument"`
		}

		err := owner.Execute(`
			mutation($input: CreateDocumentInput!) {
				createDocument(input: $input) {
					documentEdge {
						node {
							id
							defaultApprovers { id }
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"title":              "Doc With Approvers",
				"content":            testutil.ProseMirrorTextDoc("Content"),
				"documentType":       "POLICY",
				"classification":     "INTERNAL",
				"defaultApproverIds": []string{approverID},
			},
		}, &result)
		require.NoError(t, err)

		assert.Len(t, result.CreateDocument.DocumentEdge.Node.DefaultApprovers, 1)
		assert.Equal(t, approverID, result.CreateDocument.DocumentEdge.Node.DefaultApprovers[0].ID)
	})

	t.Run("update document with default approvers", func(t *testing.T) {
		t.Parallel()

		docID, _ := createTestDocument(t, owner)

		var result struct {
			UpdateDocument struct {
				Document struct {
					ID               string `json:"id"`
					DefaultApprovers []struct {
						ID string `json:"id"`
					} `json:"defaultApprovers"`
				} `json:"document"`
			} `json:"updateDocument"`
		}

		err := owner.Execute(`
			mutation($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
						defaultApprovers { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":                 docID,
				"defaultApproverIds": []string{approverID},
			},
		}, &result)
		require.NoError(t, err)

		assert.Len(t, result.UpdateDocument.Document.DefaultApprovers, 1)
		assert.Equal(t, approverID, result.UpdateDocument.Document.DefaultApprovers[0].ID)

		// Clear approvers
		err = owner.Execute(`
			mutation($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
						defaultApprovers { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":                 docID,
				"defaultApproverIds": []string{},
			},
		}, &result)
		require.NoError(t, err)

		assert.Empty(t, result.UpdateDocument.Document.DefaultApprovers)
	})
}

func TestDocumentVersion_DeleteDraft(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		mutation DeleteDocumentDraft($input: DeleteDocumentDraftInput!) {
			deleteDocumentDraft(input: $input) {
				document {
					id
				}
			}
		}
	`

	t.Run(
		"delete draft after publishing",
		func(t *testing.T) {
			t.Parallel()

			docID, _ := createTestDocument(t, owner)
			approveTestDocument(t, owner, docID)

			// Create a draft by updating content
			updateQuery := `
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document { id }
						documentVersion { id status }
					}
				}
			`
			var updateResult struct {
				UpdateDocument struct {
					Document struct {
						ID string `json:"id"`
					} `json:"document"`
					DocumentVersion *struct {
						ID     string `json:"id"`
						Status string `json:"status"`
					} `json:"documentVersion"`
				} `json:"updateDocument"`
			}
			err := owner.Execute(updateQuery, map[string]any{
				"input": map[string]any{
					"id":      docID,
					"content": testutil.ProseMirrorTextDoc("Draft content"),
				},
			}, &updateResult)
			require.NoError(t, err)
			require.NotNil(t, updateResult.UpdateDocument.DocumentVersion)
			assert.Equal(t, "DRAFT", updateResult.UpdateDocument.DocumentVersion.Status)

			// Now delete the draft
			var result struct {
				DeleteDocumentDraft struct {
					Document struct {
						ID string `json:"id"`
					} `json:"document"`
				} `json:"deleteDocumentDraft"`
			}
			err = owner.Execute(query, map[string]any{
				"input": map[string]any{"documentId": docID},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, docID, result.DeleteDocumentDraft.Document.ID)
		},
	)

	t.Run(
		"cannot delete initial v0.1 draft",
		func(t *testing.T) {
			t.Parallel()

			docID, _ := createTestDocument(t, owner)

			var result struct{}
			err := owner.Execute(query, map[string]any{
				"input": map[string]any{"documentId": docID},
			}, &result)
			require.Error(t, err)
		},
	)

	t.Run(
		"cannot delete when latest is published",
		func(t *testing.T) {
			t.Parallel()

			docID, _ := createTestDocument(t, owner)
			approveTestDocument(t, owner, docID)

			var result struct{}
			err := owner.Execute(query, map[string]any{
				"input": map[string]any{"documentId": docID},
			}, &result)
			require.Error(t, err)
		},
	)

	t.Run(
		"viewer cannot delete draft",
		func(t *testing.T) {
			t.Parallel()

			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			docID, _ := createTestDocument(t, owner)
			approveTestDocument(t, owner, docID)

			// Create a draft
			updateQuery := `
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document { id }
						documentVersion { id }
					}
				}
			`
			var updateResult struct {
				UpdateDocument struct {
					Document struct {
						ID string `json:"id"`
					} `json:"document"`
					DocumentVersion *struct {
						ID string `json:"id"`
					} `json:"documentVersion"`
				} `json:"updateDocument"`
			}
			err := owner.Execute(updateQuery, map[string]any{
				"input": map[string]any{
					"id":      docID,
					"content": testutil.ProseMirrorTextDoc("Draft content"),
				},
			}, &updateResult)
			require.NoError(t, err)

			var result struct{}
			err = viewer.Execute(query, map[string]any{
				"input": map[string]any{"documentId": docID},
			}, &result)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete document draft")
		},
	)
}
