// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func proseMirrorDoc(paragraphs ...string) string {
	content := make([]map[string]any, 0, len(paragraphs))
	for _, text := range paragraphs {
		para := map[string]any{"type": "paragraph"}
		if text != "" {
			para["content"] = []map[string]any{
				{"type": "text", "text": text},
			}
		}
		content = append(content, para)
	}

	raw, err := json.Marshal(map[string]any{
		"type":    "doc",
		"content": content,
	})
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func TestMCP_Document_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	body := proseMirrorDoc("Document body")

	var addResult struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
		DocumentVersion struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"document_version"`
	}
	mc.CallToolInto("addDocument", map[string]any{
		"organization_id": orgID,
		"title":           factory.SafeName("Document"),
		"content":         body,
		"classification":  "INTERNAL",
		"document_type":   "POLICY",
	}, &addResult)
	require.NotEmpty(t, addResult.Document.ID)
	assert.Contains(t, addResult.DocumentVersion.Content, `"type":"doc"`)
	assert.Contains(t, addResult.DocumentVersion.Content, "Document body")

	var getResult struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
	}
	mc.CallToolInto("getDocument", map[string]any{
		"id": addResult.Document.ID,
	}, &getResult)
	assert.Equal(t, addResult.Document.ID, getResult.Document.ID)

	var updateResult struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
		DocumentVersion struct {
			Title              string   `json:"title"`
			Content            string   `json:"content"`
			CollaborationHeads []string `json:"collaboration_heads"`
		} `json:"document_version"`
	}
	mc.CallToolInto("updateDocument", map[string]any{
		"id":    addResult.Document.ID,
		"title": "Updated Document",
	}, &updateResult)
	assert.Equal(t, addResult.Document.ID, updateResult.Document.ID)
	assert.Equal(t, "Updated Document", updateResult.DocumentVersion.Title)

	var listResult struct {
		Documents []struct {
			ID string `json:"id"`
		} `json:"documents"`
	}
	mc.CallToolInto("listDocuments", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Documents)

	var deleteResult struct {
		DeletedDocumentID string `json:"deleted_document_id"`
	}
	mc.CallToolInto("deleteDocument", map[string]any{
		"document_id": addResult.Document.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Document.ID, deleteResult.DeletedDocumentID)
}

func TestMCP_Document_CollaborationEdits(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	initial := proseMirrorDoc("Old body", "Keep me")

	var addResult struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
		DocumentVersion struct {
			ID string `json:"id"`
		} `json:"document_version"`
	}
	mc.CallToolInto("addDocument", map[string]any{
		"organization_id": orgID,
		"title":           factory.SafeName("Collab Doc"),
		"content":         initial,
		"classification":  "INTERNAL",
		"document_type":   "POLICY",
	}, &addResult)
	require.NotEmpty(t, addResult.DocumentVersion.ID)

	seeded := documentWithTable("Old body", "Keep me")
	var seedResult struct {
		DocumentVersion struct {
			Content            string   `json:"content"`
			CollaborationHeads []string `json:"collaboration_heads"`
		} `json:"document_version"`
	}
	mc.CallToolInto("updateDocument", map[string]any{
		"id":      addResult.Document.ID,
		"content": seeded,
	}, &seedResult)
	require.NotEmpty(t, seedResult.DocumentVersion.CollaborationHeads)
	assert.Contains(t, seedResult.DocumentVersion.Content, `"type":"doc"`)
	assert.Contains(t, seedResult.DocumentVersion.Content, `"type":"table"`)
	assert.NotContains(t, seedResult.DocumentVersion.Content, "# ")

	var getVersion struct {
		DocumentVersion struct {
			ID                 string   `json:"id"`
			Content            string   `json:"content"`
			CollaborationHeads []string `json:"collaboration_heads"`
		} `json:"document_version"`
	}
	mc.CallToolInto("getDocumentVersion", map[string]any{
		"id": addResult.DocumentVersion.ID,
	}, &getVersion)
	require.Equal(
		t,
		seedResult.DocumentVersion.CollaborationHeads,
		getVersion.DocumentVersion.CollaborationHeads,
	)
	assert.Contains(t, getVersion.DocumentVersion.Content, "Old body")

	seededContentErr := mc.CallToolExpectToolError("updateDocument", map[string]any{
		"id":      addResult.Document.ID,
		"content": proseMirrorDoc("Should fail"),
	})
	assert.True(
		t,
		strings.Contains(seededContentErr, "fragment edits") ||
			strings.Contains(seededContentErr, "seeded draft"),
		"unexpected error: %s",
		seededContentErr,
	)

	markdownErr := mc.CallToolExpectToolError("addDocument", map[string]any{
		"organization_id": orgID,
		"title":           factory.SafeName("MarkdownRejected"),
		"content":         "# Markdown heading",
		"classification":  "INTERNAL",
		"document_type":   "POLICY",
	})
	assert.Contains(t, markdownErr, "invalid document content")

	unknownHeadsErr := mc.CallToolExpectToolError("updateDocument", map[string]any{
		"id":         addResult.Document.ID,
		"base_heads": []string{strings.Repeat("ab", 32)},
		"edits": []map[string]any{
			{
				"target":      paragraphNodeJSON(t, getVersion.DocumentVersion.Content, "Old body"),
				"replacement": `[{"type":"paragraph","content":[{"type":"text","text":"Should not land"}]}]`,
			},
		},
	})
	assert.Contains(t, unknownHeadsErr, "base heads")

	var unchanged struct {
		DocumentVersion struct {
			Content string `json:"content"`
		} `json:"document_version"`
	}
	mc.CallToolInto("getDocumentVersion", map[string]any{
		"id": addResult.DocumentVersion.ID,
	}, &unchanged)
	assert.Contains(t, unchanged.DocumentVersion.Content, "Old body")
	assert.NotContains(t, unchanged.DocumentVersion.Content, "Should not land")

	oldTarget := paragraphNodeJSON(t, getVersion.DocumentVersion.Content, "Old body")
	var editResult struct {
		DocumentVersion struct {
			Content            string   `json:"content"`
			CollaborationHeads []string `json:"collaboration_heads"`
		} `json:"document_version"`
	}
	mc.CallToolInto("updateDocument", map[string]any{
		"id":         addResult.Document.ID,
		"base_heads": getVersion.DocumentVersion.CollaborationHeads,
		"edits": []map[string]any{
			{
				"target":      oldTarget,
				"replacement": `[{"type":"paragraph","content":[{"type":"text","text":"New body"}]}]`,
			},
		},
	}, &editResult)
	assert.Contains(t, editResult.DocumentVersion.Content, "New body")
	assert.Contains(t, editResult.DocumentVersion.Content, "Keep me")
	assert.Contains(t, editResult.DocumentVersion.Content, `"type":"table"`)
	assert.Contains(t, editResult.DocumentVersion.Content, "Cell")
	assert.NotContains(t, editResult.DocumentVersion.Content, "Old body")
	require.NotEmpty(t, editResult.DocumentVersion.CollaborationHeads)

	staleErr := mc.CallToolExpectToolError("updateDocument", map[string]any{
		"id":         addResult.Document.ID,
		"base_heads": editResult.DocumentVersion.CollaborationHeads,
		"edits": []map[string]any{
			{
				"target":      oldTarget,
				"replacement": `[{"type":"paragraph","content":[{"type":"text","text":"Lost race"}]}]`,
			},
		},
	})
	assert.True(
		t,
		strings.Contains(staleErr, "not found") || strings.Contains(staleErr, "changed"),
		"unexpected stale error: %s",
		staleErr,
	)

	var afterConflict struct {
		DocumentVersion struct {
			Content string `json:"content"`
		} `json:"document_version"`
	}
	mc.CallToolInto("getDocumentVersion", map[string]any{
		"id": addResult.DocumentVersion.ID,
	}, &afterConflict)
	assert.Contains(t, afterConflict.DocumentVersion.Content, "New body")
	assert.NotContains(t, afterConflict.DocumentVersion.Content, "Lost race")

	newTarget := paragraphNodeJSON(t, editResult.DocumentVersion.Content, "New body")
	var secondEdit struct {
		DocumentVersion struct {
			Content            string   `json:"content"`
			CollaborationHeads []string `json:"collaboration_heads"`
		} `json:"document_version"`
	}
	mc.CallToolInto("updateDocument", map[string]any{
		"id":         addResult.Document.ID,
		"base_heads": editResult.DocumentVersion.CollaborationHeads,
		"edits": []map[string]any{
			{
				"target":      newTarget,
				"replacement": `[{"type":"paragraph","content":[{"type":"text","text":"Newer body"}]}]`,
			},
		},
	}, &secondEdit)
	assert.Contains(t, secondEdit.DocumentVersion.Content, "Newer body")
	assert.Contains(t, secondEdit.DocumentVersion.Content, `"type":"table"`)

	var history struct {
		CollaborationHistorySessions []struct {
			ActorID     string   `json:"actor_id"`
			Heads       []string `json:"heads"`
			ChangeCount int      `json:"change_count"`
		} `json:"collaboration_history_sessions"`
		TotalCount int `json:"total_count"`
	}
	mc.CallToolInto("listDocumentCollaborationHistorySessions", map[string]any{
		"document_version_id": addResult.DocumentVersion.ID,
	}, &history)
	require.Equal(t, 1, history.TotalCount)
	require.Len(t, history.CollaborationHistorySessions, 1)
	assert.GreaterOrEqual(t, history.CollaborationHistorySessions[0].ChangeCount, 3)
	require.NotEmpty(t, history.CollaborationHistorySessions[0].Heads)

	var snapshot struct {
		Content string `json:"content"`
	}
	mc.CallToolInto("getDocumentCollaborationHistorySnapshot", map[string]any{
		"document_version_id": addResult.DocumentVersion.ID,
		"heads":               history.CollaborationHistorySessions[0].Heads,
	}, &snapshot)
	assert.Contains(t, snapshot.Content, `"type":"doc"`)
	assert.Contains(t, snapshot.Content, "Newer body")
	assert.NotContains(t, snapshot.Content, "# ")
}

func documentWithTable(paragraphs ...string) string {
	content := make([]any, 0, len(paragraphs)+1)
	for _, text := range paragraphs {
		content = append(content, map[string]any{
			"type": "paragraph",
			"content": []map[string]any{
				{"type": "text", "text": text},
			},
		})
	}

	content = append(content, map[string]any{
		"type": "table",
		"content": []any{
			map[string]any{
				"type": "tableRow",
				"content": []any{
					map[string]any{
						"type": "tableCell",
						"content": []any{
							map[string]any{
								"type": "paragraph",
								"content": []map[string]any{
									{"type": "text", "text": "Cell"},
								},
							},
						},
					},
				},
			},
		},
	})

	raw, err := json.Marshal(map[string]any{
		"type":    "doc",
		"content": content,
	})
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func paragraphNodeJSON(t *testing.T, docJSON, text string) string {
	t.Helper()

	var doc struct {
		Content []json.RawMessage `json:"content"`
	}
	require.NoError(t, json.Unmarshal([]byte(docJSON), &doc))

	for _, raw := range doc.Content {
		if strings.Contains(string(raw), text) && strings.Contains(string(raw), `"paragraph"`) {
			return string(raw)
		}
	}

	t.Fatalf("paragraph %q not found in %s", text, docJSON)

	return ""
}

func TestMCP_Document_PublishUsesDefaultApprovers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	approverID := owner.GetProfileID().String()

	addDocument := func(defaultApproverIDs []string) string {
		input := map[string]any{
			"organization_id": orgID,
			"title":           factory.SafeName("Document"),
			"content":         proseMirrorDoc("Body content"),
			"classification":  "INTERNAL",
			"document_type":   "POLICY",
		}
		if defaultApproverIDs != nil {
			input["default_approver_ids"] = defaultApproverIDs
		}

		var addResult struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		}
		mc.CallToolInto("addDocument", input, &addResult)
		require.NotEmpty(t, addResult.Document.ID)

		return addResult.Document.ID
	}

	type publishResult struct {
		DocumentVersion struct {
			Status string `json:"status"`
		} `json:"document_version"`
		ApprovalQuorum *struct {
			ID string `json:"id"`
		} `json:"approval_quorum"`
	}

	t.Run("requests approval from default approvers", func(t *testing.T) {
		docID := addDocument([]string{approverID})

		var result publishResult
		mc.CallToolInto("publishDocument", map[string]any{
			"document_id": docID,
			"minor":       false,
			"changelog":   "Initial major",
		}, &result)

		require.NotNil(t, result.ApprovalQuorum)
		assert.Equal(t, "PENDING_APPROVAL", result.DocumentVersion.Status)
	})

	t.Run("publishes directly without default approvers", func(t *testing.T) {
		docID := addDocument(nil)

		var result publishResult
		mc.CallToolInto("publishDocument", map[string]any{
			"document_id": docID,
			"minor":       false,
			"changelog":   "Initial major",
		}, &result)

		assert.Nil(t, result.ApprovalQuorum)
		assert.Equal(t, "PUBLISHED", result.DocumentVersion.Status)
	})
}
