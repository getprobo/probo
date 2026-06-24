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

func TestDatum_PublishDataList(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			factory.NewDatum(owner, owner.GetProfileID().String()).
				WithName("Test Data Item").
				Create()

			const query = `
				mutation($input: PublishDataListInput!) {
					publishDataList(input: $input) {
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
				PublishDataList struct {
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
				} `json:"publishDataList"`
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

			doc := result.PublishDataList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishDataList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
			assert.Contains(t, ver.Content, "Test Data Item")
		},
	)

	t.Run(
		"publish with approvers creates draft with quorum",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: PublishDataListInput!) {
					publishDataList(input: $input) {
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
				PublishDataList struct {
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
				} `json:"publishDataList"`
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

			doc := result.PublishDataList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)

			ver := result.PublishDataList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"creating second document reuses existing document",
		func(t *testing.T) {
			t.Parallel()

			secondOwner := testutil.NewClient(t, testutil.RoleOwner)

			factory.NewDatum(secondOwner, secondOwner.GetProfileID().String()).
				WithName("Reuse Test Data").
				Create()

			const query = `
				mutation($input: PublishDataListInput!) {
					publishDataList(input: $input) {
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
				PublishDataList struct {
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
				} `json:"publishDataList"`
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

			doc1 := result1.PublishDataList.DocumentEdge.Node.ID
			doc2 := result2.PublishDataList.DocumentEdge.Node.ID
			assert.Equal(t, doc1, doc2, "should reuse same document")

			ver1Major := result1.PublishDataList.DocumentVersionEdge.Node.Major
			ver2Major := result2.PublishDataList.DocumentVersionEdge.Node.Major

			assert.Equal(t, 1, ver1Major)
			assert.Equal(t, 2, ver2Major)
		},
	)

	t.Run(
		"document linked back to organization",
		func(t *testing.T) {
			t.Parallel()

			thirdOwner := testutil.NewClient(t, testutil.RoleOwner)

			factory.NewDatum(thirdOwner, thirdOwner.GetProfileID().String()).
				WithName("Link Test Data").
				Create()

			const publishQuery = `
				mutation($input: PublishDataListInput!) {
					publishDataList(input: $input) {
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
				PublishDataList struct {
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
				} `json:"publishDataList"`
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

			docID := publishResult.PublishDataList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							id
							dataListDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					ID               string `json:"id"`
					DataListDocument *struct {
						ID string `json:"id"`
					} `json:"dataListDocument"`
				} `json:"node"`
			}

			err = thirdOwner.Execute(
				orgQuery,
				map[string]any{"id": thirdOwner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.DataListDocument)
			assert.Equal(t, docID, orgResult.Node.DataListDocument.ID)
		},
	)
}

func TestDatum_PublishDataList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	factory.NewDatum(owner, owner.GetProfileID().String()).
		WithName("RBAC Test Data").
		Create()

	const query = `
		mutation($input: PublishDataListInput!) {
			publishDataList(input: $input) {
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
		"viewer cannot publish data list",
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
