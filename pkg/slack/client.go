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

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

// Slack Web API methods, joined onto the API base the caller supplies. They
// are paths rather than absolute URLs because the notification client talks to
// the workspace the SLACK connector row was minted against: its access token
// comes from that connector's Endpoints.Token, so a deployment that repoints
// the Slack provider must move these calls with it, or a sandbox-issued token
// would travel to the real slack.com.
const (
	methodPostMessage      = "chat.postMessage"
	methodUpdateMessage    = "chat.update"
	methodConversationJoin = "conversations.join"
)

// slackWebhookHost is an inbound host CHECK, not an outbound endpoint.
// UpdateInteractiveMessage posts to the response_url carried in the
// interaction payload Slack sends us, so this is the guard that stops a forged
// payload from making Probo POST to an attacker-chosen host. It is
// deliberately NOT derived from the API base: Slack serves response URLs from
// hooks.slack.com, a different host from the API's, and widening the check to
// follow an endpoint override would trade an SSRF guard for configurability.
const slackWebhookHost = "hooks.slack.com"

type (
	Client struct {
		httpClient *http.Client
		// apiBaseURL is the Slack Web API root (the SLACK registration's
		// Endpoints.APIBase), scheme-ful and without a trailing slash.
		apiBaseURL string
	}

	SlackResponse struct {
		OK      bool   `json:"ok,omitempty"`
		TS      string `json:"ts,omitempty"`
		Channel string `json:"channel,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	SlackJoinResponse struct {
		OK      bool            `json:"ok,omitempty"`
		Channel json.RawMessage `json:"channel,omitempty"`
		Error   string          `json:"error,omitempty"`
	}
)

// NewClient returns a Slack Web API client rooted at apiBaseURL, which callers
// take from the SLACK provider registration's Endpoints.APIBase rather than
// pinning, so an endpoint override moves notifications along with the OAuth
// handshake that mints the token they carry.
func NewClient(apiBaseURL string, logger *log.Logger) *Client {
	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
	}

	return &Client{
		httpClient: httpclient.DefaultPooledClient(httpClientOpts...),
		apiBaseURL: apiBaseURL,
	}
}

// methodURL joins a Web API method onto the client's base.
func (c *Client) methodURL(method string) (string, error) {
	u, err := url.JoinPath(c.apiBaseURL, method)
	if err != nil {
		return "", fmt.Errorf("cannot build Slack %s URL: %w", method, err)
	}

	return u, nil
}

func (c *Client) CreateMessage(ctx context.Context, accessToken string, channelID string, body map[string]any) (*SlackResponse, error) {
	payload := map[string]any{
		"channel": channelID,
		"text":    body["text"],
		"blocks":  body["blocks"],
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("cannot marshal message: %w", err)
	}

	endpoint, err := c.methodURL(methodPostMessage)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cannot send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody))
	}

	var slackResponse SlackResponse

	if err := json.Unmarshal(responseBody, &slackResponse); err != nil {
		return nil, fmt.Errorf("cannot parse Slack response: %w (body: %s)", err, string(responseBody))
	}

	if !slackResponse.OK {
		return nil, fmt.Errorf("slack API error: %s (channel: %s, response: %s)", slackResponse.Error, channelID, string(responseBody))
	}

	return &slackResponse, nil
}

func (c *Client) UpdateInteractiveMessage(ctx context.Context, responseURL string, body map[string]any) error {
	if err := validateSlackResponseURL(responseURL); err != nil {
		return fmt.Errorf("invalid Slack response URL: %w", err)
	}

	updatePayload := map[string]any{
		"replace_original": true,
		"text":             body["text"],
		"blocks":           body["blocks"],
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(updatePayload); err != nil {
		return fmt.Errorf("cannot marshal interactive message update: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, responseURL, &buf)
	if err != nil {
		return fmt.Errorf("cannot create interactive message update request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cannot send interactive message update request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody))
	}

	// Slack can return either plain text "ok" or JSON {"ok":true}
	bodyStr := string(responseBody)
	if bodyStr == "ok" || bodyStr == "" {
		return nil
	}

	var slackResponse SlackResponse
	if err := json.Unmarshal(responseBody, &slackResponse); err == nil {
		if slackResponse.OK {
			return nil
		}

		if slackResponse.Error != "" {
			return fmt.Errorf("slack error: %s", slackResponse.Error)
		}
	}

	return fmt.Errorf("unexpected Slack response: %s", bodyStr)
}

func (c *Client) UpdateMessage(ctx context.Context, accessToken string, channelID string, messageTS string, body map[string]any) error {
	payload := map[string]any{
		"channel": channelID,
		"ts":      messageTS,
		"text":    body["text"],
		"blocks":  body["blocks"],
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return fmt.Errorf("cannot marshal message: %w", err)
	}

	endpoint, err := c.methodURL(methodUpdateMessage)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody))
	}

	var slackResponse SlackResponse
	if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&slackResponse); err != nil {
		return fmt.Errorf("cannot parse Slack response: %w (body: %s)", err, string(responseBody))
	}

	if !slackResponse.OK {
		return fmt.Errorf("slack API error: %s", slackResponse.Error)
	}

	return nil
}

func (c *Client) JoinChannel(ctx context.Context, accessToken string, channelID string) error {
	payload := map[string]any{
		"channel": channelID,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return fmt.Errorf("cannot marshal request: %w", err)
	}

	endpoint, err := c.methodURL(methodConversationJoin)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody))
	}

	var slackResponse SlackJoinResponse

	if err := json.Unmarshal(responseBody, &slackResponse); err != nil {
		return fmt.Errorf("cannot parse Slack response: %w (body: %s)", err, string(responseBody))
	}

	if !slackResponse.OK {
		if slackResponse.Error == "already_in_channel" {
			return nil
		}

		if slackResponse.Error == "channel_not_found" || slackResponse.Error == "is_private" {
			return fmt.Errorf("cannot join private channel - bot must be invited manually")
		}

		return fmt.Errorf("slack API error: %s", slackResponse.Error)
	}

	return nil
}

func validateSlackResponseURL(responseURL string) error {
	parsedURL, err := url.Parse(responseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: must be https")
	}

	if parsedURL.Host != slackWebhookHost {
		return fmt.Errorf("invalid URL host: must be %s", slackWebhookHost)
	}

	return nil
}
