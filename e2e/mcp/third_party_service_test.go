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

type thirdPartyService struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func TestMCP_AddThirdPartyService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	var result struct {
		ThirdPartyService thirdPartyService `json:"thirdPartyService"`
	}
	mc.CallToolInto("addThirdPartyService", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Cloud Storage",
		"description":  "Object storage service",
	}, &result)

	assert.NotEmpty(t, result.ThirdPartyService.ID)
	assert.Equal(t, "Cloud Storage", result.ThirdPartyService.Name)
}

func TestMCP_UpdateThirdPartyService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyService thirdPartyService `json:"thirdPartyService"`
	}
	mc.CallToolInto("addThirdPartyService", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Original Service",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdPartyService.ID)

	// Update
	var updateResult struct {
		ThirdPartyService thirdPartyService `json:"thirdPartyService"`
	}
	mc.CallToolInto("updateThirdPartyService", map[string]any{
		"id":   addResult.ThirdPartyService.ID,
		"name": "Updated Service",
	}, &updateResult)

	assert.Equal(t, addResult.ThirdPartyService.ID, updateResult.ThirdPartyService.ID)
	assert.Equal(t, "Updated Service", updateResult.ThirdPartyService.Name)
}

func TestMCP_DeleteThirdPartyService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create
	var addResult struct {
		ThirdPartyService thirdPartyService `json:"thirdPartyService"`
	}
	mc.CallToolInto("addThirdPartyService", map[string]any{
		"thirdPartyId": thirdPartyID,
		"name":         "Service to delete",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdPartyService.ID)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
	}
	mc.CallToolInto("deleteThirdPartyService", map[string]any{
		"id": addResult.ThirdPartyService.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.ThirdPartyService.ID, deleteResult.DeletedThirdPartyServiceID)
}

func TestMCP_ListThirdPartyServices(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	// Create services
	for i := range 3 {
		var result struct {
			ThirdPartyService thirdPartyService `json:"thirdPartyService"`
		}
		mc.CallToolInto("addThirdPartyService", map[string]any{
			"thirdPartyId": thirdPartyID,
			"name":         factory.SafeName("Service"),
		}, &result)
		require.NotEmpty(t, result.ThirdPartyService.ID)

		_ = i
	}

	// List
	var listResult struct {
		ThirdPartyServices []thirdPartyService `json:"thirdPartyServices"`
	}
	mc.CallToolInto("listThirdPartyServices", map[string]any{
		"thirdPartyId": thirdPartyID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.ThirdPartyServices), 3)
}
