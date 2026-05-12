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
