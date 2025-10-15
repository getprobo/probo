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
)

type (
	Client struct {
		webhookURL string
		httpClient *http.Client
	}

	webhookMessage struct {
		Text   string  `json:"text,omitempty"`
		Blocks []block `json:"blocks,omitempty"`
	}

	block struct {
		Type string    `json:"type"`
		Text *textItem `json:"text,omitempty"`
	}

	textItem struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
)

func NewClient(webhookURL string, httpClient *http.Client) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: httpClient,
	}
}

func (c *Client) PostMessage(ctx context.Context, text string) error {
	msg := webhookMessage{
		Text: text,
		Blocks: []block{
			{
				Type: "section",
				Text: &textItem{
					Type: "mrkdwn",
					Text: text,
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("cannot marshal message: %w", err)
	}
	body := buf.Bytes()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d, failed to read response body: %w", resp.StatusCode, err)
		}

		var errorResponse map[string]any
		var buf bytes.Buffer
		buf.Write(body)
		if err := json.NewDecoder(&buf).Decode(&errorResponse); err != nil {
			return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
		}

		return fmt.Errorf("unexpected status code: %d, response: %+v", resp.StatusCode, errorResponse)
	}

	return nil
}
