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
	"net/url"
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
		refresher  *TokenRefresher
	}

	// TokenRefresher holds the information needed to automatically refresh
	// an expired access token using the OAuth2 refresh_token grant.
	TokenRefresher struct {
		RefreshToken  string
		TokenEndpoint string
		ClientID      string
		// OnRefresh is called after a successful token refresh with the new
		// access token and refresh token so the caller can persist them.
		OnRefresh func(accessToken, refreshToken string) error
	}

	Option func(*Client)

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

	tokenRefreshResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
)

func WithTokenRefresher(r *TokenRefresher) Option {
	return func(c *Client) { c.refresher = r }
}

func NewClient(host string, token string, endpoint string, timeout time.Duration, opts ...Option) *Client {
	c := &Client{
		host:       host,
		token:      token,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: timeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
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
	respBody, statusCode, err := c.doRequest(query, variables)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusUnauthorized && c.refresher != nil {
		if refreshErr := c.tryRefreshToken(); refreshErr == nil {
			respBody, statusCode, err = c.doRequest(query, variables)
			if err != nil {
				return nil, err
			}
		}
	}

	if statusCode != http.StatusOK {
		switch statusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("authentication failed (HTTP 401): token may be invalid or expired, try 'prb auth login'")
		case http.StatusForbidden:
			return nil, fmt.Errorf("access denied (HTTP 403): you do not have permission to perform this action")
		default:
			return nil, fmt.Errorf(
				"HTTP %d: %s",
				statusCode,
				string(respBody),
			)
		}
	}

	return respBody, nil
}

func (c *Client) doRequest(
	query string,
	variables map[string]any,
) ([]byte, int, error) {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot marshal GraphQL request: %w", err)
	}

	host := c.host
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	reqURL := host + c.endpoint
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("cannot create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot send HTTP request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot read HTTP response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

func (c *Client) tryRefreshToken() error {
	r := c.refresher

	values := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {r.ClientID},
		"refresh_token": {r.RefreshToken},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		r.TokenEndpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return fmt.Errorf("cannot create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send refresh request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token request failed (HTTP %d)", resp.StatusCode)
	}

	var token tokenRefreshResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return fmt.Errorf("cannot decode refresh response: %w", err)
	}

	c.token = token.AccessToken
	r.RefreshToken = token.RefreshToken

	if r.OnRefresh != nil {
		return r.OnRefresh(token.AccessToken, token.RefreshToken)
	}

	return nil
}
