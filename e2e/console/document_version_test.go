// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner)})
	approveLatestDocumentVersion(t, owner, docID)
}

// latestDocumentVersionID returns the ID of the document's most recently created version.
func latestDocumentVersionID(t *testing.T, owner *testutil.Client, docID string) string {
	t.Helper()

	var result struct {
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
	`, map[string]any{"id": docID}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.Versions.Edges)

	return result.Node.Versions.Edges[0].Node.ID
}

// requestDocumentApproval opens a major approval quorum on the document's draft.
func requestDocumentApproval(t *testing.T, owner *testutil.Client, docID string, approverIDs []string) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentId":  docID,
			"approverIds": approverIDs,
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)
}

// approveLatestDocumentVersion approves the document's most recent version.
func approveLatestDocumentVersion(t *testing.T, owner *testutil.Client, docID string) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: ApproveDocumentVersionInput!) {
			approveDocumentVersion(input: $input) {
				approvalDecision { id state }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": latestDocumentVersionID(t, owner, docID),
		},
	})
	require.NoError(t, err)
}

// publishMajorDocumentVersion publishes the document's draft as a major version
// without approvers and returns the published version ID.
func publishMajorDocumentVersion(t *testing.T, owner *testutil.Client, docID string) string {
	t.Helper()

	var result struct {
		PublishDocument struct {
			DocumentVersion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentVersion"`
		} `json:"publishDocument"`
	}

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentId":  docID,
			"approverIds": []string{},
			"changelog":   "Major release",
		},
	}, &result)
	require.NoError(t, err)
	require.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)

	return result.PublishDocument.DocumentVersion.ID
}

// publishMinorDocumentVersion publishes the document's draft as a minor version
// and returns the published version ID.
func publishMinorDocumentVersion(t *testing.T, owner *testutil.Client, docID string) string {
	t.Helper()

	var result struct {
		PublishDocument struct {
			DocumentVersion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentVersion"`
		} `json:"publishDocument"`
	}

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": docID,
			"changelog":  "Minor release",
		},
	}, &result)
	require.NoError(t, err)
	require.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)

	return result.PublishDocument.DocumentVersion.ID
}

// updateDocumentContent edits the document so a fresh draft is created.
func updateDocumentContent(t *testing.T, owner *testutil.Client, docID, content string) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				documentVersion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc(content),
		},
	})
	require.NoError(t, err)
}

// requestDocumentSignature requests a signature on the version from the signatory.
func requestDocumentSignature(t *testing.T, owner *testutil.Client, versionID, signatoryID string) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
			"signatoryId":       signatoryID,
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

func TestDocumentVersion_PublishInitialMinorVersion(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)

	publishMutation := `
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				document {
					currentPublishedMajor
					currentPublishedMinor
				}
				documentVersion {
					status
					major
					minor
				}
			}
		}
	`

	var publishResult struct {
		PublishDocument struct {
			Document struct {
				CurrentPublishedMajor *int `json:"currentPublishedMajor"`
				CurrentPublishedMinor *int `json:"currentPublishedMinor"`
			} `json:"document"`
			DocumentVersion struct {
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersion"`
		} `json:"publishDocument"`
	}

	err := owner.Execute(publishMutation, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": docID,
			"changelog":  "Initial minor release",
		},
	}, &publishResult)
	require.NoError(t, err)

	require.NotNil(t, publishResult.PublishDocument.Document.CurrentPublishedMajor)
	require.NotNil(t, publishResult.PublishDocument.Document.CurrentPublishedMinor)
	assert.Equal(t, 0, *publishResult.PublishDocument.Document.CurrentPublishedMajor)
	assert.Equal(t, 1, *publishResult.PublishDocument.Document.CurrentPublishedMinor)
	assert.Equal(t, "PUBLISHED", publishResult.PublishDocument.DocumentVersion.Status)
	assert.Equal(t, 0, publishResult.PublishDocument.DocumentVersion.Major)
	assert.Equal(t, 1, publishResult.PublishDocument.DocumentVersion.Minor)

	var updateResult struct {
		UpdateDocument struct {
			DocumentVersion *struct {
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersion"`
		} `json:"updateDocument"`
	}

	err = owner.Execute(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				documentVersion {
					status
					major
					minor
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      docID,
			"content": testutil.ProseMirrorTextDoc("Updated content for 0.2"),
		},
	}, &updateResult)
	require.NoError(t, err)
	require.NotNil(t, updateResult.UpdateDocument.DocumentVersion)
	assert.Equal(t, "DRAFT", updateResult.UpdateDocument.DocumentVersion.Status)
	assert.Equal(t, 0, updateResult.UpdateDocument.DocumentVersion.Major)
	assert.Equal(t, 2, updateResult.UpdateDocument.DocumentVersion.Minor)

	err = owner.Execute(publishMutation, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": docID,
			"changelog":  "Second minor release",
		},
	}, &publishResult)
	require.NoError(t, err)

	require.NotNil(t, publishResult.PublishDocument.Document.CurrentPublishedMajor)
	require.NotNil(t, publishResult.PublishDocument.Document.CurrentPublishedMinor)
	assert.Equal(t, 0, *publishResult.PublishDocument.Document.CurrentPublishedMajor)
	assert.Equal(t, 2, *publishResult.PublishDocument.Document.CurrentPublishedMinor)
	assert.Equal(t, "PUBLISHED", publishResult.PublishDocument.DocumentVersion.Status)
	assert.Equal(t, 0, publishResult.PublishDocument.DocumentVersion.Major)
	assert.Equal(t, 2, publishResult.PublishDocument.DocumentVersion.Minor)
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
		mutation BulkPublishDocuments($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
				documentVersions {
					id
					status
				}
			}
		}
	`

	var result struct {
		BulkPublishDocuments struct {
			DocumentVersions []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentVersions"`
		} `json:"bulkPublishDocuments"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentIds": []string{docID1, docID2},
			"changelog":   "Bulk publish release",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.BulkPublishDocuments.DocumentVersions))

	for _, dv := range result.BulkPublishDocuments.DocumentVersions {
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
		mutation BulkPublishDocuments($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
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
		BulkPublishDocuments struct {
			DocumentVersions []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersions"`
		} `json:"bulkPublishDocuments"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentIds": []string{docID},
			"changelog":   "Needs approval",
		},
	}, &result)
	require.NoError(t, err)

	require.Len(t, result.BulkPublishDocuments.DocumentVersions, 1)
	dv := result.BulkPublishDocuments.DocumentVersions[0]
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
		mutation($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentIds": []string{docID},
			"changelog":   "First approval request",
		},
	})
	require.NoError(t, err)

	// Bulk publish again — should skip the pending document and return empty
	var result struct {
		BulkPublishDocuments struct {
			DocumentVersions []struct {
				ID string `json:"id"`
			} `json:"documentVersions"`
		} `json:"bulkPublishDocuments"`
	}

	err = owner.Execute(`
		mutation($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentIds": []string{docID},
			"changelog":   "Second attempt",
		},
	}, &result)
	require.NoError(t, err)

	assert.Empty(t, result.BulkPublishDocuments.DocumentVersions)
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
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Approval request",
		},
	})
	require.NoError(t, err)

	// Bulk publish minor — should skip the pending document
	var result struct {
		BulkPublishDocuments struct {
			DocumentVersions []struct {
				ID string `json:"id"`
			} `json:"documentVersions"`
		} `json:"bulkPublishDocuments"`
	}

	err = owner.Execute(`
		mutation($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
				documentVersions { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       true,
			"documentIds": []string{docID},
			"changelog":   "Minor publish attempt",
		},
	}, &result)
	require.NoError(t, err)

	assert.Empty(t, result.BulkPublishDocuments.DocumentVersions)
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
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
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
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
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
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       false,
			"documentId":  docID,
			"approverIds": []string{approverID},
			"changelog":   "Test changelog",
		},
	})
	require.NoError(t, err)

	t.Run("publish major blocked", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: PublishDocumentInput!) {
				publishDocument(input: $input) {
					documentVersion { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  docID,
				"approverIds": []string{},
				"changelog":   "Major release",
			},
		})
		require.Error(t, err)
	})

	t.Run("publish minor blocked", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(`
			mutation($input: PublishDocumentInput!) {
				publishDocument(input: $input) {
					documentVersion { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"minor":      true,
				"documentId": docID,
				"changelog":  "Minor release",
			},
		})
		require.Error(t, err)
	})
}

// TestDocumentVersion_PublishApproverIDsContract verifies that approver_ids must
// be an explicit choice when publishing: required for a major version (empty
// list publishes directly, non-empty requests approval) and rejected for a minor
// version (which ignores approvers).
func TestDocumentVersion_PublishApproverIDsContract(t *testing.T) {
	t.Parallel()

	const publishMutation = `
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { status }
				approvalQuorum { id }
			}
		}
	`

	type publishResult struct {
		PublishDocument struct {
			DocumentVersion struct {
				Status string `json:"status"`
			} `json:"documentVersion"`
			ApprovalQuorum *struct {
				ID string `json:"id"`
			} `json:"approvalQuorum"`
		} `json:"publishDocument"`
	}

	t.Run("major publish without approver_ids is rejected", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID, _ := createTestDocument(t, owner)

		_, err := owner.Do(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":      false,
				"documentId": docID,
				"changelog":  "Major release",
			},
		})
		require.Error(t, err)
	})

	t.Run("major publish with empty approver_ids publishes directly", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID, _ := createTestDocument(t, owner)

		var result publishResult

		err := owner.Execute(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  docID,
				"approverIds": []string{},
				"changelog":   "Direct major release",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)
		assert.Nil(t, result.PublishDocument.ApprovalQuorum)
	})

	t.Run("major publish with approver_ids requests approval", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID, _ := createTestDocument(t, owner)

		var result publishResult

		err := owner.Execute(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  docID,
				"approverIds": []string{getOwnerProfileID(t, owner)},
				"changelog":   "Major release needing approval",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PENDING_APPROVAL", result.PublishDocument.DocumentVersion.Status)
		require.NotNil(t, result.PublishDocument.ApprovalQuorum)
	})

	t.Run("minor publish without approver_ids is allowed", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID, _ := createTestDocument(t, owner)
		publishMajorDocumentVersion(t, owner, docID)
		updateDocumentContent(t, owner, docID, "Updated content for minor")

		var result publishResult

		err := owner.Execute(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":      true,
				"documentId": docID,
				"changelog":  "Minor release",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)
		assert.Nil(t, result.PublishDocument.ApprovalQuorum)
	})

	t.Run("minor publish with approver_ids is rejected", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID, _ := createTestDocument(t, owner)
		publishMajorDocumentVersion(t, owner, docID)
		updateDocumentContent(t, owner, docID, "Updated content for rejected minor")

		_, err := owner.Do(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":       true,
				"documentId":  docID,
				"approverIds": []string{getOwnerProfileID(t, owner)},
				"changelog":   "Minor release",
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

func TestDocumentVersion_ExportPDFSignatures(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"can export draft with signatures",
		func(t *testing.T) {
			t.Parallel()

			_, docVersionID := createTestDocument(t, owner)

			var result struct {
				ExportDocumentVersionPDF struct {
					Data string `json:"data"`
				} `json:"exportDocumentVersionPDF"`
			}

			err := owner.Execute(`
				mutation ExportPDF($input: ExportDocumentVersionPDFInput!) {
					exportDocumentVersionPDF(input: $input) {
						data
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"documentVersionId": docVersionID,
					"withWatermark":     false,
					"withSignatures":    true,
				},
			}, &result)
			require.NoError(t, err)
			assert.NotEmpty(t, result.ExportDocumentVersionPDF.Data)
		},
	)

	t.Run(
		"can export published with signatures",
		func(t *testing.T) {
			t.Parallel()

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
				query GetVersions($id: ID!) {
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

			var result struct {
				ExportDocumentVersionPDF struct {
					Data string `json:"data"`
				} `json:"exportDocumentVersionPDF"`
			}

			err = owner.Execute(`
				mutation ExportPDF($input: ExportDocumentVersionPDFInput!) {
					exportDocumentVersionPDF(input: $input) {
						data
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"documentVersionId": publishedVersionID,
					"withWatermark":     false,
					"withSignatures":    true,
				},
			}, &result)
			require.NoError(t, err)
			assert.NotEmpty(t, result.ExportDocumentVersionPDF.Data)
		},
	)
}

// TestDocumentVersion_MajorPublishCancelsSignatureRequests verifies that
// publishing a new major version directly (no approvers) cancels the still
// pending signature requests attached to the previous major version.
func TestDocumentVersion_MajorPublishCancelsSignatureRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)

	v1ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v1ID, signer.GetProfileID().String())
	assertRequestedSignatureCount(t, owner, v1ID, 1)

	// A new major supersedes the previous one, so its pending request is cancelled.
	updateDocumentContent(t, owner, docID, "Updated content for v2")
	v2ID := publishMajorDocumentVersion(t, owner, docID)

	assertRequestedSignatureCount(t, owner, v1ID, 0)
	assertRequestedSignatureCount(t, owner, v2ID, 0)
}

// TestDocumentVersion_MinorPublishKeepsSignatureRequests verifies that
// publishing a minor version stays within the same major and therefore keeps
// pending signature requests intact.
func TestDocumentVersion_MinorPublishKeepsSignatureRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)

	v1ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v1ID, signer.GetProfileID().String())
	assertRequestedSignatureCount(t, owner, v1ID, 1)

	// A minor bump keeps the same major; the request must survive.
	updateDocumentContent(t, owner, docID, "Updated content for 1.1")
	publishMinorDocumentVersion(t, owner, docID)

	assertRequestedSignatureCount(t, owner, v1ID, 1)
}

// TestDocumentVersion_MajorApprovalPublishCancelsSignatureRequests verifies
// that on the approval path the pending signature requests from the previous
// major are cancelled only once the quorum succeeds and the new major is
// actually published, not while the approval is still pending.
func TestDocumentVersion_MajorApprovalPublishCancelsSignatureRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	v1ID := latestDocumentVersionID(t, owner, docID)
	requestDocumentSignature(t, owner, v1ID, signer.GetProfileID().String())
	assertRequestedSignatureCount(t, owner, v1ID, 1)

	// Open a major approval for v2. While the quorum is pending the request
	// from the previous major must remain untouched.
	updateDocumentContent(t, owner, docID, "Updated content for v2")
	requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner)})
	assertRequestedSignatureCount(t, owner, v1ID, 1)

	// Once the quorum succeeds and v2.0 is published, the request is cancelled.
	approveLatestDocumentVersion(t, owner, docID)
	assertRequestedSignatureCount(t, owner, v1ID, 0)
}

// TestDocumentVersion_RequestSignatureIsIdempotentWithinVersion verifies that
// requesting a signature twice for the same signatory on the same version does
// not create a duplicate signature row.
func TestDocumentVersion_RequestSignatureIsIdempotentWithinVersion(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)
	signerID := signer.GetProfileID().String()

	v1ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v1ID, signerID)
	requestDocumentSignature(t, owner, v1ID, signerID)

	assertRequestedSignatureCount(t, owner, v1ID, 1)
}

// TestDocumentVersion_RequestSignatureDeduplicatesAcrossMinors verifies that
// re-requesting a signature on a newer minor of the same major reuses the
// signatory's existing signature instead of creating a second one, so the
// person is not listed twice in the major's export.
func TestDocumentVersion_RequestSignatureDeduplicatesAcrossMinors(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)
	signerID := signer.GetProfileID().String()

	v10ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v10ID, signerID)
	assertRequestedSignatureCount(t, owner, v10ID, 1)

	// A minor bump stays in the same major; the request stays on v1.0.
	updateDocumentContent(t, owner, docID, "Updated content for 1.1")
	v11ID := publishMinorDocumentVersion(t, owner, docID)

	// Re-requesting on the newer minor must reuse the existing signature
	// instead of inserting a duplicate within the same major. The signature
	// query aggregates across every minor of the major, so a duplicate would
	// surface as a major-wide REQUESTED count of 2; the fix keeps it at 1.
	requestDocumentSignature(t, owner, v11ID, signerID)

	assertRequestedSignatureCount(t, owner, v10ID, 1)
	assertRequestedSignatureCount(t, owner, v11ID, 1)
}

// signDocumentVersion signs the version as the authenticated client and returns
// the resulting signature node's id, state and signing time.
func signDocumentVersion(t *testing.T, signer *testutil.Client, versionID string) (id, state, signedAt string) {
	t.Helper()

	var result struct {
		SignDocument struct {
			DocumentVersionSignature struct {
				ID       string `json:"id"`
				State    string `json:"state"`
				SignedAt string `json:"signedAt"`
			} `json:"documentVersionSignature"`
		} `json:"signDocument"`
	}

	err := signer.Execute(`
		mutation($input: SignDocumentInput!) {
			signDocument(input: $input) {
				documentVersionSignature { id state signedAt }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": versionID,
		},
	}, &result)
	require.NoError(t, err)

	return result.SignDocument.DocumentVersionSignature.ID,
		result.SignDocument.DocumentVersionSignature.State,
		result.SignDocument.DocumentVersionSignature.SignedAt
}

// signDocumentVersionMutation is the raw mutation used by the negative-path
// signing tests so they can assert the request is rejected.
const signDocumentVersionMutation = `
	mutation($input: SignDocumentInput!) {
		signDocument(input: $input) {
			documentVersionSignature { id state }
		}
	}
`

// TestDocumentVersion_SignDocument verifies that a requested signature can be
// signed by the signatory, transitioning REQUESTED -> SIGNED and recording the
// signing time. Signing exercises the electronic-signature integration end to
// end: PDF export, upload, and esign create-and-accept.
func TestDocumentVersion_SignDocument(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	ownerProfileID := owner.GetProfileID().String()

	requestDocumentSignature(t, owner, publishedVersionID, ownerProfileID)

	id, state, signedAt := signDocumentVersion(t, owner, publishedVersionID)

	assert.NotEmpty(t, id)
	assert.Equal(t, "SIGNED", state)
	assert.NotEmpty(t, signedAt)
}

// TestDocumentVersion_SignDocumentWithoutRequestFails verifies that a version
// cannot be signed unless a signature was first requested for the signatory.
func TestDocumentVersion_SignDocumentWithoutRequestFails(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)

	_ = owner.ExecuteShouldFail(signDocumentVersionMutation, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
		},
	})
}

// TestDocumentVersion_SignDocumentTwiceFails verifies that an already-signed
// signature cannot be signed again.
func TestDocumentVersion_SignDocumentTwiceFails(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	ownerProfileID := owner.GetProfileID().String()

	requestDocumentSignature(t, owner, publishedVersionID, ownerProfileID)
	signDocumentVersion(t, owner, publishedVersionID)

	_ = owner.ExecuteShouldFail(signDocumentVersionMutation, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
		},
	})
}

// TestDocumentVersion_SignArchivedDocumentFails verifies that a document
// archived after its signature was requested can no longer be signed. This
// guards the archived/published preconditions that are re-validated inside the
// signing transaction so a state change between request and sign is honored.
func TestDocumentVersion_SignArchivedDocumentFails(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	ownerProfileID := owner.GetProfileID().String()

	requestDocumentSignature(t, owner, publishedVersionID, ownerProfileID)

	_, err := owner.Do(`
		mutation($input: ArchiveDocumentInput!) {
			archiveDocument(input: $input) {
				document { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId": docID,
		},
	})
	require.NoError(t, err)

	_ = owner.ExecuteShouldFail(signDocumentVersionMutation, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
		},
	})
}
