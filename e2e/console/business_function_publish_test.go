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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestBusinessFunction_PublishBusinessFunctionList(t *testing.T) {
	t.Parallel()

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			createBusinessFunctionForPublish(t, owner, "Test Business Function")

			const query = `
				mutation($input: PublishBusinessFunctionListInput!) {
					publishBusinessFunctionList(input: $input) {
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
				PublishBusinessFunctionList struct {
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
				} `json:"publishBusinessFunctionList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishBusinessFunctionList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishBusinessFunctionList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
			assert.Contains(t, ver.Content, "Test Business Function")
		},
	)

	t.Run(
		"publish with approvers creates draft with quorum",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			const query = `
				mutation($input: PublishBusinessFunctionListInput!) {
					publishBusinessFunctionList(input: $input) {
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
				PublishBusinessFunctionList struct {
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
				} `json:"publishBusinessFunctionList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
						"approverIds":    []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishBusinessFunctionList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)

			ver := result.PublishBusinessFunctionList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"creating second document reuses existing document",
		func(t *testing.T) {
			t.Parallel()

			secondOwner := testutil.NewClient(t, testutil.RoleOwner)

			createBusinessFunctionForPublish(t, secondOwner, "Reuse Test Business Function")

			const query = `
				mutation($input: PublishBusinessFunctionListInput!) {
					publishBusinessFunctionList(input: $input) {
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
				PublishBusinessFunctionList struct {
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
				} `json:"publishBusinessFunctionList"`
			}

			input := map[string]any{
				"input": map[string]any{
					"minor":          false,
					"organizationId": secondOwner.GetOrganizationID(),
				},
			}

			err := secondOwner.Execute(query, input, &result1)
			require.NoError(t, err)

			err = secondOwner.Execute(query, input, &result2)
			require.NoError(t, err)

			doc1 := result1.PublishBusinessFunctionList.DocumentEdge.Node.ID
			doc2 := result2.PublishBusinessFunctionList.DocumentEdge.Node.ID
			assert.Equal(t, doc1, doc2, "should reuse same document")

			ver1Major := result1.PublishBusinessFunctionList.DocumentVersionEdge.Node.Major
			ver2Major := result2.PublishBusinessFunctionList.DocumentVersionEdge.Node.Major

			assert.Equal(t, 1, ver1Major)
			assert.Equal(t, 2, ver2Major)
		},
	)

	t.Run(
		"document linked back to organization",
		func(t *testing.T) {
			t.Parallel()

			thirdOwner := testutil.NewClient(t, testutil.RoleOwner)

			createBusinessFunctionForPublish(t, thirdOwner, "Link Test Business Function")

			const publishQuery = `
				mutation($input: PublishBusinessFunctionListInput!) {
					publishBusinessFunctionList(input: $input) {
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
				PublishBusinessFunctionList struct {
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
				} `json:"publishBusinessFunctionList"`
			}

			err := thirdOwner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": thirdOwner.GetOrganizationID(),
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishBusinessFunctionList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							id
							businessFunctionsDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					ID                        string `json:"id"`
					BusinessFunctionsDocument *struct {
						ID string `json:"id"`
					} `json:"businessFunctionsDocument"`
				} `json:"node"`
			}

			err = thirdOwner.Execute(
				orgQuery,
				map[string]any{"id": thirdOwner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.BusinessFunctionsDocument)
			assert.Equal(t, docID, orgResult.Node.BusinessFunctionsDocument.ID)
		},
	)
}

func TestBusinessFunction_PublishBusinessFunctionList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	createBusinessFunctionForPublish(t, owner, "RBAC Test Business Function")

	const query = `
		mutation($input: PublishBusinessFunctionListInput!) {
			publishBusinessFunctionList(input: $input) {
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
		"viewer cannot publish business function list",
		func(t *testing.T) {
			t.Parallel()

			err := viewer.ExecuteShouldFail(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
					},
				},
			)
			testutil.RequireForbiddenError(t, err)
		},
	)
}

func createBusinessFunctionForPublish(t *testing.T, client *testutil.Client, name string) string {
	t.Helper()

	return createBusinessFunction(t, client, map[string]any{
		"referenceId":    "BF-PUB-001",
		"name":           name,
		"classification": "CRITICAL",
		"mtdMinutes":     60,
		"rtoMinutes":     30,
		"rpoMinutes":     15,
	})
}
