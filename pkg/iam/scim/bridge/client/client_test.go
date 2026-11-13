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

package scimclient_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
)

func TestUser_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run(
		"all fields populated",
		func(t *testing.T) {
			t.Parallel()

			data := []byte(`{
				"id": "user-123",
				"userName": "john@example.com",
				"displayName": "John Doe",
				"active": true,
				"title": "Engineer",
				"externalId": "ext-456",
				"userType": "Employee",
				"preferredLanguage": "en",
				"name": {
					"givenName": "John",
					"familyName": "Doe"
				},
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
					"employeeNumber": "EMP001",
					"department": "Engineering",
					"costCenter": "CC100",
					"organization": "Acme Corp",
					"division": "Platform",
					"manager": {
						"value": "jane@example.com"
					}
				},
				"urn:ietf:params:scim:schemas:extension:probo:2.0:User": {
					"photoETag": "etag-1"
				}
			}`)

			var user scimclient.User

			err := json.Unmarshal(data, &user)

			require.NoError(t, err)
			assert.Equal(t, "user-123", user.ID)
			assert.Equal(t, "john@example.com", user.UserName)
			assert.Equal(t, "John Doe", user.DisplayName)
			assert.Equal(t, true, user.Active)
			assert.Equal(t, "Engineer", user.Title)
			assert.Equal(t, "ext-456", user.ExternalID)
			assert.Equal(t, "Employee", user.UserType)
			assert.Equal(t, "en", user.PreferredLanguage)
			assert.Equal(t, "John", user.GivenName)
			assert.Equal(t, "Doe", user.FamilyName)
			assert.Equal(t, "EMP001", user.EmployeeNumber)
			assert.Equal(t, "Engineering", user.Department)
			assert.Equal(t, "CC100", user.CostCenter)
			assert.Equal(t, "Acme Corp", user.EnterpriseOrganization)
			assert.Equal(t, "Platform", user.Division)
			assert.Equal(t, "jane@example.com", user.ManagerValue)
			assert.Equal(t, "etag-1", user.PhotoETag)
		},
	)

	t.Run(
		"no enterprise extension",
		func(t *testing.T) {
			t.Parallel()

			data := []byte(`{
				"id": "user-789",
				"userName": "alice@example.com",
				"displayName": "Alice Smith",
				"active": false,
				"name": {
					"givenName": "Alice",
					"familyName": "Smith"
				}
			}`)

			var user scimclient.User

			err := json.Unmarshal(data, &user)

			require.NoError(t, err)
			assert.Equal(t, "user-789", user.ID)
			assert.Equal(t, "alice@example.com", user.UserName)
			assert.Equal(t, "Alice Smith", user.DisplayName)
			assert.Equal(t, false, user.Active)
			assert.Equal(t, "Alice", user.GivenName)
			assert.Equal(t, "Smith", user.FamilyName)
			assert.Equal(t, "", user.EmployeeNumber)
			assert.Equal(t, "", user.Department)
			assert.Equal(t, "", user.ManagerValue)
		},
	)

	t.Run(
		"minimal fields",
		func(t *testing.T) {
			t.Parallel()

			data := []byte(`{
				"id": "user-min",
				"userName": "bob@example.com",
				"active": true,
				"photoETag": "not-a-core-attribute"
			}`)

			var user scimclient.User

			err := json.Unmarshal(data, &user)

			require.NoError(t, err)
			assert.Equal(t, "user-min", user.ID)
			assert.Equal(t, "bob@example.com", user.UserName)
			assert.Equal(t, true, user.Active)
			assert.Equal(t, "", user.DisplayName)
			assert.Equal(t, "", user.GivenName)
			assert.Equal(t, "", user.FamilyName)
			assert.Equal(t, "", user.Title)
			assert.Empty(t, user.PhotoETag)
		},
	)
}

func TestClient_CreateUser_SendsPhotoWhenPresent(t *testing.T) {
	t.Parallel()

	var body []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	client := scimclient.NewClient(server.Client(), server.URL, "token")
	err := client.CreateUser(context.Background(), &scimclient.User{
		UserName:  "ada@example.com",
		PhotoETag: "etag-1",
		Photo:     []byte("face"),
		PhotoMIME: "image/png",
	})

	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.NotContains(t, payload, "photoETag")

	schemas, ok := payload["schemas"].([]any)
	require.True(t, ok)
	assert.Contains(t, schemas, scimclient.PhotoExtensionURN)

	extension, ok := payload[scimclient.PhotoExtensionURN].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "etag-1", extension["photoETag"])

	photos, ok := payload["photos"].([]any)
	require.True(t, ok)
	require.Len(t, photos, 1)

	photo, ok := photos[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "data:image/png;base64,ZmFjZQ==", photo["value"])
	assert.Equal(t, true, photo["primary"])
}
