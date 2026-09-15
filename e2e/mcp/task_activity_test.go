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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type mcpTaskActivity struct {
	ID           string  `json:"id"`
	TaskID       string  `json:"task_id"`
	ActorID      *string `json:"actor_id"`
	ActivityType string  `json:"activity_type"`
	Field        *string `json:"field"`
	OldValue     *string `json:"old_value"`
	NewValue     *string `json:"new_value"`
}

func TestMCP_TaskActivity_ListNewestFirst(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	mc.CallToolInto("updateTask", map[string]any{
		"id":    taskID,
		"state": "IN_PROGRESS",
	}, &struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}{})
	mc.CallToolInto("updateTask", map[string]any{
		"id":       taskID,
		"priority": "HIGH",
	}, &struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}{})

	var listResult struct {
		TaskActivities []mcpTaskActivity `json:"task_activities"`
	}
	mc.CallToolInto("listTaskActivities", map[string]any{
		"task_id": taskID,
	}, &listResult)
	require.GreaterOrEqual(t, len(listResult.TaskActivities), 3)

	assert.Equal(t, "UPDATED", listResult.TaskActivities[0].ActivityType)
	require.NotNil(t, listResult.TaskActivities[0].Field)
	assert.Equal(t, "PRIORITY", *listResult.TaskActivities[0].Field)

	assert.Equal(t, "UPDATED", listResult.TaskActivities[1].ActivityType)
	require.NotNil(t, listResult.TaskActivities[1].Field)
	assert.Equal(t, "STATE", *listResult.TaskActivities[1].Field)
	require.NotNil(t, listResult.TaskActivities[1].OldValue)
	require.NotNil(t, listResult.TaskActivities[1].NewValue)
	assert.Equal(t, "TODO", *listResult.TaskActivities[1].OldValue)
	assert.Equal(t, "IN_PROGRESS", *listResult.TaskActivities[1].NewValue)

	assert.Equal(t, "CREATED", listResult.TaskActivities[2].ActivityType)
	assert.Nil(t, listResult.TaskActivities[2].Field)
	assert.Equal(t, taskID, listResult.TaskActivities[2].TaskID)
	require.NotNil(t, listResult.TaskActivities[2].ActorID)
	assert.Equal(t, owner.GetProfileID().String(), *listResult.TaskActivities[2].ActorID)
}

func TestMCP_TaskActivity_Get(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	var listResult struct {
		TaskActivities []mcpTaskActivity `json:"task_activities"`
	}
	mc.CallToolInto("listTaskActivities", map[string]any{
		"task_id": taskID,
	}, &listResult)
	require.Len(t, listResult.TaskActivities, 1)
	assert.Equal(t, "CREATED", listResult.TaskActivities[0].ActivityType)

	var getResult struct {
		TaskActivity mcpTaskActivity `json:"task_activity"`
	}
	mc.CallToolInto("getTaskActivity", map[string]any{
		"id": listResult.TaskActivities[0].ID,
	}, &getResult)
	assert.Equal(t, listResult.TaskActivities[0].ID, getResult.TaskActivity.ID)
	assert.Equal(t, taskID, getResult.TaskActivity.TaskID)
	assert.Equal(t, "CREATED", getResult.TaskActivity.ActivityType)
}
