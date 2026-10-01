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

func TestMCP_TaskPicture_UploadListAndDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	const png = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	var uploaded struct {
		TaskPicture struct {
			ID   string `json:"id"`
			File struct {
				FileName string `json:"file_name"`
				MimeType string `json:"mime_type"`
			} `json:"file"`
		} `json:"task_picture"`
	}
	mc.CallToolInto("uploadTaskPicture", map[string]any{
		"task_id":        taskID,
		"file_name":      "pixel.png",
		"content_type":   "image/png",
		"content_base64": png,
	}, &uploaded)
	require.NotEmpty(t, uploaded.TaskPicture.ID)
	assert.Equal(t, "pixel.png", uploaded.TaskPicture.File.FileName)
	assert.Equal(t, "image/png", uploaded.TaskPicture.File.MimeType)

	var listed struct {
		TaskPictures []struct {
			ID string `json:"id"`
		} `json:"task_pictures"`
	}
	mc.CallToolInto("listTaskPictures", map[string]any{
		"task_id": taskID,
	}, &listed)
	require.Len(t, listed.TaskPictures, 1)
	assert.Equal(t, uploaded.TaskPicture.ID, listed.TaskPictures[0].ID)

	var deleted struct {
		DeletedTaskPictureID string `json:"deleted_task_picture_id"`
	}
	mc.CallToolInto("deleteTaskPicture", map[string]any{
		"id": uploaded.TaskPicture.ID,
	}, &deleted)
	assert.Equal(t, uploaded.TaskPicture.ID, deleted.DeletedTaskPictureID)
}
