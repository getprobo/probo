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
	"encoding/json"
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
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type fakeSlashCommander struct {
	cmd      slackchannel.SlashCommand
	response slackchannel.SlashCommandResponse
}

func (f *fakeSlashCommander) HandleSlashCommand(
	_ context.Context,
	cmd slackchannel.SlashCommand,
) slackchannel.SlashCommandResponse {
	f.cmd = cmd

	return f.response
}

func TestSlackCommandReturnsEphemeralBindResponse(t *testing.T) {
	t.Parallel()

	commander := &fakeSlashCommander{
		response: slackchannel.SlashCommandResponse{
			ResponseType: slackchannel.SlashResponseTypeEphemeral,
			Text:         "Link your Probo account",
		},
	}
	body := url.Values{
		"command":     []string{"/probot"},
		"text":        []string{"bind"},
		"user_id":     []string{"U123"},
		"user_name":   []string{"ada"},
		"team_id":     []string{"T123"},
		"team_domain": []string{"acme"},
	}.Encode()
	response := performSlackCommand(t, commander, body, testSigningSecret)

	assert.Equal(t, http.StatusOK, response.Code)

	var payload slackchannel.SlashCommandResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	assert.Equal(t, slackchannel.SlashResponseTypeEphemeral, payload.ResponseType)
	assert.Equal(t, "Link your Probo account", payload.Text)
	assert.Equal(t, "/probot", commander.cmd.Command)
	assert.Equal(t, "bind", commander.cmd.Text)
	assert.Equal(t, "U123", commander.cmd.UserID)
	assert.Equal(t, "ada", commander.cmd.UserName)
	assert.Equal(t, "T123", commander.cmd.TeamID)
	assert.Equal(t, "acme", commander.cmd.TeamDomain)
}

func TestSlackCommandUnavailableWhenDisabled(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"command": []string{"/probot"},
		"text":    []string{"bind"},
		"user_id": []string{"U123"},
		"team_id": []string{"T123"},
	}.Encode()
	response := performSlackCommand(t, nil, body, testSigningSecret)

	assert.Equal(t, http.StatusOK, response.Code)

	var payload slackchannel.SlashCommandResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	assert.Equal(t, slackchannel.SlashResponseTypeEphemeral, payload.ResponseType)
	assert.Contains(t, payload.Text, "not available")
}

func TestSlackCommandRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	commander := &fakeSlashCommander{}
	body := url.Values{
		"command": []string{"/probot"},
		"text":    []string{"bind"},
		"user_id": []string{"U123"},
		"team_id": []string{"T123"},
	}.Encode()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/commands",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0=deadbeef")

	response := httptest.NewRecorder()
	SlackCommandHandler(
		commander,
		[]string{testSigningSecret},
		log.NewLogger(),
	).ServeHTTP(response, req)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, commander.cmd.UserID)

	var payload slackchannel.SlashCommandResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	assert.Equal(t, slackchannel.SlashResponseTypeEphemeral, payload.ResponseType)
}

func TestSlackCommandRejectsLegacyConnectorSecret(t *testing.T) {
	t.Parallel()

	const (
		slackbotSecret = "slackbot-signing-secret"
		legacySecret   = "legacy-connector-signing-secret"
	)

	commander := &fakeSlashCommander{
		response: slackchannel.SlashCommandResponse{
			ResponseType: slackchannel.SlashResponseTypeEphemeral,
			Text:         "ok",
		},
	}
	body := url.Values{
		"command": []string{"/probot"},
		"text":    []string{"bind"},
		"user_id": []string{"U123"},
		"team_id": []string{"T123"},
	}.Encode()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/commands",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(legacySecret))
	_, err := mac.Write([]byte("v0:" + timestamp + ":" + body))
	require.NoError(t, err)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))

	response := httptest.NewRecorder()
	SlackCommandHandler(
		commander,
		[]string{slackbotSecret},
		log.NewLogger(),
	).ServeHTTP(response, req)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, commander.cmd.UserID)
}

func performSlackCommand(
	t *testing.T,
	commander slackSlashCommander,
	body string,
	signingSecret string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/commands",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, err := mac.Write([]byte("v0:" + timestamp + ":" + body))
	require.NoError(t, err)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))

	response := httptest.NewRecorder()
	SlackCommandHandler(
		commander,
		[]string{signingSecret},
		log.NewLogger(),
	).ServeHTTP(response, req)

	return response
}
