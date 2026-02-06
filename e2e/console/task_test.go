// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
					}
				}
			}
		}
	`

	var result struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"measureId":      measureID,
			"name":           "Owner Task",
			"description":    "Created by owner",
		},
	}, &result)
	require.NoError(t, err)

	task := result.CreateTask.TaskEdge.Node
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "Owner Task", task.Name)
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
			"description":    "Created without a measure",
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
		WithDescription("Original description").
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
			"taskId":      taskID,
			"name":        "Updated by Owner",
			"description": "Owner updated this",
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
				"name": "Test Task",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing name",
			input: map[string]any{
				"organizationId": "placeholder",
			},
			wantErrorContains: "name",
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
		"TODO",
		"DONE",
	}

	for _, state := range states {
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
						description
						state
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				State       string  `json:"state"`
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

func TestTask_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(owner).
		WithName("Task Description Test").
		Create()

	taskID := factory.NewTask(owner, measureID).
		WithName("Description Test Task").
		WithDescription("Initial description").
		Create()

	t.Run("set description", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":      taskID,
				"description": "Updated description",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateTask.Task.Description)
		assert.Equal(t, "Updated description", *result.UpdateTask.Task.Description)
	})

	t.Run("clear description with null", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"task"`
			} `json:"updateTask"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId":      taskID,
				"description": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateTask.Task.Description)
	})

	t.Run("update without description preserves value", func(t *testing.T) {
		// First set a description
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
				"taskId":      taskID,
				"description": "Should persist",
			},
		}, nil)
		require.NoError(t, err)

		// Update only name
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task {
						id
						name
						description
					}
				}
			}
		`

		var result struct {
			UpdateTask struct {
				Task struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
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
		require.NotNil(t, result.UpdateTask.Task.Description)
		assert.Equal(t, "Should persist", *result.UpdateTask.Task.Description)
	})
}

func TestTask_OmittableAssignee(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for assignee
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
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
				"assignedToId": profileID.String(),
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID.String(), result.UpdateTask.Task.AssignedTo.ID)
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

func TestTask_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(org1Owner).WithName("Org1 Measure").Create()
	taskID := factory.NewTask(org1Owner, measureID).WithName("Org1 Task").Create()

	t.Run("cannot read task from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": taskID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "task")
	})

	t.Run("cannot update task from another organization", func(t *testing.T) {
		query := `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"name":   "Hijacked Task",
			},
		})
		require.Error(t, err, "Should not be able to update task from another org")
	})

	t.Run("cannot delete task from another organization", func(t *testing.T) {
		query := `
			mutation DeleteTask($input: DeleteTaskInput!) {
				deleteTask(input: $input) {
					deletedTaskId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"taskId": taskID,
			},
		})
		require.Error(t, err, "Should not be able to delete task from another org")
	})
}
