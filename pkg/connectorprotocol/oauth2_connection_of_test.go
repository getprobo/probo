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

func TestOAuth2ConnectionOf(t *testing.T) {
	t.Parallel()

	t.Run("returns the oauth2 connection itself", func(t *testing.T) {
		t.Parallel()

		conn := &connectorprotocol.OAuth2Connection{AccessToken: "plain"}

		assert.Same(t, conn, connectorprotocol.OAuth2ConnectionOf(conn))
	})

	t.Run("returns the credential embedded in a slack connection", func(t *testing.T) {
		t.Parallel()

		conn := &connectorprotocol.SlackConnection{
			AccessToken:  "xoxe.xoxp-test",
			RefreshToken: "xoxe-1-refresh",
		}

		got := connectorprotocol.OAuth2ConnectionOf(conn)
		require.NotNil(t, got)
		assert.Same(t, &conn.OAuth2Connection, got)

		got.AccessToken = "xoxe.xoxp-refreshed"
		assert.Equal(t, "xoxe.xoxp-refreshed", conn.AccessToken)
	})

	t.Run("returns nil for a non-oauth2 connection", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, connectorprotocol.OAuth2ConnectionOf(&connectorprotocol.APIKeyConnection{}))
		assert.Nil(t, connectorprotocol.OAuth2ConnectionOf(nil))
	})
}
