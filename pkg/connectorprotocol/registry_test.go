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

package connectorprotocol_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connectorprotocol"
)

func TestConfiguredProtocols(t *testing.T) {
	t.Parallel()

	registry := connectorprotocol.NewConnectorRegistry()

	assert.Empty(t, registry.ConfiguredProtocols("GITHUB"))

	require.NoError(t, registry.Register("GITHUB", &connectorprotocol.OAuth2Connector{}))
	assert.Equal(
		t,
		[]connectorprotocol.ProtocolType{connectorprotocol.ProtocolOAuth2},
		registry.ConfiguredProtocols("GITHUB"),
	)

	require.NoError(
		t,
		registry.RegisterProtocol(
			"GITHUB",
			connectorprotocol.ProtocolGitHubApp,
			&connectorprotocol.GitHubAppConnector{},
		),
	)
	assert.Equal(
		t,
		[]connectorprotocol.ProtocolType{
			connectorprotocol.ProtocolGitHubApp,
			connectorprotocol.ProtocolOAuth2,
		},
		registry.ConfiguredProtocols("GITHUB"),
	)
}

func TestConnectorRegistry_ConfigureConnection(t *testing.T) {
	t.Parallel()

	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(
		t,
		registry.RegisterProtocol(
			connectorprotocol.GitHubProvider,
			connectorprotocol.ProtocolGitHubApp,
			&connectorprotocol.GitHubAppConnector{
				AppID:      "current-app-id",
				PrivateKey: "current-private-key",
			},
		),
	)

	conn := &connectorprotocol.GitHubAppConnection{
		InstallationID: 42,
		APIBase:        "https://api.github.com",
	}
	require.NoError(t, registry.ConfigureConnection(connectorprotocol.GitHubProvider, conn))

	assert.Equal(t, "current-app-id", conn.AppID)
	assert.Equal(t, "current-private-key", conn.PrivateKey)
	assert.Equal(t, int64(42), conn.InstallationID)
}

func TestConnectorRegistry_GetOAuth2RefreshConfig(t *testing.T) {
	t.Parallel()

	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(
		t,
		registry.Register(
			"PLANETSCALE",
			&connectorprotocol.OAuth2Connector{
				ClientID:      "cid",
				ClientSecret:  "secret",
				TokenURL:      "https://auth.planetscale.com/oauth/token",
				RefreshScopes: []string{"read_organizations", "read_user"},
			},
		),
	)

	cfg := registry.GetOAuth2RefreshConfig("PLANETSCALE")
	require.NotNil(t, cfg)
	assert.Equal(t, "cid", cfg.ClientID)
	assert.Equal(t, "secret", cfg.ClientSecret)
	assert.Equal(t, "https://auth.planetscale.com/oauth/token", cfg.TokenURL)
	assert.Equal(t, []string{"read_organizations", "read_user"}, cfg.RefreshScopes)

	assert.Nil(t, registry.GetOAuth2RefreshConfig("UNKNOWN"))
}
