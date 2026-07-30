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

const createDeviceMutation = `
mutation CreateDevice($input: CreateDeviceInput!) {
	createDevice(input: $input) {
		device { id state }
	}
}
`

type mcpDevice struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	State          string  `json:"state"`
	OwnerID        *string `json:"owner_id"`
	LatestPostures []struct {
		ID       string `json:"id"`
		CheckKey string `json:"check_key"`
		Status   string `json:"status"`
	} `json:"latest_postures"`
}

func createDeviceViaGraphQL(t *testing.T, owner *testutil.Client, orgID string) string {
	t.Helper()

	var result struct {
		CreateDevice struct {
			Device struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"device"`
		} `json:"createDevice"`
	}
	owner.MustExecute(createDeviceMutation, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
		},
	}, &result)
	require.NotEmpty(t, result.CreateDevice.Device.ID)

	return result.CreateDevice.Device.ID
}

func TestMCP_Device_Lifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	deviceID := createDeviceViaGraphQL(t, owner, orgID)

	// Get
	var getResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("getDevice", map[string]any{
		"id": deviceID,
	}, &getResult)
	assert.Equal(t, deviceID, getResult.Device.ID)
	assert.Equal(t, "PENDING", getResult.Device.State)
	assert.NotNil(t, getResult.Device.LatestPostures)
	assert.Empty(t, getResult.Device.LatestPostures)

	// List
	var listResult struct {
		Devices []mcpDevice `json:"devices"`
	}
	mc.CallToolInto("listDevices", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	require.NotEmpty(t, listResult.Devices)

	found := false

	for _, d := range listResult.Devices {
		if d.ID == deviceID {
			found = true

			assert.NotNil(t, d.LatestPostures)

			break
		}
	}

	assert.True(t, found, "created device should appear in listDevices")

	// Set owner
	var setOwnerResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("setDeviceOwner", map[string]any{
		"id":       deviceID,
		"owner_id": profileID,
	}, &setOwnerResult)
	require.NotNil(t, setOwnerResult.Device.OwnerID)
	assert.Equal(t, profileID, *setOwnerResult.Device.OwnerID)

	// Clear owner
	mc.CallToolInto("setDeviceOwner", map[string]any{
		"id":       deviceID,
		"owner_id": nil,
	}, &setOwnerResult)
	assert.Nil(t, setOwnerResult.Device.OwnerID)

	// Revoke
	var revokeResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("revokeDevice", map[string]any{
		"id": deviceID,
	}, &revokeResult)
	assert.Equal(t, "REVOKED", revokeResult.Device.State)

	// Delete
	var deleteResult struct {
		DeletedDeviceID string `json:"deleted_device_id"`
	}
	mc.CallToolInto("deleteDevice", map[string]any{
		"id": deviceID,
	}, &deleteResult)
	assert.Equal(t, deviceID, deleteResult.DeletedDeviceID)

	msg := mc.CallToolExpectToolError("getDevice", map[string]any{
		"id": deviceID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Device_CannotDeletePending(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	deviceID := createDeviceViaGraphQL(t, owner, orgID)

	msg := mc.CallToolExpectToolError("deleteDevice", map[string]any{
		"id": deviceID,
	})
	assert.Equal(t, "device cannot be deleted", msg)
}

func TestMCP_Device_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	deviceID := createDeviceViaGraphQL(t, owner, orgID)

	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("revokeDevice", map[string]any{
		"id": deviceID,
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_Device_EmployeeCannotIncludePostures(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	ownerMC := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	employeeMC := testutil.NewMCPClient(t, employee)
	employeeProfileID := employee.GetProfileID().String()

	deviceID := createDeviceViaGraphQL(t, owner, orgID)

	var setOwnerResult struct {
		Device mcpDevice `json:"device"`
	}
	ownerMC.CallToolInto("setDeviceOwner", map[string]any{
		"id":       deviceID,
		"owner_id": employeeProfileID,
	}, &setOwnerResult)
	require.NotNil(t, setOwnerResult.Device.OwnerID)
	assert.Equal(t, employeeProfileID, *setOwnerResult.Device.OwnerID)

	var getResult struct {
		Device mcpDevice `json:"device"`
	}
	employeeMC.CallToolInto("getDevice", map[string]any{
		"id": deviceID,
	}, &getResult)
	assert.Equal(t, deviceID, getResult.Device.ID)
	assert.NotNil(t, getResult.Device.LatestPostures)
	assert.Empty(t, getResult.Device.LatestPostures)

	msg := employeeMC.CallToolExpectToolError("getDevice", map[string]any{
		"id":               deviceID,
		"include_postures": true,
	})
	assert.Contains(t, msg, "permission denied")
}
