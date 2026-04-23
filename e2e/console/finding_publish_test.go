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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestFinding_PublishFindingList(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			createFindingForPublish(t, owner, "Test Finding")

			const query = `
				mutation($input: PublishFindingListInput!) {
					publishFindingList(input: $input) {
						documentEdge {
							node {
								id
								writeMode
								status
							}
						}
						documentVersionEdge {
							node {
								id
								title
								documentType
								status
								major
								minor
								content
							}
						}
					}
				}
			`

			var result struct {
				PublishFindingList struct {
					DocumentEdge struct {
						Node struct {
							ID        string `json:"id"`
							WriteMode string `json:"writeMode"`
							Status    string `json:"status"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID           string `json:"id"`
							Title        string `json:"title"`
							DocumentType string `json:"documentType"`
							Status       string `json:"status"`
							Major        int    `json:"major"`
							Minor        int    `json:"minor"`
							Content      string `json:"content"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishFindingList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&result,
			)

			require.NoError(t, err)

			doc := result.PublishFindingList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishFindingList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
			assert.Contains(t, ver.Content, "Test Finding")
		},
	)

	t.Run(
		"publish with approvers creates draft with quorum",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: PublishFindingListInput!) {
					publishFindingList(input: $input) {
						documentEdge {
							node {
								id
								writeMode
							}
						}
						documentVersionEdge {
							node {
								id
								status
								major
							}
						}
					}
				}
			`

			var result struct {
				PublishFindingList struct {
					DocumentEdge struct {
						Node struct {
							ID        string `json:"id"`
							WriteMode string `json:"writeMode"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
							Major  int    `json:"major"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishFindingList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
						"approverIds":    []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)

			require.NoError(t, err)

			doc := result.PublishFindingList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)

			ver := result.PublishFindingList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"creating second document reuses existing document",
		func(t *testing.T) {
			t.Parallel()

			secondOwner := testutil.NewClient(t, testutil.RoleOwner)

			createFindingForPublish(t, secondOwner, "Reuse Test Finding")

			const query = `
				mutation($input: PublishFindingListInput!) {
					publishFindingList(input: $input) {
						documentEdge {
							node { id }
						}
						documentVersionEdge {
							node { id major }
						}
					}
				}
			`

			var result1, result2 struct {
				PublishFindingList struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID    string `json:"id"`
							Major int    `json:"major"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishFindingList"`
			}

			input := map[string]any{
				"input": map[string]any{
					"organizationId": secondOwner.GetOrganizationID(),
				},
			}

			err := secondOwner.Execute(query, input, &result1)
			require.NoError(t, err)

			err = secondOwner.Execute(query, input, &result2)
			require.NoError(t, err)

			doc1 := result1.PublishFindingList.DocumentEdge.Node.ID
			doc2 := result2.PublishFindingList.DocumentEdge.Node.ID
			assert.Equal(t, doc1, doc2, "should reuse same document")

			ver1Major := result1.PublishFindingList.DocumentVersionEdge.Node.Major
			ver2Major := result2.PublishFindingList.DocumentVersionEdge.Node.Major
			assert.Equal(t, 1, ver1Major)
			assert.Equal(t, 2, ver2Major)
		},
	)

	t.Run(
		"document linked back to organization",
		func(t *testing.T) {
			t.Parallel()

			thirdOwner := testutil.NewClient(t, testutil.RoleOwner)

			createFindingForPublish(t, thirdOwner, "Link Test Finding")

			const publishQuery = `
				mutation($input: PublishFindingListInput!) {
					publishFindingList(input: $input) {
						documentEdge {
							node { id }
						}
						documentVersionEdge {
							node { id }
						}
					}
				}
			`

			var publishResult struct {
				PublishFindingList struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishFindingList"`
			}

			err := thirdOwner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId": thirdOwner.GetOrganizationID(),
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishFindingList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							id
							findingsDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					ID               string `json:"id"`
					FindingsDocument *struct {
						ID string `json:"id"`
					} `json:"findingsDocument"`
				} `json:"node"`
			}

			err = thirdOwner.Execute(
				orgQuery,
				map[string]any{"id": thirdOwner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.FindingsDocument)
			assert.Equal(t, docID, orgResult.Node.FindingsDocument.ID)
		},
	)
}

func TestFinding_PublishFindingList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	createFindingForPublish(t, owner, "RBAC Test Finding")

	const query = `
		mutation($input: PublishFindingListInput!) {
			publishFindingList(input: $input) {
				documentEdge {
					node { id }
				}
				documentVersionEdge {
					node { id }
				}
			}
		}
	`

	t.Run(
		"viewer cannot publish finding list",
		func(t *testing.T) {
			t.Parallel()

			err := viewer.ExecuteShouldFail(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
					},
				},
			)
			testutil.RequireForbiddenError(t, err)
		},
	)
}

func createFindingForPublish(t *testing.T, client *testutil.Client, description string) string {
	t.Helper()

	const query = `
		mutation($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	var result struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": client.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"description":    description,
			"status":         "OPEN",
			"priority":       "MEDIUM",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateFinding.FindingEdge.Node.ID
}
