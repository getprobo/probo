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
	"time"

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

	deleteDeviceMutation = `
		mutation DeleteDevice($input: DeleteDeviceInput!) {
			deleteDevice(input: $input) {
				deletedDeviceId
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

	devicePostureReportsQuery = `
		query DevicePostureReports($id: ID!) {
			node(id: $id) {
				... on Device {
					latestPostures {
						id
						checkKey
						status
						value { kind text number }
					}
					postureReports(first: 10) {
						totalCount
						edges {
							cursor
							node {
								id
								createdAt
								postures {
									id
									checkKey
									value { kind text number }
								}
							}
						}
					}
				}
			}
		}`
)

type devicePostureValue struct {
	Kind   string `json:"kind"`
	Text   string `json:"text"`
	Number *int   `json:"number"`
}

type devicePostureReportsResult struct {
	Node struct {
		LatestPostures []struct {
			ID       string             `json:"id"`
			CheckKey string             `json:"checkKey"`
			Status   string             `json:"status"`
			Value    devicePostureValue `json:"value"`
		} `json:"latestPostures"`
		PostureReports struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Cursor string `json:"cursor"`
				Node   struct {
					ID        string `json:"id"`
					CreatedAt string `json:"createdAt"`
					Postures  []struct {
						ID       string             `json:"id"`
						CheckKey string             `json:"checkKey"`
						Value    devicePostureValue `json:"value"`
					} `json:"postures"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"postureReports"`
	} `json:"node"`
}

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

func enrollActivateAndAuthenticateDevice(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
) (enrollDeviceResult, string) {
	t.Helper()

	enrolled := enrollDevice(t, client, organizationID)

	status, payload := exchangeEnrollmentToken(
		t,
		enrolled.EnrollDevice.EnrollmentToken,
	)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, payload.APIKey)

	require.Equal(
		t,
		http.StatusOK,
		sendHeartbeat(t, payload.APIKey, enrolled.EnrollDevice.Device.ID+"-hw"),
	)

	return enrolled, payload.APIKey
}

func reportPostures(t *testing.T, apiKey string, results []map[string]any) int {
	t.Helper()

	body, err := json.Marshal(map[string]any{"results": results})
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		testutil.GetBaseURL()+"/api/agent/v1/postures",
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

func newPostureCorrelationID(t *testing.T, deviceID string) string {
	t.Helper()

	id, err := gid.ParseGID(deviceID)
	require.NoError(t, err)

	return gid.New(id.TenantID(), coredata.DevicePostureReportEntityType).String()
}

// ufwInactivePosture is the evidence a Linux host with a disabled firewall
// reports. The server must read it as OFF: "inactive" contains "active", so a
// substring test inverts the signal.
func ufwInactivePosture(
	observedAt time.Time,
	correlationID string,
) map[string]any {
	result := map[string]any{
		"check_key":   "FIREWALL_ENABLED",
		"status":      "FAIL",
		"observed_at": observedAt.Format(time.RFC3339Nano),
		"evidence": map[string]any{
			"backend": "ufw",
			"raw":     "Status: inactive",
		},
	}
	if correlationID != "" {
		result["correlation_id"] = correlationID
	}

	return result
}

func osVersionPosture(
	observedAt time.Time,
	version string,
	correlationID string,
) map[string]any {
	result := map[string]any{
		"check_key":   "OS_VERSION",
		"status":      "PASS",
		"observed_at": observedAt.Format(time.RFC3339Nano),
		"evidence": map[string]any{
			"pretty_name": version,
		},
	}
	if correlationID != "" {
		result["correlation_id"] = correlationID
	}

	return result
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

func TestDeviceDelete(t *testing.T) {
	t.Parallel()

	t.Run("cannot delete pending device", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)
		created := createDevice(t, owner, orgID, nil)

		_, err := owner.Do(deleteDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": created.CreateDevice.Device.ID,
			},
		})
		testutil.RequireErrorCode(t, err, "CONFLICT", "pending device must be revoked before delete")
	})

	t.Run("owner can delete revoked device", func(t *testing.T) {
		t.Parallel()

		owner, _, _, _, orgID, _ := setupDeviceEnrollmentClients(t)
		created := createDevice(t, owner, orgID, nil)
		deviceID := created.CreateDevice.Device.ID

		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{"deviceId": deviceID},
		}, &struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}{})

		var deleteResult struct {
			DeleteDevice struct {
				DeletedDeviceID string `json:"deletedDeviceId"`
			} `json:"deleteDevice"`
		}
		owner.MustExecute(deleteDeviceMutation, map[string]any{
			"input": map[string]any{"deviceId": deviceID},
		}, &deleteResult)
		require.Equal(t, deviceID, deleteResult.DeleteDevice.DeletedDeviceID)

		_, err := owner.Do(getDeviceQuery, map[string]any{"id": deviceID})
		testutil.RequireErrorCode(t, err, "NOT_FOUND", "soft-deleted device must not be readable")
	})

	t.Run("cannot delete active device", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)
		enrolled, _ := enrollActivateAndAuthenticateDevice(t, employee, orgID)

		_, err := owner.Do(deleteDeviceMutation, map[string]any{
			"input": map[string]any{
				"deviceId": enrolled.EnrollDevice.Device.ID,
			},
		})
		testutil.RequireErrorCode(t, err, "CONFLICT", "active device must not be deletable")
	})

	t.Run("employee cannot delete device", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)
		created := createDevice(t, owner, orgID, nil)
		deviceID := created.CreateDevice.Device.ID

		owner.MustExecute(revokeDeviceMutation, map[string]any{
			"input": map[string]any{"deviceId": deviceID},
		}, &struct {
			RevokeDevice struct {
				Device struct {
					State string `json:"state"`
				} `json:"device"`
			} `json:"revokeDevice"`
		}{})

		_, err := employee.Do(deleteDeviceMutation, map[string]any{
			"input": map[string]any{"deviceId": deviceID},
		})
		testutil.RequireForbiddenError(t, err, "employee should not delete devices")
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

func TestDevicePostureReports(t *testing.T) {
	t.Parallel()

	t.Run("one agent run becomes one report", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled, apiKey := enrollActivateAndAuthenticateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		observedAt := time.Now().UTC()
		correlationID := newPostureCorrelationID(t, deviceID)
		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				ufwInactivePosture(observedAt, correlationID),
				osVersionPosture(observedAt, "Ubuntu 24.04.2 LTS", correlationID),
			}),
		)

		var result devicePostureReportsResult
		owner.MustExecute(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
			&result,
		)

		reports := result.Node.PostureReports
		require.Equal(t, 1, reports.TotalCount)
		require.Len(t, reports.Edges, 1)

		report := reports.Edges[0].Node
		require.Equal(t, correlationID, report.ID)
		require.NotEmpty(t, report.CreatedAt)
		require.NotEmpty(t, reports.Edges[0].Cursor)
		require.Len(t, report.Postures, 2)

		values := map[string]devicePostureValue{}
		for _, posture := range report.Postures {
			values[posture.CheckKey] = posture.Value
		}

		require.Equal(t, "OFF", values["FIREWALL_ENABLED"].Kind)
		require.Equal(t, "TEXT", values["OS_VERSION"].Kind)
		require.Equal(t, "Ubuntu 24.04.2 LTS", values["OS_VERSION"].Text)

		require.Len(t, result.Node.LatestPostures, 2)

		for _, posture := range result.Node.LatestPostures {
			require.Equal(
				t,
				values[posture.CheckKey].Kind,
				posture.Value.Kind,
				"latest posture and report disagree on %s",
				posture.CheckKey,
			)
		}
	})

	t.Run("legacy agent without correlation_id still groups one report", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled, apiKey := enrollActivateAndAuthenticateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		observedAt := time.Now().UTC()
		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				ufwInactivePosture(observedAt, ""),
				osVersionPosture(observedAt, "Ubuntu 24.04.2 LTS", ""),
			}),
		)

		var result devicePostureReportsResult
		owner.MustExecute(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
			&result,
		)

		reports := result.Node.PostureReports
		require.Equal(t, 1, reports.TotalCount)
		require.Len(t, reports.Edges, 1)

		report := reports.Edges[0].Node
		reportID, err := gid.ParseGID(report.ID)
		require.NoError(t, err)
		require.Equal(t, coredata.DevicePostureReportEntityType, reportID.EntityType())

		deviceGID, err := gid.ParseGID(deviceID)
		require.NoError(t, err)
		require.Equal(t, deviceGID.TenantID(), reportID.TenantID())

		require.Len(t, report.Postures, 2)

		values := map[string]devicePostureValue{}
		for _, posture := range report.Postures {
			values[posture.CheckKey] = posture.Value
		}

		require.Equal(t, "OFF", values["FIREWALL_ENABLED"].Kind)
		require.Equal(t, "TEXT", values["OS_VERSION"].Kind)
		require.Equal(t, "Ubuntu 24.04.2 LTS", values["OS_VERSION"].Text)
	})

	t.Run("each agent run adds a report", func(t *testing.T) {
		t.Parallel()

		owner, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled, apiKey := enrollActivateAndAuthenticateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		firstCorrelationID := newPostureCorrelationID(t, deviceID)
		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				osVersionPosture(time.Now().UTC(), "Ubuntu 24.04.1 LTS", firstCorrelationID),
			}),
		)

		var first devicePostureReportsResult
		owner.MustExecute(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
			&first,
		)
		require.Equal(t, 1, first.Node.PostureReports.TotalCount)

		secondCorrelationID := newPostureCorrelationID(t, deviceID)
		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				osVersionPosture(time.Now().UTC(), "Ubuntu 24.04.2 LTS", secondCorrelationID),
			}),
		)

		var second devicePostureReportsResult
		owner.MustExecute(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
			&second,
		)
		require.Equal(t, 2, second.Node.PostureReports.TotalCount)
		require.Len(t, second.Node.PostureReports.Edges, 2)

		newest := second.Node.PostureReports.Edges[0].Node
		oldest := second.Node.PostureReports.Edges[1].Node

		require.Equal(t, secondCorrelationID, newest.ID)
		require.Equal(t, firstCorrelationID, oldest.ID)
		require.NotEqual(t, newest.CreatedAt, oldest.CreatedAt)
		require.Equal(
			t,
			first.Node.PostureReports.Edges[0].Node.CreatedAt,
			oldest.CreatedAt,
			"the first run's report must survive the second",
		)
		require.Equal(
			t,
			"Ubuntu 24.04.2 LTS",
			newest.Postures[0].Value.Text,
		)
	})

	t.Run("viewer can read posture reports", func(t *testing.T) {
		t.Parallel()

		_, _, employee, viewer, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled, apiKey := enrollActivateAndAuthenticateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				osVersionPosture(
					time.Now().UTC(),
					"Ubuntu 24.04.2 LTS",
					newPostureCorrelationID(t, deviceID),
				),
			}),
		)

		var result devicePostureReportsResult
		viewer.MustExecute(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
			&result,
		)
		require.Equal(t, 1, result.Node.PostureReports.TotalCount)

		err := employee.ExecuteShouldFail(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
		)
		require.Error(t, err, "employee must not read device postures")
	})

	t.Run("another organization cannot read posture reports", func(t *testing.T) {
		t.Parallel()

		_, _, employee, _, orgID, _ := setupDeviceEnrollmentClients(t)

		enrolled, apiKey := enrollActivateAndAuthenticateDevice(t, employee, orgID)
		deviceID := enrolled.EnrollDevice.Device.ID

		require.Equal(
			t,
			http.StatusNoContent,
			reportPostures(t, apiKey, []map[string]any{
				osVersionPosture(
					time.Now().UTC(),
					"Ubuntu 24.04.2 LTS",
					newPostureCorrelationID(t, deviceID),
				),
			}),
		)

		outsider := testutil.NewClient(t, testutil.RoleOwner)

		err := outsider.ExecuteShouldFail(
			devicePostureReportsQuery,
			map[string]any{"id": deviceID},
		)
		require.Error(t, err, "another organization must not read device postures")
	})
}
