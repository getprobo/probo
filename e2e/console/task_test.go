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

package console_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTask_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).WithName("Internal control for Task Tests").Create()

	query := `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						name
						state
					}
				}
			}
		}
	`

	var result struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					State string `json:"state"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"internalControlId": internalControlID,
			"name":              "Owner Task",
			"content":           factory.ProseMirrorPlainText("Created by owner"),
			"priority":          "MEDIUM",
		},
	}, &result)
	require.NoError(t, err)

	task := result.CreateTask.TaskEdge.Node
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "Owner Task", task.Name)
	assert.Equal(t, "TODO", task.State)
}

func TestTask_CreateWithoutInternalControl(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						name
						internalControl {
							id
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID              string `json:"id"`
					Name            string `json:"name"`
					InternalControl *struct {
						ID string `json:"id"`
					} `json:"internalControl"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Task without internal control",
			"content":        factory.ProseMirrorPlainText("Created without an internal control"),
			"priority":       "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	task := result.CreateTask.TaskEdge.Node
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "Task without internal control", task.Name)
	assert.Nil(t, task.InternalControl)
}

func TestTask_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()
	taskID := factory.NewTask(owner, internalControlID).
		WithName("Task to Update").
		WithContent("Original description").
		Create()

	query := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateTask struct {
			Task struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":  taskID,
			"name":    "Updated by Owner",
			"content": factory.ProseMirrorPlainText("Owner updated this"),
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, taskID, result.UpdateTask.Task.ID)
	assert.Equal(t, "Updated by Owner", result.UpdateTask.Task.Name)
}

func TestTask_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()
	taskID := factory.NewTask(owner, internalControlID).
		WithName("Task to Delete").
		Create()

	query := `
		mutation DeleteTask($input: DeleteTaskInput!) {
			deleteTask(input: $input) {
				deletedTaskId
			}
		}
	`

	var result struct {
		DeleteTask struct {
			DeletedTaskID string `json:"deletedTaskId"`
		} `json:"deleteTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId": taskID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, taskID, result.DeleteTask.DeletedTaskID)
}

func TestTask_ListByInternalControl(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()

	// Create multiple tasks
	taskNames := []string{"Task A", "Task B", "Task C"}
	for _, name := range taskNames {
		factory.NewTask(owner, internalControlID).WithName(name).Create()
	}

	query := `
		query GetInternalControlTasks($id: ID!) {
			node(id: $id) {
				... on InternalControl {
					id
					tasks(first: 10) {
						edges {
							node {
								id
								name
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID    string `json:"id"`
			Tasks struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tasks"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": internalControlID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Tasks.TotalCount, 3)
}

func TestTask_Filter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()
	matchingTaskID := factory.NewTask(owner, internalControlID).
		WithName("Quarterly access review").
		Create()
	inProgressDecoyID := factory.NewTask(owner, internalControlID).
		WithName("Prepare security training").
		Create()
	factory.NewTask(owner, internalControlID).
		WithName("Annual access review").
		Create()

	updateQuery := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
				}
			}
		}
	`
	for _, taskID := range []string{matchingTaskID, inProgressDecoyID} {
		err := owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"state":  "IN_PROGRESS",
			},
		}, &struct{}{})
		require.NoError(t, err)
	}

	query := `
		query FilterTasks($organizationId: ID!, $filter: TaskFilter) {
			node(id: $organizationId) {
				... on Organization {
					tasks(first: 10, filter: $filter) {
						edges {
							node {
								id
								name
								state
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Tasks struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tasks"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
		"filter": map[string]any{
			"query": "access",
			"state": "IN_PROGRESS",
		},
	}, &result)
	require.NoError(t, err)
	require.Len(t, result.Node.Tasks.Edges, 1)
	assert.Equal(t, matchingTaskID, result.Node.Tasks.Edges[0].Node.ID)
	assert.Equal(t, "Quarterly access review", result.Node.Tasks.Edges[0].Node.Name)
	assert.Equal(t, "IN_PROGRESS", result.Node.Tasks.Edges[0].Node.State)
	assert.Equal(t, 1, result.Node.Tasks.TotalCount)
}

func TestTask_FilterByAssignee(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()
	assigneeID := factory.CreateUser(owner)
	otherAssigneeID := factory.CreateUser(owner)
	assignedTaskID := factory.NewTask(owner, internalControlID).
		WithName("Collect assigned evidence").
		Create()
	otherTaskID := factory.NewTask(owner, internalControlID).
		WithName("Collect other evidence").
		Create()
	factory.NewTask(owner, internalControlID).
		WithName("Collect unassigned evidence").
		Create()

	updateQuery := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
				}
			}
		}
	`
	for _, assignment := range []struct {
		taskID    string
		profileID string
	}{
		{taskID: assignedTaskID, profileID: assigneeID},
		{taskID: otherTaskID, profileID: otherAssigneeID},
	} {
		err := owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"taskId":       assignment.taskID,
				"assignedToId": assignment.profileID,
			},
		}, &struct{}{})
		require.NoError(t, err)
	}

	query := `
		query FilterTasks($organizationId: ID!, $filter: TaskFilter) {
			node(id: $organizationId) {
				... on Organization {
					tasks(first: 10, filter: $filter) {
						edges {
							node {
								id
								assignedTo {
									id
								}
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Tasks struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						AssignedTo struct {
							ID string `json:"id"`
						} `json:"assignedTo"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tasks"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
		"filter": map[string]any{
			"assignedToId": assigneeID,
		},
	}, &result)
	require.NoError(t, err)
	require.Len(t, result.Node.Tasks.Edges, 1)
	assert.Equal(t, assignedTaskID, result.Node.Tasks.Edges[0].Node.ID)
	assert.Equal(t, assigneeID, result.Node.Tasks.Edges[0].Node.AssignedTo.ID)
	assert.Equal(t, 1, result.Node.Tasks.TotalCount)

	err = owner.Execute(query, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
		"filter": map[string]any{
			"assignedToId": assigneeID,
			"query":        "assigned",
			"state":        "TODO",
		},
	}, &result)
	require.NoError(t, err)
	require.Len(t, result.Node.Tasks.Edges, 1)
	assert.Equal(t, assignedTaskID, result.Node.Tasks.Edges[0].Node.ID)
	assert.Equal(t, 1, result.Node.Tasks.TotalCount)

	err = owner.Execute(query, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
		"filter": map[string]any{
			"assignedToId": assigneeID,
			"state":        "IN_PROGRESS",
		},
	}, &result)
	require.NoError(t, err)
	assert.Empty(t, result.Node.Tasks.Edges)
	assert.Equal(t, 0, result.Node.Tasks.TotalCount)

	internalControlQuery := `
		query FilterInternalControlTasks($internalControlId: ID!, $filter: TaskFilter) {
			node(id: $internalControlId) {
				... on InternalControl {
					tasks(first: 10, filter: $filter) {
						edges {
							node {
								id
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var internalControlResult struct {
		Node struct {
			Tasks struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tasks"`
		} `json:"node"`
	}

	err = owner.Execute(internalControlQuery, map[string]any{
		"internalControlId": internalControlID,
		"filter": map[string]any{
			"assignedToId": assigneeID,
		},
	}, &internalControlResult)
	require.NoError(t, err)
	require.Len(t, internalControlResult.Node.Tasks.Edges, 1)
	assert.Equal(t, assignedTaskID, internalControlResult.Node.Tasks.Edges[0].Node.ID)
	assert.Equal(t, 1, internalControlResult.Node.Tasks.TotalCount)

	err = owner.Execute(internalControlQuery, map[string]any{
		"internalControlId": internalControlID,
		"filter": map[string]any{
			"assignedToId": assigneeID,
			"query":        "assigned",
			"state":        "TODO",
		},
	}, &internalControlResult)
	require.NoError(t, err)
	require.Len(t, internalControlResult.Node.Tasks.Edges, 1)
	assert.Equal(t, assignedTaskID, internalControlResult.Node.Tasks.Edges[0].Node.ID)
	assert.Equal(t, 1, internalControlResult.Node.Tasks.TotalCount)

	err = owner.Execute(internalControlQuery, map[string]any{
		"internalControlId": internalControlID,
		"filter": map[string]any{
			"assignedToId": assigneeID,
			"state":        "IN_PROGRESS",
		},
	}, &internalControlResult)
	require.NoError(t, err)
	assert.Empty(t, internalControlResult.Node.Tasks.Edges)
	assert.Equal(t, 0, internalControlResult.Node.Tasks.TotalCount)
}

func TestTask_Filter_LiteralWildcards(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).Create()
	percentTaskID := factory.NewTask(owner, internalControlID).
		WithName("Quarterly 100% review").
		Create()
	underscoreTaskID := factory.NewTask(owner, internalControlID).
		WithName("Q3_access review").
		Create()
	factory.NewTask(owner, internalControlID).
		WithName("Quarterly access review").
		Create()

	query := `
		query FilterTasks($organizationId: ID!, $filter: TaskFilter) {
			node(id: $organizationId) {
				... on Organization {
					tasks(first: 10, filter: $filter) {
						edges {
							node {
								id
								name
							}
						}
						totalCount
					}
				}
			}
		}
	`

	type result struct {
		Node struct {
			Tasks struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tasks"`
		} `json:"node"`
	}

	t.Run("percent is a literal substring", func(t *testing.T) {
		t.Parallel()

		var got result

		err := owner.Execute(query, map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"filter":         map[string]any{"query": "%"},
		}, &got)
		require.NoError(t, err)
		require.Len(t, got.Node.Tasks.Edges, 1)
		assert.Equal(t, percentTaskID, got.Node.Tasks.Edges[0].Node.ID)
		assert.Equal(t, "Quarterly 100% review", got.Node.Tasks.Edges[0].Node.Name)
		assert.Equal(t, 1, got.Node.Tasks.TotalCount)
	})

	t.Run("underscore is a literal substring", func(t *testing.T) {
		t.Parallel()

		var got result

		err := owner.Execute(query, map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"filter":         map[string]any{"query": "_"},
		}, &got)
		require.NoError(t, err)
		require.Len(t, got.Node.Tasks.Edges, 1)
		assert.Equal(t, underscoreTaskID, got.Node.Tasks.Edges[0].Node.ID)
		assert.Equal(t, "Q3_access review", got.Node.Tasks.Edges[0].Node.Name)
		assert.Equal(t, 1, got.Node.Tasks.TotalCount)
	})
}

func TestTask_RequiredFields(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":     "Test Task",
				"priority": "MEDIUM",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing name",
			input: map[string]any{
				"organizationId": "placeholder",
				"priority":       "MEDIUM",
			},
			wantErrorContains: "name",
		},
		{
			name: "missing priority",
			input: map[string]any{
				"organizationId": "placeholder",
				"name":           "Test Task",
			},
			wantErrorContains: "priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateTask($input: CreateTaskInput!) {
					createTask(input: $input) {
						taskEdge {
							node {
								id
							}
						}
					}
				}
			`

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}

			for k, v := range tt.input {
				if v == "placeholder" {
					continue // Skip placeholder values
				}

				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestTask_StateEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("Task State Test").
		Create()

	states := []string{
		"BACKLOG",
		"TODO",
		"IN_PROGRESS",
		"DONE",
		"CANCELED",
		"DUPLICATE",
	}

	for _, state := range states {
		t.Run("create with state "+state, func(t *testing.T) {
			query := `
				mutation CreateTask($input: CreateTaskInput!) {
					createTask(input: $input) {
						taskEdge {
							node {
								id
								state
							}
						}
					}
				}
			`

			var result struct {
				CreateTask struct {
					TaskEdge struct {
						Node struct {
							ID    string `json:"id"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"taskEdge"`
				} `json:"createTask"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId":    owner.GetOrganizationID().String(),
					"internalControlId": internalControlID,
					"name":              "Create State " + state,
					"priority":          "MEDIUM",
					"state":             state,
				},
			}, &result)
			require.NoError(t, err, "State %s should be valid on create", state)
			assert.Equal(t, state, result.CreateTask.TaskEdge.Node.State)
		})

		t.Run("update to state "+state, func(t *testing.T) {
			taskID := factory.NewTask(owner, internalControlID).
				WithName("State Test " + state).
				Create()

			query := `
				mutation UpdateTask($input: UpdateTaskInput!) {
					updateTask(input: $input) {
						task {
							id
							state
						}
					}
				}
			`

			var result struct {
				UpdateTask struct {
					Task struct {
						ID    string `json:"id"`
						State string `json:"state"`
					} `json:"task"`
				} `json:"updateTask"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"taskId": taskID,
					"state":  state,
				},
			}, &result)
			require.NoError(t, err, "State %s should be valid", state)
			assert.Equal(t, state, result.UpdateTask.Task.State)
		})
	}
}

func TestTask_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("Task SubResolver Test").
		Create()

	taskID := factory.NewTask(owner, internalControlID).
		WithName("SubResolver Test Task").
		Create()

	t.Run("task node query", func(t *testing.T) {
		query := `
			query GetTask($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						name
						content
						state
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID      string  `json:"id"`
				Name    string  `json:"name"`
				Content *string `json:"content"`
				State   string  `json:"state"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": taskID}, &result)
		require.NoError(t, err)
		assert.Equal(t, taskID, result.Node.ID)
		assert.Equal(t, "SubResolver Test Task", result.Node.Name)
	})

	t.Run("internal control sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						internalControl {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID              string `json:"id"`
				InternalControl struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"internalControl"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": taskID}, &result)
		require.NoError(t, err)
		assert.Equal(t, internalControlID, result.Node.InternalControl.ID)
		assert.NotEmpty(t, result.Node.InternalControl.Name)
	})

	t.Run("assignedTo sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						assignedTo {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID         string `json:"id"`
				AssignedTo *struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"assignedTo"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": taskID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.AssignedTo)
	})
}

func TestTask_InvalidID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update with invalid ID", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskId": "invalid-id-format",
				"name":   "Test",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		query := `
			mutation DeleteTask($input: DeleteTaskInput!) {
				deleteTask(input: $input) {
					deletedTaskId
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskId": "invalid-id-format",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("query with non-existent ID", func(t *testing.T) {
		query := `
			query GetTask($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo",
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestTask_OmittableContent(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("Task Description Test").
		Create()

	taskID := factory.NewTask(owner, internalControlID).
		WithName("Description Test Task").
		WithContent("Initial description").
		Create()

	t.Run("set content", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						content
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID      string  `json:"id"`
					Content *string `json:"content"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": factory.ProseMirrorPlainText("Updated description"),
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateTask.Task.Content)
		factory.AssertProseMirrorPlainText(t, "Updated description", *result.UpdateTask.Task.Content)
	})

	t.Run("clear content with null", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						content
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID      string `json:"id"`
					Content string `json:"content"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": nil,
			},
		}, &result)
		require.NoError(t, err)
		factory.AssertProseMirrorPlainText(t, "", result.UpdateTask.Task.Content)
	})

	t.Run("update without content preserves value", func(t *testing.T) {
		setQuery := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
					}
				}
			}
		`

		err := owner.Execute(setQuery, map[string]any{
			"input": map[string]any{
				"taskId":  taskID,
				"content": factory.ProseMirrorPlainText("Should persist"),
			},
		}, nil)
		require.NoError(t, err)

		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						name
						content
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID      string  `json:"id"`
					Name    string  `json:"name"`
					Content *string `json:"content"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"name":   "Updated Name",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateTask.Task.Content)
		factory.AssertProseMirrorPlainText(t, "Should persist", *result.UpdateTask.Task.Content)
	})
}

func TestTask_OmittableAssignee(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for assignee
	profileID := factory.CreateUser(owner)
	internalControlID := factory.NewInternalControl(owner).
		WithName("Task Assignee Test").
		Create()

	taskID := factory.NewTask(owner, internalControlID).
		WithName("Assignee Test Task").
		Create()

	t.Run("set assignee", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						assignedTo {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID         string `json:"id"`
					AssignedTo struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"assignedTo"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":       taskID,
				"assignedToId": profileID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID, result.UpdateTask.Task.AssignedTo.ID)
	})

	t.Run("clear assignee", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						assignedTo {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID         string `json:"id"`
					AssignedTo *struct {
						ID string `json:"id"`
					} `json:"assignedTo"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":       taskID,
				"assignedToId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateTask.Task.AssignedTo)
	})
}

func TestTask_OmittableDeadline(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("Task Deadline Test").
		Create()

	// Create task with deadline via mutation (factory doesn't support deadline)
	query := `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						deadline
					}
				}
			}
		}
	`

	var createResult struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID       string  `json:"id"`
					Deadline *string `json:"deadline"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"internalControlId": internalControlID,
			"name":              "Deadline Test Task",
			"priority":          "MEDIUM",
			"deadline":          "2025-12-31T00:00:00Z",
		},
	}, &createResult)
	require.NoError(t, err)

	taskID := createResult.CreateTask.TaskEdge.Node.ID
	require.NotNil(t, createResult.CreateTask.TaskEdge.Node.Deadline)

	t.Run("update deadline", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						deadline
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID       string  `json:"id"`
					Deadline *string `json:"deadline"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":   taskID,
				"deadline": "2026-01-15T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateTask.Task.Deadline)
		assert.Contains(t, *result.UpdateTask.Task.Deadline, "2026-01-15")
	})

	t.Run("clear deadline with null", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						deadline
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID       string  `json:"id"`
					Deadline *string `json:"deadline"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":   taskID,
				"deadline": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateTask.Task.Deadline)
	})
}

func TestTask_TimeEstimate(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).
		WithName("Task Time Estimate Test").
		Create()

	const createQuery = `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						timeEstimate
					}
				}
			}
		}
	`

	var createResult struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID           string  `json:"id"`
					TimeEstimate *string `json:"timeEstimate"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":    owner.GetOrganizationID().String(),
			"internalControlId": internalControlID,
			"name":              factory.SafeName("Estimate"),
			"priority":          "MEDIUM",
			"timeEstimate":      "PT1H",
		},
	}, &createResult)
	require.NoError(t, err)
	require.NotNil(t, createResult.CreateTask.TaskEdge.Node.TimeEstimate)
	assert.Equal(t, "PT1H", *createResult.CreateTask.TaskEdge.Node.TimeEstimate)

	taskID := createResult.CreateTask.TaskEdge.Node.ID

	const updateQuery = `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
					timeEstimate
				}
			}
		}
	`

	var updateResult struct {
		UpdateTask struct {
			Task struct {
				ID           string  `json:"id"`
				TimeEstimate *string `json:"timeEstimate"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err = owner.Execute(updateQuery, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"timeEstimate": "P1M",
		},
	}, &updateResult)
	require.NoError(t, err)
	require.NotNil(t, updateResult.UpdateTask.Task.TimeEstimate)
	assert.Equal(t, "P1M", *updateResult.UpdateTask.Task.TimeEstimate)

	const getQuery = `
		query GetTask($id: ID!) {
			node(id: $id) {
				... on Task {
					id
					timeEstimate
				}
			}
		}
	`

	var getResult struct {
		Node struct {
			ID           string  `json:"id"`
			TimeEstimate *string `json:"timeEstimate"`
		} `json:"node"`
	}

	err = owner.Execute(getQuery, map[string]any{"id": taskID}, &getResult)
	require.NoError(t, err)
	require.NotNil(t, getResult.Node.TimeEstimate)
	assert.Equal(t, "P1M", *getResult.Node.TimeEstimate)

	err = owner.Execute(updateQuery, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"timeEstimate": nil,
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Nil(t, updateResult.UpdateTask.Task.TimeEstimate)

	t.Run("rejects two calendar months", func(t *testing.T) {
		t.Parallel()

		err := owner.ExecuteShouldFail(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"internalControlId": internalControlID,
				"name":              factory.SafeName("Overlong estimate"),
				"priority":          "MEDIUM",
				"timeEstimate":      "P2M",
			},
		})
		require.Error(t, err)
	})

	t.Run("rejects incomplete duration", func(t *testing.T) {
		t.Parallel()

		err := owner.ExecuteShouldFail(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"internalControlId": internalControlID,
				"name":              factory.SafeName("Incomplete estimate"),
				"priority":          "MEDIUM",
				"timeEstimate":      "P1YT",
			},
		})
		require.Error(t, err)
	})
}

func TestTask_Recurrence(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).WithName("Task Recurrence Test").Create()

	createQuery := `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						name
						deadline
						recurrenceInterval
					}
				}
			}
		}
	`

	t.Run("create with recurrence and deadline round-trips", func(t *testing.T) {
		t.Parallel()

		var result struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID                 string  `json:"id"`
						Deadline           *string `json:"deadline"`
						RecurrenceInterval *string `json:"recurrenceInterval"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateTask.TaskEdge.Node
		assert.NotEmpty(t, node.ID)
		require.NotNil(t, node.Deadline)
		assert.Equal(t, "2026-01-15T00:00:00Z", *node.Deadline)
		require.NotNil(t, node.RecurrenceInterval)
		assert.Equal(t, "P21D", *node.RecurrenceInterval)
	})

	t.Run("create with a calendar month keeps P1M", func(t *testing.T) {
		t.Parallel()

		var result struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID                 string  `json:"id"`
						RecurrenceInterval *string `json:"recurrenceInterval"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Monthly Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-31T00:00:00Z",
				"recurrenceInterval": "P1M",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.CreateTask.TaskEdge.Node.RecurrenceInterval)
		assert.Equal(t, "P1M", *result.CreateTask.TaskEdge.Node.RecurrenceInterval)
	})

	t.Run("create with recurrence but no deadline fails", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"recurrenceInterval": "P21D",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deadline")
	})

	t.Run("create as done with recurrence fails", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"state":              "DONE",
				"deadline":           "2026-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "state")
	})

	t.Run("create with a zero interval fails", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-15T00:00:00Z",
				"recurrenceInterval": "PT0S",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recurrence_interval")
	})

	t.Run("create with an interval that does not advance the deadline fails", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-31T12:00:00Z",
				"recurrenceInterval": "P1M-29D",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recurrence_interval")
	})

	t.Run("clones the next occurrence once the deadline has passed", func(t *testing.T) {
		t.Parallel()

		deadline := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)

		name := factory.SafeName("Due Recurring Task")
		description := "Carry this description"
		assigneeID := factory.CreateUser(owner)

		var created struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               name,
				"content":            factory.ProseMirrorPlainText(description),
				"priority":           "HIGH",
				"timeEstimate":       "PT1H",
				"assignedToId":       assigneeID,
				"deadline":           deadline.Format(time.RFC3339),
				"recurrenceInterval": "P21D",
			},
		}, &created)
		require.NoError(t, err)

		sourceID := created.CreateTask.TaskEdge.Node.ID

		require.NotEmpty(t, sourceID)

		expectedDeadline := deadline.Add(21 * 24 * time.Hour)

		type recurringTask struct {
			ID                 string  `json:"id"`
			Name               string  `json:"name"`
			Content            string  `json:"content"`
			State              string  `json:"state"`
			Priority           string  `json:"priority"`
			TimeEstimate       *string `json:"timeEstimate"`
			Deadline           *string `json:"deadline"`
			RecurrenceInterval *string `json:"recurrenceInterval"`
			AssignedTo         *struct {
				ID string `json:"id"`
			} `json:"assignedTo"`
		}

		var (
			source recurringTask
			next   recurringTask
		)

		ok := testutil.Poll(
			t,
			30*time.Second,
			200*time.Millisecond,
			func() bool {
				var listed struct {
					Node struct {
						Tasks struct {
							Edges []struct {
								Node recurringTask `json:"node"`
							} `json:"edges"`
						} `json:"tasks"`
					} `json:"node"`
				}

				err := owner.Execute(`
					query DueRecurringTasks($id: ID!, $filter: TaskFilter) {
						node(id: $id) {
							... on Organization {
								tasks(first: 10, filter: $filter) {
									edges {
										node {
											id
											name
											content
											state
											priority
											timeEstimate
											deadline
											recurrenceInterval
											assignedTo {
												id
											}
										}
									}
								}
							}
						}
					}
				`, map[string]any{
					"id": owner.GetOrganizationID().String(),
					"filter": map[string]any{
						"query": name,
					},
				}, &listed)
				if err != nil {
					return false
				}

				var (
					current   recurringTask
					candidate recurringTask
				)

				for _, edge := range listed.Node.Tasks.Edges {
					node := edge.Node
					if node.ID == sourceID {
						current = node

						continue
					}

					if node.RecurrenceInterval != nil && *node.RecurrenceInterval == "P21D" {
						candidate = node
					}
				}

				if current.ID == "" || current.RecurrenceInterval != nil || candidate.ID == "" {
					return false
				}

				if candidate.Deadline == nil {
					return false
				}

				parsed, err := time.Parse(time.RFC3339Nano, *candidate.Deadline)
				if err != nil || !parsed.Equal(expectedDeadline) {
					return false
				}

				source = current

				next = candidate

				return true
			},
		)
		require.True(t, ok, "next occurrence did not appear")

		assert.Equal(t, name, next.Name)
		assert.Equal(t, "TODO", next.State)
		assert.Equal(t, "HIGH", next.Priority)
		factory.AssertProseMirrorPlainText(t, description, next.Content)
		require.NotNil(t, next.TimeEstimate)
		assert.Equal(t, "PT1H", *next.TimeEstimate)
		require.NotNil(t, next.AssignedTo)
		assert.Equal(t, assigneeID, next.AssignedTo.ID)

		require.NotNil(t, next.RecurrenceInterval)
		assert.Equal(t, "P21D", *next.RecurrenceInterval)

		require.NotNil(t, next.Deadline)

		parsedNext, err := time.Parse(time.RFC3339Nano, *next.Deadline)
		require.NoError(t, err)
		assert.True(t, parsedNext.Equal(expectedDeadline))

		assert.Nil(t, source.RecurrenceInterval)

		require.NotNil(t, source.Deadline)

		parsedSource, err := time.Parse(time.RFC3339Nano, *source.Deadline)
		require.NoError(t, err)
		assert.True(t, parsedSource.Equal(deadline))
	})

	t.Run("completing before the deadline does not clone", func(t *testing.T) {
		t.Parallel()

		var created struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID                 string  `json:"id"`
						Name               string  `json:"name"`
						Deadline           *string `json:"deadline"`
						RecurrenceInterval *string `json:"recurrenceInterval"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2027-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		}, &created)
		require.NoError(t, err)

		var updated struct {
			UpdateTask struct {
				Task struct {
					ID                 string  `json:"id"`
					State              string  `json:"state"`
					Deadline           *string `json:"deadline"`
					RecurrenceInterval *string `json:"recurrenceInterval"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err = owner.Execute(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						state
						deadline
						recurrenceInterval
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": created.CreateTask.TaskEdge.Node.ID,
				"state":  "DONE",
			},
		}, &updated)
		require.NoError(t, err)

		completed := updated.UpdateTask.Task
		assert.Equal(t, "DONE", completed.State)
		require.NotNil(t, completed.RecurrenceInterval)
		assert.Equal(t, "P21D", *completed.RecurrenceInterval)
		require.NotNil(t, completed.Deadline)
		assert.Equal(t, "2027-01-15T00:00:00Z", *completed.Deadline)
	})

	t.Run("completing a monthly task does not clone", func(t *testing.T) {
		t.Parallel()

		var created struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Monthly Task"),
				"priority":           "MEDIUM",
				"deadline":           "2027-01-31T00:00:00Z",
				"recurrenceInterval": "P1M",
			},
		}, &created)
		require.NoError(t, err)

		var updated struct {
			UpdateTask struct {
				Task struct {
					State              string  `json:"state"`
					Deadline           *string `json:"deadline"`
					RecurrenceInterval *string `json:"recurrenceInterval"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err = owner.Execute(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						state
						deadline
						recurrenceInterval
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": created.CreateTask.TaskEdge.Node.ID,
				"state":  "DONE",
			},
		}, &updated)
		require.NoError(t, err)
		assert.Equal(t, "DONE", updated.UpdateTask.Task.State)
		require.NotNil(t, updated.UpdateTask.Task.RecurrenceInterval)
		assert.Equal(t, "P1M", *updated.UpdateTask.Task.RecurrenceInterval)
		require.NotNil(t, updated.UpdateTask.Task.Deadline)
		assert.Equal(t, "2027-01-31T00:00:00Z", *updated.UpdateTask.Task.Deadline)
	})

	t.Run("canceling a recurring task does not clone", func(t *testing.T) {
		t.Parallel()

		var created struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2027-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		}, &created)
		require.NoError(t, err)

		var updated struct {
			UpdateTask struct {
				Task struct {
					ID                 string  `json:"id"`
					State              string  `json:"state"`
					RecurrenceInterval *string `json:"recurrenceInterval"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err = owner.Execute(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						state
						recurrenceInterval
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": created.CreateTask.TaskEdge.Node.ID,
				"state":  "CANCELED",
			},
		}, &updated)
		require.NoError(t, err)
		assert.Equal(t, "CANCELED", updated.UpdateTask.Task.State)
		require.NotNil(t, updated.UpdateTask.Task.RecurrenceInterval)
		assert.Equal(t, "P21D", *updated.UpdateTask.Task.RecurrenceInterval)
	})

	t.Run("clearing the deadline also clears recurrence", func(t *testing.T) {
		t.Parallel()

		var created struct {
			CreateTask struct {
				TaskEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"taskEdge"`
			} `json:"createTask"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"internalControlId":  internalControlID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2027-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		}, &created)
		require.NoError(t, err)

		var updated struct {
			UpdateTask struct {
				Task struct {
					Deadline           *string `json:"deadline"`
					RecurrenceInterval *string `json:"recurrenceInterval"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err = owner.Execute(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						deadline
						recurrenceInterval
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":   created.CreateTask.TaskEdge.Node.ID,
				"deadline": nil,
			},
		}, &updated)
		require.NoError(t, err)
		assert.Nil(t, updated.UpdateTask.Task.Deadline)
		assert.Nil(t, updated.UpdateTask.Task.RecurrenceInterval)
	})

	t.Run("update to add recurrence without a deadline fails", func(t *testing.T) {
		t.Parallel()

		taskID := factory.NewTask(owner, internalControlID).
			WithName(factory.SafeName("Task without deadline")).
			Create()

		_, err := owner.Do(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId":             taskID,
				"recurrenceInterval": "P30D",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deadline")
	})
}
