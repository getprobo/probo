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

func TestMCP_Document_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Document struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"document"`
	}
	mc.CallToolInto("addDocument", map[string]any{
		"organizationId": orgID,
		"title":          factory.SafeName("Document"),
		"documentType":   "POLICY",
	}, &addResult)
	require.NotEmpty(t, addResult.Document.ID)

	// Get
	var getResult struct {
		Document struct {
			ID string `json:"id"`
		} `json:"document"`
	}
	mc.CallToolInto("getDocument", map[string]any{
		"id": addResult.Document.ID,
	}, &getResult)
	assert.Equal(t, addResult.Document.ID, getResult.Document.ID)

	// Update
	var updateResult struct {
		Document struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"document"`
	}
	mc.CallToolInto("updateDocument", map[string]any{
		"id":    addResult.Document.ID,
		"title": "Updated Document",
	}, &updateResult)
	assert.Equal(t, "Updated Document", updateResult.Document.Title)

	// List
	var listResult struct {
		Documents []struct {
			ID string `json:"id"`
		} `json:"documents"`
	}
	mc.CallToolInto("listDocuments", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Documents)

	// Delete
	var deleteResult struct {
		DeletedDocumentID string `json:"deletedDocumentId"`
	}
	mc.CallToolInto("deleteDocument", map[string]any{
		"id": addResult.Document.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Document.ID, deleteResult.DeletedDocumentID)
}

func TestMCP_Document_PublishUsesDefaultApprovers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	approverID := owner.GetProfileID().String()

	addDocument := func(defaultApproverIDs []string) string {
		input := map[string]any{
			"organizationId": orgID,
			"title":          factory.SafeName("Document"),
			"content":        "Body content",
			"classification": "INTERNAL",
			"documentType":   "POLICY",
		}
		if defaultApproverIDs != nil {
			input["defaultApproverIds"] = defaultApproverIDs
		}

		var addResult struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
		}
		mc.CallToolInto("addDocument", input, &addResult)
		require.NotEmpty(t, addResult.Document.ID)

		return addResult.Document.ID
	}

	type publishResult struct {
		DocumentVersion struct {
			Status string `json:"status"`
		} `json:"documentVersion"`
		ApprovalQuorum *struct {
			ID string `json:"id"`
		} `json:"approvalQuorum"`
	}

	t.Run("requests approval from default approvers", func(t *testing.T) {
		docID := addDocument([]string{approverID})

		var result publishResult
		mc.CallToolInto("publishDocument", map[string]any{
			"documentId": docID,
			"minor":      false,
			"changelog":  "Initial major",
		}, &result)

		require.NotNil(t, result.ApprovalQuorum)
		assert.Equal(t, "PENDING_APPROVAL", result.DocumentVersion.Status)
	})

	t.Run("publishes directly without default approvers", func(t *testing.T) {
		docID := addDocument(nil)

		var result publishResult
		mc.CallToolInto("publishDocument", map[string]any{
			"documentId": docID,
			"minor":      false,
			"changelog":  "Initial major",
		}, &result)

		assert.Nil(t, result.ApprovalQuorum)
		assert.Equal(t, "PUBLISHED", result.DocumentVersion.Status)
	})
}
