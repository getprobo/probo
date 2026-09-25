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

package connector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
)

func TestConfiguredProtocols(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()

	assert.Empty(t, registry.ConfiguredProtocols("GITHUB"))

	require.NoError(t, registry.Register("GITHUB", &connector.OAuth2Connector{}))
	assert.Equal(
		t,
		[]connector.ProtocolType{connector.ProtocolOAuth2},
		registry.ConfiguredProtocols("GITHUB"),
	)

	require.NoError(
		t,
		registry.RegisterProtocol(
			"GITHUB",
			connector.ProtocolGitHubApp,
			&connector.GitHubAppConnector{},
		),
	)
	assert.Equal(
		t,
		[]connector.ProtocolType{
			connector.ProtocolGitHubApp,
			connector.ProtocolOAuth2,
		},
		registry.ConfiguredProtocols("GITHUB"),
	)
}

func TestConnectorRegistry_ConfigureConnection(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	require.NoError(
		t,
		registry.RegisterProtocol(
			connector.GitHubProvider,
			connector.ProtocolGitHubApp,
			&connector.GitHubAppConnector{
				AppID:      "current-app-id",
				PrivateKey: "current-private-key",
			},
		),
	)

	conn := &connector.GitHubAppConnection{
		InstallationID: 42,
		APIBase:        "https://api.github.com",
	}
	require.NoError(t, registry.ConfigureConnection(connector.GitHubProvider, conn))

	assert.Equal(t, "current-app-id", conn.AppID)
	assert.Equal(t, "current-private-key", conn.PrivateKey)
	assert.Equal(t, int64(42), conn.InstallationID)
}
