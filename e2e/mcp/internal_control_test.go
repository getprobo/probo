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

func TestMCP_Measure_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		InternalControl struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"internal_control"`
	}
	mc.CallToolInto("addInternalControl", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Measure"),
		"category":        "POLICY",
	}, &addResult)
	require.NotEmpty(t, addResult.InternalControl.ID)

	// Get
	var getResult struct {
		InternalControl struct {
			ID string `json:"id"`
		} `json:"internal_control"`
	}
	mc.CallToolInto("getInternalControl", map[string]any{
		"id": addResult.InternalControl.ID,
	}, &getResult)
	assert.Equal(t, addResult.InternalControl.ID, getResult.InternalControl.ID)

	// Update
	var updateResult struct {
		InternalControl struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"internal_control"`
	}
	mc.CallToolInto("updateInternalControl", map[string]any{
		"id":   addResult.InternalControl.ID,
		"name": "Updated Internal control",
	}, &updateResult)
	assert.Equal(t, "Updated Internal control", updateResult.InternalControl.Name)

	// List
	var listResult struct {
		InternalControls []struct {
			ID string `json:"id"`
		} `json:"internal_controls"`
	}
	mc.CallToolInto("listInternalControls", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.InternalControls)

	// Sub-resources (empty lists are fine, just verify the tools work)
	var risksResult struct {
		Risks []struct{ ID string } `json:"risks"`
	}
	mc.CallToolInto("listInternalControlRisks", map[string]any{
		"internal_control_id": addResult.InternalControl.ID,
	}, &risksResult)

	var controlsResult struct {
		Controls []struct{ ID string } `json:"controls"`
	}
	mc.CallToolInto("listInternalControlControls", map[string]any{
		"internal_control_id": addResult.InternalControl.ID,
	}, &controlsResult)

	var tasksResult struct {
		Tasks []struct{ ID string } `json:"tasks"`
	}
	mc.CallToolInto("listInternalControlTasks", map[string]any{
		"internal_control_id": addResult.InternalControl.ID,
	}, &tasksResult)

	// Delete
	var deleteResult struct {
		DeletedInternalControlID string `json:"deleted_internal_control_id"`
	}
	mc.CallToolInto("deleteInternalControl", map[string]any{
		"id": addResult.InternalControl.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.InternalControl.ID, deleteResult.DeletedInternalControlID)
}
