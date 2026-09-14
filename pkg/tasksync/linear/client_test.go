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

package linear

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientListTeamsCreateAndUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req graphqlRequest
		require.NoError(t, json.Unmarshal(body, &req))

		switch {
		case strings.Contains(req.Query, "TaskSyncLinearTeams"):
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"teams": map[string]any{
						"nodes": []map[string]any{
							{"id": "team-1", "name": "Engineering", "key": "ENG"},
						},
					},
				},
			})
		case strings.Contains(req.Query, "TaskSyncLinearWorkflowStates"):
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"workflowStates": map[string]any{
						"nodes": []map[string]any{
							{"id": "state-1", "name": "Todo", "type": "unstarted", "position": 1},
						},
					},
				},
			})
		case strings.Contains(req.Query, "TaskSyncLinearIssueCreate"):
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"issueCreate": map[string]any{
						"success": true,
						"issue": map[string]any{
							"id":         "issue-1",
							"identifier": "ENG-1",
							"url":        "https://linear.app/eng/issue/ENG-1",
							"title":      "Hello",
							"updatedAt":  "2026-09-14T12:00:00Z",
						},
					},
				},
			})
		case strings.Contains(req.Query, "TaskSyncLinearIssueUpdate"):
			writeJSON(t, w, map[string]any{
				"data": map[string]any{
					"issueUpdate": map[string]any{
						"success": true,
						"issue": map[string]any{
							"id":         "issue-1",
							"identifier": "ENG-1",
							"url":        "https://linear.app/eng/issue/ENG-1",
							"title":      "Updated",
							"updatedAt":  "2026-09-14T13:00:00Z",
						},
					},
				},
			})
		default:
			http.Error(w, "unexpected query", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.Client(), server.URL)
	ctx := context.Background()

	teams, err := client.ListTeams(ctx)
	require.NoError(t, err)
	require.Len(t, teams, 1)
	assert.Equal(t, "ENG", teams[0].Key)

	states, err := client.ListWorkflowStates(ctx, "team-1")
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Equal(t, "unstarted", states[0].Type)

	issue, err := client.CreateIssue(ctx, IssueInput{
		TeamID:      "team-1",
		Title:       "Hello",
		Description: "Body",
		StateID:     "state-1",
		Priority:    2,
	})
	require.NoError(t, err)
	assert.Equal(t, "ENG-1", issue.Identifier)

	title := "Updated"
	updated, err := client.UpdateIssue(ctx, "issue-1", IssueUpdateInput{Title: &title})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
}

func TestClient_OrganizationID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req graphqlRequest
		require.NoError(t, json.Unmarshal(body, &req))
		require.Contains(t, req.Query, "TaskSyncLinearOrganization")

		writeJSON(t, w, map[string]any{
			"data": map[string]any{
				"organization": map[string]any{"id": "lin-org-1"},
			},
		})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.Client(), server.URL)
	id, err := client.OrganizationID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "lin-org-1", id)
}

func TestClient_UpdateIssueClearsDueDate(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req graphqlRequest
		require.NoError(t, json.Unmarshal(body, &req))
		require.Contains(t, req.Query, "TaskSyncLinearIssueUpdate")

		vars, err := json.Marshal(req.Variables)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(vars, &captured))

		writeJSON(t, w, map[string]any{
			"data": map[string]any{
				"issueUpdate": map[string]any{
					"success": true,
					"issue": map[string]any{
						"id":         "issue-1",
						"identifier": "ENG-1",
						"url":        "https://linear.app/eng/issue/ENG-1",
						"title":      "Updated",
						"updatedAt":  "2026-09-14T13:00:00Z",
					},
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.Client(), server.URL)
	_, err := client.UpdateIssue(
		context.Background(),
		"issue-1",
		IssueUpdateInput{DueDateSet: true},
	)
	require.NoError(t, err)

	input, ok := captured["input"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, input, "dueDate")
	assert.Nil(t, input["dueDate"])
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload map[string]any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}
