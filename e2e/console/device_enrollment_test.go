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

package console_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	enrollDeviceMutation = `
		mutation EnrollDevice($input: EnrollDeviceInput!) {
			enrollDevice(input: $input) {
				enrollmentToken
				serverUrl
				enrollmentUrl
				device { id }
			}
		}`

	createDeviceMutation = `
		mutation CreateDevice($input: CreateDeviceInput!) {
			createDevice(input: $input) {
				enrollmentToken
				serverUrl
				enrollmentUrl
				device { id }
			}
		}`

	revokeDeviceMutation = `
		mutation RevokeDevice($input: RevokeDeviceInput!) {
			revokeDevice(input: $input) {
				device { id state }
			}
		}`

	devicePermissionQuery = `
		query DevicePermission($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					canEnrollDevice: permission(action: "itam:device:enroll")
				}
			}
		}`

	getDeviceQuery = `
		query GetDevice($id: ID!) {
			node(id: $id) {
				... on Device {
					id
					state
					owner {
						id
						fullName
					}
				}
			}
		}`

	listDevicesQuery = `
		query ListDevices($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					devices(first: 1) {
						totalCount
					}
				}
			}
		}`

	listEnrolledDevicesQuery = `
		query ListEnrolledDevices($orgId: ID!) {
			viewer {
				enrolledDevices(organizationId: $orgId, first: 100) {
					edges {
						node {
							id
							state
						}
					}
				}
			}
		}`

	getEnrolledDeviceQuery = `
		query GetEnrolledDevice($id: ID!) {
			viewer {
				enrolledDevice(id: $id) {
					id
					state
					hostname
				}
			}
		}`
)

type enrollDeviceResult struct {
	EnrollDevice struct {
		EnrollmentToken string `json:"enrollmentToken"`
		ServerURL       string `json:"serverUrl"`
		EnrollmentURL   string `json:"enrollmentUrl"`
		Device          struct {
			ID string `json:"id"`
		} `json:"device"`
	} `json:"enrollDevice"`
}

type createDeviceResult struct {
	CreateDevice struct {
		EnrollmentToken string `json:"enrollmentToken"`
		ServerURL       string `json:"serverUrl"`
		EnrollmentURL   string `json:"enrollmentUrl"`
		Device          struct {
			ID string `json:"id"`
		} `json:"device"`
	} `json:"createDevice"`
}

type enrollAPIResponse struct {
	APIKey string `json:"api_key"`
}

func exchangeEnrollmentToken(t *testing.T, token string) (int, enrollAPIResponse) {
	t.Helper()

	body, err := json.Marshal(map[string]string{"token": token})
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/agent/v1/enroll",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload enrollAPIResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.Unmarshal(raw, &payload))
	}

	return resp.StatusCode, payload
}

func assertEnrollmentURLs(t *testing.T, serverURL, enrollmentURL, enrollmentToken string) {
	t.Helper()

	require.Equal(t, testutil.GetBaseURL(), serverURL)

	parsed, err := url.Parse(enrollmentURL)
	require.NoError(t, err)
	require.Equal(t, "probo", parsed.Scheme)
	require.Equal(t, "enroll", parsed.Host)
	require.Equal(t, serverURL, parsed.Query().Get("server"))
	require.Equal(t, enrollmentToken, parsed.Query().Get("token"))
}

func enrollDevice(t *testing.T, client *testutil.Client, organizationID string) enrollDeviceResult {
	t.Helper()

	var result enrollDeviceResult
	client.MustExecute(enrollDeviceMutation, map[string]any{
		"input": map[string]any{
			"organizationId": organizationID,
		},
	}, &result)
	require.NotEmpty(t, result.EnrollDevice.EnrollmentToken)
	require.NotEmpty(t, result.EnrollDevice.Device.ID)
	assertEnrollmentURLs(
		t,
		result.EnrollDevice.ServerURL,
		result.EnrollDevice.EnrollmentURL,
		result.EnrollDevice.EnrollmentToken,
	)

	return result
}

func activateEnrolledDevice(t *testing.T, enrollmentToken, hardwareUUID string) {
	t.Helper()

	status, payload := exchangeEnrollmentToken(t, enrollmentToken)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, payload.APIKey)

	body, err := json.Marshal(map[string]any{
		"hardware_uuid": hardwareUUID,
		"hostname":      "e2e-host",
		"platform":      "DARWIN",
		"os_version":    "14.0",
		"agent_version": "1.0.0",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/agent/v1/heartbeat",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", payload.APIKey))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendHeartbeat(t *testing.T, apiKey, hardwareUUID string) int {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"hardware_uuid": hardwareUUID,
		"hostname":      "e2e-host",
		"platform":      "DARWIN",
		"os_version":    "14.0",
		"agent_version": "1.0.0",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/agent/v1/heartbeat",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func enrollAndActivateDevice(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
) enrollDeviceResult {
	t.Helper()

	enrolled := enrollDevice(t, client, organizationID)
	activateEnrolledDevice(
		t,
		enrolled.EnrollDevice.EnrollmentToken,
		enrolled.EnrollDevice.Device.ID+"-hw",
	)

	return enrolled
}

func createDevice(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	ownerProfileID *string,
) createDeviceResult {
	t.Helper()

	input := map[string]any{
		"organizationId": organizationID,
	}
	if ownerProfileID != nil {
		input["ownerId"] = *ownerProfileID
	}

	var result createDeviceResult
	client.MustExecute(createDeviceMutation, map[string]any{"input": input}, &result)
	require.NotEmpty(t, result.CreateDevice.EnrollmentToken)
	require.NotEmpty(t, result.CreateDevice.Device.ID)
	assertEnrollmentURLs(
		t,
		result.CreateDevice.ServerURL,
		result.CreateDevice.EnrollmentURL,
		result.CreateDevice.EnrollmentToken,
	)

	return result
}

func listEnrolledDeviceIDs(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
) []string {
	t.Helper()

	var result struct {
		Viewer struct {
			EnrolledDevices struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"enrolledDevices"`
		} `json:"viewer"`
	}
	client.MustExecute(listEnrolledDevicesQuery, map[string]any{"orgId": organizationID}, &result)

	ids := make([]string, len(result.Viewer.EnrolledDevices.Edges))
	for i, edge := range result.Viewer.EnrolledDevices.Edges {
		ids[i] = edge.Node.ID
	}

	return ids
}

func setupDeviceEnrollmentClients(t *testing.T) (
	owner, admin, employee, viewer *testutil.Client,
	orgID, ownerProfileID string,
) {
	t.Helper()

	owner = testutil.NewClient(t, testutil.RoleOwner)
	admin = testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	employee = testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	viewer = testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID = owner.GetOrganizationID().String()
	ownerProfileID = owner.GetProfileID().String()

	return owner, admin, employee, viewer, orgID, ownerProfileID
}

func TestDeviceEnrollment(t *testing.T) {
	t.Parallel()

	t.Run("enrollment token can be exchanged once", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)

		status, payload := exchangeEnrollmentToken(t, enrolled.EnrollDevice.EnrollmentToken)
		require.Equal(t, http.StatusOK, status)
		require.NotEmpty(t, payload.APIKey)

		replayStatus, _ := exchangeEnrollmentToken(t, enrolled.EnrollDevice.EnrollmentToken)
		require.Equal(t, http.StatusUnauthorized, replayStatus)
	})

	t.Run("revoked device enrollment token returns unauthorized", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)

		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": enrolled.EnrollDevice.Device.ID,
			},
		}, &struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}{})

		status, _ := exchangeEnrollmentToken(t, enrolled.EnrollDevice.EnrollmentToken)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("revoked device API key is rejected on heartbeat", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)
		status, payload := exchangeEnrollmentToken(t, enrolled.EnrollDevice.EnrollmentToken)
		require.Equal(t, http.StatusOK, status)
		require.NotEmpty(t, payload.APIKey)

		deviceID := enrolled.EnrollDevice.Device.ID
		require.Equal(t, http.StatusOK, sendHeartbeat(t, payload.APIKey, deviceID+"-hw"))

		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": deviceID,
			},
		}, &struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}{})

		require.Equal(
			t,
			http.StatusUnauthorized,
			sendHeartbeat(t, payload.APIKey, deviceID+"-hw"),
		)
	})

	t.Run("re-enrollment succeeds after revoke with same hardware UUID", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollAndActivateDevice(t, employee, orgID)
		hardwareUUID := enrolled.EnrollDevice.Device.ID + "-hw"

		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": enrolled.EnrollDevice.Device.ID,
			},
		}, &struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}{})

		reEnrolled := enrollDevice(t, employee, orgID)
		activateEnrolledDevice(t, reEnrolled.EnrollDevice.EnrollmentToken, hardwareUUID)

		var result struct {
			Node struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"node"`
		}
		employee.MustExecute(getDeviceQuery, map[string]any{"id": reEnrolled.EnrollDevice.Device.ID}, &result)
		require.Equal(t, reEnrolled.EnrollDevice.Device.ID, result.Node.ID)
		require.Equal(t, "ACTIVE", result.Node.State)
	})

	t.Run("owner can enroll device", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrollDevice(t, owner, orgID)
	})

	t.Run("admin can enroll device", func(t *testing.T) {
		t.Parallel()

		_, admin, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrollDevice(t, admin, orgID)
	})

	t.Run("employee can enroll device", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrollDevice(t, employee, orgID)
	})

	t.Run("employee permission gate", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		var result struct {
			Node struct {
				CanEnrollDevice bool `json:"canEnrollDevice"`
			} `json:"node"`
		}
		employee.MustExecute(devicePermissionQuery, map[string]any{"orgId": orgID}, &result)
		require.True(t, result.Node.CanEnrollDevice)
	})

	t.Run("employee can read own device", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)

		var result struct {
			Node struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"node"`
		}
		employee.MustExecute(getDeviceQuery, map[string]any{"id": enrolled.EnrollDevice.Device.ID}, &result)
		require.Equal(t, enrolled.EnrollDevice.Device.ID, result.Node.ID)
		require.Equal(t, "PENDING", result.Node.State)
	})

	t.Run("employee cannot read another users device via node", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		employeeA := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		employeeB := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		enrolledB := enrollDevice(t, employeeB, orgID)

		_, err := employeeA.Do(getDeviceQuery, map[string]any{
			"id": enrolledB.EnrollDevice.Device.ID,
		})
		testutil.RequireForbiddenError(t, err, "employee cannot read another users device via node")
	})

	t.Run("employee cannot list org devices", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		_, err := employee.Do(listDevicesQuery, map[string]any{"orgId": orgID})
		testutil.RequireForbiddenError(t, err, "employee should not list org devices")
	})

	t.Run("employee can list own enrolled devices", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollAndActivateDevice(t, employee, orgID)

		var result struct {
			Viewer struct {
				EnrolledDevices struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"enrolledDevices"`
			} `json:"viewer"`
		}
		employee.MustExecute(listEnrolledDevicesQuery, map[string]any{"orgId": orgID}, &result)

		require.Len(t, result.Viewer.EnrolledDevices.Edges, 1)
		require.Equal(t, enrolled.EnrollDevice.Device.ID, result.Viewer.EnrolledDevices.Edges[0].Node.ID)
		require.Equal(t, "ACTIVE", result.Viewer.EnrolledDevices.Edges[0].Node.State)
	})

	t.Run("employee enrolled devices exclude pending devices", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)

		ids := listEnrolledDeviceIDs(t, employee, orgID)
		require.NotContains(t, ids, enrolled.EnrollDevice.Device.ID)
	})

	t.Run("employee only sees own enrolled devices", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		employeeA := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		employeeB := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		employeeBID := employeeB.GetProfileID().String()

		enrolledA := enrollAndActivateDevice(t, employeeA, orgID)
		createdB := createDevice(t, owner, orgID, &employeeBID)
		activateEnrolledDevice(
			t,
			createdB.CreateDevice.EnrollmentToken,
			createdB.CreateDevice.Device.ID+"-hw",
		)

		var result struct {
			Viewer struct {
				EnrolledDevices struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"enrolledDevices"`
			} `json:"viewer"`
		}
		employeeA.MustExecute(listEnrolledDevicesQuery, map[string]any{"orgId": orgID}, &result)

		require.Len(t, result.Viewer.EnrolledDevices.Edges, 1)
		require.Equal(t, enrolledA.EnrollDevice.Device.ID, result.Viewer.EnrolledDevices.Edges[0].Node.ID)
	})

	t.Run("owner can list own enrolled devices", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollAndActivateDevice(t, owner, orgID)

		ids := listEnrolledDeviceIDs(t, owner, orgID)
		require.Contains(t, ids, enrolled.EnrollDevice.Device.ID)
	})

	t.Run("admin can list own enrolled devices", func(t *testing.T) {
		t.Parallel()

		_, admin, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollAndActivateDevice(t, admin, orgID)

		ids := listEnrolledDeviceIDs(t, admin, orgID)
		require.Contains(t, ids, enrolled.EnrollDevice.Device.ID)
	})

	t.Run("owner only sees own enrolled devices", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolledOwner := enrollAndActivateDevice(t, owner, orgID)
		enrollDevice(t, employee, orgID)

		ids := listEnrolledDeviceIDs(t, owner, orgID)
		require.Len(t, ids, 1)
		require.Equal(t, enrolledOwner.EnrollDevice.Device.ID, ids[0])
	})

	t.Run("employee sees device when owner was set with profile id", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		employeeA := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		profileID := employeeA.GetProfileID().String()

		created := createDevice(t, owner, orgID, &profileID)
		activateEnrolledDevice(
			t,
			created.CreateDevice.EnrollmentToken,
			created.CreateDevice.Device.ID+"-hw",
		)

		ids := listEnrolledDeviceIDs(t, employeeA, orgID)
		require.Contains(t, ids, created.CreateDevice.Device.ID)

		var deviceResult struct {
			Node struct {
				Owner *struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"owner"`
			} `json:"node"`
		}
		owner.MustExecute(
			getDeviceQuery,
			map[string]any{"id": created.CreateDevice.Device.ID},
			&deviceResult,
		)
		require.NotNil(t, deviceResult.Node.Owner)
		require.Equal(t, profileID, deviceResult.Node.Owner.ID)
		require.NotEmpty(t, deviceResult.Node.Owner.FullName)
	})

	t.Run("employee cannot revoke device", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollDevice(t, employee, orgID)

		_, err := employee.Do(revokeDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": enrolled.EnrollDevice.Device.ID,
			},
		})
		testutil.RequireForbiddenError(t, err, "employee should not revoke devices")
	})

	t.Run("employee cannot create device for another user", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, ownerProfileID := setupDeviceEnrollmentClients(t)

		_, err := employee.Do(createDeviceMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"ownerId":        ownerProfileID,
			},
		})
		testutil.RequireForbiddenError(t, err, "employee should not create device for another user")
	})

	t.Run("viewer cannot enroll device", func(t *testing.T) {
		t.Parallel()

		_, _, _, viewer, orgID, _ := setupDeviceEnrollmentClients(t)

		_, err := viewer.Do(enrollDeviceMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not enroll devices")
	})

	t.Run("unassumed session can poll own enrolledDevice", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled := enrollAndActivateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		unassumed := testutil.NewClientWithNewSession(t, employee)

		_, err := unassumed.Do(getDeviceQuery, map[string]any{"id": deviceID})
		testutil.RequireErrorCode(t, err, "ASSUMPTION_REQUIRED", "node get requires assumption")

		var result struct {
			Viewer struct {
				EnrolledDevice struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"enrolledDevice"`
			} `json:"viewer"`
		}
		unassumed.MustExecute(getEnrolledDeviceQuery, map[string]any{"id": deviceID}, &result)
		require.Equal(t, deviceID, result.Viewer.EnrolledDevice.ID)
		require.Equal(t, "ACTIVE", result.Viewer.EnrolledDevice.State)
	})

	t.Run("unassumed session cannot read another users enrolledDevice", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		employeeA := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		employeeB := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		enrolledB := enrollDevice(t, employeeB, orgID)

		unassumedA := testutil.NewClientWithNewSession(t, employeeA)

		_, err := unassumedA.Do(getEnrolledDeviceQuery, map[string]any{
			"id": enrolledB.EnrollDevice.Device.ID,
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "employee cannot read another users enrolledDevice")
	})

	t.Run("enrolledDevice does not disclose foreign org device existence", func(t *testing.T) {
		t.Parallel()

		_, _, employeeA, _, _, _ := setupDeviceEnrollmentClients(t)
		_, _, employeeB, _, orgBID, _ := setupDeviceEnrollmentClients(t)

		enrolledB := enrollDevice(t, employeeB, orgBID)

		unassumedA := testutil.NewClientWithNewSession(t, employeeA)

		_, err := unassumedA.Do(getEnrolledDeviceQuery, map[string]any{
			"id": enrolledB.EnrollDevice.Device.ID,
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "foreign org enrolledDevice must look like not found")

		unknownID := gid.New(employeeA.GetOrganizationID().TenantID(), coredata.DeviceEntityType).String()
		_, err = unassumedA.Do(getEnrolledDeviceQuery, map[string]any{
			"id": unknownID,
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "unknown enrolledDevice must look like not found")
	})

	t.Run("owner retains admin access", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)

		created := createDevice(t, owner, orgID, nil)

		var deviceResult struct {
			Node struct {
				Owner *struct {
					FullName string `json:"fullName"`
				} `json:"owner"`
			} `json:"node"`
		}
		owner.MustExecute(
			getDeviceQuery,
			map[string]any{"id": created.CreateDevice.Device.ID},
			&deviceResult,
		)
		require.Nil(t, deviceResult.Node.Owner)

		var listResult struct {
			Node struct {
				Devices struct {
					TotalCount int `json:"totalCount"`
				} `json:"devices"`
			} `json:"node"`
		}
		owner.MustExecute(listDevicesQuery, map[string]any{"orgId": orgID}, &listResult)
		require.GreaterOrEqual(t, listResult.Node.Devices.TotalCount, 1)

		var revokeResult struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}
		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": created.CreateDevice.Device.ID,
			},
		}, &revokeResult)
		require.Equal(t, "REVOKED", revokeResult.RevokeDevice.Device.State)
	})
}

func TestDeviceEnrollmentPermissionQueryShape(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	orgID := owner.GetOrganizationID().String()

	for _, tc := range []struct {
		name   string
		client *testutil.Client
	}{
		{name: "owner", client: owner},
		{name: "admin", client: admin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := tc.client.Do(devicePermissionQuery, map[string]any{"orgId": orgID})
			require.NoError(t, err)

			var result struct {
				Node struct {
					CanEnrollDevice bool `json:"canEnrollDevice"`
				} `json:"node"`
			}
			require.NoError(t, json.Unmarshal(resp.Data, &result))
			require.True(t, result.Node.CanEnrollDevice)
		})
	}
}
