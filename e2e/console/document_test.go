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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestDocument_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"title":          "Security Policy",
				"content":        "This is the security policy content.",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			assertField: "title",
			assertValue: "Security Policy",
		},
		{
			name: "with POLICY type",
			input: map[string]any{
				"title":          "Policy Document",
				"content":        "Policy content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			assertField: "documentType",
			assertValue: "POLICY",
		},
		{
			name: "with PROCEDURE type",
			input: map[string]any{
				"title":          "Procedure Document",
				"content":        "Procedure content",
				"documentType":   "PROCEDURE",
				"classification": "INTERNAL",
			},
			assertField: "documentType",
			assertValue: "PROCEDURE",
		},
		{
			name: "with ISMS type",
			input: map[string]any{
				"title":          "ISMS Document",
				"content":        "ISMS content",
				"documentType":   "ISMS",
				"classification": "INTERNAL",
			},
			assertField: "documentType",
			assertValue: "ISMS",
		},
		{
			name: "with OTHER type",
			input: map[string]any{
				"title":          "Other Document",
				"content":        "Other content",
				"documentType":   "OTHER",
				"classification": "INTERNAL",
			},
			assertField: "documentType",
			assertValue: "OTHER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateDocument($input: CreateDocumentInput!) {
					createDocument(input: $input) {
						documentEdge {
							node {
								id
								title
								documentType
								classification
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"approverIds":    []string{approverProfileID.String()},
			}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateDocument struct {
					DocumentEdge struct {
						Node struct {
							ID             string `json:"id"`
							Title          string `json:"title"`
							DocumentType   string `json:"documentType"`
							Classification string `json:"classification"`
						} `json:"node"`
					} `json:"documentEdge"`
				} `json:"createDocument"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateDocument.DocumentEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "title":
				assert.Equal(t, tt.assertValue, node.Title)
			case "documentType":
				assert.Equal(t, tt.assertValue, node.DocumentType)
			}
		})
	}
}

func TestDocument_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		skipApprover      bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"title":          "Test Document",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing approverIds",
			input: map[string]any{
				"title":          "Test Document",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			skipApprover:      true,
			wantErrorContains: "approverIds",
		},
		{
			name: "title with HTML tags",
			input: map[string]any{
				"title":          "<script>alert('xss')</script>",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "title with angle brackets",
			input: map[string]any{
				"title":          "Test < Document",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "angle brackets",
		},
		{
			name: "title with newline",
			input: map[string]any{
				"title":          "Test\nDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "newline",
		},
		{
			name: "title with carriage return",
			input: map[string]any{
				"title":          "Test\rDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "carriage return",
		},
		{
			name: "title with null byte",
			input: map[string]any{
				"title":          "Test\x00Document",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "control character",
		},
		{
			name: "title with tab character",
			input: map[string]any{
				"title":          "Test\tDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "control character",
		},
		{
			name: "title with zero-width space",
			input: map[string]any{
				"title":          "Test\u200BDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "title with zero-width joiner",
			input: map[string]any{
				"title":          "Test\u200DDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "title with right-to-left override",
			input: map[string]any{
				"title":          "Test\u202EDocument",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}
			if !tt.skipApprover {
				input["approverIds"] = []string{approverProfileID.String()}
			}
			for k, v := range tt.input {
				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDocument_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "update title",
			setup: func() string {
				return factory.NewDocument(owner, approverProfileID.String()).
					WithTitle("Document to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":    id,
					"title": "Updated Document Title",
				}
			},
			assertField: "title",
			assertValue: "Updated Document Title",
		},
		{
			name: "update document type",
			setup: func() string {
				return factory.NewDocument(owner, approverProfileID.String()).
					WithTitle("Type Test").
					WithDocumentType("POLICY").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "documentType": "PROCEDURE"}
			},
			assertField: "documentType",
			assertValue: "PROCEDURE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documentID := tt.setup()

			query := `
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document {
							id
							title
							documentType
						}
					}
				}
			`

			var result struct {
				UpdateDocument struct {
					Document struct {
						ID           string `json:"id"`
						Title        string `json:"title"`
						DocumentType string `json:"documentType"`
					} `json:"document"`
				} `json:"updateDocument"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(documentID)}, &result)
			require.NoError(t, err)

			doc := result.UpdateDocument.Document
			switch tt.assertField {
			case "title":
				assert.Equal(t, tt.assertValue, doc.Title)
			case "documentType":
				assert.Equal(t, tt.assertValue, doc.DocumentType)
			}
		})
	}
}

func TestDocument_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	baseDocumentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("Validation Test Document").Create()

	tests := []struct {
		name              string
		setup             func() string
		input             func(id string) map[string]any
		wantErrorContains string
	}{
		{
			name:  "invalid ID format",
			setup: func() string { return "invalid-id-format" },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test"}
			},
			wantErrorContains: "base64",
		},
		{
			name:  "title with HTML tags",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "title with angle brackets",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test < Document"}
			},
			wantErrorContains: "angle brackets",
		},
		{
			name:  "title with newline",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test\nDocument"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "title with carriage return",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test\rDocument"}
			},
			wantErrorContains: "carriage return",
		},
		{
			name:  "title with null byte",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test\x00Document"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "title with zero-width space",
			setup: func() string { return baseDocumentID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "title": "Test\u200BDocument"}
			},
			wantErrorContains: "zero-width",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documentID := tt.setup()

			query := `
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(documentID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDocument_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	t.Run("delete existing document", func(t *testing.T) {
		documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("Document to Delete").Create()

		query := `
			mutation DeleteDocument($input: DeleteDocumentInput!) {
				deleteDocument(input: $input) {
					deletedDocumentId
				}
			}
		`

		var result struct {
			DeleteDocument struct {
				DeletedDocumentID string `json:"deletedDocumentId"`
			} `json:"deleteDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"documentId": documentID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, documentID, result.DeleteDocument.DeletedDocumentID)
	})
}

func TestDocument_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		documentID        string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			documentID:        "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteDocument($input: DeleteDocumentInput!) {
					deleteDocument(input: $input) {
						deletedDocumentId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"documentId": tt.documentID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDocument_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	documentTitles := []string{"Document A", "Document B", "Document C"}
	for _, title := range documentTitles {
		factory.NewDocument(owner, approverProfileID.String()).WithTitle(title).Create()
	}

	query := `
		query GetDocuments($id: ID!) {
			node(id: $id) {
				... on Organization {
					documents(first: 10) {
						edges {
							node {
								id
								title
							}
						}
						totalCount
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
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"documents"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Documents.TotalCount, 3)
}

func TestDocument_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						title
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo",
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestDocument_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateDocument($input: CreateDocumentInput!) {
				createDocument(input: $input) {
					documentEdge {
						node {
							id
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		var result struct {
			CreateDocument struct {
				DocumentEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"documentEdge"`
			} `json:"createDocument"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"approverIds":    []string{approverProfileID.String()},
				"title":          "Timestamp Test Document",
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateDocument.DocumentEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						createdAt
						updatedAt
					}
				}
			}
		`

		var getResult struct {
			Node struct {
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"node"`
		}

		err := owner.Execute(getQuery, map[string]any{"id": documentID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateDocument struct {
				Document struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"document"`
			} `json:"updateDocument"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":    documentID,
				"title": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		doc := updateResult.UpdateDocument.Document
		testutil.AssertTimestampsOnUpdate(t, doc.CreatedAt, doc.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestDocument_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("SubResolver Test Document").Create()

	t.Run("approvers sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						approvers {
							totalCount
							edges {
								node {
									id
									fullName
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID        string `json:"id"`
				Approvers struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID       string `json:"id"`
							FullName string `json:"fullName"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"approvers"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": documentID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.Approvers.TotalCount)
		require.Len(t, result.Node.Approvers.Edges, 1)
		assert.Equal(t, approverProfileID.String(), result.Node.Approvers.Edges[0].Node.ID)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						organization {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": documentID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestDocument_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

			_, err := owner.Do(`
				mutation CreateDocument($input: CreateDocumentInput!) {
					createDocument(input: $input) {
						documentEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"approverIds":    []string{approverProfileID.String()},
					"title":          "RBAC Test Document",
					"content":        "Test content",
					"documentType":   "POLICY",
					"classification": "INTERNAL",
				},
			})
			require.NoError(t, err, "owner should be able to create document")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

			_, err := admin.Do(`
				mutation CreateDocument($input: CreateDocumentInput!) {
					createDocument(input: $input) {
						documentEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": admin.GetOrganizationID().String(),
					"approverIds":    []string{approverProfileID.String()},
					"title":          "RBAC Test Document",
					"content":        "Test content",
					"documentType":   "POLICY",
					"classification": "INTERNAL",
				},
			})
			require.NoError(t, err, "admin should be able to create document")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

			_, err := viewer.Do(`
				mutation CreateDocument($input: CreateDocumentInput!) {
					createDocument(input: $input) {
						documentEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": viewer.GetOrganizationID().String(),
					"approverIds":    []string{approverProfileID.String()},
					"title":          "RBAC Test Document",
					"content":        "Test content",
					"documentType":   "POLICY",
					"classification": "INTERNAL",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create document")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":    documentID,
					"title": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update document")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":    documentID,
					"title": "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update document")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateDocument($input: UpdateDocumentInput!) {
					updateDocument(input: $input) {
						document { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":    documentID,
					"title": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update document")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteDocument($input: DeleteDocumentInput!) {
					deleteDocument(input: $input) {
						deletedDocumentId
					}
				}
			`, map[string]any{
				"input": map[string]any{"documentId": documentID},
			})
			require.NoError(t, err, "owner should be able to delete document")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteDocument($input: DeleteDocumentInput!) {
					deleteDocument(input: $input) {
						deletedDocumentId
					}
				}
			`, map[string]any{
				"input": map[string]any{"documentId": documentID},
			})
			require.NoError(t, err, "admin should be able to delete document")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteDocument($input: DeleteDocumentInput!) {
					deleteDocument(input: $input) {
						deletedDocumentId
					}
				}
			`, map[string]any{
				"input": map[string]any{"documentId": documentID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete document")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Document { id title }
					}
				}
			`, map[string]any{"id": documentID}, &result)
			require.NoError(t, err, "owner should be able to read document")
			require.NotNil(t, result.Node, "owner should receive document data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Document { id title }
					}
				}
			`, map[string]any{"id": documentID}, &result)
			require.NoError(t, err, "admin should be able to read document")
			require.NotNil(t, result.Node, "admin should receive document data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
			documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Document { id title }
					}
				}
			`, map[string]any{"id": documentID}, &result)
			require.NoError(t, err, "viewer should be able to read document")
			require.NotNil(t, result.Node, "viewer should receive document data")
		})
	})
}

func TestDocument_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	longTitle := strings.Repeat("a", 1001)

	t.Run("create", func(t *testing.T) {
		query := `
			mutation CreateDocument($input: CreateDocumentInput!) {
				createDocument(input: $input) {
					documentEdge {
						node { id }
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"approverIds":    []string{approverProfileID.String()},
				"title":          longTitle,
				"content":        "Test content",
				"documentType":   "POLICY",
				"classification": "INTERNAL",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("update", func(t *testing.T) {
		documentID := factory.NewDocument(owner, approverProfileID.String()).WithTitle("Max Length Test").Create()

		query := `
			mutation UpdateDocument($input: UpdateDocumentInput!) {
				updateDocument(input: $input) {
					document { id }
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":    documentID,
				"title": longTitle,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})
}

func TestDocument_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	for i := 0; i < 5; i++ {
		factory.NewDocument(owner, approverProfileID.String()).
			WithTitle(fmt.Sprintf("Pagination Document %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						documents(first: 2) {
							edges {
								node { id title }
								cursor
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
								startCursor
								endCursor
							}
							totalCount
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
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := owner.Execute(
			query,
			map[string]any{
				"id": owner.GetOrganizationID().String(),
			},
			&result,
		)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.Documents.Edges), result.Node.Documents.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Documents.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.Documents.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						documents(first: 2, after: $after) {
							edges {
								node { id title }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var resultAfter struct {
			Node struct {
				Documents struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"documents"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Documents.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Documents.Edges), resultAfter.Node.Documents.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						documents(last: 2) {
							edges {
								node { id title }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
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
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Documents.Edges), result.Node.Documents.PageInfo, 2, true)
	})
}

func TestDocument_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, org1Owner).GetProfileID()
	documentID := factory.NewDocument(org1Owner, approverProfileID.String()).WithTitle("Org1 Document").Create()

	t.Run("cannot read document from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Document {
						id
						title
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID    string `json:"id"`
				Title string `json:"title"`
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
				"id":    documentID,
				"title": "Hijacked Document",
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
									title
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
							ID    string `json:"id"`
							Title string `json:"title"`
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
}

func TestDocument_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	approverProfileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	factory.NewDocument(owner, approverProfileID.String()).WithTitle("AAA Order Test").Create()
	factory.NewDocument(owner, approverProfileID.String()).WithTitle("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: DocumentOrder) {
				node(id: $id) {
					... on Organization {
						documents(first: 100, orderBy: $orderBy) {
							edges {
								node {
									id
									createdAt
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
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"documents"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "CREATED_AT",
				"direction": "DESC",
			},
		}, &result)
		require.NoError(t, err)

		times := make([]time.Time, len(result.Node.Documents.Edges))
		for i, edge := range result.Node.Documents.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}
