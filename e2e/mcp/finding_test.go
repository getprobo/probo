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

func TestMCP_Finding_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Finding struct {
			ID          string  `json:"id"`
			Kind        string  `json:"kind"`
			Description *string `json:"description"`
		} `json:"finding"`
	}
	mc.CallToolInto("addFinding", map[string]any{
		"organization_id": orgID,
		"kind":            "OBSERVATION",
		"description":     factory.SafeName("Finding"),
	}, &addResult)
	require.NotEmpty(t, addResult.Finding.ID)
	assert.Equal(t, "OBSERVATION", addResult.Finding.Kind)

	// Get
	var getResult struct {
		Finding struct {
			ID string `json:"id"`
		} `json:"finding"`
	}
	mc.CallToolInto("getFinding", map[string]any{
		"id": addResult.Finding.ID,
	}, &getResult)
	assert.Equal(t, addResult.Finding.ID, getResult.Finding.ID)

	// Update
	updatedDescription := "Updated finding description"

	var updateResult struct {
		Finding struct {
			ID          string  `json:"id"`
			Description *string `json:"description"`
		} `json:"finding"`
	}
	mc.CallToolInto("updateFinding", map[string]any{
		"id":          addResult.Finding.ID,
		"description": updatedDescription,
	}, &updateResult)
	require.NotNil(t, updateResult.Finding.Description)
	assert.Equal(t, updatedDescription, *updateResult.Finding.Description)

	// List
	var listResult struct {
		Findings []struct {
			ID string `json:"id"`
		} `json:"findings"`
	}
	mc.CallToolInto("listFindings", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Findings)

	// Delete
	var deleteResult struct {
		DeletedFindingID string `json:"deleted_finding_id"`
	}
	mc.CallToolInto("deleteFinding", map[string]any{
		"id": addResult.Finding.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Finding.ID, deleteResult.DeletedFindingID)
}
