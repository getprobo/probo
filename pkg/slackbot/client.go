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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

const (
	slackAPIPostMessage        = "https://slack.com/api/chat.postMessage"
	slackAPIConversationsOpen  = "https://slack.com/api/conversations.open"
	slackAPIReactionsAdd       = "https://slack.com/api/reactions.add"
	slackAPIAssistantSetStatus = "https://slack.com/api/assistant.threads.setStatus"
)

type (
	Client struct {
		token      string
		httpClient *http.Client
	}

	slackResponse struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	postMessageRequest struct {
		Channel  string `json:"channel"`
		Text     string `json:"text,omitempty"`
		ThreadTS string `json:"thread_ts,omitempty"`
		Blocks   []any  `json:"blocks,omitempty"`
	}

	addReactionRequest struct {
		Channel   string `json:"channel"`
		Name      string `json:"name"`
		Timestamp string `json:"timestamp"`
	}

	setStatusRequest struct {
		ChannelID string `json:"channel_id"`
		ThreadTS  string `json:"thread_ts"`
		Status    string `json:"status"`
	}

	openConversationRequest struct {
		Users string `json:"users"`
	}

	openConversationResponse struct {
		slackResponse
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}
)

func NewClient(botToken string, logger *log.Logger) *Client {
	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
	}

	return &Client{
		token:      botToken,
		httpClient: httpclient.DefaultPooledClient(httpClientOpts...),
	}
}

func (c *Client) OpenIM(ctx context.Context, userID string) (string, error) {
	body, err := json.Marshal(openConversationRequest{Users: userID})
	if err != nil {
		return "", fmt.Errorf("cannot marshal open conversation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPIConversationsOpen, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot call slack api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read slack response: %w", err)
	}

	var result openConversationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("cannot decode slack response: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("slack api error: %s", result.Error)
	}
	if result.Channel.ID == "" {
		return "", fmt.Errorf("slack api returned empty dm channel id")
	}

	return result.Channel.ID, nil
}

func (c *Client) PostMessage(ctx context.Context, channel, text, threadTS string) error {
	return c.PostMessageWithBlocks(ctx, channel, text, threadTS, nil)
}

func (c *Client) PostMessageWithBlocks(
	ctx context.Context,
	channel, text, threadTS string,
	blocks []any,
) error {
	body, err := json.Marshal(
		postMessageRequest{
			Channel:  channel,
			Text:     text,
			ThreadTS: threadTS,
			Blocks:   blocks,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal post message request: %w", err)
	}

	return c.post(ctx, slackAPIPostMessage, body)
}

func (c *Client) AddReaction(ctx context.Context, channel, name, timestamp string) error {
	body, err := json.Marshal(
		addReactionRequest{
			Channel:   channel,
			Name:      name,
			Timestamp: timestamp,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal add reaction request: %w", err)
	}

	return c.post(ctx, slackAPIReactionsAdd, body)
}

// SetStatus sets the assistant thread status visible to users (e.g. "is thinking...").
// Pass an empty status to clear it.
func (c *Client) SetStatus(ctx context.Context, channelID, threadTS, status string) error {
	body, err := json.Marshal(
		setStatusRequest{
			ChannelID: channelID,
			ThreadTS:  threadTS,
			Status:    status,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal set status request: %w", err)
	}

	return c.post(ctx, slackAPIAssistantSetStatus, body)
}

func (c *Client) post(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot call slack api: %w", err)
	}
	defer resp.Body.Close()

	return decodeSlackResponse(resp.Body)
}

func decodeSlackResponse(body io.Reader) error {
	var result slackResponse
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return fmt.Errorf("cannot decode slack response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack api error: %s", result.Error)
	}

	return nil
}
