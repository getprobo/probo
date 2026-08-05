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

func mcpAddProcessingActivityInput(orgID, name string) map[string]any {
	return map[string]any{
		"organization_id":          orgID,
		"name":                     name,
		"special_or_criminal_data": "NO",
		"lawful_basis":             "CONSENT",
		"international_transfers":  false,
		"data_protection_impact_assessment_needed": "NOT_NEEDED",
		"transfer_impact_assessment_needed":        "NOT_NEEDED",
		"role":                                     "CONTROLLER",
	}
}

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
		} `json:"processing_activity"`
	}
	mc.CallToolInto(
		"addProcessingActivity",
		mcpAddProcessingActivityInput(orgID, factory.SafeName("PA")),
		&addResult,
	)
	require.NotEmpty(t, addResult.ProcessingActivity.ID)

	// Get
	var getResult struct {
		ProcessingActivity struct {
			ID string `json:"id"`
		} `json:"processing_activity"`
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
		} `json:"processing_activity"`
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
		} `json:"processing_activities"`
	}
	mc.CallToolInto("listProcessingActivities", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.ProcessingActivities)

	// Delete
	var deleteResult struct {
		DeletedProcessingActivityID string `json:"deleted_processing_activity_id"`
	}
	mc.CallToolInto("deleteProcessingActivity", map[string]any{
		"id": addResult.ProcessingActivity.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.ProcessingActivity.ID, deleteResult.DeletedProcessingActivityID)
}
