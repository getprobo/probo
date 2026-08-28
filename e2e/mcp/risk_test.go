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

func TestMCP_Risk_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("addRisk", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Risk"),
		"category":        "SECURITY",
	}, &addResult)
	require.NotEmpty(t, addResult.Risk.ID)

	// Get
	var getResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("getRisk", map[string]any{
		"id": addResult.Risk.ID,
	}, &getResult)
	assert.Equal(t, addResult.Risk.ID, getResult.Risk.ID)

	// Update
	var updateResult struct {
		Risk struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
	}
	mc.CallToolInto("updateRisk", map[string]any{
		"id":   addResult.Risk.ID,
		"name": "Updated Risk",
	}, &updateResult)
	assert.Equal(t, "Updated Risk", updateResult.Risk.Name)

	// List
	var listResult struct {
		Risks []struct {
			ID string `json:"id"`
		} `json:"risks"`
	}
	mc.CallToolInto("listRisks", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Risks)

	// Delete
	var deleteResult struct {
		DeletedRiskID string `json:"deleted_risk_id"`
	}
	mc.CallToolInto("deleteRisk", map[string]any{
		"id": addResult.Risk.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Risk.ID, deleteResult.DeletedRiskID)

	// Get deleted risk returns sanitized not-found error
	msg := mc.CallToolExpectToolError("getRisk", map[string]any{
		"id": addResult.Risk.ID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Risk_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("addRisk", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Risk"),
		"category":        "SECURITY",
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_Risk_ListMeasures(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var riskResult struct {
		Risk struct {
			ID string `json:"id"`
		} `json:"risk"`
	}
	mc.CallToolInto("addRisk", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Risk"),
		"category":        "SECURITY",
	}, &riskResult)
	require.NotEmpty(t, riskResult.Risk.ID)

	var measureResult struct {
		Measure struct {
			ID string `json:"id"`
		} `json:"measure"`
	}
	mc.CallToolInto("addMeasure", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Measure"),
		"category":        "POLICY",
	}, &measureResult)
	require.NotEmpty(t, measureResult.Measure.ID)

	var emptyList struct {
		Measures []struct {
			ID string `json:"id"`
		} `json:"measures"`
	}
	mc.CallToolInto("listRiskMeasures", map[string]any{
		"risk_id": riskResult.Risk.ID,
	}, &emptyList)
	assert.Empty(t, emptyList.Measures)

	mc.CallToolInto("linkMeasure", map[string]any{
		"measure_id":  measureResult.Measure.ID,
		"resource_id": riskResult.Risk.ID,
	}, &struct{}{})

	var linkedList struct {
		Measures []struct {
			ID string `json:"id"`
		} `json:"measures"`
	}
	mc.CallToolInto("listRiskMeasures", map[string]any{
		"risk_id": riskResult.Risk.ID,
	}, &linkedList)
	require.Len(t, linkedList.Measures, 1)
	assert.Equal(t, measureResult.Measure.ID, linkedList.Measures[0].ID)

	msg := mc.CallToolExpectToolError("deleteRisk", map[string]any{
		"id": riskResult.Risk.ID,
	})
	assert.Equal(t, "resource is in use", msg)

	mc.CallToolInto("unlinkMeasure", map[string]any{
		"measure_id":  measureResult.Measure.ID,
		"resource_id": riskResult.Risk.ID,
	}, &struct{}{})

	var deleteResult struct {
		DeletedRiskID string `json:"deleted_risk_id"`
	}
	mc.CallToolInto("deleteRisk", map[string]any{
		"id": riskResult.Risk.ID,
	}, &deleteResult)
	assert.Equal(t, riskResult.Risk.ID, deleteResult.DeletedRiskID)
}

func TestMCP_Risk_ListScenarioRisks(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	plannedRiskID := factory.CreateRisk(owner)
	unplannedRiskID := factory.CreateRisk(owner, factory.Attrs{"name": "FilterableScenarioRiskAlpha"})
	otherUnplannedID := factory.CreateRisk(owner, factory.Attrs{"name": "UnrelatedScenarioRiskBeta"})
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, plannedRiskID, analysisID)
	factory.LinkRiskToAnalysis(owner, unplannedRiskID, analysisID)
	factory.LinkRiskToAnalysis(owner, otherUnplannedID, analysisID)
	factory.CreateTreatmentPlan(owner, plannedRiskID, analysisID)

	var listResult struct {
		Risks []struct {
			ID string `json:"id"`
		} `json:"risks"`
	}
	mc.CallToolInto("listRisks", map[string]any{
		"organization_id":  orgID,
		"risk_analysis_id": analysisID,
	}, &listResult)
	require.Len(t, listResult.Risks, 2)
	ids := []string{listResult.Risks[0].ID, listResult.Risks[1].ID}
	assert.Contains(t, ids, unplannedRiskID)
	assert.Contains(t, ids, otherUnplannedID)
	assert.NotContains(t, ids, plannedRiskID)

	var filtered struct {
		Risks []struct {
			ID string `json:"id"`
		} `json:"risks"`
	}
	mc.CallToolInto("listRisks", map[string]any{
		"organization_id":  orgID,
		"risk_analysis_id": analysisID,
		"filter": map[string]any{
			"query": "FilterableScenarioRiskAlpha",
		},
	}, &filtered)
	require.Len(t, filtered.Risks, 1)
	assert.Equal(t, unplannedRiskID, filtered.Risks[0].ID)
}
