// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type documentVersionSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Major  int    `json:"major"`
	Minor  int    `json:"minor"`
}

func getOwnerProfileID(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	return owner.GetProfileID().String()
}

// createTestDocument creates a document and returns its ID.
func createTestDocument(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	query := `
		mutation CreateDocument($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node {
						id
					}
				}
			}
		}
	`

	var result struct {
		CreateDocument struct {
			DocumentEdge struct {
				Node struct {
					ID string `json:"id"`
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

	return result.CreateDocument.DocumentEdge.Node.ID
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

	return latestDocumentVersion[struct {
		ID string `json:"id"`
	}](t, owner, docID, "id").ID
}

// latestDocumentVersion returns the document's most recently created
// version. A version is a tag on a revision of the document history, so the
// newest one is the last tag of the newest tagged revision.
func latestDocumentVersion[T any](t *testing.T, owner *testutil.Client, docID string, selection string) T {
	t.Helper()

	var result struct {
		Node struct {
			History struct {
				Edges []struct {
					Node struct {
						Versions []T `json:"versions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"history"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					history(first: 1, orderBy: { field: ENDED_AT, direction: DESC }, filter: { versioned: true }) {
						edges { node { versions { `+selection+` } } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.History.Edges)
	require.NotEmpty(t, result.Node.History.Edges[0].Node.Versions)

	versions := result.Node.History.Edges[0].Node.Versions

	return versions[len(versions)-1]
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

type (
	documentVersionActors struct {
		owner  *testutil.Client
		viewer *testutil.Client
	}
)

func TestDocumentVersion(t *testing.T) {
	t.Parallel()

	// Shared owner/viewer clients are intentional: each parallel subtest creates
	// its own documents and mutable resources in this organization. net/http
	// Client and cookiejar.Jar are safe for concurrent use; the full suite runs
	// without race instrumentation for speed while ./e2e/internal/... is
	// race-gated in CI.
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	actors := documentVersionActors{owner: owner, viewer: viewer}

	cases := []struct {
		name string
		run  func(*testing.T, documentVersionActors)
	}{
		{name: "PublishVersion", run: documentVersionPublishVersion},
		{name: "VersionTagsRevision", run: documentVersionVersionTagsRevision},
		{name: "PublishInitialMinorVersion", run: documentVersionPublishInitialMinorVersion},
		{name: "AutoCreateDraft", run: documentVersionAutoCreateDraft},
		{name: "RequestSignature", run: documentVersionRequestSignature},
		{name: "BulkPublish", run: documentVersionBulkPublish},
		{name: "BulkPublishRequestsApproval", run: documentVersionBulkPublishRequestsApproval},
		{name: "BulkPublishSkipsPendingApproval", run: documentVersionBulkPublishSkipsPendingApproval},
		{name: "BulkPublishMinorSkipsPendingApproval", run: documentVersionBulkPublishMinorSkipsPendingApproval},
		{name: "BulkRequestSignatures", run: documentVersionBulkRequestSignatures},
		{name: "AutoCreateDraftOnClassificationOrTypeUpdate", run: documentVersionAutoCreateDraftOnClassificationOrTypeUpdate},
		{name: "ViewerCannotUpdateDocument", run: documentVersionViewerCannotUpdateDocument},
		{name: "BulkDelete", run: documentVersionBulkDelete},
		{name: "VoidApproval", run: documentVersionVoidApproval},
		{name: "RejectApproval", run: documentVersionRejectApproval},
		{name: "PublishBlockedWhenPendingApproval", run: documentVersionPublishBlockedWhenPendingApproval},
		{name: "PublishApproverIDsContract", run: documentVersionPublishApproverIDsContract},
		{name: "DefaultApprovers", run: documentDefaultApprovers},
		{name: "ExportPDFSignatures", run: documentVersionExportPDFSignatures},
		{name: "MajorPublishCancelsSignatureRequests", run: documentVersionMajorPublishCancelsSignatureRequests},
		{name: "MinorPublishKeepsSignatureRequests", run: documentVersionMinorPublishKeepsSignatureRequests},
		{name: "MajorApprovalPublishCancelsSignatureRequests", run: documentVersionMajorApprovalPublishCancelsSignatureRequests},
		{name: "RequestSignatureIsIdempotentWithinVersion", run: documentVersionRequestSignatureIsIdempotentWithinVersion},
		{name: "RequestSignatureDeduplicatesAcrossMinors", run: documentVersionRequestSignatureDeduplicatesAcrossMinors},
		{name: "MinorPublishMovesSignatureRequestsToNewVersion", run: documentVersionMinorPublishMovesSignatureRequestsToNewVersion},
		{name: "MinorApprovalPublishesAndKeepsSignatureRequests", run: documentVersionMinorApprovalPublishesAndKeepsSignatureRequests},
		{name: "BulkPublishRequestApprovalOverride", run: documentVersionBulkPublishRequestApprovalOverride},
		{name: "StartingVersion", run: documentVersionStartingVersion},
		{name: "SignDocument", run: documentVersionSignDocument},
		{name: "SignDocumentWithoutRequestFails", run: documentVersionSignDocumentWithoutRequestFails},
		{name: "SignDocumentTwiceFails", run: documentVersionSignDocumentTwiceFails},
		{name: "ArchiveVoidsPendingQuorum", run: documentVersionArchiveVoidsPendingQuorum},
		{name: "ArchiveCancelsPendingSignatures", run: documentVersionArchiveCancelsPendingSignatures},
		{name: "UnarchiveRestoresEditableDraft", run: documentVersionUnarchiveRestoresEditableDraft},
		{name: "SignArchivedDocumentFails", run: documentVersionSignArchivedDocumentFails},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				tc.run(
					t,
					actors,
				)
			},
		)
	}
}

func documentVersionPublishVersion(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// After approval, the version is auto-published.
	version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")

	assert.Equal(t, "PUBLISHED", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)
}

// documentVersionVersionTagsRevision verifies that the history is the only
// timeline: a published version is a tag on the revision it was taken at, and
// later edits open a newer revision that carries no tag.
func documentVersionVersionTagsRevision(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)
	versionID := latestDocumentVersionID(t, owner, docID)

	updateDocumentContent(t, owner, docID, "Edited after the version")

	type revision struct {
		ID        string   `json:"id"`
		Heads     []string `json:"heads"`
		Synthetic bool     `json:"synthetic"`
		Versions  []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Major  int    `json:"major"`
			Minor  int    `json:"minor"`
		} `json:"versions"`
	}

	var result struct {
		Node struct {
			LatestPublishedVersion *struct {
				ID string `json:"id"`
			} `json:"latestPublishedVersion"`
			PendingApprovalVersion *struct {
				ID string `json:"id"`
			} `json:"pendingApprovalVersion"`
			History struct {
				Edges []struct {
					Node revision `json:"node"`
				} `json:"edges"`
			} `json:"history"`
			Versioned struct {
				Edges []struct {
					Node revision `json:"node"`
				} `json:"edges"`
			} `json:"versioned"`
		} `json:"node"`
	}

	const revisionFields = `id heads synthetic versions { id status major minor }`

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					latestPublishedVersion { id }
					pendingApprovalVersion { id }
					history(first: 50, orderBy: { field: ENDED_AT, direction: DESC }) {
						edges { node { `+revisionFields+` } }
					}
					versioned: history(first: 50, orderBy: { field: ENDED_AT, direction: DESC }, filter: { versioned: true }) {
						edges { node { `+revisionFields+` } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &result)
	require.NoError(t, err)

	require.NotNil(t, result.Node.LatestPublishedVersion)
	assert.Equal(t, versionID, result.Node.LatestPublishedVersion.ID)
	assert.Nil(t, result.Node.PendingApprovalVersion)

	require.Len(t, result.Node.Versioned.Edges, 1, "only the tagged revision is listed")
	tagged := result.Node.Versioned.Edges[0].Node
	assert.False(t, tagged.Synthetic)
	assert.NotEmpty(t, tagged.Heads)
	require.Len(t, tagged.Versions, 1)
	assert.Equal(t, versionID, tagged.Versions[0].ID)
	assert.Equal(t, "PUBLISHED", tagged.Versions[0].Status)
	assert.Equal(t, 1, tagged.Versions[0].Major)

	require.GreaterOrEqual(t, len(result.Node.History.Edges), 2, "the edit after the version opens a newer revision")
	newest := result.Node.History.Edges[0].Node
	assert.NotEqual(t, tagged.ID, newest.ID)
	assert.Empty(t, newest.Versions, "the newer revision carries no tag")

	var found bool
	for _, edge := range result.Node.History.Edges {
		if edge.Node.ID == tagged.ID {
			found = true
			require.Len(t, edge.Node.Versions, 1)
			assert.Equal(t, versionID, edge.Node.Versions[0].ID)
		}
	}
	assert.True(t, found, "the tagged revision is part of the full history")
}

func documentVersionPublishInitialMinorVersion(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)

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
			Document struct {
				Content struct {
					Prosemirror string `json:"prosemirror"`
				} `json:"content"`
			} `json:"document"`
		} `json:"updateDocument"`
	}

	err = owner.Execute(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					content { prosemirror }
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
	assert.Contains(t, updateResult.UpdateDocument.Document.Content.Prosemirror, "Updated content for 0.2")

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

func documentVersionAutoCreateDraft(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create and approve a document (auto-publishes on approval)
	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	// Updating content writes the live document and leaves published snapshots unchanged.
	query := `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					id
					content {
						prosemirror
						markdown
					}
				}
				documentVersion {
					id
				}
			}
		}
	`

	var result struct {
		UpdateDocument struct {
			Document struct {
				ID      string `json:"id"`
				Content struct {
					Prosemirror string `json:"prosemirror"`
					Markdown    string `json:"markdown"`
				} `json:"content"`
			} `json:"document"`
			DocumentVersion *struct {
				ID string `json:"id"`
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

	assert.Nil(t, result.UpdateDocument.DocumentVersion)
	assert.Contains(t, result.UpdateDocument.Document.Content.Prosemirror, "Updated content")
	assert.Contains(t, result.UpdateDocument.Document.Content.Markdown, "Updated content")
}

func documentVersionRequestSignature(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create and approve a document (auto-publishes on approval)
	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)

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

	err := owner.Execute(query, map[string]any{
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

func documentVersionBulkPublish(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create multiple draft documents (no default approvers — should publish directly)
	docID1 := createTestDocument(t, owner)
	docID2 := createTestDocument(t, owner)

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

func documentVersionBulkPublishRequestsApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

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

func documentVersionBulkPublishSkipsPendingApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

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

func documentVersionBulkPublishMinorSkipsPendingApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create and publish a document first (need a published version for minor publish)
	docID := createTestDocument(t, owner)
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

func documentVersionBulkRequestSignatures(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create and approve a document (auto-publishes on approval)
	docID := createTestDocument(t, owner)
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

func documentVersionAutoCreateDraftOnClassificationOrTypeUpdate(t *testing.T, actors documentVersionActors) {
	t.Run("documentType update creates draft", func(t *testing.T) {
		t.Parallel()

		owner := actors.owner.ForTest(t)

		// Create and approve a document (auto-publishes on approval)
		docID := createTestDocument(t, owner)
		approveTestDocument(t, owner, docID)

		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
						documentType
					}
				}
			}
		`

		var result struct {
			UpdateDocument struct {
				Document struct {
					ID           string `json:"id"`
					DocumentType string `json:"documentType"`
				} `json:"document"`
			} `json:"updateDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":           docID,
				"documentType": "PROCEDURE",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "PROCEDURE", result.UpdateDocument.Document.DocumentType)
	})

	t.Run("classification update creates draft", func(t *testing.T) {
		t.Parallel()

		owner := actors.owner.ForTest(t)

		// Create and approve a document (auto-publishes on approval)
		docID := createTestDocument(t, owner)
		approveTestDocument(t, owner, docID)

		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						id
						classification
					}
				}
			}
		`

		var result struct {
			UpdateDocument struct {
				Document struct {
					ID             string `json:"id"`
					Classification string `json:"classification"`
				} `json:"document"`
			} `json:"updateDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":             docID,
				"classification": "CONFIDENTIAL",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "CONFIDENTIAL", result.UpdateDocument.Document.Classification)
	})
}

func documentVersionViewerCannotUpdateDocument(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	viewer := actors.viewer.ForTest(t)

	// Create and approve a document (auto-publishes on approval)
	docID := createTestDocument(t, owner)
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

func documentVersionBulkDelete(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	// Create multiple documents to delete
	docID1 := createTestDocument(t, owner)
	docID2 := createTestDocument(t, owner)

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

func documentVersionVoidApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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
	version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
	assert.Equal(t, "PENDING_APPROVAL", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)

	versionID := version.ID

	// Void keeps the snapshot number and marks the row voided.
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
	assert.Equal(t, "VOIDED", voidResult.VoidDocumentVersionApproval.DocumentVersion.Status)
	assert.Equal(t, 1, voidResult.VoidDocumentVersionApproval.DocumentVersion.Major)
	assert.Equal(t, 0, voidResult.VoidDocumentVersionApproval.DocumentVersion.Minor)

	// Verify decisions are VOIDED after voiding
	quorums := latestDocumentVersion[struct {
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
	}](t, owner, docID, `
		approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
			edges {
				node {
					decisions(first: 100) {
						edges { node { state } }
					}
				}
			}
		}
	`).ApprovalQuorums.Edges
	require.NotEmpty(t, quorums)
	decisions := quorums[0].Node.Decisions.Edges
	require.NotEmpty(t, decisions)

	for _, d := range decisions {
		assert.Equal(t, "VOIDED", d.Node.State, "decisions should be VOIDED after voiding")
	}
}

func documentVersionRejectApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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
	version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
	assert.Equal(t, "PENDING_APPROVAL", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)

	versionID := version.ID

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

	// The refused snapshot keeps its number so the next publish can reuse it.
	version = latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
	assert.Equal(t, "REJECTED", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)

	// Verify decisions are VOIDED after reject
	quorums := latestDocumentVersion[struct {
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
	}](t, owner, docID, `
		approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
			edges {
				node {
					decisions(first: 100) {
						edges { node { state } }
					}
				}
			}
		}
	`).ApprovalQuorums.Edges
	require.NotEmpty(t, quorums)
	decisions := quorums[0].Node.Decisions.Edges
	require.Len(t, decisions, 1)
	assert.Equal(t, "REJECTED", decisions[0].Node.State, "rejecting approver's decision should be REJECTED")
}

func documentVersionPublishBlockedWhenPendingApproval(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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

		owner := actors.owner.ForTest(t)

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

		owner := actors.owner.ForTest(t)

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
func documentVersionPublishApproverIDsContract(t *testing.T, actors documentVersionActors) {
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
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)

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
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)

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
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)

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
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)
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

	t.Run("minor publish with approver_ids requests approval", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)
		publishMajorDocumentVersion(t, owner, docID)
		updateDocumentContent(t, owner, docID, "Updated content for approved minor")

		var result publishResult

		err := owner.Execute(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":       true,
				"documentId":  docID,
				"approverIds": []string{getOwnerProfileID(t, owner)},
				"changelog":   "Minor release needing approval",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PENDING_APPROVAL", result.PublishDocument.DocumentVersion.Status)
		require.NotNil(t, result.PublishDocument.ApprovalQuorum)

		version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, 1, version.Major)
		assert.Equal(t, 1, version.Minor)
	})

	t.Run("minor publish with empty approver_ids publishes directly", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)
		docID := createTestDocument(t, owner)
		publishMajorDocumentVersion(t, owner, docID)
		updateDocumentContent(t, owner, docID, "Updated content for direct minor")

		var result publishResult

		err := owner.Execute(publishMutation, map[string]any{
			"input": map[string]any{
				"minor":       true,
				"documentId":  docID,
				"approverIds": []string{},
				"changelog":   "Direct minor release",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)
		assert.Nil(t, result.PublishDocument.ApprovalQuorum)
	})
}

func documentVersionMinorApprovalPublishesAndKeepsSignatureRequests(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)

	v10ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v10ID, signer.GetProfileID().String())

	before := requestedSignatureVersions(t, owner, v10ID)
	require.Len(t, before, 1)

	updateDocumentContent(t, owner, docID, "Updated content for approved 1.1")

	_, err := owner.Do(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				approvalQuorum { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":       true,
			"documentId":  docID,
			"approverIds": []string{getOwnerProfileID(t, owner)},
			"changelog":   "Minor release needing approval",
		},
	})
	require.NoError(t, err)

	// Requesting approval does not touch the request on v1.0.
	assert.Equal(t, before, requestedSignatureVersions(t, owner, v10ID))

	approveLatestDocumentVersion(t, owner, docID)

	version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
	assert.Equal(t, "PUBLISHED", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 1, version.Minor)

	// Approving the minor version moves the request onto it, as a direct
	// minor publish does.
	after := requestedSignatureVersions(t, owner, version.ID)
	require.Len(t, after, 1)

	for id, versionID := range after {
		_, ok := before[id]
		assert.True(t, ok, "the signature request must keep its identity")
		assert.Equal(t, version.ID, versionID)
	}
}

func documentVersionBulkPublishRequestApprovalOverride(t *testing.T, actors documentVersionActors) {
	const mutation = `
		mutation($input: BulkPublishDocumentsInput!) {
			bulkPublishDocuments(input: $input) {
				documentVersions { status major minor }
			}
		}
	`

	type bulkResult struct {
		BulkPublishDocuments struct {
			DocumentVersions []struct {
				Status string `json:"status"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
			} `json:"documentVersions"`
		} `json:"bulkPublishDocuments"`
	}

	t.Run("major without approval publishes directly and keeps default approvers", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)
		approverID := getOwnerProfileID(t, owner)
		docID := createTestDocumentWithApprovers(t, owner, []string{approverID})

		var result bulkResult

		err := owner.Execute(mutation, map[string]any{
			"input": map[string]any{
				"minor":           false,
				"requestApproval": false,
				"documentIds":     []string{docID},
				"changelog":       "Direct major",
			},
		}, &result)
		require.NoError(t, err)
		require.Len(t, result.BulkPublishDocuments.DocumentVersions, 1)
		assert.Equal(t, "PUBLISHED", result.BulkPublishDocuments.DocumentVersions[0].Status)

		var doc struct {
			Node struct {
				DefaultApprovers []struct {
					ID string `json:"id"`
				} `json:"defaultApprovers"`
			} `json:"node"`
		}

		err = owner.Execute(`
			query($id: ID!) {
				node(id: $id) { ... on Document { defaultApprovers { id } } }
			}
		`, map[string]any{"id": docID}, &doc)
		require.NoError(t, err)
		require.Len(t, doc.Node.DefaultApprovers, 1)
		assert.Equal(t, approverID, doc.Node.DefaultApprovers[0].ID)
	})

	t.Run("minor with approval asks the default approvers", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)
		approverID := getOwnerProfileID(t, owner)
		docID := createTestDocument(t, owner)
		publishMajorDocumentVersion(t, owner, docID)

		_, err := owner.Do(`
			mutation($input: UpdateDocumentInput!) {
				updateDocument(input: $input) { document { id } }
			}
		`, map[string]any{
			"input": map[string]any{
				"id":                 docID,
				"content":            testutil.ProseMirrorTextDoc("Updated content for bulk minor approval"),
				"defaultApproverIds": []string{approverID},
			},
		})
		require.NoError(t, err)

		var result bulkResult

		err = owner.Execute(mutation, map[string]any{
			"input": map[string]any{
				"minor":           true,
				"requestApproval": true,
				"documentIds":     []string{docID},
				"changelog":       "Minor needing approval",
			},
		}, &result)
		require.NoError(t, err)
		require.Len(t, result.BulkPublishDocuments.DocumentVersions, 1)
		assert.Equal(t, "PENDING_APPROVAL", result.BulkPublishDocuments.DocumentVersions[0].Status)
		assert.Equal(t, 1, result.BulkPublishDocuments.DocumentVersions[0].Minor)
	})
}

func documentDefaultApprovers(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	approverID := getOwnerProfileID(t, owner)

	t.Run("create document with default approvers", func(t *testing.T) {
		t.Parallel()

		owner := actors.owner.ForTest(t)

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

		owner := actors.owner.ForTest(t)

		docID := createTestDocument(t, owner)

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

	t.Run("major publish with explicit approvers updates default approvers, even when empty", func(t *testing.T) {
		t.Parallel()

		owner := actors.owner.ForTest(t)

		docID := createTestDocumentWithApprovers(t, owner, []string{approverID})

		var result struct {
			PublishDocument struct {
				Document struct {
					DefaultApprovers []struct {
						ID string `json:"id"`
					} `json:"defaultApprovers"`
				} `json:"document"`
				DocumentVersion struct {
					Status string `json:"status"`
				} `json:"documentVersion"`
			} `json:"publishDocument"`
		}

		err := owner.Execute(`
			mutation($input: PublishDocumentInput!) {
				publishDocument(input: $input) {
					document { defaultApprovers { id } }
					documentVersion { status }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  docID,
				"approverIds": []string{},
				"changelog":   "Direct publish clearing approvers",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "PUBLISHED", result.PublishDocument.DocumentVersion.Status)
		assert.Empty(t, result.PublishDocument.Document.DefaultApprovers)
	})
}

func documentVersionExportPDFSignatures(t *testing.T, actors documentVersionActors) {
	t.Run(
		"can export a published snapshot with signatures",
		func(t *testing.T) {
			t.Parallel()

			owner := actors.owner.ForTest(t)

			docID := createTestDocument(t, owner)
			docVersionID := publishMajorDocumentVersion(t, owner, docID)

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

			owner := actors.owner.ForTest(t)

			docID := createTestDocument(t, owner)
			approveTestDocument(t, owner, docID)

			publishedVersionID := latestDocumentVersionID(t, owner, docID)

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
func documentVersionMajorPublishCancelsSignatureRequests(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)

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
func documentVersionMinorPublishKeepsSignatureRequests(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)

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
func documentVersionMajorApprovalPublishCancelsSignatureRequests(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)
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
func documentVersionRequestSignatureIsIdempotentWithinVersion(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)
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
func documentVersionRequestSignatureDeduplicatesAcrossMinors(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)
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

// requestedSignatureVersions returns a map of signature node ID -> the document
// version ID it is currently attached to, for the REQUESTED signatures visible
// from versionID (the signatures connection aggregates across every minor of
// the major).
func requestedSignatureVersions(
	t *testing.T,
	owner *testutil.Client,
	versionID string,
) map[string]string {
	t.Helper()

	var result struct {
		Node struct {
			Signatures struct {
				Edges []struct {
					Node struct {
						ID              string `json:"id"`
						State           string `json:"state"`
						DocumentVersion struct {
							ID string `json:"id"`
						} `json:"documentVersion"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"signatures"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 50, filter: { states: [REQUESTED] }) {
						edges { node { id state documentVersion { id } } }
					}
				}
			}
		}
	`, map[string]any{"id": versionID}, &result)
	require.NoError(t, err)

	out := make(map[string]string, len(result.Node.Signatures.Edges))
	for _, edge := range result.Node.Signatures.Edges {
		out[edge.Node.ID] = edge.Node.DocumentVersion.ID
	}

	return out
}

// TestDocumentVersion_MinorPublishMovesSignatureRequestsToNewVersion verifies
// that publishing a new minor version carries the still-pending signature
// requests from the previous version onto the newly published version. The same
// signature row is reused (its ID is preserved), so its notification schedule
// (count and last-notified time) is kept intact rather than reset.
func documentVersionMinorPublishMovesSignatureRequestsToNewVersion(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)
	signer := actors.viewer.ForTest(t)

	docID := createTestDocument(t, owner)

	v10ID := publishMajorDocumentVersion(t, owner, docID)
	requestDocumentSignature(t, owner, v10ID, signer.GetProfileID().String())

	// The pending request starts attached to v1.0.
	before := requestedSignatureVersions(t, owner, v10ID)
	require.Len(t, before, 1)

	var signatureID string
	for id, versionID := range before {
		signatureID = id

		assert.Equal(t, v10ID, versionID, "the request must start on v1.0")
	}

	// A minor bump publishes v1.1 within the same major.
	updateDocumentContent(t, owner, docID, "Updated content for 1.1")
	v11ID := publishMinorDocumentVersion(t, owner, docID)
	require.NotEqual(t, v10ID, v11ID)

	// The same request row is now attached to the newly published v1.1.
	after := requestedSignatureVersions(t, owner, v11ID)
	require.Len(t, after, 1)

	movedVersionID, ok := after[signatureID]
	require.True(t, ok, "the signature request must keep its identity across the minor publish")
	assert.Equal(t, v11ID, movedVersionID, "the pending request must move to the newly published minor version")
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

func archiveDocument(t *testing.T, owner *testutil.Client, docID string) {
	t.Helper()

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
}

func unarchiveDocument(t *testing.T, owner *testutil.Client, docID string) {
	t.Helper()

	_, err := owner.Do(`
		mutation($input: UnarchiveDocumentInput!) {
			unarchiveDocument(input: $input) {
				document { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId": docID,
		},
	})
	require.NoError(t, err)
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
func documentVersionSignDocument(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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
func documentVersionSignDocumentWithoutRequestFails(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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
func documentVersionSignDocumentTwiceFails(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
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

func documentVersionArchiveVoidsPendingQuorum(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner)})

	archiveDocument(t, owner, docID)

	var result struct {
		Node struct {
			Status string `json:"status"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					status
				}
			}
		}
	`, map[string]any{"id": docID}, &result)
	require.NoError(t, err)

	version := latestDocumentVersion[struct {
		Status          string `json:"status"`
		Major           int    `json:"major"`
		Minor           int    `json:"minor"`
		ApprovalQuorums struct {
			Edges []struct {
				Node struct {
					Status    string `json:"status"`
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
	}](t, owner, docID, `
		status
		major
		minor
		approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
			edges {
				node {
					status
					decisions(first: 10) {
						edges { node { state } }
					}
				}
			}
		}
	`)
	assert.Equal(t, "ARCHIVED", result.Node.Status)
	assert.Equal(t, "VOIDED", version.Status)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)
	require.NotEmpty(t, version.ApprovalQuorums.Edges)
	assert.Equal(t, "VOIDED", version.ApprovalQuorums.Edges[0].Node.Status)

	for _, edge := range version.ApprovalQuorums.Edges[0].Node.Decisions.Edges {
		assert.Equal(t, "VOIDED", edge.Node.State)
	}
}

func documentVersionArchiveCancelsPendingSignatures(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	requestDocumentSignature(t, owner, publishedVersionID, getOwnerProfileID(t, owner))

	archiveDocument(t, owner, docID)

	var result struct {
		Node struct {
			Signatures struct {
				Edges []struct {
					Node struct {
						State string `json:"state"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"signatures"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 10) {
						edges { node { state } }
					}
				}
			}
		}
	`, map[string]any{"id": publishedVersionID}, &result)
	require.NoError(t, err)
	assert.Empty(t, result.Node.Signatures.Edges)
}

func documentVersionUnarchiveRestoresEditableDraft(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner)})

	archiveDocument(t, owner, docID)
	unarchiveDocument(t, owner, docID)

	var result struct {
		Node struct {
			Status string `json:"status"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					status
				}
			}
		}
	`, map[string]any{"id": docID}, &result)
	require.NoError(t, err)

	version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
	assert.Equal(t, "ACTIVE", result.Node.Status)
	assert.Equal(t, "VOIDED", version.Status)

	updateDocumentContent(t, owner, docID, "Updated after unarchive")
}

// TestDocumentVersion_SignArchivedDocumentFails verifies that a document
// archived after its signature was requested can no longer be signed. This
// guards the archived/published preconditions that are re-validated inside the
// signing transaction so a state change between request and sign is honored.
func documentVersionSignArchivedDocumentFails(t *testing.T, actors documentVersionActors) {
	owner := actors.owner.ForTest(t)

	docID := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	publishedVersionID := latestDocumentVersionID(t, owner, docID)
	ownerProfileID := owner.GetProfileID().String()

	requestDocumentSignature(t, owner, publishedVersionID, ownerProfileID)

	archiveDocument(t, owner, docID)

	var signatureResult struct {
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

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 10) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": publishedVersionID}, &signatureResult)
	require.NoError(t, err)
	assert.Empty(t, signatureResult.Node.Signatures.Edges)

	_ = owner.ExecuteShouldFail(signDocumentVersionMutation, map[string]any{
		"input": map[string]any{
			"documentVersionId": publishedVersionID,
		},
	})
}

func documentVersionStartingVersion(t *testing.T, actors documentVersionActors) {
	const createMutation = `
		mutation($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge { node { id } }
			}
		}
	`

	create := func(t *testing.T, owner *testutil.Client, extra map[string]any) (string, error) {
		t.Helper()

		input := map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"title":          "Imported document",
			"content":        testutil.ProseMirrorTextDoc("Imported content"),
			"documentType":   "POLICY",
			"classification": "INTERNAL",
		}
		maps.Copy(input, extra)

		var result struct {
			CreateDocument struct {
				DocumentEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"documentEdge"`
			} `json:"createDocument"`
		}

		err := owner.Execute(createMutation, map[string]any{"input": input}, &result)

		return result.CreateDocument.DocumentEdge.Node.ID, err
	}

	t.Run("first publish takes the starting version", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)

		docID, err := create(t, owner, map[string]any{"startingMajor": 3, "startingMinor": 2})
		require.NoError(t, err)

		publishMajorDocumentVersion(t, owner, docID)

		version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, "PUBLISHED", version.Status)
		assert.Equal(t, 3, version.Major)
		assert.Equal(t, 2, version.Minor)

		updateDocumentContent(t, owner, docID, "Second revision")
		publishMinorDocumentVersion(t, owner, docID)

		version = latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, 3, version.Major)
		assert.Equal(t, 3, version.Minor)

		updateDocumentContent(t, owner, docID, "Third revision")
		publishMajorDocumentVersion(t, owner, docID)

		version = latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, 4, version.Major)
		assert.Equal(t, 0, version.Minor)
	})

	t.Run("a minor first publish also takes the starting version", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)

		docID, err := create(t, owner, map[string]any{"startingMajor": 7, "startingMinor": 0})
		require.NoError(t, err)

		publishMinorDocumentVersion(t, owner, docID)

		version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, 7, version.Major)
		assert.Equal(t, 0, version.Minor)
	})

	t.Run("the first approval request takes the starting version", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)

		docID, err := create(t, owner, map[string]any{"startingMajor": 2, "startingMinor": 5})
		require.NoError(t, err)

		requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner)})

		version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, "PENDING_APPROVAL", version.Status)
		assert.Equal(t, 2, version.Major)
		assert.Equal(t, 5, version.Minor)

		approveLatestDocumentVersion(t, owner, docID)

		version = latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, "PUBLISHED", version.Status)
		assert.Equal(t, 2, version.Major)
		assert.Equal(t, 5, version.Minor)
	})

	t.Run("without a starting version the first publish is 1.0", func(t *testing.T) {
		t.Parallel()
		owner := actors.owner.ForTest(t)

		docID, err := create(t, owner, nil)
		require.NoError(t, err)

		publishMajorDocumentVersion(t, owner, docID)

		version := latestDocumentVersion[documentVersionSummary](t, owner, docID, "id status major minor")
		assert.Equal(t, 1, version.Major)
		assert.Equal(t, 0, version.Minor)
	})

	for name, extra := range map[string]map[string]any{
		"major without minor": {"startingMajor": 3},
		"minor without major": {"startingMinor": 2},
		"zero version":        {"startingMajor": 0, "startingMinor": 0},
		"negative major":      {"startingMajor": -1, "startingMinor": 0},
		"too large":           {"startingMajor": 100000, "startingMinor": 0},
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			owner := actors.owner.ForTest(t)

			_, err := create(t, owner, extra)
			require.Error(t, err)
		})
	}
}
