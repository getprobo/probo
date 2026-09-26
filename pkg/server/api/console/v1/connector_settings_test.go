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

package console_v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

// Langfuse and 1Password are the two providers whose declared settings failed
// to reach these resolvers: the console dropped Langfuse's Base URL, and
// 1Password's single flat list made the API-key dialog collect the
// client-credentials fields. Each test below walks the whole chain the console
// walks — the key a Registration declares, the mutation input field it is
// submitted as, the settings struct that is persisted — so a key renamed on one
// side and not the other fails here instead of at connect time.
func TestApiKeyConnectorSettings_LangfuseBaseURL(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok)
	require.Len(t, reg.APIKeyExtraSettings(), 1)
	require.Equal(t, "baseUrl", reg.APIKeyExtraSettings()[0].Key)

	baseURL := "https://cloud.langfuse.com"

	raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:        coredata.ConnectorProviderLangfuse,
		LangfuseBaseURL: &baseURL,
	})
	require.NoError(t, err)

	var settings coredata.LangfuseConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, baseURL, settings.BaseURL)

	_, err = apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider: coredata.ConnectorProviderLangfuse,
	})
	require.Error(t, err)
}

func TestApiKeyConnectorSettings_OnePasswordSCIMBridgeURL(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok)
	require.Len(t, reg.APIKeyExtraSettings(), 1)
	require.Equal(t, "scimBridgeUrl", reg.APIKeyExtraSettings()[0].Key)

	scimBridgeURL := "https://scim.example.test"

	raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:                 coredata.ConnectorProviderOnePassword,
		OnePasswordScimBridgeURL: &scimBridgeURL,
	})
	require.NoError(t, err)

	var settings coredata.OnePasswordConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, scimBridgeURL, settings.SCIMBridgeURL)

	// The old shared settings list made this dialog collect Account ID and
	// Region instead, which CreateAPIKeyConnectorInput has no fields for at
	// all: whatever the customer typed was dropped and the create failed here.
	_, err = apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider: coredata.ConnectorProviderOnePassword,
	})
	require.Error(t, err)
}

func TestApiKeyConnectorSettings_ElasticCloudOrganizationID(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderElasticCloud)
	require.True(t, ok)
	require.Len(t, reg.APIKeyExtraSettings(), 1)
	require.Equal(t, "organization_id", reg.APIKeyExtraSettings()[0].Key)

	organizationID := "00000000000000000000000000000000"

	raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:                   coredata.ConnectorProviderElasticCloud,
		ElasticCloudOrganizationID: &organizationID,
	})
	require.NoError(t, err)

	var settings coredata.ElasticCloudConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, organizationID, settings.OrganizationID)

	_, err = apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider: coredata.ConnectorProviderElasticCloud,
	})
	require.Error(t, err)
}

func TestClientCredentialsConnectorSettings_OnePassword(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok)
	require.Len(t, reg.ClientCredentialsExtraSettings(), 2)
	require.Equal(t, "accountId", reg.ClientCredentialsExtraSettings()[0].Key)
	require.Equal(t, "region", reg.ClientCredentialsExtraSettings()[1].Key)

	accountID, region := "acme", "EU"

	raw, err := clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
		OnePasswordRegion:    &region,
	})
	require.NoError(t, err)

	var settings coredata.OnePasswordUsersAPISettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, accountID, settings.AccountID)
	assert.Equal(t, region, settings.Region)

	_, err = clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
	})
	require.Error(t, err)
}

// TestClientCredentialsTokenURL pins which endpoint a client-credentials
// connector POSTs its client secret to. The decision is split between a
// provider that fixes the endpoint and one that cannot, and getting it wrong
// either sends a credential to a client-supplied address or refuses a connect
// the customer has no way to complete.
func TestClientCredentialsTokenURL(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	supplied := func(s string) *string { return &s }

	t.Run("a pinning provider ignores what the client sends", func(t *testing.T) {
		t.Parallel()

		got, err := clientCredentialsTokenURL(
			registry,
			coredata.ConnectorProviderMongoDBAtlas,
			supplied("https://attacker.example/token"),
		)
		require.NoError(t, err)
		assert.Equal(t, "https://cloud.mongodb.com/api/oauth/token", got)
	})

	t.Run("a pinning provider needs nothing from the client", func(t *testing.T) {
		t.Parallel()

		got, err := clientCredentialsTokenURL(registry, coredata.ConnectorProviderMongoDBAtlas, nil)
		require.NoError(t, err)
		assert.Equal(t, "https://cloud.mongodb.com/api/oauth/token", got)
	})

	// 1Password declares no Endpoints.Token: its token host follows the region
	// the customer picks, so the value still comes from the form.
	t.Run("a non-pinning provider keeps taking the input", func(t *testing.T) {
		t.Parallel()

		got, err := clientCredentialsTokenURL(
			registry,
			coredata.ConnectorProviderOnePassword,
			supplied("https://api.1password.eu/v1beta1/users/oauth2/token"),
		)
		require.NoError(t, err)
		assert.Equal(t, "https://api.1password.eu/v1beta1/users/oauth2/token", got)
	})

	// An unregistered provider has no registration to consult, so there is no
	// basis for honouring a client-supplied endpoint either.
	t.Run("an unknown provider is rejected outright", func(t *testing.T) {
		t.Parallel()

		_, err := clientCredentialsTokenURL(
			provider.NewRegistry(),
			coredata.ConnectorProviderMongoDBAtlas,
			supplied("https://cloud.mongodb.com/api/oauth/token"),
		)
		assert.Error(t, err)
	})

	t.Run("a non-pinning provider rejects a missing or unusable URL", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name     string
			supplied *string
		}{
			{name: "absent", supplied: nil},
			{name: "blank", supplied: supplied("   ")},
			{name: "not a URL", supplied: supplied("not-a-url")},
			{name: "relative", supplied: supplied("/token")},
			// Host is non-empty here but carries only a port, so the exchange
			// could never reach it.
			{name: "port-only authority", supplied: supplied("https://:443/token")},
			{name: "port-only authority, no path", supplied: supplied("https://:443")},
			// A client secret must not travel in cleartext, whatever the host.
			{name: "plaintext http", supplied: supplied("http://api.1password.com/token")},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				_, err := clientCredentialsTokenURL(
					registry,
					coredata.ConnectorProviderOnePassword,
					tc.supplied,
				)
				assert.Error(t, err)
			})
		}
	})
}

// TestPinnedClientCredentialsTokenURL covers the value the connect form reads
// to decide whether to render a Token URL field at all, so the form and the
// mutation cannot disagree about who supplies it.
func TestPinnedClientCredentialsTokenURL(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	atlas, ok := registry.Get(coredata.ConnectorProviderMongoDBAtlas)
	require.True(t, ok)
	assert.Equal(t, "https://cloud.mongodb.com/api/oauth/token", pinnedClientCredentialsTokenURL(atlas))

	onePassword, ok := registry.Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok)
	assert.Empty(t, pinnedClientCredentialsTokenURL(onePassword))

	// A provider with no client-credentials path pins nothing, whatever its
	// registration happens to declare.
	slack, ok := registry.Get(coredata.ConnectorProviderSlack)
	require.True(t, ok)
	assert.Empty(t, pinnedClientCredentialsTokenURL(slack))
}

func TestApiKeyConnectorSettings_InstanceBaseURL(t *testing.T) {
	t.Parallel()

	providers := []struct {
		name     string
		provider coredata.ConnectorProvider
		set      func(*types.CreateAPIKeyConnectorInput, *string)
	}{
		{
			name:     "grafana",
			provider: coredata.ConnectorProviderGrafana,
			set: func(input *types.CreateAPIKeyConnectorInput, value *string) {
				input.GrafanaBaseURL = value
			},
		},
		{
			name:     "signoz",
			provider: coredata.ConnectorProviderSigNoz,
			set: func(input *types.CreateAPIKeyConnectorInput, value *string) {
				input.SignozBaseURL = value
			},
		},
		{
			name:     "langfuse",
			provider: coredata.ConnectorProviderLangfuse,
			set: func(input *types.CreateAPIKeyConnectorInput, value *string) {
				input.LangfuseBaseURL = value
			},
		},
		{
			name:     "retool",
			provider: coredata.ConnectorProviderRetool,
			set: func(input *types.CreateAPIKeyConnectorInput, value *string) {
				input.RetoolBaseURL = value
			},
		},
		{
			name:     "authentik",
			provider: coredata.ConnectorProviderAuthentik,
			set: func(input *types.CreateAPIKeyConnectorInput, value *string) {
				input.AuthentikBaseURL = value
			},
		},
	}

	for _, tc := range providers {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clean := "https://" + tc.name + ".example.com"
			stored := func(t *testing.T, raw string) string {
				t.Helper()

				input := types.CreateAPIKeyConnectorInput{Provider: tc.provider}
				tc.set(&input, &raw)

				encoded, err := apiKeyConnectorSettings(input)
				require.NoError(t, err)

				var settings struct {
					BaseURL string `json:"base_url"`
				}
				require.NoError(t, json.Unmarshal(encoded, &settings))

				return settings.BaseURL
			}

			assert.Equal(t, clean, stored(t, clean))
			assert.Equal(t, clean, stored(t, "  "+clean+"/  "))
			assert.Equal(t, clean+"/if/admin", stored(t, clean+"/if/admin/"))

			for _, raw := range []string{
				clean + "?x=1",
				clean + "#frag",
				clean + "/?x=1#frag",
				clean + "?",
			} {
				input := types.CreateAPIKeyConnectorInput{Provider: tc.provider}
				tc.set(&input, &raw)

				_, err := apiKeyConnectorSettings(input)
				require.Error(t, err)
			}
		})
	}

	t.Run("retool cloud", func(t *testing.T) {
		t.Parallel()

		blank := ""
		spaces := "   "

		for _, raw := range []*string{nil, &blank, &spaces} {
			encoded, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
				Provider:      coredata.ConnectorProviderRetool,
				RetoolBaseURL: raw,
			})
			require.NoError(t, err)

			var settings coredata.RetoolConnectorSettings
			require.NoError(t, json.Unmarshal(encoded, &settings))
			assert.Empty(t, settings.BaseURL)
		}
	})
}
