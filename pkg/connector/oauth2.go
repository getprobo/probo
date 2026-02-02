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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
	"golang.org/x/oauth2"
)

// NOTE: I use client_secret as a salt for the state token, it's an antipattern to
// avoid having add configuration key for now. In the future, we should use a random
// string as a salt. It does not compromise security, because the client_secret is
// private to the connector and not exposed to the client but using the same secret for
// two different connectors may not expected by other developers and can lead to confusion
// and bugs.

type (
	OAuth2Connector struct {
		ClientID        string
		ClientSecret    string
		RedirectURI     string
		Scopes          []string
		AuthURL         string
		TokenURL        string
		ExtraAuthParams map[string]string // Optional: extra params for auth URL (e.g., access_type=offline for Google)
	}

	OAuth2State struct {
		OrganizationID string `json:"oid"`
		Provider       string `json:"provider"`
		ContinueURL    string `json:"continue,omitempty"`
	}

	OAuth2Connection struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		ExpiresAt    time.Time `json:"expires_at"`
		TokenType    string    `json:"token_type"`
		Scope        string    `json:"scope,omitempty"`
	}

	// OAuth2RefreshConfig contains the OAuth2 credentials needed for token refresh.
	OAuth2RefreshConfig struct {
		ClientID     string
		ClientSecret string
		TokenURL     string
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
	if r != nil {
		if continueURL := r.URL.Query().Get("continue"); continueURL != "" {
			stateData.ContinueURL = continueURL
		}
	}
	return c.InitiateWithState(ctx, stateData, r)
}

// InitiateWithState generates an OAuth2 authorization URL with a custom state.
// This allows callers to include additional context (like SCIMBridgeID) in the state.
func (c *OAuth2Connector) InitiateWithState(ctx context.Context, stateData OAuth2State, r *http.Request) (string, error) {
	state, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL, stateData)
	if err != nil {
		return "", fmt.Errorf("cannot create state token: %w", err)
	}

	// Build redirect URI with provider (fixed per provider, so can be registered in OAuth console)
	redirectURI := c.RedirectURI
	redirectURIParsed, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("cannot parse redirect URI: %w", err)
	}
	q := redirectURIParsed.Query()
	q.Set("provider", stateData.Provider)
	redirectURIParsed.RawQuery = q.Encode()
	redirectURI = redirectURIParsed.String()

	authCodeQuery := url.Values{}
	authCodeQuery.Set("state", state)
	authCodeQuery.Set("client_id", c.ClientID)
	authCodeQuery.Set("redirect_uri", redirectURI)
	authCodeQuery.Set("response_type", "code")
	authCodeQuery.Set("scope", strings.Join(c.Scopes, " "))

	// Add any extra auth params (e.g., access_type=offline, prompt=consent for Google)
	for k, v := range c.ExtraAuthParams {
		authCodeQuery.Set(k, v)
	}

	u, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse auth URL: %w", err)
	}

	u.RawQuery = authCodeQuery.Encode()

	return u.String(), nil
}

func (c *OAuth2Connector) Complete(ctx context.Context, r *http.Request) (Connection, *gid.GID, string, error) {
	conn, state, err := c.CompleteWithState(ctx, r)
	if err != nil {
		return nil, nil, "", err
	}

	organizationID, err := gid.ParseGID(state.OrganizationID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot parse organization ID: %w", err)
	}

	return conn, &organizationID, state.ContinueURL, nil
}

// CompleteWithState completes the OAuth2 flow and returns the full state.
// This allows callers to access additional context (like SCIMBridgeID) from the state.
func (c *OAuth2Connector) CompleteWithState(ctx context.Context, r *http.Request) (Connection, *OAuth2State, error) {
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

	// Build redirect URI with provider (must match what was sent to auth endpoint)
	redirectURI := c.RedirectURI
	redirectURIParsed, err := url.Parse(redirectURI)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse redirect URI: %w", err)
	}
	q := redirectURIParsed.Query()
	q.Set("provider", provider)
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
	defer func() { _ = tokenResp.Body.Close() }()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("token response status: %d", tokenResp.StatusCode)
	}

	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read token response body: %w", err)
	}

	// Parse the raw token response (OAuth2 uses expires_in, not expires_at)
	var rawToken struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &rawToken); err != nil {
		return nil, nil, fmt.Errorf("cannot decode token response: %w", err)
	}

	oauth2Conn := OAuth2Connection{
		AccessToken:  rawToken.AccessToken,
		RefreshToken: rawToken.RefreshToken,
		TokenType:    rawToken.TokenType,
		Scope:        rawToken.Scope,
	}

	// Convert expires_in (seconds) to expires_at (absolute time)
	if rawToken.ExpiresIn > 0 {
		oauth2Conn.ExpiresAt = time.Now().Add(time.Duration(rawToken.ExpiresIn) * time.Second)
	}

	if provider == SlackProvider {
		conn, _, err := ParseSlackTokenResponse(body, oauth2Conn, organizationID)
		return conn, &payload.Data, err
	}

	return &oauth2Conn, &payload.Data, nil
}

func (c *OAuth2Connection) Type() ProtocolType {
	return ProtocolOAuth2
}

func (c *OAuth2Connection) Client(ctx context.Context) (*http.Client, error) {
	return c.ClientWithOptions(ctx)
}

// ClientWithOptions returns an HTTP client with the given options.
// Use this to add logging and tracing to the HTTP client.
func (c *OAuth2Connection) ClientWithOptions(ctx context.Context, opts ...httpclient.Option) (*http.Client, error) {
	transport := &oauth2Transport{
		token:      c.AccessToken,
		tokenType:  c.TokenType,
		underlying: httpclient.DefaultPooledTransport(opts...),
	}
	client := &http.Client{
		Transport: transport,
	}
	return client, nil
}

// RefreshableClient returns an HTTP client that automatically refreshes the token when expired.
// It also updates the connection's token fields if a refresh occurs.
func (c *OAuth2Connection) RefreshableClient(ctx context.Context, cfg OAuth2RefreshConfig, opts ...httpclient.Option) (*http.Client, error) {
	if c.RefreshToken == "" {
		return c.ClientWithOptions(ctx, opts...)
	}

	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: cfg.TokenURL,
		},
	}

	// Determine the token expiry
	// If ExpiresAt is zero or in the past, set expiry to force a refresh
	expiry := c.ExpiresAt
	if expiry.IsZero() || expiry.Before(time.Now()) {
		// Set expiry to the past to force oauth2 library to refresh
		expiry = time.Now().Add(-time.Hour)
	}

	token := &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		Expiry:       expiry,
		TokenType:    c.TokenType,
	}

	// Create an HTTP client with telemetry for the oauth2 library to use
	// This ensures token refresh requests are also logged
	baseClient := &http.Client{
		Transport: httpclient.DefaultPooledTransport(opts...),
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, baseClient)

	// Create a token source that will automatically refresh when expired
	tokenSource := config.TokenSource(ctx, token)

	// Get the current (possibly refreshed) token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("cannot refresh token: %w", err)
	}

	// Update the connection with the potentially refreshed token
	c.AccessToken = newToken.AccessToken
	c.ExpiresAt = newToken.Expiry
	c.TokenType = newToken.TokenType
	if newToken.RefreshToken != "" {
		c.RefreshToken = newToken.RefreshToken
	}

	// Return a client with telemetry that uses the refreshed token
	return &http.Client{
		Transport: &oauth2Transport{
			token:      newToken.AccessToken,
			tokenType:  newToken.TokenType,
			underlying: httpclient.DefaultPooledTransport(opts...),
		},
	}, nil
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
