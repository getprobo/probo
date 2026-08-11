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

type (
	documentBulkActors struct {
		owner  *testutil.Client
		viewer *testutil.Client
	}
)

func TestDocument_Bulk(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	actors := documentBulkActors{owner: owner, viewer: viewer}

	cases := []struct {
		name string
		run  func(*testing.T, documentBulkActors)
	}{
		{name: "ArchiveUnarchive", run: documentBulkArchiveUnarchive},
		{name: "Export", run: documentBulkExport},
		{name: "GenerateChangelogInitial", run: documentBulkGenerateChangelogInitial},
		{name: "GenerateChangelogNoChanges", run: documentBulkGenerateChangelogNoChanges},
		{name: "CancelSignatureRequest", run: documentBulkCancelSignatureRequest},
		{name: "ArchiveViewerForbidden", run: documentBulkArchiveViewerForbidden},
		{name: "ExportViewerForbidden", run: documentBulkExportViewerForbidden},
		{name: "CancelSignatureViewerForbidden", run: documentBulkCancelSignatureViewerForbidden},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				tc.run(t, actors)
			},
		)
	}
}

type documentArchiveState struct {
	Status     string  `json:"status"`
	ArchivedAt *string `json:"archivedAt"`
}

func loadDocumentArchiveStates(
	t *testing.T,
	client *testutil.Client,
	documentIDs []string,
) map[string]documentArchiveState {
	t.Helper()

	out := make(map[string]documentArchiveState, len(documentIDs))

	for _, documentID := range documentIDs {
		var result struct {
			Node documentArchiveState `json:"node"`
		}

		err := client.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						status
						archivedAt
					}
				}
			}
		`, map[string]any{"id": documentID}, &result)
		require.NoError(t, err)

		out[documentID] = result.Node
	}

	return out
}

func bulkArchiveDocuments(t *testing.T, client *testutil.Client, documentIDs []string) {
	t.Helper()

	_, err := client.Do(`
		mutation($input: BulkArchiveDocumentsInput!) {
			bulkArchiveDocuments(input: $input) {
				documents { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": documentIDs,
		},
	})
	require.NoError(t, err)
}

func bulkUnarchiveDocuments(t *testing.T, client *testutil.Client, documentIDs []string) {
	t.Helper()

	_, err := client.Do(`
		mutation($input: BulkUnarchiveDocumentsInput!) {
			bulkUnarchiveDocuments(input: $input) {
				documents { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": documentIDs,
		},
	})
	require.NoError(t, err)
}

func documentBulkArchiveUnarchive(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)

	docID1 := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Bulk archive A")).
		Create()
	docID2 := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Bulk archive B")).
		Create()

	bulkArchiveDocuments(t, owner, []string{docID1, docID2})

	archived := loadDocumentArchiveStates(t, owner, []string{docID1, docID2})
	for _, documentID := range []string{docID1, docID2} {
		state := archived[documentID]
		assert.Equal(t, "ARCHIVED", state.Status, "document %s", documentID)
		require.NotNil(t, state.ArchivedAt, "document %s archivedAt", documentID)
		assert.NotEmpty(t, *state.ArchivedAt)
	}

	bulkUnarchiveDocuments(t, owner, []string{docID1, docID2})

	active := loadDocumentArchiveStates(t, owner, []string{docID1, docID2})
	for _, documentID := range []string{docID1, docID2} {
		state := active[documentID]
		assert.Equal(t, "ACTIVE", state.Status, "document %s", documentID)
		assert.Nil(t, state.ArchivedAt, "document %s archivedAt", documentID)
	}
}

func documentBulkExport(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)

	docID1, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID1)

	docID2, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID2)

	var result struct {
		BulkExportDocuments struct {
			ExportJobID string `json:"exportJobId"`
		} `json:"bulkExportDocuments"`
	}

	err := owner.Execute(`
		mutation($input: BulkExportDocumentsInput!) {
			bulkExportDocuments(input: $input) {
				exportJobId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds":    []string{docID1, docID2},
			"withWatermark":  false,
			"withSignatures": false,
		},
	}, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.BulkExportDocuments.ExportJobID)
}

const generateDocumentChangelogMutation = `
	mutation($input: GenerateDocumentChangelogInput!) {
		generateDocumentChangelog(input: $input) {
			changelog
		}
	}
`

func documentBulkGenerateChangelogInitial(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)

	docID, _ := createTestDocument(t, owner)

	var result struct {
		GenerateDocumentChangelog struct {
			Changelog string `json:"changelog"`
		} `json:"generateDocumentChangelog"`
	}

	err := owner.Execute(generateDocumentChangelogMutation, map[string]any{
		"input": map[string]any{
			"documentId": docID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "Initial version", result.GenerateDocumentChangelog.Changelog)
}

func documentBulkGenerateChangelogNoChanges(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	_, err := owner.Do(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				documentVersion { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":             docID,
			"classification": "CONFIDENTIAL",
		},
	})
	require.NoError(t, err)

	var result struct {
		GenerateDocumentChangelog struct {
			Changelog string `json:"changelog"`
		} `json:"generateDocumentChangelog"`
	}

	err = owner.Execute(generateDocumentChangelogMutation, map[string]any{
		"input": map[string]any{
			"documentId": docID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "No changes detected", result.GenerateDocumentChangelog.Changelog)
}

func firstDocumentVersionSignatureID(
	t *testing.T,
	client *testutil.Client,
	versionID string,
) string {
	t.Helper()

	var result struct {
		Node struct {
			Signatures struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"signatures"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 1) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": versionID}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.Signatures.Edges)

	return result.Node.Signatures.Edges[0].Node.ID
}

func documentBulkCancelSignatureRequest(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	ownerProfileID := getOwnerProfileID(t, owner)

	requestDocumentSignature(t, owner, publishedVersionID, ownerProfileID)
	signatureID := firstDocumentVersionSignatureID(t, owner, publishedVersionID)

	var cancelResult struct {
		CancelSignatureRequest struct {
			DeletedDocumentVersionSignatureID string `json:"deletedDocumentVersionSignatureId"`
		} `json:"cancelSignatureRequest"`
	}

	err := owner.Execute(`
		mutation($input: CancelSignatureRequestInput!) {
			cancelSignatureRequest(input: $input) {
				deletedDocumentVersionSignatureId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionSignatureId": signatureID,
		},
	}, &cancelResult)
	require.NoError(t, err)
	assert.Equal(t, signatureID, cancelResult.CancelSignatureRequest.DeletedDocumentVersionSignatureID)

	var after struct {
		Node struct {
			Signatures struct {
				TotalCount int `json:"totalCount"`
			} `json:"signatures"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 1) {
						totalCount
					}
				}
			}
		}
	`, map[string]any{"id": publishedVersionID}, &after)
	require.NoError(t, err)
	assert.Equal(t, 0, after.Node.Signatures.TotalCount)
}

func documentBulkArchiveViewerForbidden(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)
	viewer := actors.viewer.ForTest(t)

	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Bulk archive RBAC")).
		Create()

	_, err := viewer.Do(`
		mutation($input: BulkArchiveDocumentsInput!) {
			bulkArchiveDocuments(input: $input) {
				documents { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds": []string{documentID},
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer must not bulk archive documents")
}

func documentBulkExportViewerForbidden(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)
	viewer := actors.viewer.ForTest(t)

	documentID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, documentID)

	_, err := viewer.Do(`
		mutation($input: BulkExportDocumentsInput!) {
			bulkExportDocuments(input: $input) {
				exportJobId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentIds":    []string{documentID},
			"withWatermark":  false,
			"withSignatures": false,
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer must not bulk export documents")
}

func documentBulkCancelSignatureViewerForbidden(t *testing.T, actors documentBulkActors) {
	owner := actors.owner.ForTest(t)
	viewer := actors.viewer.ForTest(t)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	requestDocumentSignature(t, owner, publishedVersionID, getOwnerProfileID(t, owner))
	signatureID := firstDocumentVersionSignatureID(t, owner, publishedVersionID)

	_, err := viewer.Do(`
		mutation($input: CancelSignatureRequestInput!) {
			cancelSignatureRequest(input: $input) {
				deletedDocumentVersionSignatureId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionSignatureId": signatureID,
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer must not cancel signature requests")
}
