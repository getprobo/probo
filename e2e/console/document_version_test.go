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
			"content":        "Initial content",
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

func TestDocumentVersion_CreateDraft(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and approve a document (auto-publishes on approval)
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	query := `
		mutation CreateDraftDocumentVersion($input: CreateDraftDocumentVersionInput!) {
			createDraftDocumentVersion(input: $input) {
				documentVersionEdge {
					node {
						id
						status
					}
				}
			}
		}
	`

	var result struct {
		CreateDraftDocumentVersion struct {
			DocumentVersionEdge struct {
				Node struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"node"`
			} `json:"documentVersionEdge"`
		} `json:"createDraftDocumentVersion"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentID": docID,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, "DRAFT", result.CreateDraftDocumentVersion.DocumentVersionEdge.Node.Status)
}

func TestDocumentVersion_UpdateContent(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	_, draftVersionID := createTestDocument(t, owner)

	query := `
		mutation UpdateDocumentVersion($input: UpdateDocumentVersionContentInput!) {
			updateDocumentVersionContent(input: $input) {
				content
			}
		}
	`

	var result struct {
		UpdateDocumentVersionContent struct {
			Content string `json:"content"`
		} `json:"updateDocumentVersionContent"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":      draftVersionID,
			"content": "Updated content for the document version",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, "Updated content for the document version", result.UpdateDocumentVersionContent.Content)
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

func TestDocumentVersion_BulkPublish(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple documents
	docID1, _ := createTestDocument(t, owner)
	docID2, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID1)
	approveTestDocument(t, owner, docID2)

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
