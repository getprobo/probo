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

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) OpenIM(ctx context.Context, userID string) (string, error) {
	body, err := json.Marshal(openConversationRequest{Users: userID})
	if err != nil {
		return "", fmt.Errorf("cannot marshal open conversation request: %w", err)
	}

	endpoint, err := c.methodURL(slackMethodConversationsOpen)
	if err != nil {
		return "", fmt.Errorf("cannot build open conversation endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot call slack api: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read slack response: %w", err)
	}

	var result openConversationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("cannot decode slack response: %w", err)
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", newAPIError(resp, result.Error)
	}

	if result.Channel.ID == "" {
		return "", fmt.Errorf("slack api returned empty dm channel id")
	}

	return result.Channel.ID, nil
}

func (c *Client) ListConversations(ctx context.Context, cursor string) (*ConversationsPage, error) {
	endpoint, err := c.methodURL(slackMethodConversationsList)
	if err != nil {
		return nil, fmt.Errorf("cannot build list conversations endpoint: %w", err)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse list conversations endpoint: %w", err)
	}

	query := u.Query()
	query.Set("exclude_archived", strconv.FormatBool(true))
	query.Set("limit", strconv.Itoa(conversationsPageSize))
	query.Set("types", "public_channel,private_channel")

	if cursor != "" {
		query.Set("cursor", cursor)
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create list conversations request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot list conversations: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var result conversationsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot decode list conversations response: %w", err)
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newAPIError(resp, result.Error)
	}

	return &ConversationsPage{
		Conversations: result.Channels,
		NextCursor:    result.ResponseMetadata.NextCursor,
	}, nil
}

func (c *Client) ListThreadReplies(
	ctx context.Context,
	channel string,
	threadTS string,
) ([]ThreadReply, error) {
	if channel == "" || threadTS == "" {
		return nil, fmt.Errorf("cannot list Slack thread replies without channel and thread")
	}

	var replies []ThreadReply
	cursor := ""

	for {
		page, err := c.listThreadRepliesPage(ctx, channel, threadTS, cursor)
		if err != nil {
			if len(replies) > 0 {
				return replies, err
			}

			return nil, err
		}

		replies = append(replies, page.Messages...)
		if page.NextCursor == "" ||
			len(page.Messages) == 0 ||
			len(replies) >= threadTranscriptMaxMessages {
			return replies, nil
		}

		cursor = page.NextCursor
	}
}

func (c *Client) listThreadRepliesPage(
	ctx context.Context,
	channel string,
	threadTS string,
	cursor string,
) (*threadRepliesPage, error) {
	endpoint, err := c.methodURL(slackMethodConversationsReplies)
	if err != nil {
		return nil, fmt.Errorf("cannot build list thread replies endpoint: %w", err)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse list thread replies endpoint: %w", err)
	}

	query := u.Query()
	query.Set("channel", channel)
	query.Set("ts", threadTS)
	query.Set("limit", strconv.Itoa(threadRepliesPageSize))
	if cursor != "" {
		query.Set("cursor", cursor)
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create list thread replies request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot list Slack thread replies: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var result conversationsRepliesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot decode list thread replies response: %w", err)
	}

	if !result.OK || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newAPIError(resp, result.Error)
	}

	return &threadRepliesPage{
		Messages:   result.Messages,
		NextCursor: result.ResponseMetadata.NextCursor,
	}, nil
}
