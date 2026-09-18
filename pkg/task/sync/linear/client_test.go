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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	teamPageVariables struct {
		First int     `json:"first"`
		After *string `json:"after"`
	}
)

func decodeGraphQLRequest(r *http.Request) (graphqlRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return graphqlRequest{}, err
	}

	var req graphqlRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return graphqlRequest{}, err
	}

	return req, nil
}

func decodeGraphQLVariables[T any](req graphqlRequest) (T, error) {
	var vars T

	raw, err := json.Marshal(req.Variables)
	if err != nil {
		return vars, err
	}

	err = json.Unmarshal(raw, &vars)

	return vars, err
}

func graphQLObjectVariables(req graphqlRequest) (map[string]any, error) {
	vars, ok := req.Variables.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot decode graphql variables: not an object")
	}

	return vars, nil
}

func writeJSON(w http.ResponseWriter, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveDecodedGraphQL(
	w http.ResponseWriter,
	r *http.Request,
	handlerErr *error,
) (graphqlRequest, bool) {
	req, err := decodeGraphQLRequest(r)
	if err != nil {
		*handlerErr = err
		http.Error(w, err.Error(), http.StatusBadRequest)

		return graphqlRequest{}, false
	}

	return req, true
}

func newLinearGraphQLServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server
}

func TestClient_ListTeamsCreateAndUpdate(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		queries    []string
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			queries = append(queries, req.Query)

			switch {
			case strings.Contains(req.Query, "TaskSyncLinearTeams"):
				writeJSON(
					w,
					map[string]any{
						"data": map[string]any{
							"teams": map[string]any{
								"nodes": []map[string]any{
									{"id": "team-1", "name": "Engineering", "key": "ENG"},
								},
							},
						},
					},
				)
			case strings.Contains(req.Query, "TaskSyncLinearWorkflowStates"):
				writeJSON(
					w,
					map[string]any{
						"data": map[string]any{
							"workflowStates": map[string]any{
								"nodes": []map[string]any{
									{"id": "state-1", "name": "Todo", "type": "unstarted", "position": 1},
								},
							},
						},
					},
				)
			case strings.Contains(req.Query, "TaskSyncLinearIssueCreate"):
				writeJSON(
					w,
					map[string]any{
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
					},
				)
			case strings.Contains(req.Query, "TaskSyncLinearIssueUpdate"):
				writeJSON(
					w,
					map[string]any{
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
					},
				)
			default:
				handlerErr = fmt.Errorf("cannot handle unexpected query %q", req.Query)
				http.Error(w, "unexpected query", http.StatusBadRequest)
			}
		},
	)

	client := NewClient(server.Client(), server.URL)
	ctx := context.Background()

	teams, err := client.ListTeams(ctx)
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, teams, 1)
	assert.Equal(t, "ENG", teams[0].Key)

	states, err := client.ListWorkflowStates(ctx, "team-1")
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Equal(t, "unstarted", states[0].Type)

	issue, err := client.CreateIssue(
		ctx,
		IssueInput{
			TeamID:      "team-1",
			Title:       "Hello",
			Description: "Body",
			StateID:     "state-1",
			Priority:    2,
		},
	)
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Equal(t, "ENG-1", issue.Identifier)

	title := "Updated"
	updated, err := client.UpdateIssue(
		ctx,
		"issue-1",
		IssueUpdateInput{Title: &title},
	)
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)

	require.Len(t, queries, 4)
	assert.Contains(t, queries[0], "TaskSyncLinearTeams")
	assert.Contains(t, queries[1], "TaskSyncLinearWorkflowStates")
	assert.Contains(t, queries[2], "TaskSyncLinearIssueCreate")
	assert.Contains(t, queries[3], "TaskSyncLinearIssueUpdate")
}

func TestClient_ListTeamsFollowsPages(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		requests   []teamPageVariables
		queries    []string
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, err := decodeGraphQLVariables[teamPageVariables](req)
			if err != nil {
				handlerErr = err
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			queries = append(queries, req.Query)
			requests = append(requests, vars)

			if len(requests) == 1 {
				writeJSON(
					w,
					map[string]any{
						"data": map[string]any{
							"teams": map[string]any{
								"nodes": []map[string]any{
									{"id": "team-1", "name": "Engineering", "key": "ENG"},
								},
								"pageInfo": map[string]any{
									"hasNextPage": true,
									"endCursor":   "cursor-1",
								},
							},
						},
					},
				)

				return
			}

			writeJSON(
				w,
				map[string]any{
					"data": map[string]any{
						"teams": map[string]any{
							"nodes": []map[string]any{
								{"id": "team-2", "name": "Design", "key": "DES"},
							},
							"pageInfo": map[string]any{
								"hasNextPage": false,
								"endCursor":   "cursor-2",
							},
						},
					},
				},
			)
		},
	)

	teams, err := NewClient(server.Client(), server.URL).ListTeams(context.Background())
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, teams, 2)
	require.Len(t, requests, 2)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0], "TaskSyncLinearTeams")
	assert.Contains(t, queries[0], "hasNextPage")
	assert.Contains(t, queries[1], "TaskSyncLinearTeams")
	assert.Equal(t, 100, requests[0].First)
	assert.Nil(t, requests[0].After)
	assert.Equal(t, 100, requests[1].First)
	require.NotNil(t, requests[1].After)
	assert.Equal(t, "cursor-1", *requests[1].After)
	assert.Equal(t, "ENG", teams[0].Key)
	assert.Equal(t, "DES", teams[1].Key)
}

func TestClient_ArchiveIssue(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		query      string
		captured   map[string]any
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, err := graphQLObjectVariables(req)
			if err != nil {
				handlerErr = err
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			query = req.Query
			captured = vars

			writeJSON(
				w,
				map[string]any{
					"data": map[string]any{
						"issueArchive": map[string]any{"success": true},
					},
				},
			)
		},
	)

	err := NewClient(server.Client(), server.URL).ArchiveIssue(context.Background(), "issue-1")
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Contains(t, query, "TaskSyncLinearIssueArchive")
	assert.Equal(t, "issue-1", captured["id"])
}

func TestClient_OrganizationID(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		query      string
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			query = req.Query

			writeJSON(
				w,
				map[string]any{
					"data": map[string]any{
						"organization": map[string]any{"id": "lin-org-1"},
					},
				},
			)
		},
	)

	client := NewClient(server.Client(), server.URL)
	id, err := client.OrganizationID(context.Background())
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Contains(t, query, "TaskSyncLinearOrganization")
	assert.Equal(t, "lin-org-1", id)
}

func TestClient_UpdateIssueClearsDueDate(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		query      string
		captured   map[string]any
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, err := graphQLObjectVariables(req)
			if err != nil {
				handlerErr = err
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			query = req.Query
			captured = vars

			writeJSON(
				w,
				map[string]any{
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
				},
			)
		},
	)

	client := NewClient(server.Client(), server.URL)
	_, err := client.UpdateIssue(
		context.Background(),
		"issue-1",
		IssueUpdateInput{DueDateSet: true},
	)
	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Contains(t, query, "TaskSyncLinearIssueUpdate")

	input, ok := captured["input"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, input, "dueDate")
	assert.Nil(t, input["dueDate"])
}
