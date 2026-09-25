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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func TestSlackEventReturnsURLVerificationChallenge(t *testing.T) {
	t.Parallel()

	body := `{"type":"url_verification","challenge":"challenge-token"}`
	response := performSlackEvent(t, body)

	assert.Equal(t, http.StatusOK, response.Code)

	var payload map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	assert.Equal(t, "challenge-token", payload["challenge"])
}

func TestSlackEventRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	response := performSlackEvent(t, `{"type":`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func performSlackEvent(
	t *testing.T,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/events",
		strings.NewReader(body),
	)

	response := httptest.NewRecorder()
	slackbot := slackchannel.NewService(nil, nil, nil, log.NewLogger())
	SlackEventHandler(slackbot, log.NewLogger()).ServeHTTP(response, req)

	return response
}
