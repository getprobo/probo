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

func TestMCP_Vendor_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Vendor struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vendor"`
	}
	name := factory.SafeName("Vendor")
	mc.CallToolInto("addVendor", map[string]any{
		"organizationId": orgID,
		"name":           name,
	}, &addResult)
	require.NotEmpty(t, addResult.Vendor.ID)
	assert.Equal(t, name, addResult.Vendor.Name)

	// Update
	var updateResult struct {
		Vendor struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vendor"`
	}
	mc.CallToolInto("updateVendor", map[string]any{
		"id":   addResult.Vendor.ID,
		"name": "Updated Vendor",
	}, &updateResult)
	assert.Equal(t, "Updated Vendor", updateResult.Vendor.Name)

	// List
	var listResult struct {
		Vendors []struct {
			ID string `json:"id"`
		} `json:"vendors"`
	}
	mc.CallToolInto("listVendors", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Vendors)

	// Delete
	var deleteResult struct {
		DeletedVendorID string `json:"deletedVendorId"`
	}
	mc.CallToolInto("deleteVendor", map[string]any{
		"id": addResult.Vendor.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.Vendor.ID, deleteResult.DeletedVendorID)
}
