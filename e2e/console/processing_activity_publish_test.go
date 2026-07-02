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

func TestProcessingActivity_PublishProcessingActivityList(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			_ = factory.NewProcessingActivity(owner).
				WithName("Test Processing Activity").
				WithLawfulBasis("CONSENT").
				Create()

			const query = `
				mutation($input: PublishProcessingActivityListInput!) {
					publishProcessingActivityList(input: $input) {
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
				PublishProcessingActivityList struct {
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
				} `json:"publishProcessingActivityList"`
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

			doc := result.PublishProcessingActivityList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishProcessingActivityList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
			assert.Contains(t, ver.Content, "Test Processing Activity")
		},
	)

	t.Run(
		"publish with approvers creates draft pending approval",
		func(t *testing.T) {
			t.Parallel()

			_ = factory.NewProcessingActivity(owner).
				WithName("Approval PA").
				WithLawfulBasis("CONSENT").
				Create()

			const query = `
				mutation($input: PublishProcessingActivityListInput!) {
					publishProcessingActivityList(input: $input) {
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
				PublishProcessingActivityList struct {
					DocumentVersionEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishProcessingActivityList"`
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

			ver := result.PublishProcessingActivityList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"second publish reuses existing document and bumps major version",
		func(t *testing.T) {
			t.Parallel()

			secondOwner := testutil.NewClient(t, testutil.RoleOwner)

			_ = factory.NewProcessingActivity(secondOwner).
				WithName("Reuse PA").
				WithLawfulBasis("CONSENT").
				Create()

			const query = `
				mutation($input: PublishProcessingActivityListInput!) {
					publishProcessingActivityList(input: $input) {
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
				PublishProcessingActivityList struct {
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
				} `json:"publishProcessingActivityList"`
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

			assert.Equal(
				t,
				result1.PublishProcessingActivityList.DocumentEdge.Node.ID,
				result2.PublishProcessingActivityList.DocumentEdge.Node.ID,
				"should reuse the same document",
			)
			assert.Equal(t, 1, result1.PublishProcessingActivityList.DocumentVersionEdge.Node.Major)
			assert.Equal(t, 2, result2.PublishProcessingActivityList.DocumentVersionEdge.Node.Major)
		},
	)

	t.Run(
		"document linked back to organization via processingActivitiesDocument",
		func(t *testing.T) {
			t.Parallel()

			thirdOwner := testutil.NewClient(t, testutil.RoleOwner)

			_ = factory.NewProcessingActivity(thirdOwner).
				WithName("Linked PA").
				WithLawfulBasis("CONSENT").
				Create()

			const publishQuery = `
				mutation($input: PublishProcessingActivityListInput!) {
					publishProcessingActivityList(input: $input) {
						documentEdge {
							node { id }
						}
					}
				}
			`

			var publishResult struct {
				PublishProcessingActivityList struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
				} `json:"publishProcessingActivityList"`
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

			docID := publishResult.PublishProcessingActivityList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							id
							processingActivitiesDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					ID                           string `json:"id"`
					ProcessingActivitiesDocument *struct {
						ID string `json:"id"`
					} `json:"processingActivitiesDocument"`
				} `json:"node"`
			}

			err = thirdOwner.Execute(
				orgQuery,
				map[string]any{"id": thirdOwner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.ProcessingActivitiesDocument)
			assert.Equal(t, docID, orgResult.Node.ProcessingActivitiesDocument.ID)
		},
	)
}

func TestProcessingActivity_PublishProcessingActivityList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	_ = factory.NewProcessingActivity(owner).
		WithName("RBAC PA").
		WithLawfulBasis("CONSENT").
		Create()

	const query = `
		mutation($input: PublishProcessingActivityListInput!) {
			publishProcessingActivityList(input: $input) {
				documentEdge {
					node { id }
				}
			}
		}
	`

	t.Run(
		"viewer cannot publish processing activity list",
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
