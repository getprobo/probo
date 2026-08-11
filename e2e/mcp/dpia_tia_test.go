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

func TestMCP_DPIA_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	paID := factory.CreateProcessingActivity(owner)

	// Create
	var addResult struct {
		DataProtectionImpactAssessment struct {
			ID string `json:"id"`
		} `json:"data_protection_impact_assessment"`
	}
	mc.CallToolInto("addDataProtectionImpactAssessment", map[string]any{
		"processing_activity_id": paID,
	}, &addResult)
	require.NotEmpty(t, addResult.DataProtectionImpactAssessment.ID)

	// Get
	var getResult struct {
		DataProtectionImpactAssessment struct {
			ID string `json:"id"`
		} `json:"data_protection_impact_assessment"`
	}
	mc.CallToolInto("getDataProtectionImpactAssessment", map[string]any{
		"id": addResult.DataProtectionImpactAssessment.ID,
	}, &getResult)
	assert.Equal(t, addResult.DataProtectionImpactAssessment.ID, getResult.DataProtectionImpactAssessment.ID)

	// Update
	var updateResult struct {
		DataProtectionImpactAssessment struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"data_protection_impact_assessment"`
	}
	mc.CallToolInto("updateDataProtectionImpactAssessment", map[string]any{
		"id":          addResult.DataProtectionImpactAssessment.ID,
		"description": "Updated DPIA",
	}, &updateResult)
	assert.Equal(t, "Updated DPIA", updateResult.DataProtectionImpactAssessment.Description)

	// List
	var listResult struct {
		DataProtectionImpactAssessments []struct {
			ID string `json:"id"`
		} `json:"data_protection_impact_assessments"`
	}
	mc.CallToolInto("listDataProtectionImpactAssessments", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.DataProtectionImpactAssessments)

	// Delete
	var deleteResult struct {
		DeletedDataProtectionImpactAssessmentID string `json:"deleted_data_protection_impact_assessment_id"`
	}
	mc.CallToolInto("deleteDataProtectionImpactAssessment", map[string]any{
		"id": addResult.DataProtectionImpactAssessment.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.DataProtectionImpactAssessment.ID, deleteResult.DeletedDataProtectionImpactAssessmentID)
}

func TestMCP_TIA_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	paID := factory.CreateProcessingActivity(owner)

	// Create
	var addResult struct {
		TransferImpactAssessment struct {
			ID string `json:"id"`
		} `json:"transfer_impact_assessment"`
	}
	mc.CallToolInto("addTransferImpactAssessment", map[string]any{
		"processing_activity_id": paID,
	}, &addResult)
	require.NotEmpty(t, addResult.TransferImpactAssessment.ID)

	// Get
	var getResult struct {
		TransferImpactAssessment struct {
			ID string `json:"id"`
		} `json:"transfer_impact_assessment"`
	}
	mc.CallToolInto("getTransferImpactAssessment", map[string]any{
		"id": addResult.TransferImpactAssessment.ID,
	}, &getResult)
	assert.Equal(t, addResult.TransferImpactAssessment.ID, getResult.TransferImpactAssessment.ID)

	// Update
	var updateResult struct {
		TransferImpactAssessment struct {
			ID           string `json:"id"`
			DataSubjects string `json:"data_subjects"`
		} `json:"transfer_impact_assessment"`
	}
	mc.CallToolInto("updateTransferImpactAssessment", map[string]any{
		"id":            addResult.TransferImpactAssessment.ID,
		"data_subjects": "EU Residents",
	}, &updateResult)
	assert.Equal(t, "EU Residents", updateResult.TransferImpactAssessment.DataSubjects)

	// List
	var listResult struct {
		TransferImpactAssessments []struct {
			ID string `json:"id"`
		} `json:"transfer_impact_assessments"`
	}
	mc.CallToolInto("listTransferImpactAssessments", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.TransferImpactAssessments)

	// Delete
	var deleteResult struct {
		DeletedTransferImpactAssessmentID string `json:"deleted_transfer_impact_assessment_id"`
	}
	mc.CallToolInto("deleteTransferImpactAssessment", map[string]any{
		"id": addResult.TransferImpactAssessment.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.TransferImpactAssessment.ID, deleteResult.DeletedTransferImpactAssessmentID)
}
