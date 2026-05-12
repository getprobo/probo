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
		} `json:"dataProtectionImpactAssessment"`
	}
	mc.CallToolInto("addDataProtectionImpactAssessment", map[string]any{
		"processingActivityId": paID,
	}, &addResult)
	require.NotEmpty(t, addResult.DataProtectionImpactAssessment.ID)

	// Get
	var getResult struct {
		DataProtectionImpactAssessment struct {
			ID string `json:"id"`
		} `json:"dataProtectionImpactAssessment"`
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
		} `json:"dataProtectionImpactAssessment"`
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
		} `json:"dataProtectionImpactAssessments"`
	}
	mc.CallToolInto("listDataProtectionImpactAssessments", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.DataProtectionImpactAssessments)

	// Delete
	var deleteResult struct {
		DeletedDataProtectionImpactAssessmentID string `json:"deletedDataProtectionImpactAssessmentId"`
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
		} `json:"transferImpactAssessment"`
	}
	mc.CallToolInto("addTransferImpactAssessment", map[string]any{
		"processingActivityId": paID,
	}, &addResult)
	require.NotEmpty(t, addResult.TransferImpactAssessment.ID)

	// Get
	var getResult struct {
		TransferImpactAssessment struct {
			ID string `json:"id"`
		} `json:"transferImpactAssessment"`
	}
	mc.CallToolInto("getTransferImpactAssessment", map[string]any{
		"id": addResult.TransferImpactAssessment.ID,
	}, &getResult)
	assert.Equal(t, addResult.TransferImpactAssessment.ID, getResult.TransferImpactAssessment.ID)

	// Update
	var updateResult struct {
		TransferImpactAssessment struct {
			ID           string `json:"id"`
			DataSubjects string `json:"dataSubjects"`
		} `json:"transferImpactAssessment"`
	}
	mc.CallToolInto("updateTransferImpactAssessment", map[string]any{
		"id":           addResult.TransferImpactAssessment.ID,
		"dataSubjects": "EU Residents",
	}, &updateResult)
	assert.Equal(t, "EU Residents", updateResult.TransferImpactAssessment.DataSubjects)

	// List
	var listResult struct {
		TransferImpactAssessments []struct {
			ID string `json:"id"`
		} `json:"transferImpactAssessments"`
	}
	mc.CallToolInto("listTransferImpactAssessments", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.TransferImpactAssessments)

	// Delete
	var deleteResult struct {
		DeletedTransferImpactAssessmentID string `json:"deletedTransferImpactAssessmentId"`
	}
	mc.CallToolInto("deleteTransferImpactAssessment", map[string]any{
		"id": addResult.TransferImpactAssessment.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.TransferImpactAssessment.ID, deleteResult.DeletedTransferImpactAssessmentID)
}
