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

package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	MailpitMessage struct {
		ID string `json:"id"`
	}

	MailpitSearchResponse struct {
		Messages []MailpitMessage `json:"messages"`
	}

	MailpitLink struct {
		URL string `json:"url"`
	}

	MailpitLinkCheckResponse struct {
		Links []MailpitLink `json:"links"`
	}
)

func (c *Client) SearchMails(query string) (*MailpitSearchResponse, error) {
	req, err := http.NewRequest("GET", c.mailpitBaseURL+"/api/v1/search?query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp MailpitSearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	return &searchResp, nil
}

func (c *Client) CheckMessageLinks(messageID string) (*MailpitLinkCheckResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/message/%s/link-check", c.mailpitBaseURL, messageID), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var linkCheckResp MailpitLinkCheckResponse
	if err := json.Unmarshal(respBody, &linkCheckResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	return &linkCheckResp, nil
}
