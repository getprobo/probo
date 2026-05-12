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

type vendorContact struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
	Role  *string `json:"role"`
}

func TestMCP_AddVendorContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	var result struct {
		VendorContact vendorContact `json:"vendorContact"`
	}
	mc.CallToolInto("addVendorContact", map[string]any{
		"vendorId": vendorID,
		"name":     "Alice Smith",
		"email":    "alice@example.com",
		"phone":    "+1-555-0100",
		"role":     "Account Manager",
	}, &result)

	assert.NotEmpty(t, result.VendorContact.ID)
	assert.Equal(t, "Alice Smith", result.VendorContact.Name)
}

func TestMCP_UpdateVendorContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create
	var addResult struct {
		VendorContact vendorContact `json:"vendorContact"`
	}
	mc.CallToolInto("addVendorContact", map[string]any{
		"vendorId": vendorID,
		"name":     "Bob Jones",
		"email":    "bob@example.com",
	}, &addResult)
	require.NotEmpty(t, addResult.VendorContact.ID)

	// Update
	var updateResult struct {
		VendorContact vendorContact `json:"vendorContact"`
	}
	mc.CallToolInto("updateVendorContact", map[string]any{
		"id":   addResult.VendorContact.ID,
		"name": "Robert Jones",
		"role": "CTO",
	}, &updateResult)

	assert.Equal(t, addResult.VendorContact.ID, updateResult.VendorContact.ID)
	assert.Equal(t, "Robert Jones", updateResult.VendorContact.Name)
}

func TestMCP_DeleteVendorContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create
	var addResult struct {
		VendorContact vendorContact `json:"vendorContact"`
	}
	mc.CallToolInto("addVendorContact", map[string]any{
		"vendorId": vendorID,
		"name":     "Contact to delete",
	}, &addResult)
	require.NotEmpty(t, addResult.VendorContact.ID)

	// Delete
	var deleteResult struct {
		DeletedVendorContactID string `json:"deletedVendorContactId"`
	}
	mc.CallToolInto("deleteVendorContact", map[string]any{
		"id": addResult.VendorContact.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.VendorContact.ID, deleteResult.DeletedVendorContactID)
}

func TestMCP_ListVendorContacts(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	vendorID := factory.CreateVendor(owner)

	// Create contacts
	for i := range 3 {
		var result struct {
			VendorContact vendorContact `json:"vendorContact"`
		}
		mc.CallToolInto("addVendorContact", map[string]any{
			"vendorId": vendorID,
			"name":     factory.SafeName("Contact"),
			"email":    factory.SafeEmail(),
		}, &result)
		require.NotEmpty(t, result.VendorContact.ID)
		_ = i
	}

	// List
	var listResult struct {
		VendorContacts []vendorContact `json:"vendorContacts"`
	}
	mc.CallToolInto("listVendorContacts", map[string]any{
		"vendorId": vendorID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.VendorContacts), 3)
}
