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
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTaskComment_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).WithName("Task for comments").Create()

	query := `
		mutation CreateTaskComment($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node {
						id
						description
						owner {
							id
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateTaskComment struct {
			TaskCommentEdge struct {
				Node struct {
					ID          string `json:"id"`
					Description string `json:"description"`
					Owner       struct {
						ID string `json:"id"`
					} `json:"owner"`
				} `json:"node"`
			} `json:"taskCommentEdge"`
		} `json:"createTaskComment"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":      taskID,
			"description": "First comment",
		},
	}, &result)
	require.NoError(t, err)

	comment := result.CreateTaskComment.TaskCommentEdge.Node
	assert.NotEmpty(t, comment.ID)
	assert.Equal(t, "First comment", comment.Description)
	assert.Equal(t, owner.GetProfileID().String(), comment.Owner.ID)
}

func TestTaskComment_CreateWithOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	assignee := testutil.NewClientInOrg(t, testutil.RoleOwner, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).WithName("Task for owned comments").Create()
	ownerID := assignee.GetProfileID().String()

	commentID := factory.NewTaskComment(owner, taskID).
		WithDescription("Assigned comment").
		WithOwnerID(ownerID).
		Create()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on TaskComment {
					id
					owner { id }
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID    string `json:"id"`
			Owner struct {
				ID string `json:"id"`
			} `json:"owner"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": commentID}, &result)
	require.NoError(t, err)
	assert.Equal(t, ownerID, result.Node.Owner.ID)
}

func TestTaskComment_ListOldestFirst(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	firstID := factory.NewTaskComment(owner, taskID).WithDescription("Oldest comment").Create()
	secondID := factory.NewTaskComment(owner, taskID).WithDescription("Newest comment").Create()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					comments(first: 10, orderBy: { field: CREATED_AT, direction: ASC }) {
						totalCount
						edges {
							node {
								id
								description
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Comments struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID          string `json:"id"`
						Description string `json:"description"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"comments"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": taskID}, &result)
	require.NoError(t, err)

	require.GreaterOrEqual(t, result.Node.Comments.TotalCount, 2)
	require.GreaterOrEqual(t, len(result.Node.Comments.Edges), 2)
	assert.Equal(t, firstID, result.Node.Comments.Edges[0].Node.ID)
	assert.Equal(t, "Oldest comment", result.Node.Comments.Edges[0].Node.Description)
	assert.Equal(t, secondID, result.Node.Comments.Edges[1].Node.ID)
	assert.Equal(t, "Newest comment", result.Node.Comments.Edges[1].Node.Description)
}

func TestTaskComment_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	commentID := factory.NewTaskComment(owner, taskID).WithDescription("Original comment").Create()

	query := `
		mutation UpdateTaskComment($input: UpdateTaskCommentInput!) {
			updateTaskComment(input: $input) {
				taskComment {
					id
					description
				}
			}
		}
	`

	var result struct {
		UpdateTaskComment struct {
			TaskComment struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"taskComment"`
		} `json:"updateTaskComment"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskCommentId": commentID,
			"description":   "Updated comment",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, commentID, result.UpdateTaskComment.TaskComment.ID)
	assert.Equal(t, "Updated comment", result.UpdateTaskComment.TaskComment.Description)
}

func TestTaskComment_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	commentID := factory.NewTaskComment(owner, taskID).Create()

	query := `
		mutation DeleteTaskComment($input: DeleteTaskCommentInput!) {
			deleteTaskComment(input: $input) {
				deletedTaskCommentId
			}
		}
	`

	var result struct {
		DeleteTaskComment struct {
			DeletedTaskCommentID string `json:"deletedTaskCommentId"`
		} `json:"deleteTaskComment"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskCommentId": commentID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, commentID, result.DeleteTaskComment.DeletedTaskCommentID)
}

func TestTaskComment_ViewerCannotCreate(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	query := `
		mutation CreateTaskComment($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node { id }
				}
			}
		}
	`

	_, err := viewer.Do(query, map[string]any{
		"input": map[string]any{
			"taskId":      taskID,
			"description": "Viewer comment",
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer cannot create task comments")
}

func TestTaskComment_CreateValidation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	const query = `
		mutation CreateTaskComment($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node { id }
				}
			}
		}
	`

	tests := []struct {
		name              string
		input             map[string]any
		wantErrorContains string
	}{
		{
			name: "empty description",
			input: map[string]any{
				"description": "",
			},
			wantErrorContains: "description",
		},
		{
			name: "HTML injection",
			input: map[string]any{
				"description": "<script>xss</script>",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "control char",
			input: map[string]any{
				"description": "Test\x00",
			},
			wantErrorContains: "control",
		},
		{
			name: "max length",
			input: map[string]any{
				"description": strings.Repeat("c", 5001),
			},
			wantErrorContains: "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := map[string]any{"taskId": taskID}
			maps.Copy(input, tt.input)

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestTaskComment_AuditorCannotAccess(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	commentID := factory.NewTaskComment(owner, taskID).Create()

	t.Run("cannot create", func(t *testing.T) {
		t.Parallel()

		query := `
			mutation CreateTaskComment($input: CreateTaskCommentInput!) {
				createTaskComment(input: $input) {
					taskCommentEdge {
						node { id }
					}
				}
			}
		`

		_, err := auditor.Do(query, map[string]any{
			"input": map[string]any{
				"taskId":      taskID,
				"description": "Auditor comment",
			},
		})
		testutil.RequireForbiddenError(t, err, "auditor cannot create task comments")
	})

	t.Run("cannot list", func(t *testing.T) {
		t.Parallel()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Task {
						comments(first: 10) {
							edges {
								node { id }
							}
						}
					}
				}
			}
		`

		_, err := auditor.Do(query, map[string]any{"id": taskID})
		require.Error(t, err)
	})

	t.Run("cannot get", func(t *testing.T) {
		t.Parallel()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on TaskComment {
						id
						description
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := auditor.Execute(query, map[string]any{"id": commentID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "task comment")
	})
}

func TestTaskComment_ViewerCanList(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	commentID := factory.NewTaskComment(owner, taskID).WithDescription("Visible comment").Create()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					comments(first: 10) {
						edges {
							node {
								id
								description
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Comments struct {
				Edges []struct {
					Node struct {
						ID          string `json:"id"`
						Description string `json:"description"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"comments"`
		} `json:"node"`
	}

	err := viewer.Execute(query, map[string]any{"id": taskID}, &result)
	require.NoError(t, err)

	require.NotEmpty(t, result.Node.Comments.Edges)
	assert.Equal(t, commentID, result.Node.Comments.Edges[0].Node.ID)
	assert.Equal(t, "Visible comment", result.Node.Comments.Edges[0].Node.Description)
}

func TestTaskComment_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(org1Owner).Create()
	commentID := factory.NewTaskComment(org1Owner, taskID).Create()

	t.Run("cannot read comment from another organization", func(t *testing.T) {
		t.Parallel()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on TaskComment {
						id
						description
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": commentID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "task comment")
	})

	t.Run("cannot update comment from another organization", func(t *testing.T) {
		t.Parallel()

		query := `
			mutation UpdateTaskComment($input: UpdateTaskCommentInput!) {
				updateTaskComment(input: $input) {
					taskComment { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskCommentId": commentID,
				"description":   "Hijacked comment",
			},
		})
		require.Error(t, err)
	})

	t.Run("cannot delete comment from another organization", func(t *testing.T) {
		t.Parallel()

		query := `
			mutation DeleteTaskComment($input: DeleteTaskCommentInput!) {
				deleteTaskComment(input: $input) {
					deletedTaskCommentId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskCommentId": commentID,
			},
		})
		require.Error(t, err)
	})
}
