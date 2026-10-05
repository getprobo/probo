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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_Task_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	internalControlID := factory.CreateInternalControl(owner)

	// Create
	var addResult struct {
		Task struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"task"`
	}
	mc.CallToolInto("addTask", map[string]any{
		"organization_id":      orgID,
		"internal_control_ids": []any{internalControlID},
		"name":                 factory.SafeName("Task"),
	}, &addResult)
	require.NotEmpty(t, addResult.Task.ID)

	// Get
	var getResult struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	mc.CallToolInto("getTask", map[string]any{
		"id": addResult.Task.ID,
	}, &getResult)
	assert.Equal(t, addResult.Task.ID, getResult.Task.ID)

	// Update
	var updateResult struct {
		Task struct {
			ID      string  `json:"id"`
			Name    string  `json:"name"`
			Content *string `json:"content"`
		} `json:"task"`
	}
	mc.CallToolInto("updateTask", map[string]any{
		"id":      addResult.Task.ID,
		"name":    "Updated Task",
		"content": "Task **description**",
	}, &updateResult)
	assert.Equal(t, "Updated Task", updateResult.Task.Name)
	require.NotNil(t, updateResult.Task.Content)
	assert.Equal(t, "Task **description**\n", *updateResult.Task.Content)

	// List
	var listResult struct {
		Tasks []struct {
			ID string `json:"id"`
		} `json:"tasks"`
	}
	mc.CallToolInto("listTasks", map[string]any{
		"organization_id":     orgID,
		"internal_control_id": internalControlID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Tasks)

	// Delete
	var deleteResult struct {
		DeletedTaskID string `json:"deleted_task_id"`
	}
	mc.CallToolInto("deleteTask", map[string]any{
		"id": addResult.Task.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Task.ID, deleteResult.DeletedTaskID)
}

func TestMCP_Task_Filter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	internalControlID := factory.NewInternalControl(owner).Create()
	matchingID := factory.NewTask(owner, internalControlID).
		WithName("Quarterly access review").
		Create()
	inProgressDecoyID := factory.NewTask(owner, internalControlID).
		WithName("Prepare security training").
		Create()
	factory.NewTask(owner, internalControlID).
		WithName("Annual access review").
		Create()

	for _, taskID := range []string{matchingID, inProgressDecoyID} {
		mc.CallToolInto("updateTask", map[string]any{
			"id":    taskID,
			"state": "IN_PROGRESS",
		}, &struct{}{})
	}

	var listResult struct {
		Tasks []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"tasks"`
	}

	mc.CallToolInto("listTasks", map[string]any{
		"organization_id": orgID,
		"filter": map[string]any{
			"query": "access",
			"state": "IN_PROGRESS",
		},
	}, &listResult)
	require.Len(t, listResult.Tasks, 1)
	assert.Equal(t, matchingID, listResult.Tasks[0].ID)
	assert.Equal(t, "Quarterly access review", listResult.Tasks[0].Name)
	assert.Equal(t, "IN_PROGRESS", listResult.Tasks[0].State)

	mc.CallToolInto("listInternalControlTasks", map[string]any{
		"internal_control_id": internalControlID,
		"filter": map[string]any{
			"query": "access",
			"state": "IN_PROGRESS",
		},
	}, &listResult)
	require.Len(t, listResult.Tasks, 1)
	assert.Equal(t, matchingID, listResult.Tasks[0].ID)
}

func TestMCP_Task_FilterByAssignee(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	internalControlID := factory.NewInternalControl(owner).Create()
	assigneeID := factory.CreateUser(owner)
	otherAssigneeID := factory.CreateUser(owner)
	assignedID := factory.NewTask(owner, internalControlID).
		WithName("Assigned evidence task").
		Create()
	otherID := factory.NewTask(owner, internalControlID).
		WithName("Other assignee task").
		Create()
	factory.NewTask(owner, internalControlID).
		WithName("Unassigned evidence task").
		Create()

	mc.CallToolInto("updateTask", map[string]any{
		"id":             assignedID,
		"assigned_to_id": assigneeID,
	}, &struct{}{})
	mc.CallToolInto("updateTask", map[string]any{
		"id":             otherID,
		"assigned_to_id": otherAssigneeID,
	}, &struct{}{})

	var listResult struct {
		Tasks []struct {
			ID           string `json:"id"`
			AssignedToID string `json:"assigned_to_id"`
		} `json:"tasks"`
	}

	mc.CallToolInto("listTasks", map[string]any{
		"organization_id": orgID,
		"filter": map[string]any{
			"assigned_to_id": assigneeID,
		},
	}, &listResult)
	require.Len(t, listResult.Tasks, 1)
	assert.Equal(t, assignedID, listResult.Tasks[0].ID)
	assert.Equal(t, assigneeID, listResult.Tasks[0].AssignedToID)

	mc.CallToolInto("listInternalControlTasks", map[string]any{
		"internal_control_id": internalControlID,
		"filter": map[string]any{
			"assigned_to_id": assigneeID,
		},
	}, &listResult)
	require.Len(t, listResult.Tasks, 1)
	assert.Equal(t, assignedID, listResult.Tasks[0].ID)
}

func TestMCP_Task_Recurrence(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.CreateInternalControl(owner)

	t.Run("add with recurrence and deadline round-trips", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		var addResult struct {
			Task struct {
				ID                 string `json:"id"`
				RecurrenceInterval string `json:"recurrence_interval"`
			} `json:"task"`
		}
		mc.CallToolInto("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Recurring Task"),
			"deadline":             "2026-01-15T00:00:00Z",
			"recurrence_interval":  "P21D",
		}, &addResult)
		require.NotEmpty(t, addResult.Task.ID)
		assert.Equal(t, "P21D", addResult.Task.RecurrenceInterval)
	})

	t.Run("add with a calendar month keeps P1M", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		var addResult struct {
			Task struct {
				ID                 string `json:"id"`
				RecurrenceInterval string `json:"recurrence_interval"`
			} `json:"task"`
		}
		mc.CallToolInto("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Monthly Task"),
			"deadline":             "2026-01-31T00:00:00Z",
			"recurrence_interval":  "P1M",
		}, &addResult)
		require.NotEmpty(t, addResult.Task.ID)
		assert.Equal(t, "P1M", addResult.Task.RecurrenceInterval)
	})

	t.Run("add as done with recurrence fails", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		errText := mc.CallToolExpectToolError("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Recurring Task"),
			"state":                "DONE",
			"deadline":             "2026-01-15T00:00:00Z",
			"recurrence_interval":  "P21D",
		})
		assert.Contains(t, errText, "state")
	})

	t.Run("add with recurrence but no deadline fails", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		errText := mc.CallToolExpectToolError("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Recurring Task"),
			"recurrence_interval":  "P21D",
		})
		assert.Contains(t, errText, "deadline")
	})

	t.Run("completing before the deadline does not clone", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		var addResult struct {
			Task struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"task"`
		}
		mc.CallToolInto("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Recurring Task"),
			"deadline":             "2027-01-15T00:00:00Z",
			"recurrence_interval":  "P21D",
		}, &addResult)

		var updateResult struct {
			Task struct {
				ID                 string  `json:"id"`
				State              string  `json:"state"`
				RecurrenceInterval *string `json:"recurrence_interval"`
			} `json:"task"`
		}
		mc.CallToolInto("updateTask", map[string]any{
			"id":    addResult.Task.ID,
			"state": "DONE",
		}, &updateResult)
		assert.Equal(t, "DONE", updateResult.Task.State)
		require.NotNil(t, updateResult.Task.RecurrenceInterval)
		assert.Equal(t, "P21D", *updateResult.Task.RecurrenceInterval)
	})

	t.Run("clearing the deadline also clears recurrence", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		var addResult struct {
			Task struct {
				ID string `json:"id"`
			} `json:"task"`
		}
		mc.CallToolInto("addTask", map[string]any{
			"organization_id":      owner.GetOrganizationID().String(),
			"internal_control_ids": []any{internalControlID},
			"name":                 factory.SafeName("Recurring Task"),
			"deadline":             "2027-01-15T00:00:00Z",
			"recurrence_interval":  "P21D",
		}, &addResult)

		var updateResult struct {
			Task struct {
				Deadline           *string `json:"deadline"`
				RecurrenceInterval *string `json:"recurrence_interval"`
			} `json:"task"`
		}
		mc.CallToolInto("updateTask", map[string]any{
			"id":       addResult.Task.ID,
			"deadline": nil,
		}, &updateResult)
		assert.Nil(t, updateResult.Task.Deadline)
		assert.Nil(t, updateResult.Task.RecurrenceInterval)
	})

	t.Run("update to add recurrence without a deadline fails", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)
		taskID := factory.CreateTask(owner, &internalControlID, factory.Attrs{"name": factory.SafeName("Task")})

		errText := mc.CallToolExpectToolError("updateTask", map[string]any{
			"id":                  taskID,
			"recurrence_interval": "P30D",
		})
		assert.Contains(t, errText, "deadline")
	})
}
