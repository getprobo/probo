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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type mcpDevice struct {
	ID               string  `json:"id"`
	OrganizationID   string  `json:"organization_id"`
	State            string  `json:"state"`
	OwnerID          *string `json:"owner_id"`
	PosturesIncluded bool    `json:"postures_included"`
	LatestPostures   []struct {
		ID       string `json:"id"`
		CheckKey string `json:"check_key"`
		Status   string `json:"status"`
	} `json:"latest_postures"`
}

func createDeviceViaMCP(t *testing.T, mc *testutil.MCPClient, orgID string) string {
	t.Helper()

	var result struct {
		Device          mcpDevice `json:"device"`
		EnrollmentToken string    `json:"enrollment_token"`
		ServerURL       string    `json:"server_url"`
		EnrollmentURL   string    `json:"enrollment_url"`
	}
	mc.CallToolInto("createDevice", map[string]any{
		"organization_id": orgID,
	}, &result)
	require.NotEmpty(t, result.Device.ID)
	assert.Equal(t, "PENDING", result.Device.State)
	assert.NotEmpty(t, result.EnrollmentToken)
	assert.NotEmpty(t, result.ServerURL)
	assert.NotEmpty(t, result.EnrollmentURL)
	assert.False(t, result.Device.PosturesIncluded)
	assert.NotNil(t, result.Device.LatestPostures)
	assert.Empty(t, result.Device.LatestPostures)

	return result.Device.ID
}

func TestMCP_Device_Lifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	deviceID := createDeviceViaMCP(t, mc, orgID)

	// Get
	var getResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("getDevice", map[string]any{
		"id": deviceID,
	}, &getResult)
	assert.Equal(t, deviceID, getResult.Device.ID)
	assert.Equal(t, "PENDING", getResult.Device.State)
	assert.False(t, getResult.Device.PosturesIncluded)
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

			assert.False(t, d.PosturesIncluded)
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

	var clearOwnerResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("setDeviceOwner", map[string]any{
		"id":       deviceID,
		"owner_id": nil,
	}, &clearOwnerResult)
	assert.Nil(t, clearOwnerResult.Device.OwnerID)

	var getAfterClear struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("getDevice", map[string]any{
		"id": deviceID,
	}, &getAfterClear)
	assert.Nil(t, getAfterClear.Device.OwnerID)

	// Revoke
	var revokeResult struct {
		Devices []mcpDevice `json:"devices"`
	}
	mc.CallToolInto("revokeDevices", map[string]any{
		"device_ids": []string{deviceID},
	}, &revokeResult)
	require.Len(t, revokeResult.Devices, 1)
	assert.Equal(t, "REVOKED", revokeResult.Devices[0].State)

	// Delete
	var deleteResult struct {
		DeletedDeviceIDs []string `json:"deleted_device_ids"`
	}
	mc.CallToolInto("deleteDevices", map[string]any{
		"device_ids": []string{deviceID},
	}, &deleteResult)
	assert.Equal(t, []string{deviceID}, deleteResult.DeletedDeviceIDs)

	msg := mc.CallToolExpectToolError("getDevice", map[string]any{
		"id": deviceID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Device_CreateWithOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	profileID := factory.CreateUser(owner)

	var result struct {
		Device          mcpDevice `json:"device"`
		EnrollmentToken string    `json:"enrollment_token"`
	}
	mc.CallToolInto("createDevice", map[string]any{
		"organization_id": orgID,
		"owner_id":        profileID,
	}, &result)
	require.NotEmpty(t, result.Device.ID)
	assert.Equal(t, "PENDING", result.Device.State)
	assert.NotEmpty(t, result.EnrollmentToken)
	require.NotNil(t, result.Device.OwnerID)
	assert.Equal(t, profileID, *result.Device.OwnerID)
}

func TestMCP_Device_CreateWithInvalidOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	msg := mc.CallToolExpectToolError("createDevice", map[string]any{
		"organization_id": orgID,
		"owner_id":        orgID,
	})
	assert.Equal(t, "owner_id must reference a membership profile of the device organization", msg)

	// Each organization lives in its own tenant, so the tenant-scoped profile
	// load fails before the owner organization is ever compared.
	otherOwner := testutil.NewClient(t, testutil.RoleOwner)
	otherProfileID := factory.CreateUser(otherOwner)

	msg = mc.CallToolExpectToolError("createDevice", map[string]any{
		"organization_id": orgID,
		"owner_id":        otherProfileID,
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_Device_CannotDeletePending(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	deviceID := createDeviceViaMCP(t, mc, orgID)

	msg := mc.CallToolExpectToolError("deleteDevices", map[string]any{
		"device_ids": []string{deviceID},
	})
	assert.Equal(t, "device cannot be deleted", msg)
}

func TestMCP_Device_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	deviceID := createDeviceViaMCP(t, mc, orgID)

	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("revokeDevices", map[string]any{
		"device_ids": []string{deviceID},
	})
	assert.Contains(t, msg, "permission denied")

	msg = viewerMC.CallToolExpectToolError("createDevice", map[string]any{
		"organization_id": orgID,
	})
	assert.Contains(t, msg, "permission denied")

	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	employeeMC := testutil.NewMCPClient(t, employee)

	msg = employeeMC.CallToolExpectToolError("createDevice", map[string]any{
		"organization_id": orgID,
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

	deviceID := createDeviceViaMCP(t, ownerMC, orgID)

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
	assert.False(t, getResult.Device.PosturesIncluded)
	assert.NotNil(t, getResult.Device.LatestPostures)
	assert.Empty(t, getResult.Device.LatestPostures)

	msg := employeeMC.CallToolExpectToolError("getDevice", map[string]any{
		"id":               deviceID,
		"include_postures": true,
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_Device_IncludePosturesOnDeviceWithNoChecks(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	deviceID := createDeviceViaMCP(t, mc, orgID)

	var getResult struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("getDevice", map[string]any{
		"id":               deviceID,
		"include_postures": true,
	}, &getResult)
	assert.True(t, getResult.Device.PosturesIncluded)
	assert.NotNil(t, getResult.Device.LatestPostures)
	assert.Empty(t, getResult.Device.LatestPostures)

	var listResult struct {
		Devices []mcpDevice `json:"devices"`
	}
	mc.CallToolInto("listDevices", map[string]any{
		"organization_id":  orgID,
		"include_postures": true,
	}, &listResult)

	found := false

	for _, d := range listResult.Devices {
		if d.ID == deviceID {
			found = true

			assert.True(t, d.PosturesIncluded)
			assert.Empty(t, d.LatestPostures)

			break
		}
	}

	assert.True(t, found, "created device should appear in listDevices")
}

func TestMCP_Device_DownloadEvidence(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	deviceID := createDeviceViaMCP(t, mc, orgID)

	var result struct {
		DownloadURL string `json:"download_url"`
	}
	mc.CallToolInto("downloadDeviceEvidence", map[string]any{
		"device_ids": []string{deviceID},
	}, &result)
	require.NotEmpty(t, result.DownloadURL)

	resp, err := http.Get(result.DownloadURL)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(body), "%PDF"))

	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	employeeMC := testutil.NewMCPClient(t, employee)
	msg := employeeMC.CallToolExpectToolError("downloadDeviceEvidence", map[string]any{
		"device_ids": []string{deviceID},
	})
	assert.Contains(t, msg, "permission denied")
}

func TestMCP_Device_RevokeAndDeleteMany(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()
	firstID := createDeviceViaMCP(t, mc, orgID)
	secondID := createDeviceViaMCP(t, mc, orgID)

	var revokeResult struct {
		Devices []mcpDevice `json:"devices"`
	}
	mc.CallToolInto("revokeDevices", map[string]any{
		"device_ids": []string{firstID, secondID},
	}, &revokeResult)
	require.Len(t, revokeResult.Devices, 2)
	assert.Equal(t, "REVOKED", revokeResult.Devices[0].State)
	assert.Equal(t, "REVOKED", revokeResult.Devices[1].State)

	pendingID := createDeviceViaMCP(t, mc, orgID)
	msg := mc.CallToolExpectToolError("deleteDevices", map[string]any{
		"device_ids": []string{firstID, pendingID},
	})
	assert.Equal(t, "device cannot be deleted", msg)

	var stillThere struct {
		Device mcpDevice `json:"device"`
	}
	mc.CallToolInto("getDevice", map[string]any{"id": firstID}, &stillThere)
	assert.Equal(t, firstID, stillThere.Device.ID)

	var deleteResult struct {
		DeletedDeviceIDs []string `json:"deleted_device_ids"`
	}
	mc.CallToolInto("deleteDevices", map[string]any{
		"device_ids": []string{firstID, secondID},
	}, &deleteResult)
	assert.Equal(t, []string{firstID, secondID}, deleteResult.DeletedDeviceIDs)
}
