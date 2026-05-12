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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type scimClient struct {
	t        testing.TB
	client   *http.Client
	token    string
	endpoint string
}

func newSCIMClient(t testing.TB, owner *testutil.Client) *scimClient {
	t.Helper()

	const query = `
		mutation($input: CreateSCIMConfigurationInput!) {
			createSCIMConfiguration(input: $input) {
				scimConfiguration { id }
				token
			}
		}
	`

	var result struct {
		CreateSCIMConfiguration struct {
			ScimConfiguration struct {
				ID string `json:"id"`
			} `json:"scimConfiguration"`
			Token string `json:"token"`
		} `json:"createSCIMConfiguration"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
		},
	}, &result)
	require.NoError(t, err, "GraphQL request failed")

	require.NotEmpty(t, result.CreateSCIMConfiguration.Token)

	return &scimClient{
		t:        t,
		client:   &http.Client{},
		token:    result.CreateSCIMConfiguration.Token,
		endpoint: testutil.GetBaseURL() + "/api/connect/v1/scim/2.0",
	}
}

func (sc *scimClient) createUser(userName, fullName, externalID string, active bool) (string, int) {
	sc.t.Helper()

	payload := map[string]any{
		"schemas":    []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName":   userName,
		"active":     active,
		"externalId": externalID,
		"name": map[string]any{
			"givenName":  "Test",
			"familyName": "User",
		},
		"displayName": fullName,
		"emails": []map[string]any{
			{"value": userName, "primary": true},
		},
	}

	return sc.doRequest("POST", "/Users", payload)
}

func (sc *scimClient) listUsers() (string, int) {
	sc.t.Helper()
	return sc.doRequest("GET", "/Users", nil)
}

func (sc *scimClient) getUser(id string) (string, int) {
	sc.t.Helper()
	return sc.doRequest("GET", "/Users/"+id, nil)
}

func (sc *scimClient) deleteUser(id string) (string, int) {
	sc.t.Helper()
	return sc.doRequest("DELETE", "/Users/"+id, nil)
}

func (sc *scimClient) doRequest(method, path string, payload any) (string, int) {
	sc.t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		require.NoError(sc.t, err)
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, sc.endpoint+path, body)
	require.NoError(sc.t, err)

	req.Header.Set("Authorization", "Bearer "+sc.token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/scim+json")
	}

	resp, err := sc.client.Do(req)
	require.NoError(sc.t, err)
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(sc.t, err)

	return string(respBody), resp.StatusCode
}

func TestSCIM_CreateUser(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	sc := newSCIMClient(t, owner)

	t.Run("create a new user", func(t *testing.T) {
		t.Parallel()

		email := factory.SafeEmail()
		body, status := sc.createUser(email, "New User", "ext-create-1", true)

		assert.Equal(t, http.StatusCreated, status, body)

		var resource map[string]any
		require.NoError(t, json.Unmarshal([]byte(body), &resource))
		assert.Equal(t, email, resource["userName"])
		assert.NotEmpty(t, resource["id"])
	})

	t.Run("duplicate user returns 409", func(t *testing.T) {
		t.Parallel()

		email := factory.SafeEmail()
		_, status := sc.createUser(email, "Dup User", "ext-dup-1", true)
		require.Equal(t, http.StatusCreated, status)

		_, status = sc.createUser(email, "Dup User", "ext-dup-1", true)
		assert.Equal(t, http.StatusConflict, status)
	})
}

func TestSCIM_ListUsers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	sc := newSCIMClient(t, owner)

	email := factory.SafeEmail()
	_, status := sc.createUser(email, "List User", "ext-list-1", true)
	require.Equal(t, http.StatusCreated, status)

	body, status := sc.listUsers()
	require.Equal(t, http.StatusOK, status, body)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &response))

	resources := response["Resources"].([]any)
	assert.GreaterOrEqual(t, len(resources), 1)
}

func TestSCIM_ExternalIDFallback(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	sc := newSCIMClient(t, owner)

	t.Run("email rename reuses profile via external ID", func(t *testing.T) {
		t.Parallel()

		externalID := "google-" + factory.SafeName("")
		oldEmail := factory.SafeEmail()
		newEmail := factory.SafeEmail()

		// Create user with old email
		body, status := sc.createUser(oldEmail, "Rename User", externalID, true)
		require.Equal(t, http.StatusCreated, status, body)

		var created map[string]any
		require.NoError(t, json.Unmarshal([]byte(body), &created))
		originalID := created["id"].(string)

		// Create user with new email but same external ID (simulates email rename)
		body, status = sc.createUser(newEmail, "Rename User", externalID, true)
		require.Equal(t, http.StatusCreated, status, body)

		var updated map[string]any
		require.NoError(t, json.Unmarshal([]byte(body), &updated))

		// Should reuse the same profile (same ID)
		assert.Equal(t, originalID, updated["id"].(string), "profile ID should be preserved after email rename")
		assert.Equal(t, newEmail, updated["userName"], "email should be updated")

		// Verify via GET that the profile is consistent
		body, status = sc.getUser(originalID)
		require.Equal(t, http.StatusOK, status, body)

		var fetched map[string]any
		require.NoError(t, json.Unmarshal([]byte(body), &fetched))
		assert.Equal(t, newEmail, fetched["userName"])
	})

	t.Run("different external ID creates new profile", func(t *testing.T) {
		t.Parallel()

		email := factory.SafeEmail()

		_, status := sc.createUser(email, "User A", "ext-a-"+factory.SafeName(""), true)
		require.Equal(t, http.StatusCreated, status)

		// Same email, different external ID — should fail (email already taken)
		_, status = sc.createUser(email, "User B", "ext-b-"+factory.SafeName(""), true)
		assert.Equal(t, http.StatusConflict, status)
	})
}

func TestSCIM_DeleteUser(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	sc := newSCIMClient(t, owner)

	email := factory.SafeEmail()
	body, status := sc.createUser(email, "Delete User", "ext-del-1", true)
	require.Equal(t, http.StatusCreated, status, body)

	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	userID := created["id"].(string)

	_, status = sc.deleteUser(userID)
	assert.Equal(t, http.StatusNoContent, status)

	_, status = sc.getUser(userID)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestSCIM_Unauthorized(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	req, err := http.NewRequest("GET", testutil.GetBaseURL()+"/api/connect/v1/scim/2.0/Users", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
