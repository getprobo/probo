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

// TestSCIM_DeleteUser_ArchivesWhenProfileInUse verifies that deleting a SCIM
// user whose profile is still referenced by a completed document version
// signature archives the profile (state INACTIVE) instead of failing. The
// signature's RESTRICT foreign key makes the hard delete fail; the service must
// fall back to deactivation and return 204, not surface an opaque 500.
func TestSCIM_DeleteUser_ArchivesWhenProfileInUse(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	sc := newSCIMClient(t, owner)

	// The signer signs a published document version, creating a completed
	// signature that references the signer's profile via a RESTRICT FK.
	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)
	versionID := latestDocumentVersionID(t, owner, docID)

	requestDocumentSignature(t, owner, versionID, signer.GetProfileID().String())

	_, state, _ := signDocumentVersion(t, signer, versionID)
	require.Equal(t, "SIGNED", state)

	// Enroll the signer's existing profile into SCIM so it becomes SCIM-managed
	// (same underlying profile ID, source flipped to SCIM).
	body, status := sc.createUser(signer.GetEmail(), "Signer User", "ext-signed-1", true)
	require.Equal(t, http.StatusCreated, status, body)

	var enrolled map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &enrolled))
	scimUserID := enrolled["id"].(string)
	require.Equal(t, signer.GetProfileID().String(), scimUserID)

	// Deleting the in-use profile must archive it, not 500.
	body, status = sc.deleteUser(scimUserID)
	require.Equal(t, http.StatusNoContent, status, body)

	// The profile is archived (deactivated), not hard-deleted: it is still
	// present but inactive, and the completed signature is preserved.
	body, status = sc.getUser(scimUserID)
	require.Equal(t, http.StatusOK, status, body)

	var fetched map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &fetched))
	assert.Equal(t, false, fetched["active"], "profile should be archived (inactive), not deleted")
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
