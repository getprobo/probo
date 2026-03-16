// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/version"
)

type (
	Client struct {
		host       string
		token      string
		endpoint   string
		httpClient *http.Client
	}

	graphQLRequest struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables,omitempty"`
	}

	graphQLResponse struct {
		Data   json.RawMessage `json:"data"`
		Errors []graphQLError  `json:"errors"`
	}

	graphQLError struct {
		Message string `json:"message"`
	}
)

func NewClient(host string, token string, endpoint string, timeout time.Duration) *Client {
	return &Client{
		host:       host,
		token:      token,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Do(
	query string,
	variables map[string]any,
) (json.RawMessage, error) {
	raw, err := c.DoRaw(query, variables)
	if err != nil {
		return nil, err
	}

	var resp graphQLResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("cannot parse GraphQL response: %w", err)
	}

	if len(resp.Errors) > 0 {
		var msg strings.Builder
		msg.WriteString(resp.Errors[0].Message)
		for _, e := range resp.Errors[1:] {
			msg.WriteString("; " + e.Message)
		}
		return nil, fmt.Errorf("GraphQL error: %s", msg.String())
	}

	return resp.Data, nil
}

func (c *Client) DoRaw(
	query string,
	variables map[string]any,
) ([]byte, error) {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal GraphQL request: %w", err)
	}

	url := fmt.Sprintf("https://%s%s", c.host, c.endpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot send HTTP request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read HTTP response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("authentication failed (HTTP 401): token may be invalid or expired, try 'prb auth login'")
		case http.StatusForbidden:
			return nil, fmt.Errorf("access denied (HTTP 403): you do not have permission to perform this action")
		default:
			return nil, fmt.Errorf(
				"HTTP %d: %s",
				resp.StatusCode,
				string(respBody),
			)
		}
	}

	return respBody, nil
}
