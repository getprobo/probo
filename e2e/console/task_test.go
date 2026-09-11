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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTask_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.NewMeasure(owner).WithName("Measure for Task Tests").Create()

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
			"organizationId": owner.GetOrganizationID().String(),
			"measureId":      measureID,
			"name":           "Owner Task",
			"content":        factory.ProseMirrorPlainText("Created by owner"),
			"priority":       "MEDIUM",
		},
	}, &result)
	require.NoError(t, err)

	task := result.CreateTask.TaskEdge.Node
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "Owner Task", task.Name)
	assert.Equal(t, "TODO", task.State)
}

func TestTask_CreateWithoutMeasure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						name
						measure {
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
					ID      string `json:"id"`
					Name    string `json:"name"`
					Measure *struct {
						ID string `json:"id"`
					} `json:"measure"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Task without measure",
			"content":        factory.ProseMirrorPlainText("Created without a measure"),
			"priority":       "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	task := result.CreateTask.TaskEdge.Node
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "Task without measure", task.Name)
	assert.Nil(t, task.Measure)
}

func TestTask_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).
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
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).
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

func TestTask_ListByMeasure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.NewMeasure(owner).Create()

	// Create multiple tasks
	taskNames := []string{"Task A", "Task B", "Task C"}
	for _, name := range taskNames {
		factory.NewTask(owner, measureID).WithName(name).Create()
	}

	query := `
		query GetMeasureTasks($id: ID!) {
			node(id: $id) {
				... on Measure {
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

	err := owner.Execute(query, map[string]any{"id": measureID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Tasks.TotalCount, 3)
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

	measureID := factory.NewMeasure(owner).
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
					"organizationId": owner.GetOrganizationID().String(),
					"measureId":      measureID,
					"name":           "Create State " + state,
					"priority":       "MEDIUM",
					"state":          state,
				},
			}, &result)
			require.NoError(t, err, "State %s should be valid on create", state)
			assert.Equal(t, state, result.CreateTask.TaskEdge.Node.State)
		})

		t.Run("update to state "+state, func(t *testing.T) {
			taskID := factory.NewTask(owner, measureID).
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

	measureID := factory.NewMeasure(owner).
		WithName("Task SubResolver Test").
		Create()

	taskID := factory.NewTask(owner, measureID).
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

	t.Run("measure sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						measure {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID      string `json:"id"`
				Measure struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"measure"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": taskID}, &result)
		require.NoError(t, err)
		assert.Equal(t, measureID, result.Node.Measure.ID)
		assert.NotEmpty(t, result.Node.Measure.Name)
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

	measureID := factory.NewMeasure(owner).
		WithName("Task Description Test").
		Create()

	taskID := factory.NewTask(owner, measureID).
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
	measureID := factory.NewMeasure(owner).
		WithName("Task Assignee Test").
		Create()

	taskID := factory.NewTask(owner, measureID).
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

	measureID := factory.NewMeasure(owner).
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
			"organizationId": owner.GetOrganizationID().String(),
			"measureId":      measureID,
			"name":           "Deadline Test Task",
			"priority":       "MEDIUM",
			"deadline":       "2025-12-31T00:00:00Z",
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
	measureID := factory.NewMeasure(owner).
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
			"organizationId": owner.GetOrganizationID().String(),
			"measureId":      measureID,
			"name":           factory.SafeName("Estimate"),
			"priority":       "MEDIUM",
			"timeEstimate":   "PT1H",
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
				"organizationId": owner.GetOrganizationID().String(),
				"measureId":      measureID,
				"name":           factory.SafeName("Overlong estimate"),
				"priority":       "MEDIUM",
				"timeEstimate":   "P2M",
			},
		})
		require.Error(t, err)
	})

	t.Run("rejects incomplete duration", func(t *testing.T) {
		t.Parallel()

		err := owner.ExecuteShouldFail(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"measureId":      measureID,
				"name":           factory.SafeName("Incomplete estimate"),
				"priority":       "MEDIUM",
				"timeEstimate":   "P1YT",
			},
		})
		require.Error(t, err)
	})
}

func TestTask_Recurrence(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.NewMeasure(owner).WithName("Task Recurrence Test").Create()

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
				"measureId":          measureID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-15T00:00:00Z",
				"recurrenceInterval": "P21D",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateTask.TaskEdge.Node
		assert.NotEmpty(t, node.ID)
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
				"measureId":          measureID,
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
				"measureId":          measureID,
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
				"measureId":          measureID,
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
				"measureId":          measureID,
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
				"measureId":          measureID,
				"name":               factory.SafeName("Recurring Task"),
				"priority":           "MEDIUM",
				"deadline":           "2026-01-31T12:00:00Z",
				"recurrenceInterval": "P1M-29D",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recurrence_interval")
	})

	t.Run("completing clones the next occurrence", func(t *testing.T) {
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
				"measureId":          measureID,
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
				NextTaskEdge *struct {
					Node struct {
						ID                 string  `json:"id"`
						Name               string  `json:"name"`
						State              string  `json:"state"`
						Deadline           *string `json:"deadline"`
						RecurrenceInterval *string `json:"recurrenceInterval"`
					} `json:"node"`
				} `json:"nextTaskEdge"`
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
					nextTaskEdge {
						node {
							id
							name
							state
							deadline
							recurrenceInterval
						}
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
		assert.Nil(t, completed.RecurrenceInterval)
		require.NotNil(t, completed.Deadline)
		assert.Equal(t, "2027-01-15T00:00:00Z", *completed.Deadline)

		require.NotNil(t, updated.UpdateTask.NextTaskEdge)
		next := updated.UpdateTask.NextTaskEdge.Node
		assert.NotEqual(t, completed.ID, next.ID)
		assert.Equal(t, created.CreateTask.TaskEdge.Node.Name, next.Name)
		assert.Equal(t, "TODO", next.State)
		require.NotNil(t, next.RecurrenceInterval)
		assert.Equal(t, "P21D", *next.RecurrenceInterval)
		require.NotNil(t, next.Deadline)
		assert.Equal(t, "2027-02-05T00:00:00Z", *next.Deadline)
	})

	t.Run("completing a monthly task advances a calendar month", func(t *testing.T) {
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
				"measureId":          measureID,
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
					State string `json:"state"`
				} `json:"task"`
				NextTaskEdge *struct {
					Node struct {
						Deadline           *string `json:"deadline"`
						RecurrenceInterval *string `json:"recurrenceInterval"`
					} `json:"node"`
				} `json:"nextTaskEdge"`
			} `json:"updateTask"`
		}

		err = owner.Execute(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { state }
					nextTaskEdge {
						node {
							deadline
							recurrenceInterval
						}
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
		require.NotNil(t, updated.UpdateTask.NextTaskEdge)
		require.NotNil(t, updated.UpdateTask.NextTaskEdge.Node.RecurrenceInterval)
		assert.Equal(t, "P1M", *updated.UpdateTask.NextTaskEdge.Node.RecurrenceInterval)
		require.NotNil(t, updated.UpdateTask.NextTaskEdge.Node.Deadline)
		assert.Equal(t, "2027-02-28T00:00:00Z", *updated.UpdateTask.NextTaskEdge.Node.Deadline)
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
				"measureId":          measureID,
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
				NextTaskEdge *struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"nextTaskEdge"`
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
					nextTaskEdge {
						node { id }
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
		assert.Nil(t, updated.UpdateTask.NextTaskEdge)
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
				"measureId":          measureID,
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

		taskID := factory.NewTask(owner, measureID).
			WithName("Task without deadline").
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
