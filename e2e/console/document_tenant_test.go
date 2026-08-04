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

func TestDocument_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	documentID := factory.NewDocument(org1Owner).WithTitle("Org1 Document").Create()

	t.Run("cannot read document from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						id
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": documentID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "document")
	})

	t.Run("cannot update document from another organization", func(t *testing.T) {
		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":             documentID,
				"classification": "CONFIDENTIAL",
			},
		})
		require.Error(t, err, "Should not be able to update document from another org")
	})

	t.Run("cannot delete document from another organization", func(t *testing.T) {
		query := `
			mutation DeleteDocument($input: DeleteDocumentInput!) {
				deleteDocument(input: $input) {
					deletedDocumentId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"documentId": documentID,
			},
		})
		require.Error(t, err, "Should not be able to delete document from another org")
	})

	t.Run("cannot list documents from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						documents(first: 100) {
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

		var result struct {
			Node struct {
				Documents struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)
		if err == nil {
			for _, edge := range result.Node.Documents.Edges {
				assert.NotEqual(t, documentID, edge.Node.ID, "Should not see document from another org")
			}
		}
	})

	t.Run("cannot create document referencing a default approver from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)

		_, err := org1Owner.Do(`
			mutation($input: CreateDocumentInput!) {
				createDocument(input: $input) {
					documentEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":     org1Owner.GetOrganizationID().String(),
				"title":              factory.SafeName("Document"),
				"content":            testutil.ProseMirrorTextDoc("Document content"),
				"documentType":       "POLICY",
				"classification":     "INTERNAL",
				"defaultApproverIds": []string{org2ProfileID},
			},
		})
		require.Error(t, err, "must not accept a defaultApproverId belonging to another organization")
	})

	t.Run("cannot update document to reference a default approver from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)
		otherDocumentID := factory.NewDocument(org1Owner).WithTitle("Org1 Document for DefaultApproverIDs").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":                 otherDocumentID,
				"defaultApproverIds": []string{org2ProfileID},
			},
		})
		require.Error(t, err, "must not accept a defaultApproverId belonging to another organization")
	})

	t.Run("cannot request approval referencing an approver from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)
		otherDocumentID := factory.NewDocument(org1Owner).WithTitle("Org1 Document for ApproverIDs").Create()

		_, err := org1Owner.Do(`
			mutation($input: PublishDocumentInput!) {
				publishDocument(input: $input) {
					approvalQuorum { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  otherDocumentID,
				"approverIds": []string{org2ProfileID},
				"changelog":   "Test changelog",
			},
		})
		require.Error(t, err, "must not accept an approverId belonging to another organization")
	})
}
