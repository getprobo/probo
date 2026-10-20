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
			nil,
			coredata.ConnectorProviderSlack,
			conn,
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
			nil,
			coredata.ConnectorProviderSlack,
			conn,
		)
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}
