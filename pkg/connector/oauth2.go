// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/base64"
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
		ClientID          string
		ClientSecret      string
		RedirectURI       string
		Scopes            []string
		AuthURL           string
		TokenURL          string
		ExtraAuthParams   map[string]string // Optional: extra params for auth URL (e.g., access_type=offline for Google)
		TokenEndpointAuth string            // "post-form" (default), "basic-form", or "basic-json"
	}

	OAuth2State struct {
		OrganizationID string `json:"oid"`
		Provider       string `json:"provider"`
		ContinueURL    string `json:"continue,omitempty"`
		ConnectorID    string `json:"cid,omitempty"` // Set when reconnecting an existing connector
	}

	OAuth2Connection struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		ExpiresAt    time.Time `json:"expires_at"`
		TokenType    string    `json:"token_type"`
		Scope        string    `json:"scope,omitempty"`

		// Client Credentials fields (only set when GrantType == "client_credentials"):
		GrantType    OAuth2GrantType `json:"grant_type,omitempty"`
		ClientID     string          `json:"client_id,omitempty"`
		ClientSecret string          `json:"client_secret,omitempty"`
		TokenURL     string          `json:"token_url,omitempty"`
	}

	// OAuth2RefreshConfig contains the OAuth2 credentials needed for token refresh.
	OAuth2RefreshConfig struct {
		ClientID          string
		ClientSecret      string
		TokenURL          string
		TokenEndpointAuth string // "post-form" (default), "basic-form", or "basic-json"
	}
)

var (
	_ Connector  = (*OAuth2Connector)(nil)
	_ Connection = (*OAuth2Connection)(nil)

	OAuth2TokenType = "probo/connector/oauth2"
	OAuth2TokenTTL  = 10 * time.Minute
)

// DecodeOAuth2StatePayload decodes the OAuth2 state token payload without
// verifying the signature. This is useful when you need to inspect the
// payload to determine which secret to use for full validation (e.g.,
// extracting the provider from the state token to look up the correct
// connector).
func DecodeOAuth2StatePayload(tokenString string) (*statelesstoken.Payload[OAuth2State], error) {
	return statelesstoken.DecodePayload[OAuth2State](tokenString)
}

func (c *OAuth2Connector) Initiate(ctx context.Context, provider string, organizationID gid.GID, r *http.Request) (string, error) {
	stateData := OAuth2State{
		OrganizationID: organizationID.String(),
		Provider:       provider,
	}
	if r != nil {
		if continueURL := r.URL.Query().Get("continue"); continueURL != "" {
			stateData.ContinueURL = continueURL
		}
		if connectorID := r.URL.Query().Get("connector_id"); connectorID != "" {
			stateData.ConnectorID = connectorID
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

	authCodeQuery := url.Values{}
	authCodeQuery.Set("state", state)
	authCodeQuery.Set("client_id", c.ClientID)
	authCodeQuery.Set("redirect_uri", c.RedirectURI)
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

	organizationID, err := gid.ParseGID(payload.Data.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse organization ID: %w", err)
	}

	tokenRequest, err := c.buildTokenRequest(ctx, code, c.RedirectURI)
	if err != nil {
		return nil, nil, err
	}

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

	if payload.Data.Provider == SlackProvider {
		conn, _, err := ParseSlackTokenResponse(body, oauth2Conn, organizationID)
		return conn, &payload.Data, err
	}

	return &oauth2Conn, &payload.Data, nil
}

func basicAuthHeader(clientID, clientSecret string) string {
	credentials := clientID + ":" + clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}

// buildTokenRequest creates the HTTP request for the token exchange, branching
// on c.TokenEndpointAuth to support different provider requirements.
func (c *OAuth2Connector) buildTokenRequest(ctx context.Context, code, redirectURI string) (*http.Request, error) {
	switch c.TokenEndpointAuth {
	case "basic-json":
		// JSON body with Basic auth header (Notion).
		body := map[string]string{
			"code":         code,
			"redirect_uri": redirectURI,
			"grant_type":   "authorization_code",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal token request body: %w", err)
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.TokenURL,
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			return nil, fmt.Errorf("cannot create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Probo Connector")
		req.Header.Set("Authorization", basicAuthHeader(c.ClientID, c.ClientSecret))
		return req, nil

	case "basic-form":
		// Form-encoded body with Basic auth header (DocuSign).
		formData := url.Values{}
		formData.Set("code", code)
		formData.Set("redirect_uri", redirectURI)
		formData.Set("grant_type", "authorization_code")

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.TokenURL,
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			return nil, fmt.Errorf("cannot create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Probo Connector")
		req.Header.Set("Authorization", basicAuthHeader(c.ClientID, c.ClientSecret))
		return req, nil

	default:
		// "post-form" or empty: credentials in form body (Slack, HubSpot, GitHub, etc.).
		formData := url.Values{}
		formData.Set("client_id", c.ClientID)
		formData.Set("client_secret", c.ClientSecret)
		formData.Set("code", code)
		formData.Set("redirect_uri", redirectURI)
		formData.Set("grant_type", "authorization_code")

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.TokenURL,
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			return nil, fmt.Errorf("cannot create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Probo Connector")
		return req, nil
	}
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
//
// For client_credentials grant type, it uses the connection's own credentials
// to obtain a new token instead of refreshing via a refresh token.
func (c *OAuth2Connection) RefreshableClient(ctx context.Context, cfg OAuth2RefreshConfig, opts ...httpclient.Option) (*http.Client, error) {
	if c.GrantType == OAuth2GrantTypeClientCredentials {
		return c.clientCredentialsClient(ctx, opts...)
	}

	if c.RefreshToken == "" {
		return c.ClientWithOptions(ctx, opts...)
	}

	// Determine auth style based on TokenEndpointAuth
	authStyle := oauth2.AuthStyleInParams
	switch cfg.TokenEndpointAuth {
	case "basic-form", "basic-json":
		authStyle = oauth2.AuthStyleInHeader
	}

	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  cfg.TokenURL,
			AuthStyle: authStyle,
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

// clientCredentialsClient obtains a new access token using the client_credentials
// grant type, using the connection's own ClientID, ClientSecret, and TokenURL.
func (c *OAuth2Connection) clientCredentialsClient(ctx context.Context, opts ...httpclient.Option) (*http.Client, error) {
	// If we have a valid token that hasn't expired, reuse it
	if c.AccessToken != "" && !c.ExpiresAt.IsZero() && c.ExpiresAt.After(time.Now()) {
		return c.ClientWithOptions(ctx, opts...)
	}

	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	if c.Scope != "" {
		formData.Set("scope", c.Scope)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.TokenURL,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create client credentials token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Probo Connector")
	req.Header.Set("Authorization", basicAuthHeader(c.ClientID, c.ClientSecret))

	httpClient := &http.Client{
		Transport: httpclient.DefaultPooledTransport(opts...),
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot post client credentials token URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client credentials token response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read client credentials token response body: %w", err)
	}

	var rawToken struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &rawToken); err != nil {
		return nil, fmt.Errorf("cannot decode client credentials token response: %w", err)
	}

	c.AccessToken = rawToken.AccessToken
	if rawToken.TokenType != "" {
		c.TokenType = rawToken.TokenType
	}
	if c.TokenType == "" {
		c.TokenType = "Bearer"
	}
	if rawToken.ExpiresIn > 0 {
		c.ExpiresAt = time.Now().Add(time.Duration(rawToken.ExpiresIn) * time.Second)
	}

	return &http.Client{
		Transport: &oauth2Transport{
			token:      c.AccessToken,
			tokenType:  c.TokenType,
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
