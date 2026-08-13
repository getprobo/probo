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

package slack_v1

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

const testSigningSecret = "slack-signing-secret"

type fakeInteractiveInbox struct {
	payloads [][]byte
	inserted bool
	err      error
}

func (f *fakeInteractiveInbox) Enqueue(
	_ context.Context,
	payload []byte,
) (bool, error) {
	f.payloads = append(f.payloads, append([]byte(nil), payload...))

	return f.inserted, f.err
}

func TestSlackInteractiveAcknowledgesAfterDurableInsert(t *testing.T) {
	t.Parallel()

	inbox := &fakeInteractiveInbox{inserted: true}
	response := performSlackAction(t, inbox)

	assert.Equal(t, http.StatusOK, response.Code)
	require.Len(t, inbox.payloads, 1)
	assert.Contains(t, string(inbox.payloads[0]), `"action_id":"accept_all"`)
}

func TestSlackInteractiveAcknowledgesDuplicate(t *testing.T) {
	t.Parallel()

	inbox := &fakeInteractiveInbox{inserted: false}
	response := performSlackAction(t, inbox)

	assert.Equal(t, http.StatusOK, response.Code)
	require.Len(t, inbox.payloads, 1)
}

func performSlackAction(
	t *testing.T,
	inbox slackInteractiveCommandInbox,
) *httptest.ResponseRecorder {
	t.Helper()

	rawPayload := `{"team":{"id":"T123"},"user":{"id":"U123"},"container":{"channel_id":"C123","message_ts":"123.456"},"actions":[{"action_id":"accept_all","action_ts":"123.789","value":"message-id"}]}`
	body := url.Values{"payload": []string{rawPayload}}.Encode()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/interactive",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(testSigningSecret))
	_, err := mac.Write([]byte("v0:" + timestamp + ":" + body))
	require.NoError(t, err)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))

	response := httptest.NewRecorder()
	SlackHandler(
		inbox,
		[]string{testSigningSecret},
		log.NewLogger(),
	).ServeHTTP(response, req)

	return response
}

func TestSlackInteractiveAcceptsLegacySigningSecret(t *testing.T) {
	t.Parallel()

	const (
		slackbotSecret = "slackbot-signing-secret"
		legacySecret   = "legacy-signing-secret"
	)

	rawPayload := `{"team":{"id":"T123"},"user":{"id":"U123"},"container":{"channel_id":"C123","message_ts":"123.456"},"actions":[{"action_id":"accept_all","action_ts":"123.789","value":"message-id"}]}`
	body := url.Values{"payload": []string{rawPayload}}.Encode()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/interactive",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(legacySecret))
	_, err := mac.Write([]byte("v0:" + timestamp + ":" + body))
	require.NoError(t, err)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))

	inbox := &fakeInteractiveInbox{inserted: true}
	response := httptest.NewRecorder()
	SlackHandler(
		inbox,
		[]string{slackbotSecret, legacySecret},
		log.NewLogger(),
	).ServeHTTP(response, req)

	assert.Equal(t, http.StatusOK, response.Code)
	require.Len(t, inbox.payloads, 1)
}
