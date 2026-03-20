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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestDocument_BulkDelete(t *testing.T) {
	t.Parallel()

	t.Run("owner can bulk delete", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		var result struct {
			BulkDeleteDocuments struct {
				DeletedDocumentIds []string `json:"deletedDocumentIds"`
			} `json:"bulkDeleteDocuments"`
		}

		err := owner.Execute(`
			mutation BulkDeleteDocuments($input: BulkDeleteDocumentsInput!) {
				bulkDeleteDocuments(input: $input) {
					deletedDocumentIds
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		}, &result)
		require.NoError(t, err, "owner should be able to bulk delete documents")
		assert.Len(t, result.BulkDeleteDocuments.DeletedDocumentIds, 2)
	})

	t.Run("admin can bulk delete", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		var result struct {
			BulkDeleteDocuments struct {
				DeletedDocumentIds []string `json:"deletedDocumentIds"`
			} `json:"bulkDeleteDocuments"`
		}

		err := admin.Execute(`
			mutation BulkDeleteDocuments($input: BulkDeleteDocumentsInput!) {
				bulkDeleteDocuments(input: $input) {
					deletedDocumentIds
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		}, &result)
		require.NoError(t, err, "admin should be able to bulk delete documents")
		assert.Len(t, result.BulkDeleteDocuments.DeletedDocumentIds, 2)
	})

	t.Run("viewer cannot bulk delete", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		approverID := factory.CreateUser(owner)
		docID := factory.NewDocument(owner, approverID).Create()

		_, err := viewer.Do(`
			mutation BulkDeleteDocuments($input: BulkDeleteDocumentsInput!) {
				bulkDeleteDocuments(input: $input) {
					deletedDocumentIds
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID},
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to bulk delete documents")
	})
}

func TestDocument_BulkPublish(t *testing.T) {
	t.Parallel()

	t.Run("owner can bulk publish", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		docID1, _ := createTestDocument(t, owner)
		docID2, _ := createTestDocument(t, owner)

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

		err := owner.Execute(`
			mutation BulkPublishDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
				bulkPublishDocumentVersions(input: $input) {
					documentVersionEdges {
						node { id status }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
				"changelog":   "Bulk publish",
			},
		}, &result)
		require.NoError(t, err, "owner should be able to bulk publish")
		assert.Len(t, result.BulkPublishDocumentVersions.DocumentVersionEdges, 2)
		for _, edge := range result.BulkPublishDocumentVersions.DocumentVersionEdges {
			assert.Equal(t, "PUBLISHED", edge.Node.Status)
		}
	})

	t.Run("admin can bulk publish", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		docID1, _ := createTestDocument(t, owner)
		docID2, _ := createTestDocument(t, owner)

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

		err := admin.Execute(`
			mutation BulkPublishDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
				bulkPublishDocumentVersions(input: $input) {
					documentVersionEdges {
						node { id status }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
				"changelog":   "Bulk publish",
			},
		}, &result)
		require.NoError(t, err, "admin should be able to bulk publish")
		assert.Len(t, result.BulkPublishDocumentVersions.DocumentVersionEdges, 2)
	})

	t.Run("viewer cannot bulk publish", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		docID, _ := createTestDocument(t, owner)

		_, err := viewer.Do(`
			mutation BulkPublishDocumentVersions($input: BulkPublishDocumentVersionsInput!) {
				bulkPublishDocumentVersions(input: $input) {
					documentVersionEdges {
						node { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID},
				"changelog":   "Bulk publish",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to bulk publish")
	})
}

func TestDocument_BulkArchive(t *testing.T) {
	t.Parallel()

	t.Run("owner can bulk archive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		_, err := owner.Do(`
			mutation BulkArchiveDocuments($input: BulkArchiveDocumentsInput!) {
				bulkArchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		})
		require.NoError(t, err, "owner should be able to bulk archive documents")
	})

	t.Run("admin can bulk archive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		_, err := admin.Do(`
			mutation BulkArchiveDocuments($input: BulkArchiveDocumentsInput!) {
				bulkArchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		})
		require.NoError(t, err, "admin should be able to bulk archive documents")
	})

	t.Run("viewer cannot bulk archive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		approverID := factory.CreateUser(owner)
		docID := factory.NewDocument(owner, approverID).Create()

		_, err := viewer.Do(`
			mutation BulkArchiveDocuments($input: BulkArchiveDocumentsInput!) {
				bulkArchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID},
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to bulk archive documents")
	})
}

func TestDocument_BulkUnarchive(t *testing.T) {
	t.Parallel()

	t.Run("owner can bulk unarchive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		for _, docID := range []string{docID1, docID2} {
			_, err := owner.Do(`
				mutation ArchiveDocument($input: ArchiveDocumentInput!) {
					archiveDocument(input: $input) {
						document { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{"documentId": docID},
			})
			require.NoError(t, err, "setup: should archive document")
		}

		_, err := owner.Do(`
			mutation BulkUnarchiveDocuments($input: BulkUnarchiveDocumentsInput!) {
				bulkUnarchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		})
		require.NoError(t, err, "owner should be able to bulk unarchive documents")
	})

	t.Run("admin can bulk unarchive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		approverID := factory.CreateUser(owner)
		docID1 := factory.NewDocument(owner, approverID).Create()
		docID2 := factory.NewDocument(owner, approverID).Create()

		for _, docID := range []string{docID1, docID2} {
			_, err := owner.Do(`
				mutation ArchiveDocument($input: ArchiveDocumentInput!) {
					archiveDocument(input: $input) {
						document { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{"documentId": docID},
			})
			require.NoError(t, err, "setup: should archive document")
		}

		_, err := admin.Do(`
			mutation BulkUnarchiveDocuments($input: BulkUnarchiveDocumentsInput!) {
				bulkUnarchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID1, docID2},
			},
		})
		require.NoError(t, err, "admin should be able to bulk unarchive documents")
	})

	t.Run("viewer cannot bulk unarchive", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		approverID := factory.CreateUser(owner)
		docID := factory.NewDocument(owner, approverID).Create()

		_, err := owner.Do(`
			mutation ArchiveDocument($input: ArchiveDocumentInput!) {
				archiveDocument(input: $input) {
					document { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"documentId": docID},
		})
		require.NoError(t, err, "setup: should archive document")

		_, err = viewer.Do(`
			mutation BulkUnarchiveDocuments($input: BulkUnarchiveDocumentsInput!) {
				bulkUnarchiveDocuments(input: $input) {
					documents { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentIds": []string{docID},
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to bulk unarchive documents")
	})
}
