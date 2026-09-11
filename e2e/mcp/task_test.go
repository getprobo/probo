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
	measureID := factory.CreateMeasure(owner)

	// Create
	var addResult struct {
		Task struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"task"`
	}
	mc.CallToolInto("addTask", map[string]any{
		"organization_id": orgID,
		"measure_id":      measureID,
		"name":            factory.SafeName("Task"),
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
		"organization_id": orgID,
		"measure_id":      measureID,
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

func TestMCP_Task_Recurrence(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureID := factory.CreateMeasure(owner)

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
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Recurring Task"),
			"deadline":            "2026-01-15T00:00:00Z",
			"recurrence_interval": "P21D",
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
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Monthly Task"),
			"deadline":            "2026-01-31T00:00:00Z",
			"recurrence_interval": "P1M",
		}, &addResult)
		require.NotEmpty(t, addResult.Task.ID)
		assert.Equal(t, "P1M", addResult.Task.RecurrenceInterval)
	})

	t.Run("add as done with recurrence fails", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		errText := mc.CallToolExpectToolError("addTask", map[string]any{
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Recurring Task"),
			"state":               "DONE",
			"deadline":            "2026-01-15T00:00:00Z",
			"recurrence_interval": "P21D",
		})
		assert.Contains(t, errText, "state")
	})

	t.Run("add with recurrence but no deadline fails", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		errText := mc.CallToolExpectToolError("addTask", map[string]any{
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Recurring Task"),
			"recurrence_interval": "P21D",
		})
		assert.Contains(t, errText, "deadline")
	})

	t.Run("completing clones the next occurrence", func(t *testing.T) {
		t.Parallel()

		mc := testutil.NewMCPClient(t, owner)

		var addResult struct {
			Task struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"task"`
		}
		mc.CallToolInto("addTask", map[string]any{
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Recurring Task"),
			"deadline":            "2027-01-15T00:00:00Z",
			"recurrence_interval": "P21D",
		}, &addResult)

		var updateResult struct {
			Task struct {
				ID                 string  `json:"id"`
				State              string  `json:"state"`
				RecurrenceInterval *string `json:"recurrence_interval"`
			} `json:"task"`
			NextTask *struct {
				ID                 string  `json:"id"`
				Name               string  `json:"name"`
				State              string  `json:"state"`
				Deadline           *string `json:"deadline"`
				RecurrenceInterval *string `json:"recurrence_interval"`
			} `json:"next_task"`
		}
		mc.CallToolInto("updateTask", map[string]any{
			"id":    addResult.Task.ID,
			"state": "DONE",
		}, &updateResult)
		assert.Equal(t, "DONE", updateResult.Task.State)
		assert.Nil(t, updateResult.Task.RecurrenceInterval)

		require.NotNil(t, updateResult.NextTask)
		next := updateResult.NextTask
		assert.NotEqual(t, addResult.Task.ID, next.ID)
		assert.Equal(t, addResult.Task.Name, next.Name)
		assert.Equal(t, "TODO", next.State)
		require.NotNil(t, next.RecurrenceInterval)
		assert.Equal(t, "P21D", *next.RecurrenceInterval)
		require.NotNil(t, next.Deadline)
		assert.Equal(t, "2027-02-05T00:00:00Z", *next.Deadline)
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
			"organization_id":     owner.GetOrganizationID().String(),
			"measure_id":          measureID,
			"name":                factory.SafeName("Recurring Task"),
			"deadline":            "2027-01-15T00:00:00Z",
			"recurrence_interval": "P21D",
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
		taskID := factory.CreateTask(owner, &measureID, factory.Attrs{"name": factory.SafeName("Task")})

		errText := mc.CallToolExpectToolError("updateTask", map[string]any{
			"id":                  taskID,
			"recurrence_interval": "P30D",
		})
		assert.Contains(t, errText, "deadline")
	})
}
