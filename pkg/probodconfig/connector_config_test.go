// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package probodconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/probodconfig"
)

// TestConnectorConfig_APIKeyRoundTrip pins the bootstrap-to-probod path
// for a ManagedAPIKey (Model B) connector: bootstrap emits an api_key
// ConnectorConfig, it is marshalled to JSON, and probod's UnmarshalJSON
// must recover the key on ConnectorConfig.APIKey. This is what lets the
// Crisp plugin token reach the provider registry.
func TestConnectorConfig_APIKeyRoundTrip(t *testing.T) {
	t.Parallel()

	original := probodconfig.ConnectorConfig{
		Provider: "CRISP",
		Protocol: connector.ProtocolType("api_key"),
		RawConfig: probodconfig.ConnectorConfigAPIKey{
			APIKey:     "identifier:secret",
			ResourceID: "plugin-id",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got probodconfig.ConnectorConfig
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "CRISP", got.Provider)
	assert.Equal(t, connector.ProtocolAPIKey, got.Protocol)
	assert.Equal(t, "identifier:secret", got.APIKey)
	assert.Equal(t, "plugin-id", got.ResourceID)
	assert.Nil(t, got.Config, "api_key connectors carry no OAuth2 Connector")
}

// TestConnectorConfig_OAuth2RoundTrip guards the pre-existing OAuth2
// path against regressions from the added api_key branch.
func TestConnectorConfig_OAuth2RoundTrip(t *testing.T) {
	t.Parallel()

	original := probodconfig.ConnectorConfig{
		Provider:  "SLACK",
		Protocol:  connector.ProtocolType("oauth2"),
		RawConfig: probodconfig.ConnectorConfigOAuth2{ClientID: "cid", ClientSecret: "secret"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got probodconfig.ConnectorConfig
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "SLACK", got.Provider)
	assert.Equal(t, connector.ProtocolOAuth2, got.Protocol)

	oauth2c, ok := got.Config.(*connector.OAuth2Connector)
	require.True(t, ok)
	assert.Equal(t, "cid", oauth2c.ClientID)
	assert.Equal(t, "secret", oauth2c.ClientSecret)
}
