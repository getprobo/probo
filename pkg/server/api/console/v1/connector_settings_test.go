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

	"go.probo.inc/probo/pkg/connectorprotocol/provider"
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

func TestClientCredentialsConnectorSettings_OnePassword(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok)
	require.Len(t, reg.ClientCredentialsExtraSettings(), 2)
	require.Equal(t, "accountId", reg.ClientCredentialsExtraSettings()[0].Key)
	require.Equal(t, "region", reg.ClientCredentialsExtraSettings()[1].Key)

	accountID, region := "WX52YXL2MVEPBK43P3R4XMWKXI", "EU"

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

	ent := "ent"
	raw, err = clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
		OnePasswordRegion:    &ent,
	})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, "ENT", settings.Region)

	_, err = clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
	})
	require.Error(t, err)

	lower := "eu"
	raw, err = clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
		OnePasswordRegion:    &lower,
	})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, "EU", settings.Region)

	unknown := "uk"
	_, err = clientCredentialsConnectorSettings(types.CreateClientCredentialsConnectorInput{
		Provider:             coredata.ConnectorProviderOnePassword,
		OnePasswordAccountID: &accountID,
		OnePasswordRegion:    &unknown,
	})
	require.Error(t, err)
}

// TestClientCredentialsTokenURL pins which endpoint a client-credentials
// connector POSTs its client secret to: always one the registration names,
// pinned or derived from the connector's settings, never one the client sends.
func TestClientCredentialsTokenURL(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	for _, tc := range []struct {
		provider coredata.ConnectorProvider
		want     string
	}{
		{provider: coredata.ConnectorProviderMongoDBAtlas, want: "https://cloud.mongodb.com/api/oauth/token"},
		{provider: coredata.ConnectorProviderOVHcloud, want: "https://www.ovh.com/auth/oauth2/token"},
		{provider: coredata.ConnectorProviderOVHcloudUS, want: "https://us.ovhcloud.com/auth/oauth2/token"},
	} {
		t.Run(
			"pinned by "+string(tc.provider),
			func(t *testing.T) {
				t.Parallel()

				got, err := clientCredentialsTokenURL(registry, &coredata.Connector{Provider: tc.provider})
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			},
		)
	}

	// 1Password's token endpoint shares the host of the Users API the
	// customer's region names.
	for _, tc := range []struct {
		region string
		want   string
	}{
		{region: "US", want: "https://api.1password.com/v1/oauth/token"},
		{region: "CA", want: "https://api.1password.ca/v1/oauth/token"},
		{region: "EU", want: "https://api.1password.eu/v1/oauth/token"},
		{region: "ENT", want: "https://api.ent.1password.com/v1/oauth/token"},
	} {
		t.Run(
			"1password derives it from region "+tc.region,
			func(t *testing.T) {
				t.Parallel()

				c := &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword}
				require.NoError(
					t,
					c.SetSettings(
						&coredata.OnePasswordUsersAPISettings{
							AccountID: "WX52YXL2MVEPBK43P3R4XMWKXI",
							Region:    tc.region,
						},
					),
				)

				got, err := clientCredentialsTokenURL(registry, c)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			},
		)
	}

	t.Run(
		"a provider without client credentials is rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := clientCredentialsTokenURL(registry, &coredata.Connector{Provider: coredata.ConnectorProviderSlack})
			assert.Error(t, err)
		},
	)

	t.Run(
		"an unknown provider is rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := clientCredentialsTokenURL(provider.NewRegistry(), &coredata.Connector{Provider: coredata.ConnectorProviderMongoDBAtlas})
			assert.Error(t, err)
		},
	)
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

func TestClientCredentialsScope(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()
	supplied := func(s string) *string { return &s }

	t.Run("a provider that declares scopes wins over client input", func(t *testing.T) {
		t.Parallel()

		// OVHcloud's token endpoint refuses a grant with no scope, so the
		// registration has to reach the exchange whatever the dialog sent.
		got := clientCredentialsScope(registry, coredata.ConnectorProviderOVHcloud, supplied("account/read"))
		assert.Equal(t, "account/all", got)
	})

	t.Run("and still wins when the client sends nothing", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "account/all", clientCredentialsScope(registry, coredata.ConnectorProviderOVHcloud, nil))
	})

	t.Run("a provider that declares none falls back to the supplied value", func(t *testing.T) {
		t.Parallel()

		// MongoDB Atlas offers client credentials without an OAuth2 block, so
		// nothing is pinned and the customer's value is authoritative.
		got := clientCredentialsScope(registry, coredata.ConnectorProviderMongoDBAtlas, supplied("  read:users  "))
		assert.Equal(t, "read:users", got, "the supplied value is trimmed")
	})

	t.Run("nil input on an unpinned provider yields no scope", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, clientCredentialsScope(registry, coredata.ConnectorProviderMongoDBAtlas, nil))
	})

	// 1Password's token endpoint takes no scope: the app's own scopes ride on
	// every token it issues.
	t.Run("a provider that derives its token endpoint sends no scope", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, clientCredentialsScope(registry, coredata.ConnectorProviderOnePassword, supplied("users.view")))
	})

	t.Run("an unknown provider cannot pin anything", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "x", clientCredentialsScope(registry, coredata.ConnectorProvider("NOPE"), supplied("x")))
	})
}
