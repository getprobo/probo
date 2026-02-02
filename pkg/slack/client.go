// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

const (
	slackAPIPostMessage      = "https://slack.com/api/chat.postMessage"
	slackAPIUpdateMessage    = "https://slack.com/api/chat.update"
	slackAPIConversationJoin = "https://slack.com/api/conversations.join"
	slackWebhookHost         = "hooks.slack.com"
)

type (
	Client struct {
		httpClient *http.Client
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

func NewClient(logger *log.Logger) *Client {
	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
	}

	return &Client{
		httpClient: httpclient.DefaultPooledClient(httpClientOpts...),
	}
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPIPostMessage, &buf)
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPIUpdateMessage, &buf)
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPIConversationJoin, &buf)
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
