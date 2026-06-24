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
