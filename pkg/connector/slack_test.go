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

package connector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestParseSlackTokenResponse(t *testing.T) {
	t.Parallel()

	orgID := gid.New(gid.NewTenantID(), 0)
	base := OAuth2Connection{
		AccessToken: "xoxb-test",
		TokenType:   "bot",
		ExpiresAt:   time.Now().Add(time.Hour),
		Scope:       "chat:write channels:join incoming-webhook",
	}

	t.Run("with incoming webhook", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"ok":true,"incoming_webhook":{"url":"https://hooks.slack.com/services/T/B/X","channel":"#general","channel_id":"C123"}}`)

		conn, returnedOrgID, err := ParseSlackTokenResponse(body, base, orgID)
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NotNil(t, returnedOrgID)

		assert.Equal(t, orgID, *returnedOrgID)
		assert.Equal(t, "https://hooks.slack.com/services/T/B/X", conn.Settings.WebhookURL)
		assert.Equal(t, "#general", conn.Settings.Channel)
		assert.Equal(t, "C123", conn.Settings.ChannelID)
		assert.Equal(t, "xoxb-test", conn.AccessToken)
	})

	t.Run("without incoming webhook", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"ok":true}`)

		conn, returnedOrgID, err := ParseSlackTokenResponse(body, base, orgID)
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NotNil(t, returnedOrgID)

		assert.Empty(t, conn.Settings.WebhookURL)
		assert.Empty(t, conn.Settings.Channel)
		assert.Empty(t, conn.Settings.ChannelID)
		assert.Equal(t, "xoxb-test", conn.AccessToken)
	})

	t.Run("slack error response", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"ok":false,"error":"invalid_code"}`)

		conn, returnedOrgID, err := ParseSlackTokenResponse(body, base, orgID)
		require.Error(t, err)
		assert.Nil(t, conn)
		assert.Nil(t, returnedOrgID)
		assert.ErrorContains(t, err, "invalid_code")
	})

	t.Run("missing access token", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"ok":true}`)
		connWithoutToken := base
		connWithoutToken.AccessToken = ""

		conn, returnedOrgID, err := ParseSlackTokenResponse(body, connWithoutToken, orgID)
		require.Error(t, err)
		assert.Nil(t, conn)
		assert.Nil(t, returnedOrgID)
		assert.ErrorContains(t, err, "missing access token")
	})
}
