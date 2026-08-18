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

func TestMCP_TreatmentPlan_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)

	var addResult struct {
		TreatmentPlan struct {
			ID                 string `json:"id"`
			Treatment          string `json:"treatment"`
			InherentLikelihood int    `json:"inherent_likelihood"`
			InherentImpact     int    `json:"inherent_impact"`
			InherentRiskScore  int    `json:"inherent_risk_score"`
			ResidualLikelihood int    `json:"residual_likelihood"`
			ResidualImpact     int    `json:"residual_impact"`
		} `json:"treatment_plan"`
	}
	mc.CallToolInto("addTreatmentPlan", map[string]any{
		"risk_id":             riskID,
		"risk_analysis_id":    analysisID,
		"treatment":           "MITIGATED",
		"owner_id":            owner.GetProfileID().String(),
		"inherent_likelihood": 3,
		"inherent_impact":     4,
	}, &addResult)
	require.NotEmpty(t, addResult.TreatmentPlan.ID)
	assert.Equal(t, "MITIGATED", addResult.TreatmentPlan.Treatment)
	assert.Equal(t, 3, addResult.TreatmentPlan.InherentLikelihood)
	assert.Equal(t, 4, addResult.TreatmentPlan.InherentImpact)
	assert.Equal(t, 12, addResult.TreatmentPlan.InherentRiskScore)
	assert.Equal(t, 3, addResult.TreatmentPlan.ResidualLikelihood)
	assert.Equal(t, 4, addResult.TreatmentPlan.ResidualImpact)

	var getResult struct {
		TreatmentPlan struct {
			ID        string `json:"id"`
			Treatment string `json:"treatment"`
		} `json:"treatment_plan"`
	}
	mc.CallToolInto("getTreatmentPlan", map[string]any{
		"id": addResult.TreatmentPlan.ID,
	}, &getResult)
	assert.Equal(t, addResult.TreatmentPlan.ID, getResult.TreatmentPlan.ID)
	assert.Equal(t, "MITIGATED", getResult.TreatmentPlan.Treatment)

	var updateResult struct {
		TreatmentPlan struct {
			ID                 string `json:"id"`
			Treatment          string `json:"treatment"`
			ResidualLikelihood int    `json:"residual_likelihood"`
			ResidualImpact     int    `json:"residual_impact"`
		} `json:"treatment_plan"`
	}
	mc.CallToolInto("updateTreatmentPlan", map[string]any{
		"id":                  addResult.TreatmentPlan.ID,
		"treatment":           "AVOIDED",
		"residual_likelihood": 2,
		"residual_impact":     2,
	}, &updateResult)
	assert.Equal(t, "AVOIDED", updateResult.TreatmentPlan.Treatment)
	assert.Equal(t, 2, updateResult.TreatmentPlan.ResidualLikelihood)
	assert.Equal(t, 2, updateResult.TreatmentPlan.ResidualImpact)

	var listResult struct {
		TreatmentPlans []struct {
			ID string `json:"id"`
		} `json:"treatment_plans"`
	}
	mc.CallToolInto("listTreatmentPlans", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.TreatmentPlans)

	var listByAnalysis struct {
		TreatmentPlans []struct {
			ID string `json:"id"`
		} `json:"treatment_plans"`
	}
	mc.CallToolInto("listTreatmentPlans", map[string]any{
		"organization_id":  orgID,
		"risk_analysis_id": analysisID,
	}, &listByAnalysis)
	require.Len(t, listByAnalysis.TreatmentPlans, 1)
	assert.Equal(t, addResult.TreatmentPlan.ID, listByAnalysis.TreatmentPlans[0].ID)

	var deleteResult struct {
		DeletedTreatmentPlanID string `json:"deleted_treatment_plan_id"`
	}
	mc.CallToolInto("deleteTreatmentPlan", map[string]any{
		"id": addResult.TreatmentPlan.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.TreatmentPlan.ID, deleteResult.DeletedTreatmentPlanID)

	msg := mc.CallToolExpectToolError("getTreatmentPlan", map[string]any{
		"id": addResult.TreatmentPlan.ID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_TreatmentPlan_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)

	msg := viewerMC.CallToolExpectToolError("addTreatmentPlan", map[string]any{
		"risk_id":             riskID,
		"risk_analysis_id":    analysisID,
		"treatment":           "MITIGATED",
		"owner_id":            owner.GetProfileID().String(),
		"inherent_likelihood": 2,
		"inherent_impact":     2,
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_TreatmentPlan_Duplicate(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	riskID := factory.CreateRisk(owner)
	analysisID := factory.CreateRiskAnalysis(owner)
	factory.LinkRiskToAnalysis(owner, riskID, analysisID)

	var addResult struct {
		TreatmentPlan struct {
			ID string `json:"id"`
		} `json:"treatment_plan"`
	}
	mc.CallToolInto("addTreatmentPlan", map[string]any{
		"risk_id":             riskID,
		"risk_analysis_id":    analysisID,
		"treatment":           "MITIGATED",
		"owner_id":            owner.GetProfileID().String(),
		"inherent_likelihood": 2,
		"inherent_impact":     2,
	}, &addResult)
	require.NotEmpty(t, addResult.TreatmentPlan.ID)

	msg := mc.CallToolExpectToolError("addTreatmentPlan", map[string]any{
		"risk_id":             riskID,
		"risk_analysis_id":    analysisID,
		"treatment":           "ACCEPTED",
		"owner_id":            owner.GetProfileID().String(),
		"inherent_likelihood": 2,
		"inherent_impact":     2,
	})
	assert.Equal(t, "resource already exists", msg)
}
