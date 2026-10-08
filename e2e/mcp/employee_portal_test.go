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
	"go.probo.inc/probo/e2e/internal/testutil"
)

type employeePortal struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Capabilities struct {
		DeviceAgent bool `json:"device_agent"`
	} `json:"capabilities"`
}

func TestMCP_CreateGetUpdateListDeleteEmployeePortal(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	var createResult struct {
		EmployeePortal employeePortal `json:"employee_portal"`
	}
	mc.CallToolInto("createEmployeePortal", map[string]any{
		"organization_id": owner.GetOrganizationID().String(),
		"name":            "MCP Employee Portal",
	}, &createResult)

	require.NotEmpty(t, createResult.EmployeePortal.ID)
	assert.Equal(t, "MCP Employee Portal", createResult.EmployeePortal.Name)
	assert.True(t, createResult.EmployeePortal.Active)
	assert.True(t, createResult.EmployeePortal.Capabilities.DeviceAgent)

	var getResult struct {
		EmployeePortal employeePortal `json:"employee_portal"`
	}
	mc.CallToolInto("getEmployeePortal", map[string]any{
		"employee_portal_id": createResult.EmployeePortal.ID,
	}, &getResult)

	assert.Equal(t, createResult.EmployeePortal.ID, getResult.EmployeePortal.ID)
	assert.Equal(t, "MCP Employee Portal", getResult.EmployeePortal.Name)

	var updateResult struct {
		EmployeePortal employeePortal `json:"employee_portal"`
	}
	mc.CallToolInto("updateEmployeePortal", map[string]any{
		"employee_portal_id": createResult.EmployeePortal.ID,
		"name":               "Updated MCP Portal",
		"capabilities": map[string]any{
			"device_agent": false,
		},
	}, &updateResult)

	assert.Equal(t, createResult.EmployeePortal.ID, updateResult.EmployeePortal.ID)
	assert.Equal(t, "Updated MCP Portal", updateResult.EmployeePortal.Name)
	assert.False(t, updateResult.EmployeePortal.Capabilities.DeviceAgent)

	var listResult struct {
		EmployeePortals []employeePortal `json:"employee_portals"`
	}
	mc.CallToolInto("listEmployeePortals", map[string]any{
		"organization_id": owner.GetOrganizationID().String(),
	}, &listResult)

	found := false

	for _, portal := range listResult.EmployeePortals {
		if portal.ID == createResult.EmployeePortal.ID {
			found = true

			assert.Equal(t, "Updated MCP Portal", portal.Name)

			break
		}
	}

	assert.True(t, found, "created employee portal should appear in list")

	var deleteResult struct {
		DeletedEmployeePortalID string `json:"deleted_employee_portal_id"`
	}
	mc.CallToolInto("deleteEmployeePortal", map[string]any{
		"employee_portal_id": createResult.EmployeePortal.ID,
	}, &deleteResult)

	assert.Equal(t, createResult.EmployeePortal.ID, deleteResult.DeletedEmployeePortalID)
}
