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

	workflowStatePageVariables struct {
		TeamID string  `json:"teamId"`
		First  int     `json:"first"`
		After  *string `json:"after"`
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
		variables  []map[string]any
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

			queries = append(queries, req.Query)
			variables = append(variables, vars)

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
	require.Len(t, variables, 4)
	assert.Contains(t, queries[0], "TaskSyncLinearTeams")
	assert.Contains(t, queries[1], "TaskSyncLinearWorkflowStates")
	assert.Contains(t, queries[2], "TaskSyncLinearIssueCreate")
	assert.Contains(t, queries[3], "TaskSyncLinearIssueUpdate")
	assert.Equal(t, "team-1", variables[1]["teamId"])
	assert.Equal(t, float64(100), variables[1]["first"])
	assert.Nil(t, variables[1]["after"])

	createInput, ok := variables[2]["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "team-1", createInput["teamId"])
	assert.Equal(t, "Hello", createInput["title"])
	assert.Equal(t, "Body", createInput["description"])
	assert.Equal(t, "state-1", createInput["stateId"])
	assert.Equal(t, float64(2), createInput["priority"])

	assert.Equal(t, "issue-1", variables[3]["id"])

	updateInput, ok := variables[3]["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Updated", updateInput["title"])
}

func TestClient_UserIDByEmail(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns the only matching user",
		func(t *testing.T) {
			t.Parallel()

			var (
				handlerErr error
				email      string
			)

			server := newLinearGraphQLServer(
				t,
				func(w http.ResponseWriter, r *http.Request) {
					req, ok := serveDecodedGraphQL(w, r, &handlerErr)
					if !ok {
						return
					}

					assert.Contains(t, req.Query, "TaskSyncLinearUserByEmail")

					vars, err := graphQLObjectVariables(req)
					if err != nil {
						handlerErr = err
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					got, _ := vars["email"].(string)
					email = got

					writeJSON(
						w,
						map[string]any{
							"data": map[string]any{
								"users": map[string]any{
									"nodes": []map[string]any{
										{"id": "user-1", "email": "jane@example.com"},
									},
								},
							},
						},
					)
				},
			)

			id, err := NewClient(server.Client(), server.URL).UserIDByEmail(
				context.Background(),
				"jane@example.com",
			)

			require.NoError(t, handlerErr)
			require.NoError(t, err)
			assert.Equal(t, "jane@example.com", email)
			assert.Equal(t, "user-1", id)
		},
	)

	t.Run(
		"skips when no user or several users match",
		func(t *testing.T) {
			t.Parallel()

			nodes := [][]map[string]any{
				{},
				{
					{"id": "user-1", "email": "jane@example.com"},
					{"id": "user-2", "email": "jane@example.com"},
				},
			}

			for _, want := range nodes {
				server := newLinearGraphQLServer(
					t,
					func(w http.ResponseWriter, r *http.Request) {
						writeJSON(
							w,
							map[string]any{
								"data": map[string]any{
									"users": map[string]any{"nodes": want},
								},
							},
						)
					},
				)

				id, err := NewClient(server.Client(), server.URL).UserIDByEmail(
					context.Background(),
					"jane@example.com",
				)

				require.NoError(t, err)
				assert.Empty(t, id)
			}
		},
	)

	t.Run(
		"empty email skips the request",
		func(t *testing.T) {
			t.Parallel()

			id, err := NewClient(http.DefaultClient, "http://127.0.0.1").UserIDByEmail(
				context.Background(),
				"  ",
			)

			require.NoError(t, err)
			assert.Empty(t, id)
		},
	)
}

func TestClient_CreateIssue_IncludesAssignee(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		input      map[string]any
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

			got, _ := vars["input"].(map[string]any)
			input = got

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
		},
	)

	_, err := NewClient(server.Client(), server.URL).CreateIssue(
		context.Background(),
		IssueInput{
			TeamID:     "team-1",
			Title:      "Hello",
			StateID:    "state-1",
			Priority:   2,
			AssigneeID: new("user-1"),
		},
	)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.NotNil(t, input)
	assert.Equal(t, "user-1", input["assigneeId"])
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

func TestClient_ListWorkflowStatesFollowsPages(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		requests   []workflowStatePageVariables
		queries    []string
	)

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, err := decodeGraphQLVariables[workflowStatePageVariables](req)
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
							"workflowStates": map[string]any{
								"nodes": []map[string]any{
									{"id": "state-1", "name": "Todo", "type": "unstarted", "position": 1},
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
						"workflowStates": map[string]any{
							"nodes": []map[string]any{
								{"id": "state-2", "name": "Done", "type": "completed", "position": 2},
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

	states, err := NewClient(server.Client(), server.URL).ListWorkflowStates(
		context.Background(),
		"team-1",
	)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, states, 2)
	require.Len(t, requests, 2)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0], "TaskSyncLinearWorkflowStates")
	assert.Contains(t, queries[0], "hasNextPage")
	assert.Contains(t, queries[1], "TaskSyncLinearWorkflowStates")
	assert.Equal(t, "team-1", requests[0].TeamID)
	assert.Equal(t, 100, requests[0].First)
	assert.Nil(t, requests[0].After)
	assert.Equal(t, "team-1", requests[1].TeamID)
	assert.Equal(t, 100, requests[1].First)
	require.NotNil(t, requests[1].After)
	assert.Equal(t, "cursor-1", *requests[1].After)
	assert.Equal(t, "state-1", states[0].ID)
	assert.Equal(t, "state-2", states[1].ID)
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

func TestClient_LinkAttachment(t *testing.T) {
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
						"attachmentCreate": map[string]any{
							"success":    true,
							"attachment": map[string]any{"id": "attachment-1"},
						},
					},
				},
			)
		},
	)

	attachmentID, err := NewClient(server.Client(), server.URL).LinkAttachment(
		context.Background(),
		"issue-1",
		"https://app.example/organizations/org/governance/tasks/task",
		"Probo task",
	)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Equal(t, "attachment-1", attachmentID)
	assert.Contains(t, query, "attachmentCreate")
	assert.NotContains(t, query, "attachmentLinkURL")

	input, ok := captured["input"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "issue-1", input["issueId"])
	assert.Equal(t, "https://app.example/organizations/org/governance/tasks/task", input["url"])
	assert.Equal(t, "Probo task", input["title"])
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

func TestClient_DoIncludesGraphQLErrorMessage(t *testing.T) {
	t.Parallel()

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			writeJSON(
				w,
				map[string]any{
					"errors": []map[string]any{
						{"message": "You cannot update this issue"},
						{"message": "ignored second error"},
					},
				},
			)
		},
	)

	_, err := NewClient(server.Client(), server.URL).ViewerID(context.Background())
	require.Error(t, err)
	assert.Equal(t, "cannot call Linear graphql: You cannot update this issue", err.Error())
}

func TestClient_HasTeam(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		variables  map[string]any
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

			variables = vars

			nodes := []map[string]any{}
			if vars["id"] == "team-1" {
				nodes = []map[string]any{{"id": "team-1"}}
			}

			writeJSON(w, map[string]any{
				"data": map[string]any{
					"teams": map[string]any{
						"nodes": nodes,
					},
				},
			})
		},
	)

	client := NewClient(server.Client(), server.URL)
	ctx := context.Background()

	found, err := client.HasTeam(ctx, "team-1")

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "team-1", variables["id"])

	found, err = client.HasTeam(ctx, "team-missing")

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, "team-missing", variables["id"])
}

func TestClient_SearchTeamsAndIssues(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		queries    []string
		variables  []map[string]any
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

			queries = append(queries, req.Query)
			variables = append(variables, vars)

			switch {
			case strings.Contains(req.Query, "TaskSyncLinearTeamSearch"):
				writeJSON(w, map[string]any{
					"data": map[string]any{
						"teams": map[string]any{
							"nodes": []map[string]any{
								{"id": "team-1", "name": "Engineering", "key": "ENG"},
							},
							"pageInfo": map[string]any{
								"hasNextPage": false,
								"endCursor":   "",
							},
						},
					},
				})
			case strings.Contains(req.Query, "TaskSyncLinearIssueSearch"):
				writeJSON(w, map[string]any{
					"data": map[string]any{
						"searchIssues": map[string]any{
							"nodes": []map[string]any{
								{
									"id":         "issue-1",
									"identifier": "ENG-12",
									"title":      "Fix login",
									"url":        "https://linear.app/eng/issue/ENG-12",
									"updatedAt":  "2026-09-14T12:00:00Z",
								},
							},
							"pageInfo": map[string]any{
								"hasNextPage": true,
								"endCursor":   "issue-cursor",
							},
						},
					},
				})
			case strings.Contains(req.Query, "TaskSyncLinearIssue("):
				writeJSON(w, map[string]any{
					"data": map[string]any{
						"issue": map[string]any{
							"id":         "issue-1",
							"identifier": "ENG-12",
							"title":      "Fix login",
							"url":        "https://linear.app/eng/issue/ENG-12",
							"updatedAt":  "2026-09-14T12:00:00Z",
							"archivedAt": nil,
							"team":       map[string]any{"id": "team-1"},
						},
					},
				})
			default:
				handlerErr = fmt.Errorf("cannot handle unexpected query %q", req.Query)

				http.Error(w, "unexpected query", http.StatusBadRequest)
			}
		},
	)

	client := NewClient(server.Client(), server.URL)
	ctx := context.Background()

	teams, err := client.SearchTeams(ctx, "eng", 20, nil)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, teams.Teams, 1)
	assert.Equal(t, "ENG", teams.Teams[0].Key)
	assert.False(t, teams.HasNextPage)

	filter, ok := variables[0]["filter"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, filter["or"])

	issues, err := client.SearchIssues(ctx, "team-1", "ENG-12", 20, nil)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.Len(t, issues.Issues, 1)
	assert.Equal(t, "ENG-12", issues.Issues[0].Identifier)
	assert.Equal(t, "Fix login", issues.Issues[0].Title)
	assert.True(t, issues.HasNextPage)
	assert.Equal(t, "issue-cursor", issues.EndCursor)

	issue, err := client.GetIssue(ctx, "issue-1")

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	assert.Equal(t, "team-1", issue.TeamID)
	assert.Equal(t, "ENG-12", issue.Identifier)
}
