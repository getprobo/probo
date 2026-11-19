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
			Title       string  `json:"title"`
			Description *string `json:"description"`
		} `json:"finding"`
	}
	mc.CallToolInto("addFinding", map[string]any{
		"organization_id": orgID,
		"kind":            "OBSERVATION",
		"title":           "Access review gap",
		"description":     factory.SafeName("Finding"),
	}, &addResult)
	require.NotEmpty(t, addResult.Finding.ID)
	assert.Equal(t, "OBSERVATION", addResult.Finding.Kind)
	assert.Equal(t, "Access review gap", addResult.Finding.Title)

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

func TestMCP_CorrectiveActionPlan(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	internalControlID := factory.CreateInternalControl(owner)
	taskID := factory.CreateTask(owner, &internalControlID)

	var finding struct {
		Finding struct {
			ID string `json:"id"`
		} `json:"finding"`
	}
	mc.CallToolInto("addFinding", map[string]any{
		"organization_id": orgID,
		"kind":            "MINOR_NONCONFORMITY",
		"title":           factory.SafeName("Finding"),
	}, &finding)
	findingID := finding.Finding.ID

	type plan struct {
		ID          string  `json:"id"`
		FindingID   string  `json:"finding_id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"due_date"`
	}

	var added struct {
		CorrectiveActionPlan plan `json:"corrective_action_plan"`
	}
	mc.CallToolInto("addCorrectiveActionPlan", map[string]any{
		"finding_id":  findingID,
		"title":       "Restore access reviews",
		"description": "Quarterly access reviews",
		"due_date":    "2027-01-15T00:00:00Z",
	}, &added)
	planID := added.CorrectiveActionPlan.ID
	require.NotEmpty(t, planID)
	assert.Equal(t, findingID, added.CorrectiveActionPlan.FindingID)
	assert.Equal(t, "Restore access reviews", added.CorrectiveActionPlan.Title)

	var updated struct {
		CorrectiveActionPlan plan `json:"corrective_action_plan"`
	}
	mc.CallToolInto("updateCorrectiveActionPlan", map[string]any{
		"id":          planID,
		"title":       "Restore quarterly access reviews",
		"description": nil,
		"due_date":    nil,
	}, &updated)
	assert.Equal(t, "Restore quarterly access reviews", updated.CorrectiveActionPlan.Title)
	assert.Nil(t, updated.CorrectiveActionPlan.Description)
	assert.Nil(t, updated.CorrectiveActionPlan.DueDate)

	var got struct {
		CorrectiveActionPlan plan `json:"corrective_action_plan"`
	}
	mc.CallToolInto("getCorrectiveActionPlan", map[string]any{"id": planID}, &got)
	assert.Equal(t, planID, got.CorrectiveActionPlan.ID)

	var listed struct {
		CorrectiveActionPlans []plan `json:"corrective_action_plans"`
	}
	mc.CallToolInto("listFindingCorrectiveActionPlans", map[string]any{"finding_id": findingID}, &listed)
	require.Len(t, listed.CorrectiveActionPlans, 1)
	assert.Equal(t, planID, listed.CorrectiveActionPlans[0].ID)

	mc.CallToolInto("linkInternalControl", map[string]any{
		"internal_control_id": internalControlID,
		"resource_id":         planID,
	}, &struct{}{})

	var controls struct {
		InternalControls []struct {
			ID string `json:"id"`
		} `json:"internal_controls"`
	}
	mc.CallToolInto("listCorrectiveActionPlanInternalControls", map[string]any{
		"corrective_action_plan_id": planID,
	}, &controls)
	require.Len(t, controls.InternalControls, 1)
	assert.Equal(t, internalControlID, controls.InternalControls[0].ID)

	type taskList struct {
		Tasks []struct {
			ID string `json:"id"`
		} `json:"tasks"`
	}

	var planTasks taskList
	mc.CallToolInto("listCorrectiveActionPlanTasks", map[string]any{
		"corrective_action_plan_id": planID,
	}, &planTasks)
	require.Len(t, planTasks.Tasks, 1)
	assert.Equal(t, taskID, planTasks.Tasks[0].ID)

	var findingTasks taskList
	mc.CallToolInto("listFindingTasks", map[string]any{"finding_id": findingID}, &findingTasks)
	require.Len(t, findingTasks.Tasks, 1)
	assert.Equal(t, taskID, findingTasks.Tasks[0].ID)

	mc.CallToolInto("unlinkInternalControl", map[string]any{
		"internal_control_id": internalControlID,
		"resource_id":         planID,
	}, &struct{}{})

	mc.CallToolInto("listFindingTasks", map[string]any{"finding_id": findingID}, &findingTasks)
	assert.Empty(t, findingTasks.Tasks)

	var deleted struct {
		DeletedCorrectiveActionPlanID string `json:"deleted_corrective_action_plan_id"`
	}
	mc.CallToolInto("deleteCorrectiveActionPlan", map[string]any{"id": planID}, &deleted)
	assert.Equal(t, planID, deleted.DeletedCorrectiveActionPlanID)
}
