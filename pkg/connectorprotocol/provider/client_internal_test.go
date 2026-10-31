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

package provider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAlignClientCredentials(t *testing.T) {
	t.Parallel()

	registry := NewBuiltinRegistry()
	expiresAt := time.Now().Add(10 * time.Minute)

	onePassword := func(t *testing.T, conn *connectorprotocol.OAuth2Connection) *coredata.Connector {
		t.Helper()

		c := &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword, Connection: conn}
		require.NoError(
			t,
			c.SetSettings(
				&coredata.OnePasswordUsersAPISettings{
					AccountID: "WX52YXL2MVEPBK43P3R4XMWKXI",
					Region:    "EU",
				},
			),
		)

		return c
	}

	t.Run(
		"stored legacy endpoint is replaced and its token dropped",
		func(t *testing.T) {
			t.Parallel()

			conn := &connectorprotocol.OAuth2Connection{
				GrantType:   connectorprotocol.OAuth2GrantTypeClientCredentials,
				TokenURL:    "https://api.1password.eu/v1beta1/users/oauth2/token",
				AccessToken: "ory_at_legacy",
				ExpiresAt:   expiresAt,
				Scope:       "openid",
			}

			require.NoError(t, registry.alignClientCredentials(onePassword(t, conn), conn))

			assert.Equal(t, "https://api.1password.eu/v1/oauth/token", conn.TokenURL)
			assert.Empty(t, conn.AccessToken)
			assert.True(t, conn.ExpiresAt.IsZero())
			assert.Empty(t, conn.Scope)
		},
	)

	t.Run(
		"derived endpoint keeps its token",
		func(t *testing.T) {
			t.Parallel()

			conn := &connectorprotocol.OAuth2Connection{
				GrantType:   connectorprotocol.OAuth2GrantTypeClientCredentials,
				TokenURL:    "https://api.1password.eu/v1/oauth/token",
				AccessToken: "op_o_c_current",
				ExpiresAt:   expiresAt,
			}

			require.NoError(t, registry.alignClientCredentials(onePassword(t, conn), conn))

			assert.Equal(t, "op_o_c_current", conn.AccessToken)
			assert.Equal(t, expiresAt, conn.ExpiresAt)
		},
	)

	t.Run(
		"pinned endpoint is left alone",
		func(t *testing.T) {
			t.Parallel()

			conn := &connectorprotocol.OAuth2Connection{
				GrantType:   connectorprotocol.OAuth2GrantTypeClientCredentials,
				TokenURL:    "https://cloud.mongodb.com/api/oauth/token",
				AccessToken: "atlas-token",
				Scope:       "custom",
			}
			c := &coredata.Connector{Provider: coredata.ConnectorProviderMongoDBAtlas, Connection: conn}

			require.NoError(t, registry.alignClientCredentials(c, conn))

			assert.Equal(t, "https://cloud.mongodb.com/api/oauth/token", conn.TokenURL)
			assert.Equal(t, "atlas-token", conn.AccessToken)
			assert.Equal(t, "custom", conn.Scope)
		},
	)
}
