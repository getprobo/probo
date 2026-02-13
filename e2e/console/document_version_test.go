// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

// createTestDocument creates a document and returns its ID and the document version ID
func createTestDocument(t *testing.T, owner *testutil.Client) (docID string, docVersionID string) {
	t.Helper()
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
			"approverIds":    []string{profileID.String()},
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

func TestDocumentVersion_PublishVersion(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)

	query := `
		mutation PublishDocumentVersion($input: PublishDocumentVersionInput!) {
			publishDocumentVersion(input: $input) {
				documentVersion {
					id
					status
					version
					changelog
				}
				document {
					id
				}
			}
		}
	`

	var result struct {
		PublishDocumentVersion struct {
			DocumentVersion struct {
				ID        string `json:"id"`
				Status    string `json:"status"`
				Version   int    `json:"version"`
				Changelog string `json:"changelog"`
			} `json:"documentVersion"`
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		} `json:"publishDocumentVersion"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentId": docID,
			"changelog":  "Initial release",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, "PUBLISHED", result.PublishDocumentVersion.DocumentVersion.Status)
	assert.Equal(t, 1, result.PublishDocumentVersion.DocumentVersion.Version)
	assert.Equal(t, "Initial release", result.PublishDocumentVersion.DocumentVersion.Changelog)
}

func TestDocumentVersion_CreateDraft(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and publish a document first
	docID, _ := createTestDocument(t, owner)

	publishQuery := `
		mutation PublishDocumentVersion($input: PublishDocumentVersionInput!) {
			publishDocumentVersion(input: $input) {
				documentVersion {
					id
				}
			}
		}
	`

	_, err := owner.Do(publishQuery, map[string]any{
		"input": map[string]any{
			"documentId": docID,
			"changelog":  "Initial release",
		},
	})
	require.NoError(t, err)

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

	err = owner.Execute(query, map[string]any{
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
		mutation UpdateDocumentVersion($input: UpdateDocumentVersionInput!) {
			updateDocumentVersion(input: $input) {
				documentVersion {
					id
					content
				}
			}
		}
	`

	var result struct {
		UpdateDocumentVersion struct {
			DocumentVersion struct {
				ID      string `json:"id"`
				Content string `json:"content"`
			} `json:"documentVersion"`
		} `json:"updateDocumentVersion"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentVersionId": draftVersionID,
			"content":           "Updated content for the document",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, "Updated content for the document", result.UpdateDocumentVersion.DocumentVersion.Content)
}

func TestDocumentVersion_RequestSignature(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and publish a document
	docID, _ := createTestDocument(t, owner)

	publishQuery := `
		mutation PublishDocumentVersion($input: PublishDocumentVersionInput!) {
			publishDocumentVersion(input: $input) {
				documentVersion {
					id
				}
			}
		}
	`

	var publishResult struct {
		PublishDocumentVersion struct {
			DocumentVersion struct {
				ID string `json:"id"`
			} `json:"documentVersion"`
		} `json:"publishDocumentVersion"`
	}

	err := owner.Execute(publishQuery, map[string]any{
		"input": map[string]any{
			"documentId": docID,
			"changelog":  "Initial release",
		},
	}, &publishResult)
	require.NoError(t, err)

	publishedVersionID := publishResult.PublishDocumentVersion.DocumentVersion.ID

	// Create a person to sign
	signerProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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
			"signatoryId":       signerProfileID.String(),
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.RequestSignature.DocumentVersionSignatureEdge.Node.ID)
	assert.Equal(t, "REQUESTED", result.RequestSignature.DocumentVersionSignatureEdge.Node.State)
	assert.Equal(t, signerProfileID.String(), result.RequestSignature.DocumentVersionSignatureEdge.Node.SignedBy.ID)
}

func TestDocumentVersion_BulkPublish(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple documents
	docID1, _ := createTestDocument(t, owner)
	docID2, _ := createTestDocument(t, owner)

	query := `
		mutation BulkPublishDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
			bulkPublishDocumentVersions(input: $input) {
				documentVersionEdges {
					node {
						id
						status
					}
				}
			}
		}
	`

	var result struct {
		BulkPublishDocumentVersions struct {
			DocumentVersionEdges []struct {
				Node struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"node"`
			} `json:"documentVersionEdges"`
		} `json:"bulkPublishDocumentVersions"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds": []string{docID1, docID2},
			"changelog":   "Bulk publish release",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.BulkPublishDocumentVersions.DocumentVersionEdges))
	for _, edge := range result.BulkPublishDocumentVersions.DocumentVersionEdges {
		assert.Equal(t, "PUBLISHED", edge.Node.Status)
	}
}

func TestDocumentVersion_BulkRequestSignatures(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create and publish a document
	docID, _ := createTestDocument(t, owner)

	publishQuery := `
		mutation PublishDocumentVersion($input: PublishDocumentVersionInput!) {
			publishDocumentVersion(input: $input) {
				documentVersion {
					id
				}
			}
		}
	`

	_, err := owner.Do(publishQuery, map[string]any{
		"input": map[string]any{
			"documentId": docID,
			"changelog":  "Initial release",
		},
	})
	require.NoError(t, err)

	// Create multiple signers
	signer1ProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	signer2ProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

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

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"documentIds":  []string{docID},
			"signatoryIds": []string{signer1ProfileID.String(), signer2ProfileID.String()},
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
