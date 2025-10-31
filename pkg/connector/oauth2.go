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

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

// NOTE: I use client_secret as a salt for the state token, it's an antipattern to
// avoid having add configuration key for now. In the future, we should use a random
// string as a salt. It does not compromise security, because the client_secret is
// private to the connector and not exposed to the client but using the same secret for
// two different connectors may not expected by other developers and can lead to confusion
// and bugs.

type (
	OAuth2Connector struct {
		ClientID     string
		ClientSecret string
		RedirectURI  string
		Scopes       []string
		AuthURL      string
		TokenURL     string
	}

	OAuth2State struct {
		OrganizationID string `json:"oid"`
		Provider       string `json:"provider"`
	}

	OAuth2Connection struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		ExpiresAt    time.Time `json:"expires_at"`
		TokenType    string    `json:"token_type"`
		Scope        string    `json:"scope,omitempty"`
	}
)

var (
	_ Connector  = (*OAuth2Connector)(nil)
	_ Connection = (*OAuth2Connection)(nil)

	OAuth2TokenType = "probo/connector/oauth2"
	OAuth2TokenTTL  = 10 * time.Minute
)

func (c *OAuth2Connector) Initiate(ctx context.Context, provider string, organizationID gid.GID, r *http.Request) (string, error) {
	stateData := OAuth2State{
		OrganizationID: organizationID.String(),
		Provider:       provider,
	}
	state, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL, stateData)
	if err != nil {
		return "", fmt.Errorf("cannot create state token: %w", err)
	}

	redirectURI := c.RedirectURI
	redirectURIParsed, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("cannot parse redirect URI: %w", err)
	}
	q := redirectURIParsed.Query()
	q.Set("provider", provider)
	if continueURL := r.URL.Query().Get("continue"); continueURL != "" {
		q.Set("continue", continueURL)
	}
	redirectURIParsed.RawQuery = q.Encode()
	redirectURI = redirectURIParsed.String()

	authCodeQuery := url.Values{}
	authCodeQuery.Set("state", state)
	authCodeQuery.Set("client_id", c.ClientID)
	authCodeQuery.Set("redirect_uri", redirectURI)
	authCodeQuery.Set("response_type", "code")
	authCodeQuery.Set("scope", strings.Join(c.Scopes, " "))

	u, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse auth URL: %w", err)
	}

	u.RawQuery = authCodeQuery.Encode()

	return u.String(), nil
}

func (c *OAuth2Connector) Complete(ctx context.Context, r *http.Request) (Connection, *gid.GID, error) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		return nil, nil, fmt.Errorf("missing provider in query parameters")
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("no code in request")
	}

	stateToken := r.URL.Query().Get("state")
	if stateToken == "" {
		return nil, nil, fmt.Errorf("no state in request")
	}

	payload, err := statelesstoken.ValidateToken[OAuth2State](c.ClientSecret, OAuth2TokenType, stateToken)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot validate state token: %w", err)
	}

	if payload.Data.Provider != provider {
		return nil, nil, fmt.Errorf("provider mismatch: state has %q, query has %q", payload.Data.Provider, provider)
	}

	organizationID, err := gid.ParseGID(payload.Data.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse organization ID: %w", err)
	}

	redirectURI := c.RedirectURI
	redirectURIParsed, err := url.Parse(redirectURI)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse redirect URI: %w", err)
	}
	q := redirectURIParsed.Query()
	q.Set("provider", provider)
	if continueURL := r.URL.Query().Get("continue"); continueURL != "" {
		q.Set("continue", continueURL)
	}
	redirectURIParsed.RawQuery = q.Encode()
	redirectURI = redirectURIParsed.String()

	tokenRequestData := url.Values{}
	tokenRequestData.Set("client_id", c.ClientID)
	tokenRequestData.Set("client_secret", c.ClientSecret)
	tokenRequestData.Set("code", code)
	tokenRequestData.Set("redirect_uri", redirectURI)
	tokenRequestData.Set("grant_type", "authorization_code")

	tokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(tokenRequestData.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create token request: %w", err)
	}

	tokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	tokenRequest.Header.Set("Accept", "application/json")
	tokenRequest.Header.Set("User-Agent", "Probo Connector")

	tokenResp, err := http.DefaultClient.Do(tokenRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot post token URL: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("token response status: %d", tokenResp.StatusCode)
	}

	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read token response body: %w", err)
	}

	var oauth2Conn OAuth2Connection
	var buf bytes.Buffer
	buf.Write(body)
	err = json.NewDecoder(&buf).Decode(&oauth2Conn)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode token response: %w", err)
	}

	if provider == SlackProvider {
		return ParseSlackTokenResponse(body, oauth2Conn, organizationID)
	}

	return &oauth2Conn, &organizationID, nil
}

func (c *OAuth2Connection) Type() ProtocolType {
	return ProtocolOAuth2
}

func (c OAuth2Connection) Client(ctx context.Context) (*http.Client, error) {
	client := &http.Client{
		Transport: &oauth2Transport{
			token:      c.AccessToken,
			tokenType:  c.TokenType,
			underlying: http.DefaultTransport,
		},
	}
	return client, nil
}

func (c OAuth2Connection) MarshalJSON() ([]byte, error) {
	type Alias OAuth2Connection
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  string(ProtocolOAuth2),
		Alias: Alias(c),
	})
}

func (c *OAuth2Connection) UnmarshalJSON(data []byte) error {
	type Alias OAuth2Connection
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(data, &aux)
}

// OAuth transport for adding authorization header
type oauth2Transport struct {
	token      string
	tokenType  string
	underlying http.RoundTripper
}

func (t *oauth2Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", t.tokenType+" "+t.token)
	return t.underlying.RoundTrip(req2)
}
