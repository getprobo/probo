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

func TestStatementOfApplicability_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"create a statement of applicability",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: CreateStatementOfApplicabilityInput!) {
					createStatementOfApplicability(input: $input) {
						statementOfApplicabilityEdge {
							node {
								id
								name
								createdAt
								updatedAt
							}
						}
					}
				}
			`

			name := factory.SafeName("SOA")

			var result struct {
				CreateStatementOfApplicability struct {
					StatementOfApplicabilityEdge struct {
						Node struct {
							ID        string `json:"id"`
							Name      string `json:"name"`
							CreatedAt string `json:"createdAt"`
							UpdatedAt string `json:"updatedAt"`
						} `json:"node"`
					} `json:"statementOfApplicabilityEdge"`
				} `json:"createStatementOfApplicability"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID().String(),
						"name":           name,
					},
				},
				&result,
			)

			require.NoError(t, err)
			node := result.CreateStatementOfApplicability.StatementOfApplicabilityEdge.Node
			assert.NotEmpty(t, node.ID)
			assert.Equal(t, name, node.Name)
		},
	)

	t.Run(
		"create with default approvers",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: CreateStatementOfApplicabilityInput!) {
					createStatementOfApplicability(input: $input) {
						statementOfApplicabilityEdge {
							node {
								id
								name
								defaultApprovers { id }
							}
						}
					}
				}
			`

			name := factory.SafeName("SOA")

			var result struct {
				CreateStatementOfApplicability struct {
					StatementOfApplicabilityEdge struct {
						Node struct {
							ID               string `json:"id"`
							Name             string `json:"name"`
							DefaultApprovers []struct {
								ID string `json:"id"`
							} `json:"defaultApprovers"`
						} `json:"node"`
					} `json:"statementOfApplicabilityEdge"`
				} `json:"createStatementOfApplicability"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId":     owner.GetOrganizationID().String(),
						"name":               name,
						"defaultApproverIds": []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)

			require.NoError(t, err)
			node := result.CreateStatementOfApplicability.StatementOfApplicabilityEdge.Node
			assert.NotEmpty(t, node.ID)
			assert.Len(t, node.DefaultApprovers, 1)
			assert.Equal(t, owner.GetProfileID().String(), node.DefaultApprovers[0].ID)
		},
	)
}

func TestStatementOfApplicability_CreateDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"create document without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			frameworkID := factory.NewFramework(owner).Create()
			controlID := factory.NewControl(owner, frameworkID).Create()

			soaID := factory.NewStatementOfApplicability(owner).Create()
			factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

			const query = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
						documentEdge {
							node {
								id
								contentSource
								status
							}
						}
						documentVersionEdge {
							node {
								id
								title
								documentType
								orientation
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
				PublishStatementOfApplicability struct {
					DocumentEdge struct {
						Node struct {
							ID            string `json:"id"`
							ContentSource string `json:"contentSource"`
							Status        string `json:"status"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID           string `json:"id"`
							Title        string `json:"title"`
							DocumentType string `json:"documentType"`
							Orientation  string `json:"orientation"`
							Status       string `json:"status"`
							Major        int    `json:"major"`
							Minor        int    `json:"minor"`
							Content      string `json:"content"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishStatementOfApplicability"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"statementOfApplicabilityId": soaID,
					},
				},
				&result,
			)

			require.NoError(t, err)

			doc := result.PublishStatementOfApplicability.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.ContentSource)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishStatementOfApplicability.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "STATEMENT_OF_APPLICABILITY", ver.DocumentType)
			assert.Equal(t, "LANDSCAPE", ver.Orientation)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "Purpose")
		},
	)

	t.Run(
		"create document with approvers creates draft with quorum",
		func(t *testing.T) {
			t.Parallel()

			frameworkID := factory.NewFramework(owner).Create()
			controlID := factory.NewControl(owner, frameworkID).Create()

			soaID := factory.NewStatementOfApplicability(owner).Create()
			factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

			const query = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
						documentEdge {
							node {
								id
								contentSource
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
				PublishStatementOfApplicability struct {
					DocumentEdge struct {
						Node struct {
							ID            string `json:"id"`
							ContentSource string `json:"contentSource"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
							Major  int    `json:"major"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishStatementOfApplicability"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"statementOfApplicabilityId": soaID,
						"approverIds":                []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)

			require.NoError(t, err)

			doc := result.PublishStatementOfApplicability.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.ContentSource)

			ver := result.PublishStatementOfApplicability.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "PENDING_APPROVAL", ver.Status)
		},
	)

	t.Run(
		"creating second document reuses existing document",
		func(t *testing.T) {
			t.Parallel()

			frameworkID := factory.NewFramework(owner).Create()
			controlID := factory.NewControl(owner, frameworkID).Create()

			soaID := factory.NewStatementOfApplicability(owner).Create()
			factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

			const query = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
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
				PublishStatementOfApplicability struct {
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
				} `json:"publishStatementOfApplicability"`
			}

			input := map[string]any{
				"input": map[string]any{
					"statementOfApplicabilityId": soaID,
				},
			}

			err := owner.Execute(query, input, &result1)
			require.NoError(t, err)

			err = owner.Execute(query, input, &result2)
			require.NoError(t, err)

			doc1 := result1.PublishStatementOfApplicability.DocumentEdge.Node.ID
			doc2 := result2.PublishStatementOfApplicability.DocumentEdge.Node.ID
			assert.Equal(t, doc1, doc2, "should reuse same document")

			ver1Major := result1.PublishStatementOfApplicability.DocumentVersionEdge.Node.Major
			ver2Major := result2.PublishStatementOfApplicability.DocumentVersionEdge.Node.Major
			assert.Equal(t, 1, ver1Major)
			assert.Equal(t, 2, ver2Major)
		},
	)

	t.Run(
		"document linked back to SOA",
		func(t *testing.T) {
			t.Parallel()

			frameworkID := factory.NewFramework(owner).Create()
			controlID := factory.NewControl(owner, frameworkID).Create()

			soaID := factory.NewStatementOfApplicability(owner).Create()
			factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

			const createQuery = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
						documentEdge {
							node { id }
						}
						documentVersionEdge {
							node { id }
						}
					}
				}
			`

			var createResult struct {
				PublishStatementOfApplicability struct {
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
				} `json:"publishStatementOfApplicability"`
			}

			err := owner.Execute(
				createQuery,
				map[string]any{
					"input": map[string]any{
						"statementOfApplicabilityId": soaID,
					},
				},
				&createResult,
			)
			require.NoError(t, err)

			docID := createResult.PublishStatementOfApplicability.DocumentEdge.Node.ID

			const soaQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on StatementOfApplicability {
							id
							document { id }
						}
					}
				}
			`

			var soaResult struct {
				Node struct {
					ID       string `json:"id"`
					Document *struct {
						ID string `json:"id"`
					} `json:"document"`
				} `json:"node"`
			}

			err = owner.Execute(soaQuery, map[string]any{"id": soaID}, &soaResult)
			require.NoError(t, err)
			require.NotNil(t, soaResult.Node.Document)
			assert.Equal(t, docID, soaResult.Node.Document.ID)
		},
	)
}

func TestStatementOfApplicability_CreateDocument_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	frameworkID := factory.NewFramework(owner).Create()
	controlID := factory.NewControl(owner, frameworkID).Create()

	soaID := factory.NewStatementOfApplicability(owner).Create()
	factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

	const query = `
		mutation($input: PublishStatementOfApplicabilityInput!) {
			publishStatementOfApplicability(input: $input) {
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
		"viewer cannot create document",
		func(t *testing.T) {
			t.Parallel()

			err := viewer.ExecuteShouldFail(
				query,
				map[string]any{
					"input": map[string]any{
						"statementOfApplicabilityId": soaID,
					},
				},
			)
			testutil.RequireForbiddenError(t, err)
		},
	)
}

func TestStatementOfApplicability_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	soaID := factory.NewStatementOfApplicability(org1Owner).Create()

	t.Run(
		"cannot create document for another org SOA",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
						documentEdge {
							node { id }
						}
						documentVersionEdge {
							node { id }
						}
					}
				}
			`

			err := org2Owner.ExecuteShouldFail(
				query,
				map[string]any{
					"input": map[string]any{
						"statementOfApplicabilityId": soaID,
					},
				},
			)
			require.Error(t, err)
		},
	)
}
