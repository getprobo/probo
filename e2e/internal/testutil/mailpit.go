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
		IDLower string `json:"id"`
		IDUpper string `json:"ID"`
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

	MailpitAddress struct {
		Address string `json:"Address"`
	}

	MailpitMessageDetail struct {
		ID      string           `json:"ID"`
		Subject string           `json:"Subject"`
		Text    string           `json:"Text"`
		HTML    string           `json:"HTML"`
		To      []MailpitAddress `json:"To"`
	}
)

func (m MailpitMessage) ResolvedID() string {
	if m.IDUpper != "" {
		return m.IDUpper
	}

	return m.IDLower
}

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

func (c *Client) GetMailpitMessage(messageID string) (*MailpitMessageDetail, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/message/%s", c.mailpitBaseURL, messageID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

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

	var detail MailpitMessageDetail
	if err := json.Unmarshal(respBody, &detail); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	return &detail, nil
}

// FindMailpitMessage loads every message returned by search and returns the first
// detail for which accept returns true.
func (c *Client) FindMailpitMessage(
	searchQuery string,
	accept func(*MailpitMessageDetail) bool,
) (*MailpitMessageDetail, error) {
	searchMails, err := c.SearchMails(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("mailpit search failed: %w", err)
	}

	if len(searchMails.Messages) == 0 {
		return nil, fmt.Errorf("mailpit search returned no messages for query %q", searchQuery)
	}

	var lastErr error

	for _, msg := range searchMails.Messages {
		messageID := msg.ResolvedID()
		if messageID == "" {
			lastErr = fmt.Errorf("mailpit search hit missing message id")

			continue
		}

		detail, err := c.GetMailpitMessage(messageID)
		if err != nil {
			lastErr = fmt.Errorf("cannot load mailpit message %s: %w", messageID, err)

			continue
		}

		if accept(detail) {
			return detail, nil
		}

		lastErr = fmt.Errorf("mailpit message %s did not match accept predicate", messageID)
	}

	return nil, lastErr
}

// FindLinkTokenFromMailpitSearch scans all messages from search and returns the
// first token query parameter found among their links.
func (c *Client) FindLinkTokenFromMailpitSearch(searchQuery string) (string, error) {
	searchMails, err := c.SearchMails(searchQuery)
	if err != nil {
		return "", fmt.Errorf("mailpit search failed: %w", err)
	}

	if len(searchMails.Messages) == 0 {
		return "", fmt.Errorf("mailpit search returned no messages for query %q", searchQuery)
	}

	var lastErr error

	for _, msg := range searchMails.Messages {
		messageID := msg.ResolvedID()
		if messageID == "" {
			lastErr = fmt.Errorf("mailpit search hit missing message id")

			continue
		}

		linksCheck, err := c.CheckMessageLinks(messageID)
		if err != nil {
			lastErr = fmt.Errorf("mailpit link check failed for %s: %w", messageID, err)

			continue
		}

		for _, link := range linksCheck.Links {
			linkURL, err := url.Parse(link.URL)
			if err != nil {
				continue
			}

			if token := linkURL.Query().Get("token"); token != "" {
				return token, nil
			}
		}

		lastErr = fmt.Errorf("mailpit message %s contained no token link", messageID)
	}

	return "", lastErr
}
