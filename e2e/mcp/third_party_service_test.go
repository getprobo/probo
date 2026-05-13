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
