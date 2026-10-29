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
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/collaboration"
)

func TestDocumentReview_PublishBlockedByOpenSuggestion(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Review publish")).
		Create()
	updateDocumentContent(t, owner, documentID, "Keep this sentence.")

	t.Run("open comment does not block publish", func(t *testing.T) {
		commentDocumentID := factory.NewDocument(owner).
			WithTitle(factory.SafeName("Review comment publish")).
			Create()
		updateDocumentContent(t, owner, commentDocumentID, "Keep this sentence.")

		var created struct {
			CreateDocumentCommentThread struct {
				DocumentCommentThreadEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"documentCommentThreadEdge"`
			} `json:"createDocumentCommentThread"`
		}

		err := owner.Execute(`
			mutation($input: CreateDocumentCommentThreadInput!) {
				createDocumentCommentThread(input: $input) {
					documentCommentThreadEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId": commentDocumentID,
				"content":    "A comment",
				"quote":      "Keep this sentence.",
			},
		}, &created)
		require.NoError(t, err)
		require.NotEmpty(t, created.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID)

		publishMajorDocumentVersion(t, owner, commentDocumentID)
	})

	t.Run("open suggestion mark blocks publish", func(t *testing.T) {
		paintSuggestionInsert(t, documentID, "suggestion-open", "inserted")

		_, err := owner.Do(`
			mutation($input: PublishDocumentInput!) {
				publishDocument(input: $input) {
					documentVersion { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"minor":       false,
				"documentId":  documentID,
				"approverIds": []string{},
				"changelog":   "Blocked release",
			},
		})
		testutil.RequireConflictError(t, err, "publish should fail while a suggestion is open")
		require.ErrorContains(t, err, "resolve open suggestions before publishing")

		var content struct {
			Node struct {
				Content struct {
					Prosemirror string `json:"prosemirror"`
				} `json:"content"`
			} `json:"node"`
		}
		err = owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on Document { content { prosemirror } }
				}
			}
		`, map[string]any{"id": documentID}, &content)
		require.NoError(t, err)
		require.NotContains(t, content.Node.Content.Prosemirror, "inserted")
		require.NotContains(t, content.Node.Content.Prosemirror, collaboration.SuggestionInsertMark)
	})
}

func TestDocumentReview_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)
	owner := org.Client(t, testutil.RoleOwner)
	viewer := org.Client(t, testutil.RoleViewer)
	admin := org.Client(t, testutil.RoleAdmin)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Review rbac")).
		Create()
	updateDocumentContent(t, owner, documentID, "Keep this sentence.")

	var thread struct {
		CreateDocumentCommentThread struct {
			DocumentCommentThreadEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentCommentThreadEdge"`
		} `json:"createDocumentCommentThread"`
	}
	err := viewer.Execute(`
		mutation($input: CreateDocumentCommentThreadInput!) {
			createDocumentCommentThread(input: $input) {
				documentCommentThreadEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId": documentID,
			"content":    "Viewer comment",
			"quote":      "Keep this sentence.",
		},
	}, &thread)
	require.NoError(t, err, "viewer should comment")
	require.NotEmpty(t, thread.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID)

	var started struct {
		StartDocumentSuggestion struct {
			DocumentSuggestionEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentSuggestionEdge"`
		} `json:"startDocumentSuggestion"`
	}
	err = viewer.Execute(`
		mutation($input: StartDocumentSuggestionInput!) {
			startDocumentSuggestion(input: $input) {
				documentSuggestionEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, &started)
	require.NoError(t, err, "viewer should suggest")

	suggestionID := started.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID
	require.NotEmpty(t, suggestionID)
	paintSuggestionInsert(t, documentID, suggestionID, "suggested")

	var reply struct {
		CreateDocumentCommentThread struct {
			DocumentCommentThreadEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentCommentThreadEdge"`
		} `json:"createDocumentCommentThread"`
	}
	err = viewer.Execute(`
		mutation($input: CreateDocumentCommentThreadInput!) {
			createDocumentCommentThread(input: $input) {
				documentCommentThreadEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":   documentID,
			"suggestionId": suggestionID,
			"content":      "Please consider this",
		},
	}, &reply)
	require.NoError(t, err, "viewer should reply on a redline")

	redlineThreadID := reply.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID
	require.NotEmpty(t, redlineThreadID)

	_, err = viewer.Do(`
		mutation($input: AcceptDocumentSuggestionInput!) {
			acceptDocumentSuggestion(input: $input) {
				documentSuggestion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{"suggestionId": suggestionID},
	})
	testutil.RequireForbiddenError(t, err, "viewer should not accept a suggestion")

	_, err = employee.Do(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					commentThreads(first: 10) { edges { node { id } } }
				}
			}
		}
	`, map[string]any{"id": documentID})
	testutil.RequireForbiddenError(t, err, "employee should not list comment threads")

	var accepted struct {
		AcceptDocumentSuggestion struct {
			DocumentSuggestion struct {
				Status string `json:"status"`
			} `json:"documentSuggestion"`
		} `json:"acceptDocumentSuggestion"`
	}
	err = admin.Execute(`
		mutation($input: AcceptDocumentSuggestionInput!) {
			acceptDocumentSuggestion(input: $input) {
				documentSuggestion { status }
			}
		}
	`, map[string]any{
		"input": map[string]any{"suggestionId": suggestionID},
	}, &accepted)
	require.NoError(t, err, "editor should accept a suggestion")
	require.Equal(t, "ACCEPTED", accepted.AcceptDocumentSuggestion.DocumentSuggestion.Status)

	var resolved struct {
		Node struct {
			ResolvedAt *string `json:"resolvedAt"`
		} `json:"node"`
	}
	err = admin.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentCommentThread { resolvedAt }
			}
		}
	`, map[string]any{"id": redlineThreadID}, &resolved)
	require.NoError(t, err)
	require.NotNil(t, resolved.Node.ResolvedAt, "accepting a suggestion resolves its thread")

	var content struct {
		Node struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
		} `json:"node"`
	}
	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document { content { prosemirror } }
			}
		}
	`, map[string]any{"id": documentID}, &content)
	require.NoError(t, err)
	require.Contains(t, content.Node.Content.Prosemirror, "suggested")
	require.NotContains(t, content.Node.Content.Prosemirror, collaboration.SuggestionInsertMark)
	require.NotContains(t, content.Node.Content.Prosemirror, collaboration.SuggestionDeleteMark)
}

func paintSuggestionInsert(t *testing.T, documentID string, suggestionID string, text string) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	var snapshot []byte
	err := conn.QueryRow(
		ctx,
		`SELECT snapshot FROM collaboration_documents WHERE document_id = $1 AND seeded = true`,
		documentID,
	).Scan(&snapshot)
	require.NoError(t, err)

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.Load(snapshot, actorID)
	require.NoError(t, err)
	defer func() { _ = document.Close() }()

	spans, err := collaboration.BodySpans(document)
	require.NoError(t, err)

	mark, err := json.Marshal(map[string]string{"id": suggestionID})
	require.NoError(t, err)

	spans = append(spans, automerge.Span{
		Type: automerge.SpanTypeText,
		Text: text,
		Marks: map[string]any{
			collaboration.SuggestionInsertMark: string(mark),
		},
	})
	require.NoError(t, collaboration.WriteBodySpans(document, spans))

	_, err = document.Commit("Suggest", time.Now())
	require.NoError(t, err)

	saved, err := document.Save()
	require.NoError(t, err)

	heads, err := document.Heads()
	require.NoError(t, err)

	encodedHeads := make([]byte, 0, len(heads)*len(automerge.Hash{}))
	for _, head := range heads {
		encodedHeads = append(encodedHeads, head[:]...)
	}

	_, err = conn.Exec(
		ctx,
		`UPDATE collaboration_documents SET snapshot = $2, heads = $3 WHERE document_id = $1`,
		documentID,
		saved,
		encodedHeads,
	)
	require.NoError(t, err)
}
