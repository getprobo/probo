// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
