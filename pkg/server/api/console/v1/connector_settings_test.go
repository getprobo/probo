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
	"strings"
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
	require.Len(t, reg.APIKeyExtraSettings, 1)
	require.Equal(t, "baseUrl", reg.APIKeyExtraSettings[0].Key)

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

// TestApiKeyConnectorSettings_UniFiConsoleID walks the same chain for UniFi and
// pins the extra rule its setting carries: the console ID becomes a path
// segment on api.ui.com, so a value bringing its own separators would retarget
// every request the driver makes and must be refused rather than sanitized.
func TestApiKeyConnectorSettings_UniFiConsoleID(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderUniFi)
	require.True(t, ok)
	require.Len(t, reg.APIKeyExtraSettings, 1)
	require.Equal(t, "consoleId", reg.APIKeyExtraSettings[0].Key)

	consoleID := "ABCDEF0123456789:1234567890"

	raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:       coredata.ConnectorProviderUniFi,
		UnifiConsoleID: &consoleID,
	})
	require.NoError(t, err)

	var settings coredata.UniFiConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, consoleID, settings.ConsoleID)

	// Padding is a paste artifact, not a different console.
	padded := "  " + consoleID + "  "

	raw, err = apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:       coredata.ConnectorProviderUniFi,
		UnifiConsoleID: &padded,
	})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, consoleID, settings.ConsoleID)

	for _, invalid := range []string{
		"",
		"   ",
		// A pasted URL rather than the bare identifier.
		"https://api.ui.com/v1/connector/consoles/ABCDEF:1",
		"ABCDEF:1/proxy/network/integration",
		"../../v1/hosts",
		"ABCDEF:1?x=1",
		"ABCDEF:1#frag",
		"ABC DEF:1",
	} {
		t.Run("rejects "+invalid, func(t *testing.T) {
			t.Parallel()

			_, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
				Provider:       coredata.ConnectorProviderUniFi,
				UnifiConsoleID: &invalid,
			})
			require.Error(t, err)

			// The message is surfaced verbatim to the client, so it must name
			// the field and never echo the value.
			if strings.TrimSpace(invalid) != "" {
				assert.NotContains(t, err.Error(), invalid)
			}
		})
	}

	_, err = apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider: coredata.ConnectorProviderUniFi,
	})
	require.Error(t, err)
}

func TestApiKeyConnectorSettings_OnePasswordSCIMBridgeURL(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok)
	require.Len(t, reg.APIKeyExtraSettings, 1)
	require.Equal(t, "scimBridgeUrl", reg.APIKeyExtraSettings[0].Key)

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
	require.Len(t, reg.ClientCredentialsExtraSettings, 2)
	require.Equal(t, "accountId", reg.ClientCredentialsExtraSettings[0].Key)
	require.Equal(t, "region", reg.ClientCredentialsExtraSettings[1].Key)

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
