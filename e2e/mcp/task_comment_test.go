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

func TestMCP_TaskComment_ListOldestFirst(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	firstID := factory.NewTaskComment(owner, taskID).WithContent("Oldest MCP comment").Create()
	secondID := factory.NewTaskComment(owner, taskID).WithContent("Newest MCP comment").Create()

	var listResult struct {
		TaskComments []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"task_comments"`
	}
	mc.CallToolInto("listTaskComments", map[string]any{
		"task_id": taskID,
	}, &listResult)
	require.GreaterOrEqual(t, len(listResult.TaskComments), 2)
	assert.Equal(t, firstID, listResult.TaskComments[0].ID)
	assert.Equal(t, "Oldest MCP comment\n", listResult.TaskComments[0].Content)
	assert.Equal(t, secondID, listResult.TaskComments[1].ID)
	assert.Equal(t, "Newest MCP comment\n", listResult.TaskComments[1].Content)
}

func TestMCP_TaskComment_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	var addResult struct {
		TaskComment struct {
			ID      string `json:"id"`
			Content string `json:"content"`
			OwnerID string `json:"owner_id"`
		} `json:"task_comment"`
	}
	mc.CallToolInto("addTaskComment", map[string]any{
		"task_id": taskID,
		"content": "MCP **comment**",
	}, &addResult)
	require.NotEmpty(t, addResult.TaskComment.ID)
	assert.Equal(t, "MCP **comment**\n", addResult.TaskComment.Content)
	assert.Equal(t, owner.GetProfileID().String(), addResult.TaskComment.OwnerID)

	var getResult struct {
		TaskComment struct {
			ID string `json:"id"`
		} `json:"task_comment"`
	}
	mc.CallToolInto("getTaskComment", map[string]any{
		"id": addResult.TaskComment.ID,
	}, &getResult)
	assert.Equal(t, addResult.TaskComment.ID, getResult.TaskComment.ID)

	var updateResult struct {
		TaskComment struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"task_comment"`
	}
	mc.CallToolInto("updateTaskComment", map[string]any{
		"id":      addResult.TaskComment.ID,
		"content": "Updated MCP comment",
	}, &updateResult)
	assert.Equal(t, "Updated MCP comment\n", updateResult.TaskComment.Content)

	mc.CallToolInto("updateTaskComment", map[string]any{
		"id": addResult.TaskComment.ID,
	}, &updateResult)
	assert.Equal(t, "Updated MCP comment\n", updateResult.TaskComment.Content)

	mc.CallToolInto("updateTaskComment", map[string]any{
		"id":      addResult.TaskComment.ID,
		"content": nil,
	}, &updateResult)
	assert.Equal(t, "", updateResult.TaskComment.Content)

	var listResult struct {
		TaskComments []struct {
			ID string `json:"id"`
		} `json:"task_comments"`
	}
	mc.CallToolInto("listTaskComments", map[string]any{
		"task_id": taskID,
	}, &listResult)
	assert.NotEmpty(t, listResult.TaskComments)

	var deleteResult struct {
		DeletedTaskCommentID string `json:"deleted_task_comment_id"`
	}
	mc.CallToolInto("deleteTaskComment", map[string]any{
		"id": addResult.TaskComment.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.TaskComment.ID, deleteResult.DeletedTaskCommentID)
}
