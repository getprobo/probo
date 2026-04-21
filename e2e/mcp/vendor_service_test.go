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

type vendorService struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func TestMCP_AddVendorService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	var result struct {
		VendorService vendorService `json:"vendorService"`
	}
	mc.CallToolInto("addVendorService", map[string]any{
		"vendorId":    vendorID,
		"name":        "Cloud Storage",
		"description": "Object storage service",
	}, &result)

	assert.NotEmpty(t, result.VendorService.ID)
	assert.Equal(t, "Cloud Storage", result.VendorService.Name)
}

func TestMCP_UpdateVendorService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create
	var addResult struct {
		VendorService vendorService `json:"vendorService"`
	}
	mc.CallToolInto("addVendorService", map[string]any{
		"vendorId": vendorID,
		"name":     "Original Service",
	}, &addResult)
	require.NotEmpty(t, addResult.VendorService.ID)

	// Update
	var updateResult struct {
		VendorService vendorService `json:"vendorService"`
	}
	mc.CallToolInto("updateVendorService", map[string]any{
		"id":   addResult.VendorService.ID,
		"name": "Updated Service",
	}, &updateResult)

	assert.Equal(t, addResult.VendorService.ID, updateResult.VendorService.ID)
	assert.Equal(t, "Updated Service", updateResult.VendorService.Name)
}

func TestMCP_DeleteVendorService(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create
	var addResult struct {
		VendorService vendorService `json:"vendorService"`
	}
	mc.CallToolInto("addVendorService", map[string]any{
		"vendorId": vendorID,
		"name":     "Service to delete",
	}, &addResult)
	require.NotEmpty(t, addResult.VendorService.ID)

	// Delete
	var deleteResult struct {
		DeletedVendorServiceID string `json:"deletedVendorServiceId"`
	}
	mc.CallToolInto("deleteVendorService", map[string]any{
		"id": addResult.VendorService.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.VendorService.ID, deleteResult.DeletedVendorServiceID)
}

func TestMCP_ListVendorServices(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create services
	for i := range 3 {
		var result struct {
			VendorService vendorService `json:"vendorService"`
		}
		mc.CallToolInto("addVendorService", map[string]any{
			"vendorId": vendorID,
			"name":     factory.SafeName("Service"),
		}, &result)
		require.NotEmpty(t, result.VendorService.ID)
		_ = i
	}

	// List
	var listResult struct {
		VendorServices []vendorService `json:"vendorServices"`
	}
	mc.CallToolInto("listVendorServices", map[string]any{
		"vendorId": vendorID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.VendorServices), 3)
}
