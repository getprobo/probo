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

func TestMCP_ProcessingActivity_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		ProcessingActivity struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"processingActivity"`
	}
	mc.CallToolInto("addProcessingActivity", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("PA"),
		"lawfulBasis":    "CONSENT",
	}, &addResult)
	require.NotEmpty(t, addResult.ProcessingActivity.ID)

	// Get
	var getResult struct {
		ProcessingActivity struct {
			ID string `json:"id"`
		} `json:"processingActivity"`
	}
	mc.CallToolInto("getProcessingActivity", map[string]any{
		"id": addResult.ProcessingActivity.ID,
	}, &getResult)
	assert.Equal(t, addResult.ProcessingActivity.ID, getResult.ProcessingActivity.ID)

	// Update
	var updateResult struct {
		ProcessingActivity struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"processingActivity"`
	}
	mc.CallToolInto("updateProcessingActivity", map[string]any{
		"id":   addResult.ProcessingActivity.ID,
		"name": "Updated PA",
	}, &updateResult)
	assert.Equal(t, "Updated PA", updateResult.ProcessingActivity.Name)

	// List
	var listResult struct {
		ProcessingActivities []struct {
			ID string `json:"id"`
		} `json:"processingActivities"`
	}
	mc.CallToolInto("listProcessingActivities", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.ProcessingActivities)

	// Delete
	var deleteResult struct {
		DeletedProcessingActivityID string `json:"deletedProcessingActivityId"`
	}
	mc.CallToolInto("deleteProcessingActivity", map[string]any{
		"id": addResult.ProcessingActivity.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.ProcessingActivity.ID, deleteResult.DeletedProcessingActivityID)
}
