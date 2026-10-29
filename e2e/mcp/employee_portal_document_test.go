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

func TestMCP_EmployeePortalDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	portalID := factory.DefaultEmployeePortalID(owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP shared guideline").Create()

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": documentID,
			"changelog":  "Share with employees",
		},
	}, nil)
	require.NoError(t, err)

	var added struct {
		EmployeePortalDocument struct {
			ID         string `json:"id"`
			DocumentID string `json:"document_id"`
			Title      string `json:"title"`
		} `json:"employee_portal_document"`
	}
	mc.CallToolInto("addEmployeePortalDocument", map[string]any{
		"employee_portal_id": portalID,
		"document_id":        documentID,
	}, &added)
	assert.Equal(t, documentID, added.EmployeePortalDocument.DocumentID)
	assert.Equal(t, "MCP shared guideline", added.EmployeePortalDocument.Title)

	var listed struct {
		EmployeePortalDocuments []struct {
			ID         string `json:"id"`
			DocumentID string `json:"document_id"`
		} `json:"employee_portal_documents"`
	}
	mc.CallToolInto("listEmployeePortalDocuments", map[string]any{
		"employee_portal_id": portalID,
	}, &listed)
	require.Len(t, listed.EmployeePortalDocuments, 1)
	assert.Equal(t, added.EmployeePortalDocument.ID, listed.EmployeePortalDocuments[0].ID)

	var deleted struct {
		DeletedEmployeePortalDocumentID string `json:"deleted_employee_portal_document_id"`
	}
	mc.CallToolInto("deleteEmployeePortalDocument", map[string]any{
		"id": added.EmployeePortalDocument.ID,
	}, &deleted)
	assert.Equal(t, added.EmployeePortalDocument.ID, deleted.DeletedEmployeePortalDocumentID)

	mc.CallToolInto("deleteEmployeePortalDocument", map[string]any{
		"id": added.EmployeePortalDocument.ID,
	}, &deleted)
	assert.Equal(t, added.EmployeePortalDocument.ID, deleted.DeletedEmployeePortalDocumentID)

	mc.CallToolInto("listEmployeePortalDocuments", map[string]any{
		"employee_portal_id": portalID,
	}, &listed)
	assert.Empty(t, listed.EmployeePortalDocuments)
}
