// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	measureID := factory.CreateMeasure(owner)

	// Create
	var addResult struct {
		Task struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"task"`
	}
	mc.CallToolInto("addTask", map[string]any{
		"measureId": measureID,
		"name":      factory.SafeName("Task"),
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
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"task"`
	}
	mc.CallToolInto("updateTask", map[string]any{
		"id":   addResult.Task.ID,
		"name": "Updated Task",
	}, &updateResult)
	assert.Equal(t, "Updated Task", updateResult.Task.Name)

	// List
	var listResult struct {
		Tasks []struct {
			ID string `json:"id"`
		} `json:"tasks"`
	}
	mc.CallToolInto("listTasks", map[string]any{
		"measureId": measureID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Tasks)

	// Delete
	var deleteResult struct {
		DeletedTaskID string `json:"deletedTaskId"`
	}
	mc.CallToolInto("deleteTask", map[string]any{
		"id": addResult.Task.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Task.ID, deleteResult.DeletedTaskID)
}
