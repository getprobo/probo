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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

const (
	slackMethodAssistantSetStatus   = "assistant.threads.setStatus"
	slackMethodAppsUninstall        = "apps.uninstall"
	slackMethodConversationsList    = "conversations.list"
	slackMethodConversationsOpen    = "conversations.open"
	slackMethodConversationsReplies = "conversations.replies"
	slackMethodPostEphemeral        = "chat.postEphemeral"
	slackMethodPostMessage          = "chat.postMessage"
	slackMethodReactionsAdd         = "reactions.add"
	slackMethodUpdateMessage        = "chat.update"

	conversationsPageSize = 200
	threadRepliesPageSize = 200
	slackHTTPTimeout      = 30 * time.Second
)

type (
	Client struct {
		token      string
		httpClient *http.Client
		apiBaseURL string
	}

	APIError struct {
		StatusCode int
		Code       string
		RetryAfter time.Duration
	}

	slackResponse struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	postMessageRequest struct {
		Channel     string `json:"channel"`
		Text        string `json:"text,omitempty"`
		ThreadTS    string `json:"thread_ts,omitempty"`
		Blocks      []any  `json:"blocks,omitempty"`
		ClientMsgID string `json:"client_msg_id,omitempty"`
	}

	MessageRef struct {
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	}

	postMessageResponse struct {
		slackResponse
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	}

	updateMessageRequest struct {
		Channel string `json:"channel"`
		TS      string `json:"ts"`
		Text    string `json:"text,omitempty"`
		Blocks  []any  `json:"blocks,omitempty"`
	}

	postEphemeralRequest struct {
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text,omitempty"`
		Blocks  []any  `json:"blocks,omitempty"`
	}

	addReactionRequest struct {
		Channel   string `json:"channel"`
		Name      string `json:"name"`
		Timestamp string `json:"timestamp"`
	}

	setStatusRequest struct {
		ChannelID       string   `json:"channel_id"`
		ThreadTS        string   `json:"thread_ts"`
		Status          string   `json:"status"`
		LoadingMessages []string `json:"loading_messages,omitempty"`
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

	Conversation struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		IsMember   bool   `json:"is_member"`
		IsArchived bool   `json:"is_archived"`
	}

	ConversationsPage struct {
		Conversations []Conversation
		NextCursor    string
	}

	conversationsListResponse struct {
		slackResponse
		Channels         []Conversation `json:"channels"`
		ResponseMetadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}

	ThreadReply struct {
		User    string `json:"user,omitempty"`
		BotID   string `json:"bot_id,omitempty"`
		Text    string `json:"text,omitempty"`
		TS      string `json:"ts,omitempty"`
		Subtype string `json:"subtype,omitempty"`
	}

	threadRepliesPage struct {
		Messages   []ThreadReply
		NextCursor string
	}

	conversationsRepliesResponse struct {
		slackResponse
		Messages         []ThreadReply `json:"messages"`
		ResponseMetadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}
)

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("slack api error: %s", e.Code)
	}

	return fmt.Sprintf("slack api returned HTTP status %d", e.StatusCode)
}

func NewClient(botToken string, apiBaseURL string, logger *log.Logger) *Client {
	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
		httpclient.WithSSRFProtection(),
	}

	client := httpclient.DefaultPooledClient(httpClientOpts...)
	client.Timeout = slackHTTPTimeout

	return newAPIClient(botToken, apiBaseURL, client)
}

func newAPIClient(botToken string, apiBaseURL string, httpClient *http.Client) *Client {
	return &Client{
		token:      botToken,
		httpClient: httpClient,
		apiBaseURL: apiBaseURL,
	}
}

func (c *Client) UninstallApp(
	ctx context.Context,
	clientID string,
	clientSecret string,
) error {
	endpoint, err := c.methodURL(slackMethodAppsUninstall)
	if err != nil {
		return fmt.Errorf("cannot build Slack uninstall endpoint: %w", err)
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("cannot create Slack uninstall request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot uninstall Slack app: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var result slackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("cannot decode Slack uninstall response: %w", err)
	}

	if !result.OK {
		switch result.Error {
		case "account_inactive", "invalid_auth", "token_revoked":
			return nil
		}
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newAPIError(resp, result.Error)
	}

	return nil
}

func (c *Client) methodURL(method string) (string, error) {
	endpoint, err := url.JoinPath(c.apiBaseURL, method)
	if err != nil {
		return "", fmt.Errorf("cannot join Slack API URL: %w", err)
	}

	return endpoint, nil
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

	return c.post(ctx, slackMethodPostMessage, body)
}

func (c *Client) CreateMessageWithBlocks(
	ctx context.Context,
	channel string,
	text string,
	blocks []any,
	clientMsgID string,
) (*MessageRef, error) {
	return c.createMessage(ctx, channel, text, "", blocks, clientMsgID)
}

func (c *Client) CreateMessage(
	ctx context.Context,
	channel string,
	text string,
	threadTS string,
	clientMsgID string,
) (*MessageRef, error) {
	return c.createMessage(ctx, channel, text, threadTS, nil, clientMsgID)
}

func (c *Client) createMessage(
	ctx context.Context,
	channel string,
	text string,
	threadTS string,
	blocks []any,
	clientMsgID string,
) (*MessageRef, error) {
	body, err := json.Marshal(
		postMessageRequest{
			Channel:     channel,
			Text:        text,
			ThreadTS:    threadTS,
			Blocks:      blocks,
			ClientMsgID: clientMsgID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal create message request: %w", err)
	}

	endpoint, err := c.methodURL(slackMethodPostMessage)
	if err != nil {
		return nil, fmt.Errorf("cannot build create message endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create message request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot create Slack message: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var result postMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, newAPIError(resp, "")
		}

		return nil, fmt.Errorf("cannot decode create message response: %w", err)
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newAPIError(resp, result.Error)
	}

	if result.Channel == "" || result.TS == "" {
		return nil, fmt.Errorf("slack create message response is missing channel or timestamp")
	}

	return &MessageRef{Channel: result.Channel, TS: result.TS}, nil
}

func (c *Client) UpdateMessage(
	ctx context.Context,
	channel string,
	timestamp string,
	text string,
	blocks []any,
) error {
	body, err := json.Marshal(
		updateMessageRequest{
			Channel: channel,
			TS:      timestamp,
			Text:    text,
			Blocks:  blocks,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal update message request: %w", err)
	}

	if err := c.post(ctx, slackMethodUpdateMessage, body); err != nil {
		return fmt.Errorf("cannot update message: %w", err)
	}

	return nil
}

func (c *Client) PostEphemeral(
	ctx context.Context,
	channel string,
	user string,
	text string,
	blocks []any,
) error {
	body, err := json.Marshal(
		postEphemeralRequest{
			Channel: channel,
			User:    user,
			Text:    text,
			Blocks:  blocks,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal post ephemeral request: %w", err)
	}

	if err := c.post(ctx, slackMethodPostEphemeral, body); err != nil {
		return fmt.Errorf("cannot post ephemeral message: %w", err)
	}

	return nil
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

	return c.post(ctx, slackMethodReactionsAdd, body)
}

// SetStatus sets the assistant thread status visible to users (e.g. "is thinking...").
// Pass an empty status to clear it. loadingMessages are omitted when empty.

func (c *Client) post(ctx context.Context, method string, body []byte) error {
	endpoint, err := c.methodURL(method)
	if err != nil {
		return fmt.Errorf("cannot build Slack API endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot call slack api: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return decodeSlackResponse(resp)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
}

func decodeSlackResponse(resp *http.Response) error {
	var result slackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return newAPIError(resp, "")
		}

		return fmt.Errorf("cannot decode slack response: %w", err)
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newAPIError(resp, result.Error)
	}

	return nil
}

func newAPIError(resp *http.Response, code string) *APIError {
	return &APIError{
		StatusCode: resp.StatusCode,
		Code:       code,
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
	}
}

func parseRetryAfter(value string) time.Duration {
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
}
