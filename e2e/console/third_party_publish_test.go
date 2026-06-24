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

func TestThirdParty_PublishThirdPartyList(t *testing.T) {
	t.Parallel()

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			factory.CreateThirdParty(owner, factory.Attrs{"name": "Test ThirdParty"})

			const query = `
				mutation($input: PublishThirdPartyListInput!) {
					publishThirdPartyList(input: $input) {
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
				PublishThirdPartyList struct {
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
				} `json:"publishThirdPartyList"`
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

			doc := result.PublishThirdPartyList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishThirdPartyList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
		},
	)

	t.Run(
		"publish with approvers creates draft pending approval",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			const query = `
				mutation($input: PublishThirdPartyListInput!) {
					publishThirdPartyList(input: $input) {
						documentEdge {
							node { id writeMode }
						}
						documentVersionEdge {
							node { id status major }
						}
					}
				}
			`

			var result struct {
				PublishThirdPartyList struct {
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
				} `json:"publishThirdPartyList"`
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

			doc := result.PublishThirdPartyList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)

			ver := result.PublishThirdPartyList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"second publish reuses document and bumps major version",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			factory.CreateThirdParty(owner, factory.Attrs{"name": "Reuse ThirdParty"})

			const query = `
				mutation($input: PublishThirdPartyListInput!) {
					publishThirdPartyList(input: $input) {
						documentEdge { node { id } }
						documentVersionEdge { node { id major } }
					}
				}
			`

			var r1, r2 struct {
				PublishThirdPartyList struct {
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
				} `json:"publishThirdPartyList"`
			}

			input := map[string]any{
				"input": map[string]any{
					"minor":          false,
					"organizationId": owner.GetOrganizationID(),
				},
			}

			err := owner.Execute(query, input, &r1)
			require.NoError(t, err)

			err = owner.Execute(query, input, &r2)
			require.NoError(t, err)

			assert.Equal(t,
				r1.PublishThirdPartyList.DocumentEdge.Node.ID,
				r2.PublishThirdPartyList.DocumentEdge.Node.ID,
				"should reuse same document",
			)
			assert.Equal(t, 1, r1.PublishThirdPartyList.DocumentVersionEdge.Node.Major)
			assert.Equal(t, 2, r2.PublishThirdPartyList.DocumentVersionEdge.Node.Major)
		},
	)

	t.Run(
		"organization thirdPartiesDocument links to published document",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			factory.CreateThirdParty(owner, factory.Attrs{"name": "Linked ThirdParty"})

			const publishQuery = `
				mutation($input: PublishThirdPartyListInput!) {
					publishThirdPartyList(input: $input) {
						documentEdge { node { id } }
						documentVersionEdge { node { id } }
					}
				}
			`

			var publishResult struct {
				PublishThirdPartyList struct {
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
				} `json:"publishThirdPartyList"`
			}

			err := owner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishThirdPartyList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							id
							thirdPartiesDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					ID                   string `json:"id"`
					ThirdPartiesDocument *struct {
						ID string `json:"id"`
					} `json:"thirdPartiesDocument"`
				} `json:"node"`
			}

			err = owner.Execute(
				orgQuery,
				map[string]any{"id": owner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.ThirdPartiesDocument)
			assert.Equal(t, docID, orgResult.Node.ThirdPartiesDocument.ID)
		},
	)
}

func TestThirdParty_PublishThirdPartyList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	factory.CreateThirdParty(owner, factory.Attrs{"name": "RBAC ThirdParty"})

	const query = `
		mutation($input: PublishThirdPartyListInput!) {
			publishThirdPartyList(input: $input) {
				documentEdge { node { id } }
				documentVersionEdge { node { id } }
			}
		}
	`

	t.Run(
		"viewer cannot publish thirdParty list",
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
