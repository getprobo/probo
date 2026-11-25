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

type (
	signaturePage struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node struct {
				ID       string `json:"id"`
				SignedBy struct {
					FullName string `json:"fullName"`
				} `json:"signedBy"`
			} `json:"node"`
		} `json:"edges"`
		PageInfo struct {
			HasNextPage bool    `json:"hasNextPage"`
			EndCursor   *string `json:"endCursor"`
		} `json:"pageInfo"`
	}

	decisionPage struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node struct {
				Approver struct {
					FullName string `json:"fullName"`
				} `json:"approver"`
			} `json:"node"`
		} `json:"edges"`
	}
)

func TestDocumentVersion_SignaturesPageAndSearch(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)
	versionID := latestDocumentVersionID(t, owner, docID)

	names := []string{
		factory.SafeName("Alpha"),
		factory.SafeName("Bravo"),
		factory.SafeName("Charlie"),
	}
	for _, name := range names {
		requestDocumentSignature(t, owner, versionID, factory.CreateUser(owner, factory.Attrs{"fullName": name}))
	}

	t.Run("pages past the first page", func(t *testing.T) {
		t.Parallel()

		first := loadSignaturePage(t, owner, versionID, nil, nil)
		require.Len(t, first.Edges, 2)
		require.True(t, first.PageInfo.HasNextPage)
		require.NotNil(t, first.PageInfo.EndCursor)

		next := loadSignaturePage(t, owner, versionID, first.PageInfo.EndCursor, nil)
		require.Len(t, next.Edges, 1)
		assert.False(t, next.PageInfo.HasNextPage)

		seen := map[string]bool{}

		for _, page := range []signaturePage{first, next} {
			for _, edge := range page.Edges {
				seen[edge.Node.ID] = true
			}
		}

		assert.Len(t, seen, 3, "the second page continues after the first")
	})

	t.Run("filters by signatory name", func(t *testing.T) {
		t.Parallel()

		page := loadSignaturePage(t, owner, versionID, nil, map[string]any{"query": names[1][:5]})
		require.Len(t, page.Edges, 1)
		assert.Equal(t, names[1], page.Edges[0].Node.SignedBy.FullName)
		assert.Equal(t, 1, page.TotalCount)
	})

	t.Run("treats LIKE wildcards as text", func(t *testing.T) {
		t.Parallel()

		page := loadSignaturePage(t, owner, versionID, nil, map[string]any{"query": "%"})
		assert.Empty(t, page.Edges)
	})
}

func TestDocumentVersion_ApprovalDecisionsSearch(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	docID, _ := createTestDocument(t, owner)
	approverName := factory.SafeName("Delta")
	approverID := factory.CreateUser(owner, factory.Attrs{"fullName": approverName})
	requestDocumentApproval(t, owner, docID, []string{getOwnerProfileID(t, owner), approverID})
	versionID := latestDocumentVersionID(t, owner, docID)

	all := loadDecisionPage(t, owner, versionID, nil)
	require.Len(t, all.Edges, 2)

	matched := loadDecisionPage(t, owner, versionID, map[string]any{"query": approverName})
	require.Len(t, matched.Edges, 1)
	assert.Equal(t, approverName, matched.Edges[0].Node.Approver.FullName)
	assert.Equal(t, 1, matched.TotalCount)
}

func loadSignaturePage(
	t *testing.T,
	client *testutil.Client,
	versionID string,
	after *string,
	filter map[string]any,
) signaturePage {
	t.Helper()

	var resp struct {
		Node struct {
			Signatures signaturePage `json:"signatures"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!, $after: CursorKey, $filter: DocumentVersionSignatureFilter) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 2, after: $after, filter: $filter) {
						totalCount
						edges { node { id signedBy { fullName } } }
						pageInfo { hasNextPage endCursor }
					}
				}
			}
		}
	`, map[string]any{"id": versionID, "after": after, "filter": filter}, &resp)
	require.NoError(t, err)

	return resp.Node.Signatures
}

func loadDecisionPage(
	t *testing.T,
	client *testutil.Client,
	versionID string,
	filter map[string]any,
) decisionPage {
	t.Helper()

	var resp struct {
		Node struct {
			ApprovalQuorums struct {
				Edges []struct {
					Node struct {
						Decisions decisionPage `json:"decisions"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"approvalQuorums"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!, $filter: DocumentVersionApprovalDecisionFilter) {
			node(id: $id) {
				... on DocumentVersion {
					approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges {
							node {
								decisions(first: 10, filter: $filter) {
									totalCount
									edges { node { approver { fullName } } }
								}
							}
						}
					}
				}
			}
		}
	`, map[string]any{"id": versionID, "filter": filter}, &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Node.ApprovalQuorums.Edges)

	return resp.Node.ApprovalQuorums.Edges[0].Node.Decisions
}
