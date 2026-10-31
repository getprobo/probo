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

package provider_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestHTTPClient_SlackUserToken(t *testing.T) {
	t.Parallel()

	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(t, registry.Register(string(coredata.ConnectorProviderSlack), &connectorprotocol.OAuth2Connector{
		ClientID:     "slack-client",
		ClientSecret: "slack-secret",
		TokenURL:     "http://127.0.0.1:1/token",
	}))

	t.Run("refresh token enters the refresh path", func(t *testing.T) {
		t.Parallel()

		conn := &connectorprotocol.SlackConnection{
			AccessToken:  "xoxe.xoxp-expired",
			RefreshToken: "xoxe-1-refresh",
			ExpiresAt:    time.Now().Add(-time.Hour),
			TokenType:    connectorprotocol.SlackTokenTypeUser,
		}

		_, err := provider.HTTPClient(
			context.Background(),
			registry,
			provider.NewRegistry(),
			&coredata.Connector{Provider: coredata.ConnectorProviderSlack, Connection: conn},
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot refresh token")
	})

	t.Run("token without a refresh token stays static", func(t *testing.T) {
		t.Parallel()

		conn := &connectorprotocol.SlackConnection{
			AccessToken: "xoxb-bot",
			TokenType:   "bot",
		}

		client, err := provider.HTTPClient(
			context.Background(),
			registry,
			provider.NewRegistry(),
			&coredata.Connector{Provider: coredata.ConnectorProviderSlack, Connection: conn},
		)
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

// HTTPClient applies the token endpoint a registration derives before any
// exchange, so a row stored with another endpoint never reuses its token.
func TestHTTPClient_DerivedClientCredentialsTokenURL(t *testing.T) {
	t.Parallel()

	// Loopback, which the exchange refuses to dial: a failed exchange proves
	// the client went to the derived endpoint instead of reusing the token.
	const derived = "https://127.0.0.1:1/v1/oauth/token"

	registry := provider.NewRegistry()
	require.NoError(
		t,
		registry.Register(
			&provider.Registration{
				Provider:    coredata.ConnectorProviderOnePassword,
				DisplayName: "1Password",
				ClientCredentials: &provider.ClientCredentialsConfig{
					TokenURL: func(*coredata.Connector) (string, error) {
						return derived, nil
					},
				},
			},
		),
	)

	newConnector := func(conn *connectorprotocol.OAuth2Connection) *coredata.Connector {
		return &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword, Connection: conn}
	}

	t.Run(
		"a stored legacy endpoint is replaced before the exchange",
		func(t *testing.T) {
			t.Parallel()

			conn := &connectorprotocol.OAuth2Connection{
				GrantType:    connectorprotocol.OAuth2GrantTypeClientCredentials,
				ClientID:     "client",
				ClientSecret: "secret",
				TokenURL:     "https://api.1password.com/v1beta1/users/oauth2/token",
				AccessToken:  "ory_at_legacy",
				ExpiresAt:    time.Now().Add(time.Hour),
				Scope:        "openid",
			}

			_, err := provider.HTTPClient(
				context.Background(),
				connectorprotocol.NewConnectorRegistry(),
				registry,
				newConnector(conn),
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "cannot exchange client credentials")
			assert.Equal(t, derived, conn.TokenURL)
			assert.Empty(t, conn.AccessToken)
			assert.Empty(t, conn.Scope)
		},
	)

	t.Run(
		"a token from the derived endpoint is reused",
		func(t *testing.T) {
			t.Parallel()

			conn := &connectorprotocol.OAuth2Connection{
				GrantType:   connectorprotocol.OAuth2GrantTypeClientCredentials,
				TokenURL:    derived,
				AccessToken: "op_o_c_current",
				ExpiresAt:   time.Now().Add(time.Hour),
			}

			client, err := provider.HTTPClient(
				context.Background(),
				connectorprotocol.NewConnectorRegistry(),
				registry,
				newConnector(conn),
			)
			require.NoError(t, err)
			require.NotNil(t, client)
			assert.Equal(t, "op_o_c_current", conn.AccessToken)
		},
	)
}
