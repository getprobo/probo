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
