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

func TestMCP_AiSystem_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var addResult struct {
		AiSystem struct {
			ID                 string `json:"id"`
			Name               string `json:"name"`
			Status             string `json:"status"`
			RiskClassification string `json:"risk_classification"`
		} `json:"ai_system"`
	}
	mc.CallToolInto("addAiSystem", map[string]any{
		"organization_id":     orgID,
		"name":                factory.SafeName("Document Classification Model"),
		"status":              "ACTIVE",
		"risk_classification": "LIMITED",
		"company_roles":       []string{"PROVIDER", "USER"},
	}, &addResult)
	require.NotEmpty(t, addResult.AiSystem.ID)
	assert.Equal(t, "ACTIVE", addResult.AiSystem.Status)
	assert.Equal(t, "LIMITED", addResult.AiSystem.RiskClassification)

	var getResult struct {
		AiSystem struct {
			ID string `json:"id"`
		} `json:"ai_system"`
	}
	mc.CallToolInto("getAiSystem", map[string]any{
		"id": addResult.AiSystem.ID,
	}, &getResult)
	assert.Equal(t, addResult.AiSystem.ID, getResult.AiSystem.ID)

	var updateResult struct {
		AiSystem struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"ai_system"`
	}
	mc.CallToolInto("updateAiSystem", map[string]any{
		"id":   addResult.AiSystem.ID,
		"name": "Updated AI System",
	}, &updateResult)
	assert.Equal(t, "Updated AI System", updateResult.AiSystem.Name)

	var listResult struct {
		AiSystems []struct {
			ID string `json:"id"`
		} `json:"ai_systems"`
	}
	mc.CallToolInto("listAiSystems", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	require.NotEmpty(t, listResult.AiSystems)

	var deleteResult struct {
		DeletedAiSystemID string `json:"deleted_ai_system_id"`
	}
	mc.CallToolInto("deleteAiSystem", map[string]any{
		"id": addResult.AiSystem.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.AiSystem.ID, deleteResult.DeletedAiSystemID)
}
