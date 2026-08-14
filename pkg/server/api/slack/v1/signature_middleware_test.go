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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

const testSigningSecret = "slack-signing-secret"

func TestSignatureMiddleware_RejectsMissingHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("payload"))
	response := serveWithSignature(t, okHandler(), req, testSigningSecret)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestSignatureMiddleware_RejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("payload"))
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0=deadbeef")

	response := serveWithSignature(t, okHandler(), req, testSigningSecret)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestSignatureMiddleware_RejectsOversizedPayload(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(strings.Repeat("a", maxSlackRequestBytes+1)),
	)
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0=deadbeef")

	response := serveWithSignature(t, okHandler(), req, testSigningSecret)

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
}

func TestSignatureMiddleware_AcceptsMatchingSecret(t *testing.T) {
	t.Parallel()

	const body = "payload"
	var got []byte
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			got = body
			w.WriteHeader(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	signSlackRequest(t, req, body, testSigningSecret)

	response := serveWithSignature(t, handler, req, testSigningSecret)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, body, string(got))
}

func TestSignatureMiddleware_AcceptsAnyConfiguredSecret(t *testing.T) {
	t.Parallel()

	const (
		body           = "payload"
		slackbotSecret = "slackbot-signing-secret"
		legacySecret   = "legacy-signing-secret"
	)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	signSlackRequest(t, req, body, legacySecret)

	response := serveWithSignature(t, okHandler(), req, slackbotSecret, legacySecret)

	assert.Equal(t, http.StatusOK, response.Code)
}

func TestSignatureMiddleware_RejectsUnconfiguredSecret(t *testing.T) {
	t.Parallel()

	const (
		body           = "payload"
		slackbotSecret = "slackbot-signing-secret"
		legacySecret   = "legacy-connector-signing-secret"
	)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	signSlackRequest(t, req, body, legacySecret)

	response := serveWithSignature(t, okHandler(), req, slackbotSecret)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func serveWithSignature(
	t *testing.T,
	handler http.Handler,
	req *http.Request,
	secrets ...string,
) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	newSignatureMiddleware(log.NewLogger(), secrets...)(handler).ServeHTTP(response, req)

	return response
}

func signSlackRequest(t *testing.T, req *http.Request, body, secret string) {
	t.Helper()

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte("v0:" + timestamp + ":" + body))
	require.NoError(t, err)

	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))
}
