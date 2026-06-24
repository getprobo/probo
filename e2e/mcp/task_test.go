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
