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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestDocument_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createDocumentMutation = `
		mutation CreateDocument($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge { node { id } }
			}
		}
	`

	const updateDocumentMutation = `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document { id }
			}
		}
	`

	const deleteDocumentMutation = `
		mutation DeleteDocument($input: DeleteDocumentInput!) {
			deleteDocument(input: $input) {
				deletedDocumentId
			}
		}
	`

	const readDocumentQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Document { id }
			}
		}
	`

	documentCreateInput := func(
		client *testutil.Client,
		title string,
	) map[string]any {
		return map[string]any{
			"organizationId": client.GetOrganizationID().String(),
			"title":          title,
			"content":        testutil.ProseMirrorTextDoc("Test content"),
			"documentType":   "POLICY",
			"classification": "INTERNAL",
		}
	}

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can create",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to create document",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create document",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create document",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createDocumentMutation,
							map[string]any{
								"input": documentCreateInput(
									client,
									factory.SafeName("RBAC create "+string(tt.role)),
								),
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can update",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to update document",
				},
				{
					name:     "admin can update",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to update document",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to update document",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						documentID := factory.NewDocument(ownerClient).
							WithTitle(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateDocumentMutation,
							map[string]any{
								"input": map[string]any{
									"id":             documentID,
									"classification": "CONFIDENTIAL",
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can delete",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to delete document",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete document",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete document",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						documentID := factory.NewDocument(ownerClient).
							WithTitle(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteDocumentMutation,
							map[string]any{
								"input": map[string]any{
									"documentId": documentID,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"read",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				allowMsg    string
				nodePresent string
			}{
				{
					name:        "owner can read",
					role:        testutil.RoleOwner,
					allowMsg:    "owner should be able to read document",
					nodePresent: "owner should receive document data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read document",
					nodePresent: "admin should receive document data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read document",
					nodePresent: "viewer should receive document data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						documentID := factory.NewDocument(ownerClient).
							WithTitle(factory.SafeName("RBAC read " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						var result struct {
							Node *struct {
								ID string `json:"id"`
							} `json:"node"`
						}

						err := client.Execute(
							readDocumentQuery,
							map[string]any{
								"id": documentID,
							},
							&result,
						)
						require.NoError(t, err, tt.allowMsg)
						require.NotNil(t, result.Node, tt.nodePresent)
					},
				)
			}
		},
	)
}
