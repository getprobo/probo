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

func TestMCP_Measure_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Measure struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"measure"`
	}
	mc.CallToolInto("addMeasure", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Measure"),
	}, &addResult)
	require.NotEmpty(t, addResult.Measure.ID)

	// Get
	var getResult struct {
		Measure struct {
			ID string `json:"id"`
		} `json:"measure"`
	}
	mc.CallToolInto("getMeasure", map[string]any{
		"id": addResult.Measure.ID,
	}, &getResult)
	assert.Equal(t, addResult.Measure.ID, getResult.Measure.ID)

	// Update
	var updateResult struct {
		Measure struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"measure"`
	}
	mc.CallToolInto("updateMeasure", map[string]any{
		"id":   addResult.Measure.ID,
		"name": "Updated Measure",
	}, &updateResult)
	assert.Equal(t, "Updated Measure", updateResult.Measure.Name)

	// List
	var listResult struct {
		Measures []struct {
			ID string `json:"id"`
		} `json:"measures"`
	}
	mc.CallToolInto("listMeasures", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Measures)

	// Sub-resources (empty lists are fine, just verify the tools work)
	var risksResult struct {
		Risks []struct{ ID string } `json:"risks"`
	}
	mc.CallToolInto("listMeasureRisks", map[string]any{
		"measureId": addResult.Measure.ID,
	}, &risksResult)

	var controlsResult struct {
		Controls []struct{ ID string } `json:"controls"`
	}
	mc.CallToolInto("listMeasureControls", map[string]any{
		"measureId": addResult.Measure.ID,
	}, &controlsResult)

	var tasksResult struct {
		Tasks []struct{ ID string } `json:"tasks"`
	}
	mc.CallToolInto("listMeasureTasks", map[string]any{
		"measureId": addResult.Measure.ID,
	}, &tasksResult)

	// Delete
	var deleteResult struct {
		DeletedMeasureID string `json:"deletedMeasureId"`
	}
	mc.CallToolInto("deleteMeasure", map[string]any{
		"id": addResult.Measure.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Measure.ID, deleteResult.DeletedMeasureID)
}
