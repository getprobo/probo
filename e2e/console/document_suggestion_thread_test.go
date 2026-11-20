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
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/collaboration"
)

type suggestionThreadState struct {
	Status      string  `json:"status"`
	AnchorStart *string `json:"anchorStart"`
	AnchorEnd   *string `json:"anchorEnd"`
	AnchorQuote *string `json:"anchorQuote"`
	Thread      struct {
		ID          string  `json:"id"`
		AnchorStart string  `json:"anchorStart"`
		AnchorEnd   string  `json:"anchorEnd"`
		AnchorQuote string  `json:"anchorQuote"`
		ResolvedAt  *string `json:"resolvedAt"`
		Suggestion  *struct {
			ID string `json:"id"`
		} `json:"suggestion"`
		Comments struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"comments"`
	} `json:"thread"`
}

func TestDocumentReview_SuggestionThread(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	newDocument := func(t *testing.T) string {
		t.Helper()

		documentID := factory.NewDocument(owner).
			WithTitle(factory.SafeName("Suggestion thread")).
			Create()
		updateDocumentContent(t, owner, documentID, "Keep the first sentence.")

		return documentID
	}

	t.Run("starting a suggestion opens an empty thread without anchor", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)

		state := loadSuggestionThread(t, owner, suggestionID)
		require.Equal(t, "OPEN", state.Status)
		require.Nil(t, state.AnchorStart, "an open suggestion is located by its marks")
		require.Nil(t, state.AnchorEnd)
		require.Nil(t, state.AnchorQuote)
		require.NotEmpty(t, state.Thread.ID)
		require.Nil(t, state.Thread.ResolvedAt)
		require.NotNil(t, state.Thread.Suggestion)
		require.Equal(t, suggestionID, state.Thread.Suggestion.ID)
		require.Equal(t, 0, state.Thread.Comments.TotalCount)
		require.Empty(t, state.Thread.AnchorStart)
		require.Empty(t, state.Thread.AnchorEnd)
		require.Empty(t, state.Thread.AnchorQuote)
	})

	t.Run("a start comment lands in the suggestion thread", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)

		var started struct {
			StartDocumentSuggestion struct {
				DocumentSuggestionEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"documentSuggestionEdge"`
			} `json:"startDocumentSuggestion"`
		}

		err := owner.Execute(`
			mutation($input: StartDocumentSuggestionInput!) {
				startDocumentSuggestion(input: $input) {
					documentSuggestionEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId":  documentID,
				"quote":       "first",
				"replacement": "opening",
				"comment":     "Clearer wording",
			},
		}, &started)
		require.NoError(t, err)

		state := loadSuggestionThread(t, owner, started.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID)
		require.Equal(t, 1, state.Thread.Comments.TotalCount)
		require.Empty(t, state.Thread.AnchorStart, "the suggestion thread keeps no anchor of its own")
	})

	t.Run("commenting reuses the thread and deleting the last comment keeps it", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		paintSuggestionInsert(t, documentID, suggestionID, " Added.")

		before := loadSuggestionThread(t, owner, suggestionID)

		threadID := commentOnSuggestion(t, owner, documentID, suggestionID, "First note")
		require.Equal(t, before.Thread.ID, threadID)

		secondThreadID := commentOnSuggestion(t, owner, documentID, suggestionID, "Second note")
		require.Equal(t, threadID, secondThreadID, "every comment goes to the same thread")

		state := loadSuggestionThread(t, owner, suggestionID)
		require.Equal(t, 2, state.Thread.Comments.TotalCount)
		require.Empty(t, state.Thread.AnchorStart)
		require.Empty(t, state.Thread.AnchorQuote)

		for _, edge := range state.Thread.Comments.Edges {
			var deleted struct {
				DeleteDocumentComment struct {
					DeletedDocumentCommentThreadID *string `json:"deletedDocumentCommentThreadId"`
				} `json:"deleteDocumentComment"`
			}

			err := owner.Execute(`
				mutation($input: DeleteDocumentCommentInput!) {
					deleteDocumentComment(input: $input) { deletedDocumentCommentThreadId }
				}
			`, map[string]any{
				"input": map[string]any{"commentId": edge.Node.ID},
			}, &deleted)
			require.NoError(t, err)
			require.Nil(t, deleted.DeleteDocumentComment.DeletedDocumentCommentThreadID)
		}

		state = loadSuggestionThread(t, owner, suggestionID)
		require.Equal(t, threadID, state.Thread.ID, "the suggestion keeps its thread")
		require.Equal(t, 0, state.Thread.Comments.TotalCount)
	})

	t.Run("a suggestion thread cannot be resolved or reopened", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		state := loadSuggestionThread(t, owner, suggestionID)

		_, err := owner.Do(`
			mutation($input: ResolveDocumentCommentThreadInput!) {
				resolveDocumentCommentThread(input: $input) { documentCommentThread { id } }
			}
		`, map[string]any{
			"input": map[string]any{"threadId": state.Thread.ID},
		})
		testutil.RequireConflictError(t, err, "resolving a suggestion thread should be refused")

		_, err = owner.Do(`
			mutation($input: ReopenDocumentCommentThreadInput!) {
				reopenDocumentCommentThread(input: $input) { documentCommentThread { id } }
			}
		`, map[string]any{
			"input": map[string]any{"threadId": state.Thread.ID},
		})
		testutil.RequireConflictError(t, err, "reopening a suggestion thread should be refused")
	})

	t.Run("accepting resolves the thread and records the kept text", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		paintSuggestionInsert(t, documentID, suggestionID, " Added.")

		decideSuggestion(t, owner, suggestionID, true)

		state := loadSuggestionThread(t, owner, suggestionID)
		require.Equal(t, "ACCEPTED", state.Status)
		require.NotNil(t, state.Thread.ResolvedAt, "the decision resolves the thread")
		require.NotNil(t, state.AnchorStart)
		require.NotNil(t, state.AnchorEnd)
		require.NotEqual(t, *state.AnchorStart, *state.AnchorEnd, "the anchor covers the accepted text")
		require.NotNil(t, state.AnchorQuote)
		require.Equal(t, " Added.", *state.AnchorQuote)
		require.Empty(t, state.Thread.AnchorStart)
	})

	t.Run("rejecting an insertion records a collapsed anchor", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		paintSuggestionInsert(t, documentID, suggestionID, " Added.")

		decideSuggestion(t, owner, suggestionID, false)

		state := loadSuggestionThread(t, owner, suggestionID)
		require.Equal(t, "REJECTED", state.Status)
		require.NotNil(t, state.Thread.ResolvedAt)
		require.NotNil(t, state.AnchorStart)
		require.NotNil(t, state.AnchorEnd)
		require.Equal(t, *state.AnchorStart, *state.AnchorEnd, "no text survives a rejected insertion")
		require.NotNil(t, state.AnchorQuote)
		require.Equal(t, " Added.", *state.AnchorQuote)
	})

	t.Run("a decided suggestion takes no new comments", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		paintSuggestionInsert(t, documentID, suggestionID, " Added.")
		decideSuggestion(t, owner, suggestionID, true)

		_, err := owner.Do(`
			mutation($input: CreateDocumentCommentThreadInput!) {
				createDocumentCommentThread(input: $input) {
					documentCommentThreadEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"documentId":   documentID,
				"suggestionId": suggestionID,
				"content":      factory.ProseMirrorPlainText("Too late"),
			},
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "a decided suggestion should not take comments")
	})
}

func TestDocumentReview_SuggestionAutoReject(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	newDocument := func(t *testing.T) string {
		t.Helper()

		documentID := factory.NewDocument(owner).
			WithTitle(factory.SafeName("Suggestion auto reject")).
			Create()
		updateDocumentContent(t, owner, documentID, "Keep the first sentence.")

		return documentID
	}

	t.Run("an undone suggestion is rejected and reopens on redo", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		mark, err := json.Marshal(map[string]string{"id": suggestionID})
		require.NoError(t, err)

		addMarks := func(spans []automerge.Span) []automerge.Span {
			return append(spans, automerge.Span{
				Type:  automerge.SpanTypeText,
				Text:  " Added.",
				Marks: map[string]any{collaboration.SuggestionInsertMark: string(mark)},
			})
		}
		removeMarks := func(spans []automerge.Span) []automerge.Span {
			kept := make([]automerge.Span, 0, len(spans))
			for _, span := range spans {
				if value, ok := span.Marks[collaboration.SuggestionInsertMark].(string); ok && strings.Contains(value, suggestionID) {
					continue
				}
				kept = append(kept, span)
			}

			return kept
		}

		rewriteCollaborationSpans(t, documentID, addMarks)
		applyDocumentText(t, owner, documentID, "first", "opening")

		state := loadSuggestionResolution(t, owner, suggestionID)
		require.Equal(t, "OPEN", state.Status, "a suggestion with marks stays open")

		rewriteCollaborationSpans(t, documentID, removeMarks)
		applyDocumentText(t, owner, documentID, "sentence", "line")

		state = loadSuggestionResolution(t, owner, suggestionID)
		require.Equal(t, "REJECTED", state.Status, "a suggestion left with no change is rejected")
		require.Nil(t, state.ResolvedBy, "nobody rejected it")
		require.NotNil(t, state.ResolvedAt)
		require.NotNil(t, state.Thread.ResolvedAt, "its thread leaves the open list")
		require.Nil(t, state.AnchorStart)
		require.NotNil(t, state.AnchorQuote)
		require.Empty(t, *state.AnchorQuote, "nothing was changed")

		rewriteCollaborationSpans(t, documentID, addMarks)
		applyDocumentText(t, owner, documentID, "Keep", "Hold")

		state = loadSuggestionResolution(t, owner, suggestionID)
		require.Equal(t, "OPEN", state.Status, "the marks coming back reopen it")
		require.Nil(t, state.ResolvedAt)
		require.Nil(t, state.Thread.ResolvedAt)
		require.Nil(t, state.AnchorQuote)
	})

	t.Run("a suggestion whose marks were never persisted stays open", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)

		applyDocumentText(t, owner, documentID, "first", "opening")

		state := loadSuggestionResolution(t, owner, suggestionID)
		require.Equal(t, "OPEN", state.Status, "its first marks may still be on their way")
		require.Nil(t, state.Thread.ResolvedAt)
	})

	t.Run("a reviewer's rejection is not reopened", func(t *testing.T) {
		t.Parallel()

		documentID := newDocument(t)
		suggestionID := startEmptySuggestion(t, owner, documentID)
		paintSuggestionInsert(t, documentID, suggestionID, " Added.")
		decideSuggestion(t, owner, suggestionID, false)

		mark, err := json.Marshal(map[string]string{"id": suggestionID})
		require.NoError(t, err)

		rewriteCollaborationSpans(t, documentID, func(spans []automerge.Span) []automerge.Span {
			return append(spans, automerge.Span{
				Type:  automerge.SpanTypeText,
				Text:  " Again.",
				Marks: map[string]any{collaboration.SuggestionInsertMark: string(mark)},
			})
		})
		applyDocumentText(t, owner, documentID, "first", "opening")

		state := loadSuggestionResolution(t, owner, suggestionID)
		require.Equal(t, "REJECTED", state.Status)
		require.NotNil(t, state.ResolvedBy)
	})
}

type suggestionResolutionState struct {
	Status      string  `json:"status"`
	ResolvedAt  *string `json:"resolvedAt"`
	AnchorStart *string `json:"anchorStart"`
	AnchorQuote *string `json:"anchorQuote"`
	ResolvedBy  *struct {
		ID string `json:"id"`
	} `json:"resolvedBy"`
	Thread struct {
		ResolvedAt *string `json:"resolvedAt"`
	} `json:"thread"`
}

func loadSuggestionResolution(t *testing.T, client *testutil.Client, suggestionID string) suggestionResolutionState {
	t.Helper()

	var resp struct {
		Node suggestionResolutionState `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentSuggestion {
					status
					resolvedAt
					anchorStart
					anchorQuote
					resolvedBy { id }
					thread { resolvedAt }
				}
			}
		}
	`, map[string]any{"id": suggestionID}, &resp)
	require.NoError(t, err)

	return resp.Node
}

func applyDocumentText(t *testing.T, client *testutil.Client, documentID string, quote string, replacement string) {
	t.Helper()

	err := client.Execute(`
		mutation($input: ChangeDocumentTextInput!) {
			changeDocumentText(input: $input) { document { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"mode":        "APPLY",
			"documentId":  documentID,
			"quote":       quote,
			"replacement": replacement,
		},
	}, &struct{}{})
	require.NoError(t, err)
}

// rewriteCollaborationSpans edits the stored collaboration state the way an
// editor would, without going through persist, so the next persist sees the
// edit as already there. It folds the change log into a new snapshot.
func rewriteCollaborationSpans(
	t *testing.T,
	documentID string,
	rewrite func([]automerge.Span) []automerge.Span,
) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	var (
		snapshot         []byte
		snapshotRevision int64
		changeRevision   int64
	)

	err := conn.QueryRow(
		ctx,
		`SELECT snapshot, snapshot_revision, change_revision FROM collaboration_documents WHERE document_id = $1`,
		documentID,
	).Scan(&snapshot, &snapshotRevision, &changeRevision)
	require.NoError(t, err)

	rows, err := conn.Query(
		ctx,
		`SELECT change_bytes FROM collaboration_changes WHERE document_id = $1 AND revision > $2 ORDER BY revision`,
		documentID,
		snapshotRevision,
	)
	require.NoError(t, err)

	changeBytes, err := pgx.CollectRows(rows, pgx.RowTo[[]byte])
	require.NoError(t, err)

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.Load(snapshot, actorID)
	require.NoError(t, err)

	defer func() { _ = document.Close() }()

	if len(changeBytes) > 0 {
		changes := make([]automerge.Change, len(changeBytes))
		for i, bytes := range changeBytes {
			changes[i] = automerge.Change{Bytes: bytes}
		}
		require.NoError(t, document.ApplyChanges(changes))
	}

	spans, err := collaboration.BodySpans(document)
	require.NoError(t, err)
	require.NoError(t, collaboration.WriteBodySpans(document, rewrite(spans)))

	_, err = document.Commit("Edit", time.Now())
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
		`DELETE FROM collaboration_changes WHERE document_id = $1 AND revision <= $2`,
		documentID,
		changeRevision,
	)
	require.NoError(t, err)

	_, err = conn.Exec(
		ctx,
		`UPDATE collaboration_documents SET snapshot = $2, heads = $3, snapshot_revision = $4 WHERE document_id = $1`,
		documentID,
		saved,
		encodedHeads,
		changeRevision,
	)
	require.NoError(t, err)
}

func loadSuggestionThread(t *testing.T, client *testutil.Client, suggestionID string) suggestionThreadState {
	t.Helper()

	var resp struct {
		Node suggestionThreadState `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentSuggestion {
					status
					anchorStart
					anchorEnd
					anchorQuote
					thread {
						id
						anchorStart
						anchorEnd
						anchorQuote
						resolvedAt
						suggestion { id }
						comments(first: 10) {
							totalCount
							edges { node { id } }
						}
					}
				}
			}
		}
	`, map[string]any{"id": suggestionID}, &resp)
	require.NoError(t, err)

	return resp.Node
}

func commentOnSuggestion(
	t *testing.T,
	client *testutil.Client,
	documentID string,
	suggestionID string,
	text string,
) string {
	t.Helper()

	var created struct {
		CreateDocumentCommentThread struct {
			DocumentCommentThreadEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentCommentThreadEdge"`
		} `json:"createDocumentCommentThread"`
	}

	err := client.Execute(`
		mutation($input: CreateDocumentCommentThreadInput!) {
			createDocumentCommentThread(input: $input) {
				documentCommentThreadEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentId":   documentID,
			"suggestionId": suggestionID,
			"content":      factory.ProseMirrorPlainText(text),
		},
	}, &created)
	require.NoError(t, err)

	return created.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID
}
