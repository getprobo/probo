// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package slackbot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

func signRequest(secret string, body []byte, ts int64) (timestamp, signature string) {
	timestamp = strconv.FormatInt(ts, 10)
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(baseString))

	return timestamp, "v0=" + hex.EncodeToString(h.Sum(nil))
}

func testLogger(t *testing.T) *log.Logger {
	t.Helper()

	return log.NewLogger()
}

func TestHandler_URLVerification(t *testing.T) {
	t.Parallel()

	secret := "test-signing-secret"
	body := []byte(`{"type":"url_verification","challenge":"challenge-token"}`)
	ts, sig := signRequest(secret, body, time.Now().Unix())

	handler := NewHandler(secret, nil, nil, nil, nil, t.Context(), testLogger(t))

	req := httptest.NewRequest(http.MethodPost, "/events", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("X-Slack-Request-Timestamp", ts)
	req.Header.Set("X-Slack-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "challenge-token")
}

func TestHandler_InvalidSignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"type":"event_callback","event_id":"E1","event":{"type":"app_mention","text":"hi"}}`)
	handler := NewHandler("secret", nil, nil, nil, nil, t.Context(), testLogger(t))

	req := httptest.NewRequest(http.MethodPost, "/events", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0=deadbeef")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_EventCallbackAck(t *testing.T) {
	t.Parallel()

	secret := "test-signing-secret"
	body := []byte(`{"type":"event_callback","event_id":"E1","event":{"type":"message","channel_type":"channel","text":"hi","channel":"C1","ts":"1.1"}}`)
	ts, sig := signRequest(secret, body, time.Now().Unix())

	handler := NewHandler(secret, nil, nil, nil, nil, t.Context(), testLogger(t))

	req := httptest.NewRequest(http.MethodPost, "/events", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("X-Slack-Request-Timestamp", ts)
	req.Header.Set("X-Slack-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCleanText(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hello", cleanText("<@U123> hello"))
	assert.Equal(t, "", cleanText("   "))
}

func TestBuildMessages(t *testing.T) {
	t.Parallel()

	messages := buildMessages(
		[]interaction{
			{
				Payload: EventBody{
					Channel:  "C1",
					TS:       "111.222",
					ThreadTS: "111.111",
					User:     "U1",
					Text:     "<@BOT> help me",
				},
			},
		},
	)

	require.Len(t, messages, 1)
	assert.Equal(t, llm.RoleUser, messages[0].Role)
	assert.Contains(t, messages[0].Text(), "Slack context: channel=C1")
	assert.Contains(t, messages[0].Text(), "help me")
}

func TestExtractReplyText(t *testing.T) {
	t.Parallel()

	result := &agent.Result{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}},
			{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "hello there"}}},
		},
	}

	assert.Equal(t, "hello there", extractReplyText(result))
}

func TestReplyTargetFor(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		replyTarget{channel: "C1", threadTS: "111.000"},
		replyTargetFor(EventBody{Channel: "C1", TS: "111.000"}),
	)
	assert.Equal(
		t,
		replyTarget{channel: "C1", threadTS: "222.000"},
		replyTargetFor(EventBody{Channel: "C1", TS: "111.000", ThreadTS: "222.000"}),
	)
	assert.Equal(
		t,
		replyTarget{channel: "D1", threadTS: ""},
		replyTargetFor(EventBody{Channel: "D1", ChannelType: ChannelTypeIM, TS: "111.000", ThreadTS: "222.000"}),
	)
}

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"type":"url_verification"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	_, sig := signRequest("secret", body, time.Now().Unix())

	require.NoError(t, VerifySignature("secret", ts, sig, body))
	require.Error(t, VerifySignature("secret", ts, "v0=bad", body))
}
