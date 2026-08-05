// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package console_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type scimEventNode struct {
	ID           string  `json:"id"`
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	StatusCode   int     `json:"statusCode"`
	UserName     string  `json:"userName"`
	RequestBody  *string `json:"requestBody"`
	ResponseBody *string `json:"responseBody"`
	ErrorMessage *string `json:"errorMessage"`
}

func listSCIMEvents(t *testing.T, owner *testutil.Client, configID string) []scimEventNode {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on SCIMConfiguration {
					events(first: 50, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges {
							node {
								id
								method
								path
								statusCode
								userName
								requestBody
								responseBody
								errorMessage
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Events struct {
				Edges []struct {
					Node scimEventNode `json:"node"`
				} `json:"edges"`
			} `json:"events"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{"id": configID}, &result)
	require.NoError(t, err)

	events := make([]scimEventNode, 0, len(result.Node.Events.Edges))
	for _, edge := range result.Node.Events.Edges {
		events = append(events, edge.Node)
	}

	return events
}

func findSCIMEvent(t *testing.T, events []scimEventNode, method string, userName string) scimEventNode {
	t.Helper()

	for _, event := range events {
		if event.Method == method && event.UserName == userName {
			return event
		}
	}

	t.Fatalf("SCIM event not found for method=%s among %d events", method, len(events))
	return scimEventNode{}
}

func decodeJSON(t *testing.T, raw *string) map[string]any {
	t.Helper()

	require.NotNil(t, raw)
	require.NotEmpty(t, *raw)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(*raw), &payload))

	return payload
}

func requireJSONContains(t *testing.T, raw *string, key string, want any) {
	t.Helper()

	assert.Equal(t, want, decodeJSON(t, raw)[key])
}

func TestSCIMEvent_StoresPayload(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	sc := newSCIMClient(t, owner)

	email := factory.SafeEmail()
	externalID := "ext-event-" + factory.SafeName("")

	body, status := sc.createUser(email, "Event User", externalID, true)
	require.Equal(t, http.StatusCreated, status, body)

	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	userID := created["id"].(string)

	body, status = sc.replaceUser(userID, email, "Event User Updated", externalID, true)
	require.Equal(t, http.StatusOK, status, body)

	body, status = sc.patchUser(userID, []map[string]any{
		{
			"op":    "replace",
			"path":  "active",
			"value": false,
		},
	})
	require.Equal(t, http.StatusOK, status, body)

	body, status = sc.getUser(userID)
	require.Equal(t, http.StatusOK, status, body)

	events := listSCIMEvents(t, owner, sc.configID)

	t.Run("create stores request and response bodies", func(t *testing.T) {
		t.Parallel()

		event := findSCIMEvent(t, events, "POST", email)
		assert.Equal(t, 201, event.StatusCode)
		assert.Equal(t, "/Users", event.Path)
		requireJSONContains(t, event.RequestBody, "userName", email)
		requireJSONContains(t, event.ResponseBody, "userName", email)
		requireJSONContains(t, event.ResponseBody, "id", userID)

		// The stored response is the payload served to the client, so it carries
		// the envelope fields the SCIM server adds.
		response := decodeJSON(t, event.ResponseBody)
		assert.Contains(t, response, "schemas")
		assert.Contains(t, response, "meta")
	})

	t.Run("replace stores request and response bodies", func(t *testing.T) {
		t.Parallel()

		event := findSCIMEvent(t, events, "PUT", email)
		assert.Equal(t, 200, event.StatusCode)
		assert.Equal(t, "/Users/"+userID, event.Path)
		requireJSONContains(t, event.RequestBody, "userName", email)
		requireJSONContains(t, event.RequestBody, "displayName", "Event User Updated")
		requireJSONContains(t, event.ResponseBody, "id", userID)
	})

	t.Run("patch stores request and response bodies", func(t *testing.T) {
		t.Parallel()

		event := findSCIMEvent(t, events, "PATCH", email)
		assert.Equal(t, 200, event.StatusCode)
		assert.Equal(t, "/Users/"+userID, event.Path)

		ops, ok := decodeJSON(t, event.RequestBody)["Operations"].([]any)
		require.True(t, ok)
		require.Len(t, ops, 1)

		op, ok := ops[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "replace", op["op"])
		assert.Equal(t, "active", op["path"])
		assert.Equal(t, false, op["value"])

		requireJSONContains(t, event.ResponseBody, "id", userID)
		requireJSONContains(t, event.ResponseBody, "active", false)
	})

	t.Run("get stores response body without request body", func(t *testing.T) {
		t.Parallel()

		event := findSCIMEvent(t, events, "GET", email)
		assert.Equal(t, 200, event.StatusCode)
		assert.Equal(t, "/Users/"+userID, event.Path)
		assert.Nil(t, event.RequestBody)
		requireJSONContains(t, event.ResponseBody, "id", userID)
		requireJSONContains(t, event.ResponseBody, "userName", email)
	})
}

func TestSCIMEvent_Export(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const mutation = `
		mutation($input: RequestSCIMEventExportInput!) {
			requestSCIMEventExport(input: $input) {
				exportJobId
			}
		}
	`

	t.Run("owner can request export", func(t *testing.T) {
		t.Parallel()

		var result struct {
			RequestSCIMEventExport struct {
				ExportJobID string `json:"exportJobId"`
			} `json:"requestSCIMEventExport"`
		}

		err := owner.ExecuteConnect(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.RequestSCIMEventExport.ExportJobID)
	})

	t.Run("admin can request export", func(t *testing.T) {
		t.Parallel()
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

		var result struct {
			RequestSCIMEventExport struct {
				ExportJobID string `json:"exportJobId"`
			} `json:"requestSCIMEventExport"`
		}

		err := admin.ExecuteConnect(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": admin.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.RequestSCIMEventExport.ExportJobID)
	})

	t.Run("viewer cannot request export", func(t *testing.T) {
		t.Parallel()
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.DoConnect(mutation, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"fromTime":       "2026-01-01T00:00:00Z",
				"toTime":         "2026-03-24T00:00:00Z",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot request SCIM event export")
	})
}
