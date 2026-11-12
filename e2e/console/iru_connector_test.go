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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const createIruConnectorMutation = `
	mutation($input: CreateAPIKeyConnectorInput!) {
		createAPIKeyConnector(input: $input) {
			connector {
				id
				provider
				displayName
			}
		}
	}
`

func iruConnectorInput(organizationID string, apiURL any) map[string]any {
	input := map[string]any{
		"organizationId": organizationID,
		"name":           "Iru",
		"provider":       "IRU",
		"apiKey":         "iru-test-token",
	}
	if apiURL != nil {
		input["iruApiUrl"] = apiURL
	}

	return map[string]any{"input": input}
}

// TestIruConnector_Catalog asserts that Iru is offered for connection even
// though it registers no access-review driver: it syncs devices instead.
func TestIruConnector_Catalog(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	var result struct {
		ConnectorProviders []struct {
			Provider            string `json:"provider"`
			DisplayName         string `json:"displayName"`
			APIKeySupported     bool   `json:"apiKeySupported"`
			APIKeyExtraSettings []struct {
				Key      string `json:"key"`
				Required bool   `json:"required"`
			} `json:"apiKeyExtraSettings"`
		} `json:"connectorProviders"`
	}

	owner.MustExecute(`
		query {
			connectorProviders {
				provider
				displayName
				apiKeySupported
				apiKeyExtraSettings { key required }
			}
		}
	`, nil, &result)

	for _, info := range result.ConnectorProviders {
		if info.Provider != "IRU" {
			continue
		}

		assert.Equal(t, "Iru", info.DisplayName)
		assert.True(t, info.APIKeySupported)
		require.Len(t, info.APIKeyExtraSettings, 1)
		assert.Equal(t, "apiUrl", info.APIKeyExtraSettings[0].Key)
		assert.True(t, info.APIKeyExtraSettings[0].Required)

		return
	}

	t.Fatal("expected IRU in the connector catalog")
}

func TestIruConnector_Create(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	t.Run(
		"accepts the API URL Iru shows",
		func(t *testing.T) {
			t.Parallel()

			var result struct {
				CreateAPIKeyConnector struct {
					Connector struct {
						ID          string `json:"id"`
						Provider    string `json:"provider"`
						DisplayName string `json:"displayName"`
					} `json:"connector"`
				} `json:"createAPIKeyConnector"`
			}

			owner.ForTest(t).MustExecute(
				createIruConnectorMutation,
				iruConnectorInput(orgID, " https://ACME.api.kandji.io/ "),
				&result,
			)

			connector := result.CreateAPIKeyConnector.Connector
			assert.NotEmpty(t, connector.ID)
			assert.Equal(t, "IRU", connector.Provider)
			assert.Equal(t, "Iru", connector.DisplayName)
		},
	)

	t.Run(
		"requires the API URL",
		func(t *testing.T) {
			t.Parallel()

			_, err := owner.ForTest(t).Do(createIruConnectorMutation, iruConnectorInput(orgID, nil))
			testutil.RequireErrorCode(t, err, "INVALID", "a missing iruApiUrl must return INVALID")
		},
	)

	t.Run(
		"refuses a host that is not Iru's",
		func(t *testing.T) {
			t.Parallel()

			// The host is where the token is sent, so anything else would
			// hand the customer's Iru token to another server.
			for _, apiURL := range []string{
				"https://evil.example",
				"acme.api.kandji.io.evil.example",
				"http://acme.api.kandji.io",
			} {
				_, err := owner.ForTest(t).Do(createIruConnectorMutation, iruConnectorInput(orgID, apiURL))
				testutil.RequireErrorCode(t, err, "INVALID", "an API URL outside Iru must return INVALID: "+apiURL)
			}
		},
	)

	t.Run(
		"is refused to a viewer",
		func(t *testing.T) {
			t.Parallel()

			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			_, err := viewer.Do(createIruConnectorMutation, iruConnectorInput(orgID, "acme.api.kandji.io"))
			testutil.RequireForbiddenError(t, err, "a viewer must not connect Iru")
		},
	)
}
