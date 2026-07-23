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

func TestMCP_BusinessFunction_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var addResult struct {
		BusinessFunction struct {
			ID             string `json:"id"`
			ReferenceID    string `json:"reference_id"`
			Name           string `json:"name"`
			Classification string `json:"classification"`
		} `json:"business_function"`
	}
	mc.CallToolInto("addBusinessFunction", map[string]any{
		"organization_id": orgID,
		"reference_id":    "F-10",
		"name":            factory.SafeName("SPV Payment Execution"),
		"classification":  "CRITICAL",
		"mtd_minutes":     1440,
		"rto_minutes":     720,
		"rpo_minutes":     60,
	}, &addResult)
	require.NotEmpty(t, addResult.BusinessFunction.ID)
	assert.Equal(t, "F-10", addResult.BusinessFunction.ReferenceID)
	assert.Equal(t, "CRITICAL", addResult.BusinessFunction.Classification)

	var getResult struct {
		BusinessFunction struct {
			ID string `json:"id"`
		} `json:"business_function"`
	}
	mc.CallToolInto("getBusinessFunction", map[string]any{
		"id": addResult.BusinessFunction.ID,
	}, &getResult)
	assert.Equal(t, addResult.BusinessFunction.ID, getResult.BusinessFunction.ID)

	var updateResult struct {
		BusinessFunction struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"business_function"`
	}
	mc.CallToolInto("updateBusinessFunction", map[string]any{
		"id":   addResult.BusinessFunction.ID,
		"name": "Updated Business Function",
	}, &updateResult)
	assert.Equal(t, "Updated Business Function", updateResult.BusinessFunction.Name)

	var listResult struct {
		BusinessFunctions []struct {
			ID string `json:"id"`
		} `json:"business_functions"`
	}
	mc.CallToolInto("listBusinessFunctions", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	require.NotEmpty(t, listResult.BusinessFunctions)

	var deleteResult struct {
		DeletedBusinessFunctionID string `json:"deleted_business_function_id"`
	}
	mc.CallToolInto("deleteBusinessFunction", map[string]any{
		"id": addResult.BusinessFunction.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.BusinessFunction.ID, deleteResult.DeletedBusinessFunctionID)
}
