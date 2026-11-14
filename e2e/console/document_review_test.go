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
		t.Parallel()

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
				"content":    factory.ProseMirrorPlainText("A comment"),
				"quote":      "Keep this sentence.",
			},
		}, &created)
		require.NoError(t, err)
		require.NotEmpty(t, created.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID)

		publishMajorDocumentVersion(t, owner, commentDocumentID)
	})

	t.Run("open suggestion mark blocks publish", func(t *testing.T) {
		t.Parallel()

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
			"content":    factory.ProseMirrorPlainText("Viewer comment"),
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
			"content":      factory.ProseMirrorPlainText("Please consider this"),
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

func TestDocumentReview_AddsAnotherRedlineToTheSuggestion(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Review multi suggestion")).
		Create()
	updateDocumentContent(t, owner, documentID, "Keep the first sentence. Keep the second sentence.")

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
		},
	}, &started)
	require.NoError(t, err, "owner should start a suggestion")

	suggestionID := started.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID
	require.NotEmpty(t, suggestionID)

	var added struct {
		AddDocumentSuggestionChange struct {
			DocumentSuggestion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentSuggestion"`
		} `json:"addDocumentSuggestionChange"`
	}

	err = owner.Execute(`
		mutation($input: AddDocumentSuggestionChangeInput!) {
			addDocumentSuggestionChange(input: $input) {
				documentSuggestion { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"suggestionId": suggestionID,
			"quote":        "second",
			"replacement":  "closing",
		},
	}, &added)
	require.NoError(t, err, "owner should add a second redline")
	require.Equal(t, suggestionID, added.AddDocumentSuggestionChange.DocumentSuggestion.ID)
	require.Equal(t, "OPEN", added.AddDocumentSuggestionChange.DocumentSuggestion.Status)

	_, err = owner.Do(`
		mutation($input: AddDocumentSuggestionChangeInput!) {
			addDocumentSuggestionChange(input: $input) {
				documentSuggestion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"suggestionId": suggestionID,
			"quote":        "opening",
			"replacement":  "initial",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "a redline must not cover an existing suggestion")

	err = owner.Execute(`
		mutation($input: AcceptDocumentSuggestionInput!) {
			acceptDocumentSuggestion(input: $input) {
				documentSuggestion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{"suggestionId": suggestionID},
	}, &struct{}{})
	require.NoError(t, err, "owner should accept both redlines")

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
	require.Contains(t, content.Node.Content.Prosemirror, "Keep the opening sentence. Keep the closing sentence.")
	require.NotContains(t, content.Node.Content.Prosemirror, collaboration.SuggestionInsertMark)
	require.NotContains(t, content.Node.Content.Prosemirror, collaboration.SuggestionDeleteMark)
}

func TestDocumentReview_SuggestsTableStructure(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Review table suggestion")).
		Create()

	_, err := owner.Do(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) { documentVersion { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"id": documentID,
			"content": `{"type":"doc","content":[` +
				`{"type":"paragraph","content":[{"type":"text","text":"Access rights"}]},` +
				`{"type":"table","content":[` +
				`{"type":"tableRow","content":[` +
				`{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Right"}]}]},` +
				`{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Access"}]}]}]},` +
				`{"type":"tableRow","content":[` +
				`{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Read"}]}]},` +
				`{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Yes"}]}]}]}` +
				`]}]}`,
		},
	})
	require.NoError(t, err)

	const tableChange = `
		mutation($input: AddDocumentSuggestionTableChangeInput!) {
			addDocumentSuggestionTableChange(input: $input) {
				documentSuggestion { id status }
			}
		}
	`

	insertRow := startEmptySuggestion(t, owner, documentID)

	var added struct {
		AddDocumentSuggestionTableChange struct {
			DocumentSuggestion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentSuggestion"`
		} `json:"addDocumentSuggestionTableChange"`
	}

	err = owner.Execute(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": insertRow,
			"operation":    "INSERT_ROW",
			"quote":        "Read",
			"cells":        []string{"Write", "No"},
		},
	}, &added)
	require.NoError(t, err, "owner should suggest a new row")
	require.Equal(t, insertRow, added.AddDocumentSuggestionTableChange.DocumentSuggestion.ID)
	require.Equal(t, "OPEN", added.AddDocumentSuggestionTableChange.DocumentSuggestion.Status)

	require.Equal(
		t,
		[][]string{{"Right", "Access"}, {"Read", "Yes"}},
		documentTableCells(t, owner, documentID),
		"an open row suggestion must not reach the published content",
	)

	_, err = owner.Do(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": insertRow,
			"operation":    "DELETE_ROW",
			"quote":        "Write",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "a row redline must not swallow the suggested row")

	_, err = owner.Do(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": insertRow,
			"operation":    "INSERT_COLUMN",
			"quote":        "Access rights",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "a column needs a quote inside the table")

	decideSuggestion(t, owner, insertRow, true)
	require.Equal(
		t,
		[][]string{{"Right", "Access"}, {"Read", "Yes"}, {"Write", "No"}},
		documentTableCells(t, owner, documentID),
	)

	deleteColumn := startEmptySuggestion(t, owner, documentID)
	err = owner.Execute(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": deleteColumn,
			"operation":    "DELETE_COLUMN",
			"quote":        "Yes",
		},
	}, &added)
	require.NoError(t, err, "owner should suggest removing a column")

	decideSuggestion(t, owner, deleteColumn, false)
	require.Equal(
		t,
		[][]string{{"Right", "Access"}, {"Read", "Yes"}, {"Write", "No"}},
		documentTableCells(t, owner, documentID),
		"rejecting the column redline keeps the column",
	)

	insertTable := startEmptySuggestion(t, owner, documentID)
	err = owner.Execute(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": insertTable,
			"operation":    "INSERT_TABLE",
			"quote":        "Access rights",
			"before":       true,
			"rows":         [][]string{{"Owner", "Team"}, {"Alice", "Security"}},
			"headerRow":    true,
		},
	}, &added)
	require.NoError(t, err, "owner should suggest a new table")

	decideSuggestion(t, owner, insertTable, true)
	require.Equal(
		t,
		[][]string{{"Owner", "Team"}, {"Alice", "Security"}},
		documentTableCells(t, owner, documentID),
		"the accepted table comes before the quoted paragraph",
	)

	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	foreign := startEmptySuggestion(t, owner, documentID)
	_, err = admin.Do(tableChange, map[string]any{
		"input": map[string]any{
			"suggestionId": foreign,
			"operation":    "DELETE_TABLE",
			"quote":        "Owner",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "only the author adds redlines to a suggestion")
}

func TestDocumentReview_SuggestsFormatting(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Review format suggestion")).
		Create()

	_, err := owner.Do(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) { documentVersion { id } }
		}
	`, map[string]any{
		"input": map[string]any{
			"id":      documentID,
			"content": `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Keep this sentence.","marks":[{"type":"italic"}]}]}]}`,
		},
	})
	require.NoError(t, err)

	const formatChange = `
		mutation($input: AddDocumentSuggestionFormatChangeInput!) {
			addDocumentSuggestionFormatChange(input: $input) {
				documentSuggestion { id status }
			}
		}
	`

	bold := startEmptySuggestion(t, owner, documentID)

	var added struct {
		AddDocumentSuggestionFormatChange struct {
			DocumentSuggestion struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"documentSuggestion"`
		} `json:"addDocumentSuggestionFormatChange"`
	}

	err = owner.Execute(formatChange, map[string]any{
		"input": map[string]any{
			"suggestionId": bold,
			"mark":         "BOLD",
			"action":       "ADD",
			"quote":        "Keep",
		},
	}, &added)
	require.NoError(t, err, "owner should suggest bold")
	require.Equal(t, bold, added.AddDocumentSuggestionFormatChange.DocumentSuggestion.ID)

	published := documentProsemirror(t, owner, documentID)
	require.NotContains(t, published, `"type":"bold"`)
	require.Contains(t, published, `"type":"italic"`)

	_, err = owner.Do(formatChange, map[string]any{
		"input": map[string]any{
			"suggestionId": bold,
			"mark":         "LINK",
			"action":       "ADD",
			"quote":        "Keep",
			"href":         "javascript:alert(1)",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "an unsafe link must be refused")

	decideSuggestion(t, owner, bold, true)
	require.Contains(t, documentProsemirror(t, owner, documentID), `"type":"bold"`)

	italic := startEmptySuggestion(t, owner, documentID)
	err = owner.Execute(formatChange, map[string]any{
		"input": map[string]any{
			"suggestionId": italic,
			"mark":         "ITALIC",
			"action":       "REMOVE",
			"quote":        "sentence",
		},
	}, &added)
	require.NoError(t, err, "owner should suggest removing italic")
	require.Contains(t, documentProsemirror(t, owner, documentID), `"type":"italic"`)

	decideSuggestion(t, owner, italic, true)

	var publishedNode documentContentNode
	require.NoError(t, json.Unmarshal([]byte(documentProsemirror(t, owner, documentID)), &publishedNode))
	sentence := textNode(publishedNode, "sentence")
	require.NotNil(t, sentence)
	require.NotContains(t, markTypes(sentence), "italic")

	keep := textNode(publishedNode, "Keep")
	require.NotNil(t, keep)
	require.Contains(t, markTypes(keep), "bold")
}

func documentProsemirror(t *testing.T, client *testutil.Client, documentID string) string {
	t.Helper()

	var content struct {
		Node struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document { content { prosemirror } }
			}
		}
	`, map[string]any{"id": documentID}, &content)
	require.NoError(t, err)

	return content.Node.Content.Prosemirror
}

func startEmptySuggestion(t *testing.T, client *testutil.Client, documentID string) string {
	t.Helper()

	var started struct {
		StartDocumentSuggestion struct {
			DocumentSuggestionEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"documentSuggestionEdge"`
		} `json:"startDocumentSuggestion"`
	}

	err := client.Execute(`
		mutation($input: StartDocumentSuggestionInput!) {
			startDocumentSuggestion(input: $input) {
				documentSuggestionEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, &started)
	require.NoError(t, err)

	suggestionID := started.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID
	require.NotEmpty(t, suggestionID)

	return suggestionID
}

func decideSuggestion(t *testing.T, client *testutil.Client, suggestionID string, accept bool) {
	t.Helper()

	mutation := `
		mutation($input: RejectDocumentSuggestionInput!) {
			rejectDocumentSuggestion(input: $input) { documentSuggestion { id } }
		}
	`
	if accept {
		mutation = `
			mutation($input: AcceptDocumentSuggestionInput!) {
				acceptDocumentSuggestion(input: $input) { documentSuggestion { id } }
			}
		`
	}

	err := client.Execute(mutation, map[string]any{
		"input": map[string]any{"suggestionId": suggestionID},
	}, &struct{}{})
	require.NoError(t, err)
}

type documentContentNode struct {
	Type    string                `json:"type"`
	Text    string                `json:"text"`
	Marks   []documentContentMark `json:"marks"`
	Content []documentContentNode `json:"content"`
}

type documentContentMark struct {
	Type string `json:"type"`
}

func markTypes(node *documentContentNode) []string {
	types := make([]string, 0, len(node.Marks))
	for _, mark := range node.Marks {
		types = append(types, mark.Type)
	}

	return types
}

func textNode(node documentContentNode, text string) *documentContentNode {
	if node.Text == text {
		return &node
	}

	for _, child := range node.Content {
		if found := textNode(child, text); found != nil {
			return found
		}
	}

	return nil
}

// documentTableCells returns the cell texts of the first table in the
// document's published content.
func documentTableCells(t *testing.T, client *testutil.Client, documentID string) [][]string {
	t.Helper()

	var content struct {
		Node struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document { content { prosemirror } }
			}
		}
	`, map[string]any{"id": documentID}, &content)
	require.NoError(t, err)

	var root documentContentNode
	require.NoError(t, json.Unmarshal([]byte(content.Node.Content.Prosemirror), &root))

	for _, block := range root.Content {
		if block.Type != "table" {
			continue
		}

		rows := make([][]string, 0, len(block.Content))
		for _, row := range block.Content {
			cells := make([]string, 0, len(row.Content))
			for _, cell := range row.Content {
				cells = append(cells, contentText(cell))
			}

			rows = append(rows, cells)
		}

		return rows
	}

	return nil
}

func contentText(node documentContentNode) string {
	text := node.Text
	for _, child := range node.Content {
		text += contentText(child)
	}

	return text
}

func TestDocumentReview_ChangeBesideOpenSuggestion(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Change beside suggestion")).
		Create()
	updateDocumentContent(t, owner, documentID, "Keep this sentence. Change this one.")

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
			"quote":       "Keep this sentence.",
			"replacement": "Kept this sentence.",
		},
	}, &started)
	require.NoError(t, err)

	suggestionID := started.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID
	require.NotEmpty(t, suggestionID)

	var changed struct {
		ChangeDocumentText struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		} `json:"changeDocumentText"`
	}

	err = owner.Execute(`
		mutation($input: ChangeDocumentTextInput!) {
			changeDocumentText(input: $input) {
				document { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"mode":        "APPLY",
			"documentId":  documentID,
			"quote":       "Change this one.",
			"replacement": "Changed this one.",
		},
	}, &changed)
	require.NoError(t, err)
	require.Equal(t, documentID, changed.ChangeDocumentText.Document.ID)

	text := documentPlainText(t, owner, documentID)
	require.Contains(t, text, "Changed this one.")
	require.Contains(t, text, "Keep")

	var overlap struct {
		ChangeDocumentText struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		} `json:"changeDocumentText"`
	}

	err = owner.Execute(`
		mutation($input: ChangeDocumentTextInput!) {
			changeDocumentText(input: $input) {
				document { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"mode":        "APPLY",
			"documentId":  documentID,
			"quote":       "Kept",
			"replacement": "Nope",
		},
	}, &overlap)
	require.Error(t, err)

	var heading struct {
		ChangeDocumentBlock struct {
			DocumentSuggestion struct {
				ID string `json:"id"`
			} `json:"documentSuggestion"`
		} `json:"changeDocumentBlock"`
	}

	err = owner.Execute(`
		mutation($input: ChangeDocumentBlockInput!) {
			changeDocumentBlock(input: $input) {
				documentSuggestion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"mode":         "SUGGEST",
			"suggestionId": suggestionID,
			"operation":    "SET_BLOCK",
			"block":        "HEADING",
			"level":        2,
			"quote":        "Changed this one.",
		},
	}, &heading)
	require.NoError(t, err)
	require.Equal(t, suggestionID, heading.ChangeDocumentBlock.DocumentSuggestion.ID)
}

func documentPlainText(t *testing.T, client *testutil.Client, documentID string) string {
	t.Helper()

	var content struct {
		Node struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
		} `json:"node"`
	}

	err := client.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document { content { prosemirror } }
			}
		}
	`, map[string]any{"id": documentID}, &content)
	require.NoError(t, err)

	var root documentContentNode
	require.NoError(t, json.Unmarshal([]byte(content.Node.Content.Prosemirror), &root))

	return contentText(root)
}

func paintSuggestionInsert(t *testing.T, documentID string, suggestionID string, text string) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	var snapshot []byte

	err := conn.QueryRow(
		ctx,
		`SELECT snapshot FROM collaboration_documents WHERE document_id = $1`,
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
