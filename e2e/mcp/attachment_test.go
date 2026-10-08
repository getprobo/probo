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

// 1x1 PNG.
const attachmentPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func TestMCP_UploadFile(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var uploaded struct {
		ID       string `json:"id"`
		FileName string `json:"file_name"`
		MimeType string `json:"mime_type"`
		Size     int    `json:"size"`
	}
	mc.CallToolInto("uploadAttachmentFile", map[string]any{
		"organization_id": orgID,
		"file_name":       "dot.png",
		"mime_type":       "image/png",
		"content_base64":  attachmentPNG,
	}, &uploaded)
	require.NotEmpty(t, uploaded.ID)
	assert.Equal(t, "dot.png", uploaded.FileName)
	assert.Equal(t, "image/png", uploaded.MimeType)
	assert.Positive(t, uploaded.Size)

	var added struct {
		Task struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"task"`
	}
	mc.CallToolInto("addTask", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Attachment"),
		"content":         "![diagram](/api/files/v1/attachments/" + uploaded.ID + ")",
	}, &added)
	assert.Contains(t, added.Task.Content, uploaded.ID)

	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)
	auditorMCP := testutil.NewMCPClient(t, auditor)
	denied := auditorMCP.CallToolExpectToolError("uploadAttachmentFile", map[string]any{
		"organization_id": orgID,
		"file_name":       "dot.png",
		"mime_type":       "image/png",
		"content_base64":  attachmentPNG,
	})
	assert.NotEmpty(t, denied)
}
