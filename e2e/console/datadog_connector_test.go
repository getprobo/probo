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

const createDatadogAPIKeyConnectorMutation = `
	mutation($input: CreateAPIKeyConnectorInput!) {
		createAPIKeyConnector(input: $input) {
			connector { id provider }
		}
	}
`

func TestDatadogAPIKeyCatalogEntry(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query {
			connectorProviders {
				provider
				apiKeySupported
				apiKeyExtraSettings { key required }
				apiKeyFormat { example }
			}
		}
	`

	var result struct {
		ConnectorProviders []struct {
			Provider            string `json:"provider"`
			APIKeySupported     bool   `json:"apiKeySupported"`
			APIKeyExtraSettings []struct {
				Key      string `json:"key"`
				Required bool   `json:"required"`
			} `json:"apiKeyExtraSettings"`
			APIKeyFormat *struct {
				Example string `json:"example"`
			} `json:"apiKeyFormat"`
		} `json:"connectorProviders"`
	}

	require.NoError(t, owner.Execute(query, nil, &result))

	for _, p := range result.ConnectorProviders {
		if p.Provider != "DATADOG" {
			continue
		}

		assert.True(t, p.APIKeySupported)
		require.Len(t, p.APIKeyExtraSettings, 1)
		assert.Equal(t, "site", p.APIKeyExtraSettings[0].Key)
		assert.True(t, p.APIKeyExtraSettings[0].Required)
		require.NotNil(t, p.APIKeyFormat)
		assert.Equal(t, "apiKey:applicationKey", p.APIKeyFormat.Example)

		return
	}

	t.Fatal("DATADOG is missing from the connector catalog")
}

func TestCreateAPIKeyConnectorDatadog(t *testing.T) {
	t.Parallel()

	t.Run("a key pair and a known site create the connector", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		var result struct {
			CreateAPIKeyConnector struct {
				Connector struct {
					ID       string `json:"id"`
					Provider string `json:"provider"`
				} `json:"connector"`
			} `json:"createAPIKeyConnector"`
		}

		err := owner.Execute(createDatadogAPIKeyConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Datadog",
				"provider":       "DATADOG",
				"apiKey":         "0123456789abcdef0123456789abcdef:0123456789abcdef0123456789abcdef01234567",
				"datadogSite":    "EU1",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateAPIKeyConnector.Connector.ID)
		assert.Equal(t, "DATADOG", result.CreateAPIKeyConnector.Connector.Provider)
	})

	for _, tc := range []struct {
		name   string
		apiKey string
		site   any
	}{
		{name: "a missing site is refused", apiKey: "api:app", site: nil},
		{name: "an unknown site is refused", apiKey: "api:app", site: "US2"},
		{name: "a host instead of a site is refused", apiKey: "api:app", site: "evil.example.com"},
		{name: "an API key without its application key is refused", apiKey: "0123456789abcdef0123456789abcdef", site: "US1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			owner := testutil.NewClient(t, testutil.RoleOwner)

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Datadog",
				"provider":       "DATADOG",
				"apiKey":         tc.apiKey,
			}
			if tc.site != nil {
				input["datadogSite"] = tc.site
			}

			_, err := owner.Do(createDatadogAPIKeyConnectorMutation, map[string]any{"input": input})
			testutil.RequireErrorCode(t, err, "INVALID", "must return INVALID, not create a connector")
		})
	}

	t.Run("a viewer cannot create one", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(createDatadogAPIKeyConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Datadog",
				"provider":       "DATADOG",
				"apiKey":         "api:app",
				"datadogSite":    "US1",
			},
		})
		testutil.RequireForbiddenError(t, err)
	})
}
