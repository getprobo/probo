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

type employeePortal struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Capabilities struct {
		DeviceAgent bool `json:"device_agent"`
	} `json:"capabilities"`
}

func TestMCP_GetUpdateListEmployeePortal(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	portalID := factory.DefaultEmployeePortalID(owner)

	var getResult struct {
		EmployeePortal employeePortal `json:"employee_portal"`
	}
	mc.CallToolInto("getEmployeePortal", map[string]any{
		"employee_portal_id": portalID,
	}, &getResult)

	assert.Equal(t, portalID, getResult.EmployeePortal.ID)
	assert.NotEmpty(t, getResult.EmployeePortal.Name)
	assert.True(t, getResult.EmployeePortal.Active)
	assert.True(t, getResult.EmployeePortal.Capabilities.DeviceAgent)

	var updateResult struct {
		EmployeePortal employeePortal `json:"employee_portal"`
	}
	mc.CallToolInto("updateEmployeePortal", map[string]any{
		"employee_portal_id": portalID,
		"name":               "Updated MCP Portal",
		"capabilities": map[string]any{
			"device_agent": false,
		},
	}, &updateResult)

	assert.Equal(t, portalID, updateResult.EmployeePortal.ID)
	assert.Equal(t, "Updated MCP Portal", updateResult.EmployeePortal.Name)
	assert.False(t, updateResult.EmployeePortal.Capabilities.DeviceAgent)

	var listResult struct {
		EmployeePortals []employeePortal `json:"employee_portals"`
	}
	mc.CallToolInto("listEmployeePortals", map[string]any{
		"organization_id": owner.GetOrganizationID().String(),
	}, &listResult)

	require.Len(t, listResult.EmployeePortals, 1)
	assert.Equal(t, portalID, listResult.EmployeePortals[0].ID)
	assert.Equal(t, "Updated MCP Portal", listResult.EmployeePortals[0].Name)
}
