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
