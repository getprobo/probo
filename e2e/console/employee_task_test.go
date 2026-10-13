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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const assignedTasksQuery = `
	query AssignedTasks($orgId: ID!) {
		viewer {
			assignedTasks(organizationId: $orgId, first: 50) {
				totalCount
				edges { node { id name } }
			}
		}
	}
`

func TestEmployeeTask_AssignedIsolation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	otherEmployee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	employeeTaskID := factory.NewTaskWithoutInternalControl(owner).
		WithName("Employee assigned task").
		Create()
	viewerTaskID := factory.NewTaskWithoutInternalControl(owner).
		WithName("Viewer assigned task").
		Create()
	otherTaskID := factory.NewTaskWithoutInternalControl(owner).
		WithName("Other employee assigned task").
		Create()

	assignTask(t, owner, employeeTaskID, employee.GetProfileID().String())
	assignTask(t, owner, viewerTaskID, viewer.GetProfileID().String())
	assignTask(t, owner, otherTaskID, otherEmployee.GetProfileID().String())

	t.Run("employee sees only their assigned tasks", func(t *testing.T) {
		t.Parallel()

		ids := assignedTaskIDs(t, employee, employee.GetOrganizationID().String())
		assert.Contains(t, ids, employeeTaskID)
		assert.NotContains(t, ids, viewerTaskID)
		assert.NotContains(t, ids, otherTaskID)
	})

	t.Run("viewer sees only their assigned tasks", func(t *testing.T) {
		t.Parallel()

		ids := assignedTaskIDs(t, viewer, viewer.GetOrganizationID().String())
		assert.Contains(t, ids, viewerTaskID)
		assert.NotContains(t, ids, employeeTaskID)
		assert.NotContains(t, ids, otherTaskID)
	})

	t.Run("assignedTasks ignores a client-supplied assignedToId", func(t *testing.T) {
		t.Parallel()

		const query = `
			query AssignedTasks($orgId: ID!, $assignedToId: ID!) {
				viewer {
					assignedTasks(
						organizationId: $orgId
						first: 50
						filter: { assignedToId: $assignedToId }
					) {
						edges { node { id } }
					}
				}
			}
		`

		var result struct {
			Viewer struct {
				AssignedTasks struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"assignedTasks"`
			} `json:"viewer"`
		}

		err := employee.Execute(query, map[string]any{
			"orgId":        employee.GetOrganizationID().String(),
			"assignedToId": viewer.GetProfileID().String(),
		}, &result)
		require.NoError(t, err)

		ids := make([]string, 0, len(result.Viewer.AssignedTasks.Edges))
		for _, edge := range result.Viewer.AssignedTasks.Edges {
			ids = append(ids, edge.Node.ID)
		}
		assert.Contains(t, ids, employeeTaskID)
		assert.NotContains(t, ids, viewerTaskID)
	})

	t.Run("employee cannot list organization tasks", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query ListTasks($orgId: ID!) {
				node(id: $orgId) {
					... on Organization {
						tasks(first: 10) {
							edges { node { id } }
						}
					}
				}
			}
		`, map[string]any{"orgId": employee.GetOrganizationID().String()})
		testutil.RequireForbiddenError(t, err, "employee should not list organization tasks")
	})

	t.Run("employee cannot access another assignee task via node", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query GetTask($id: ID!) {
				node(id: $id) {
					... on Task {
						id
					}
				}
			}
		`, map[string]any{"id": viewerTaskID})
		testutil.RequireForbiddenError(t, err, "employee should not access another assignee task")
	})

	t.Run("employee can access their assigned task via node", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			query GetTask($id: ID!) {
				node(id: $id) {
					... on Task {
						id
					}
				}
			}
		`, map[string]any{"id": employeeTaskID})
		require.NoError(t, err, "employee should access their assigned task")
	})

	t.Run("employee cannot update another assignee task", func(t *testing.T) {
		t.Parallel()

		_, err := employee.Do(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": viewerTaskID,
				"state":  "DONE",
			},
		})
		testutil.RequireForbiddenError(t, err, "employee should not update another assignee task")
	})

	t.Run("employee can update state on their assigned task", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id state }
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

		err := employee.Execute(query, map[string]any{
			"input": map[string]any{
				"taskId": employeeTaskID,
				"state":  "DONE",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, employeeTaskID, result.UpdateTask.Task.ID)
		assert.Equal(t, "DONE", result.UpdateTask.Task.State)
	})

	t.Run("viewer can open their assigned task", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(`
			query GetTask($id: ID!) {
				node(id: $id) {
					... on Task {
						id
						name
					}
				}
			}
		`, map[string]any{"id": viewerTaskID})
		require.NoError(t, err, "viewer should open their assigned task")
	})

	t.Run("viewer cannot update their assigned task", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(`
			mutation UpdateTask($input: UpdateTaskInput!) {
				updateTask(input: $input) {
					task { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"taskId": viewerTaskID,
				"state":  "DONE",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not update tasks")
	})
}

func assignTask(t *testing.T, owner *testutil.Client, taskID, profileID string) {
	t.Helper()

	_, err := owner.Do(`
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"assignedToId": profileID,
		},
	})
	require.NoError(t, err)
}

func assignedTaskIDs(t *testing.T, client *testutil.Client, organizationID string) []string {
	t.Helper()

	var result struct {
		Viewer struct {
			AssignedTasks struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"assignedTasks"`
		} `json:"viewer"`
	}

	err := client.Execute(assignedTasksQuery, map[string]any{
		"orgId": organizationID,
	}, &result)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Viewer.AssignedTasks.Edges))
	for _, edge := range result.Viewer.AssignedTasks.Edges {
		ids = append(ids, edge.Node.ID)
	}

	assert.Equal(t, len(ids), result.Viewer.AssignedTasks.TotalCount)

	return ids
}
