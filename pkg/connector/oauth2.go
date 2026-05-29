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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
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

// NOTE: the OAuth2 state token (and, for PKCE providers, the code verifier)
// is keyed by stateSalt(). Public clients (CIMD, no client_secret) set
// StateSigningKey to a server-side derived key; confidential clients fall
// back to the client_secret, which is private to the connector and not
// exposed to the client. Reusing the client_secret as the salt is a legacy
// path retained for confidential providers; new public clients always carry
// an explicit StateSigningKey.

type (
	OAuth2Connector struct {
		ClientID                string
		ClientSecret            string
		RedirectURI             string
		AuthURL                 string
		TokenURL                string
		ExtraAuthParams         map[string]string // Optional: extra params for auth URL (e.g., access_type=offline for Google)
		TokenEndpointAuth       string            // "post-form" (default), "basic-form", or "basic-json"
		SupportsIncrementalAuth bool
		// RequiresPKCE enables RFC 7636 PKCE (S256). When true,
		// InitiateWithState generates a verifier, persists it in the
		// OAuth2State, and adds code_challenge / code_challenge_method
		// to the authorize URL; CompleteWithState replays the verifier
		// on the token exchange.
		RequiresPKCE bool
		// IntegrationSlug is an operator-supplied identifier used by
		// providers whose authorization URL embeds it as a path segment
		// (Vercel-style integrations). It is consumed by the provider's
		// Registration.BuildAuthURL in
		// (*provider.Registry).ApplyOAuth2Defaults. Empty for the vast
		// majority of providers.
		IntegrationSlug string

		// StateSigningKey is the HMAC key used to sign the OAuth2 state
		// token and to derive the PKCE verifier. Public clients (CIMD: no
		// client_secret, authenticated by PKCE) MUST set it to a
		// server-side secret; confidential clients leave it empty and fall
		// back to ClientSecret (see stateSalt). It is set by the probod
		// wiring, never serialized.
		StateSigningKey string

		// HTTPClient is used for the OAuth2 token-exchange request
		// issued from CompleteWithState. It must be set by callers;
		// (*provider.Registry).ApplyOAuth2Defaults assigns an
		// SSRF-protected client for production use. Tests may inject a
		// loopback-friendly one.
		HTTPClient *http.Client
	}

	OAuth2State struct {
		OrganizationID  string   `json:"oid"`
		Provider        string   `json:"provider"`
		ContinueURL     string   `json:"continue,omitempty"`
		ConnectorID     string   `json:"cid,omitempty"` // Set when reconnecting an existing connector
		RequestedScopes []string `json:"scopes,omitempty"`
		// PKCENonce carries a random per-flow nonce between Initiate and
		// Complete for providers that require PKCE. The actual
		// code_verifier is DERIVED server-side from the state salt and this
		// nonce (derivePKCEVerifier), so the verifier never appears in the
		// signed-but-unencrypted state token. The nonce is safe to expose
		// in the state parameter — it is useless without the server-side
		// salt. Set only when RequiresPKCE = true.
		PKCENonce string `json:"pn,omitempty"`
		// ProviderMetadata surfaces provider-specific extras parsed
		// from the token-exchange response (e.g. PagerDuty's
		// `subdomain`). It is populated by CompleteWithState and is
		// NEVER serialized into the state token (the field is for
		// in-process plumbing only). Consumers that need to persist
		// these values (typically the OAuth callback handler) read
		// them off the returned *OAuth2State.
		ProviderMetadata map[string]string `json:"-"`
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

func (c *OAuth2Connector) Initiate(
	ctx context.Context,
	provider string,
	organizationID gid.GID,
	opts InitiateOptions,
	r *http.Request,
) (string, error) {
	stateData := OAuth2State{
		OrganizationID:  organizationID.String(),
		Provider:        provider,
		ConnectorID:     opts.ConnectorID,
		RequestedScopes: opts.Scopes,
	}

	if r != nil {
		if continueURL := r.URL.Query().Get("continue"); continueURL != "" {
			stateData.ContinueURL = continueURL
		}
	}

	return c.InitiateWithState(ctx, stateData, opts)
}

// InitiateWithState generates an OAuth2 authorization URL with a custom state.
// This allows callers to include additional context (like SCIMBridgeID) in the state.
func (c *OAuth2Connector) InitiateWithState(
	ctx context.Context,
	stateData OAuth2State,
	opts InitiateOptions,
) (string, error) {
	// An empty salt would HMAC the state token (and derive the PKCE
	// verifier) with an empty key, making both forgeable. probod always
	// sets one, but guard at the type level so a misconfigured connector
	// fails loudly instead of issuing a forgeable state.
	salt := c.stateSalt()
	if salt == "" {
		return "", fmt.Errorf("cannot create state token: connector has no state signing key or client secret")
	}

	// For PKCE providers a per-flow nonce is generated and stored in the
	// signed state; the code verifier itself is DERIVED from salt+nonce
	// (derivePKCEVerifier) and never serialized, so it stays secret even
	// though the state is signed-not-encrypted. Non-PKCE providers skip this.
	if c.RequiresPKCE {
		nonce, err := generatePKCENonce()
		if err != nil {
			return "", fmt.Errorf("cannot generate PKCE nonce: %w", err)
		}

		stateData.PKCENonce = nonce
	}

	state, err := statelesstoken.NewToken(salt, OAuth2TokenType, OAuth2TokenTTL, stateData)
	if err != nil {
		return "", fmt.Errorf("cannot create state token: %w", err)
	}

	authCodeQuery := url.Values{}
	authCodeQuery.Set("state", state)
	authCodeQuery.Set("client_id", c.ClientID)
	authCodeQuery.Set("redirect_uri", c.RedirectURI)
	authCodeQuery.Set("response_type", "code")

	if len(opts.Scopes) > 0 {
		authCodeQuery.Set("scope", strings.Join(opts.Scopes, " "))
	}

	if c.RequiresPKCE {
		verifier := derivePKCEVerifier(c.stateSalt(), stateData.PKCENonce)
		authCodeQuery.Set("code_challenge", pkceChallenge(verifier))
		authCodeQuery.Set("code_challenge_method", "S256")
	}

	incrementalAuth := c.SupportsIncrementalAuth && opts.IncludeGrantedScopes
	if incrementalAuth {
		authCodeQuery.Set("include_granted_scopes", "true")
	}

	// Skip prompt=consent when doing incremental auth so the user sees
	// only the delta, not a full re-consent. First-install flows keep it
	// because IncludeGrantedScopes is false there.
	for k, v := range c.ExtraAuthParams {
		if incrementalAuth && k == "prompt" && v == "consent" {
			continue
		}

		authCodeQuery.Set(k, v)
	}

	u, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse auth URL: %w", err)
	}

	u.RawQuery = authCodeQuery.Encode()

	return u.String(), nil
}

// generatePKCENonce produces a 32-byte cryptographically random nonce
// encoded as base64url without padding. The nonce travels in the (signed)
// state token and is combined with the server-side state salt by
// derivePKCEVerifier to produce the actual RFC 7636 code_verifier.
func generatePKCENonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cannot read random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// stateSalt returns the HMAC key used to sign the OAuth2 state token and to
// derive the PKCE verifier. Public clients (CIMD: no client_secret,
// authenticated by PKCE) set StateSigningKey to a server-side secret;
// confidential clients fall back to ClientSecret. It must never be empty
// for a connector that issues state tokens.
func (c *OAuth2Connector) stateSalt() string {
	if c.StateSigningKey != "" {
		return c.StateSigningKey
	}

	return c.ClientSecret
}

// derivePKCEVerifier deterministically derives the RFC 7636 code_verifier
// from the server-side state salt and a per-flow nonce. Because the verifier
// is recomputed server-side at both Initiate and Complete — and never placed
// in the signed-but-unencrypted state token — it stays secret even though
// the nonce is exposed in the state parameter. This is what makes PKCE
// meaningful for public clients, whose only secret is the verifier.
func derivePKCEVerifier(salt, nonce string) string {
	return deriveHMACKey(salt, "pkce:"+nonce)
}

// deriveHMACKey derives a base64url-encoded key from a server-side secret and
// a domain-separation label via HMAC-SHA256. Distinct labels yield independent
// keys, so the same secret can safely back several purposes (the PKCE verifier
// and the connector state-signing key).
func deriveHMACKey(secret, info string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(info))

	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// pkceChallenge derives the S256 PKCE challenge from a verifier: it is
// the base64url-without-padding encoding of SHA-256(verifier) (RFC 7636
// §4.2).
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
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

	salt := c.stateSalt()
	if salt == "" {
		return nil, nil, fmt.Errorf("cannot validate state token: connector has no state signing key or client secret")
	}

	payload, err := statelesstoken.ValidateToken[OAuth2State](salt, OAuth2TokenType, stateToken)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot validate state token: %w", err)
	}

	organizationID, err := gid.ParseGID(payload.Data.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse organization ID: %w", err)
	}

	codeVerifier := ""
	if c.RequiresPKCE {
		codeVerifier = derivePKCEVerifier(c.stateSalt(), payload.Data.PKCENonce)
	}

	tokenRequest, err := c.buildTokenRequest(ctx, code, c.RedirectURI, codeVerifier)
	if err != nil {
		return nil, nil, err
	}

	tokenResp, err := c.HTTPClient.Do(tokenRequest)
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

	grantedScope := rawToken.Scope
	if grantedScope == "" {
		// RFC 6749 §5.1: scope is OPTIONAL when identical to the
		// requested scope. Fall back to what we asked for so
		// subsequent reconnect diffs have a meaningful base.
		grantedScope = FormatScopeString(payload.Data.RequestedScopes)
	}

	oauth2Conn := OAuth2Connection{
		AccessToken:  rawToken.AccessToken,
		RefreshToken: rawToken.RefreshToken,
		TokenType:    rawToken.TokenType,
		Scope:        grantedScope,
	}

	// Convert expires_in (seconds) to expires_at (absolute time)
	if rawToken.ExpiresIn > 0 {
		oauth2Conn.ExpiresAt = time.Now().Add(time.Duration(rawToken.ExpiresIn) * time.Second)
	}

	if payload.Data.Provider == SlackProvider {
		conn, _, err := ParseSlackTokenResponse(body, oauth2Conn, organizationID)
		return conn, &payload.Data, err
	}

	AbsorbPagerDutyTokenResponse(&payload.Data, body)

	return &oauth2Conn, &payload.Data, nil
}

// DeriveConnectorStateKey derives the HMAC key used to sign connector
// OAuth2 state tokens (and PKCE verifiers) for public clients, from a
// server-side secret (the active OAuth2 server signing key). The domain
// separator avoids reusing the raw server key directly for an unrelated
// purpose. probod calls this once at startup and assigns the result to
// each public client's StateSigningKey.
//
// NOTE: the key is derived from the single ACTIVE OAuth2 server signing key.
// Rotating that key changes the derived state key, so connector OAuth flows
// started within the state token's 10-minute TTL window across a rotation
// will fail validation and must be retried. A dedicated, independently
// rotated connector-state key (HMAC key set) is a future improvement.
func DeriveConnectorStateKey(serverSecret string) string {
	return deriveHMACKey(serverSecret, "probo/connector/oauth2-state-key")
}

func basicAuthHeader(clientID, clientSecret string) string {
	credentials := clientID + ":" + clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}

// buildTokenRequest creates the HTTP request for the token exchange, branching
// on c.TokenEndpointAuth to support different provider requirements. When
// codeVerifier is non-empty (PKCE-enabled providers), it is replayed as
// `code_verifier` in the request body.
func (c *OAuth2Connector) buildTokenRequest(ctx context.Context, code, redirectURI, codeVerifier string) (*http.Request, error) {
	switch c.TokenEndpointAuth {
	case "basic-json":
		// JSON body with Basic auth header (Notion).
		body := map[string]string{
			"code":         code,
			"redirect_uri": redirectURI,
			"grant_type":   "authorization_code",
		}
		if codeVerifier != "" {
			body["code_verifier"] = codeVerifier
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

		if codeVerifier != "" {
			formData.Set("code_verifier", codeVerifier)
		}

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

	case "none":
		// Public client (CIMD): client_id in the body, authenticated by
		// the PKCE code_verifier. No client_secret is sent — the provider
		// advertises token_endpoint_auth_method "none".
		formData := url.Values{}
		formData.Set("client_id", c.ClientID)
		formData.Set("code", code)
		formData.Set("redirect_uri", redirectURI)
		formData.Set("grant_type", "authorization_code")

		if codeVerifier != "" {
			formData.Set("code_verifier", codeVerifier)
		}

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

	default:
		// "post-form" or empty: credentials in form body (Slack, HubSpot, GitHub, etc.).
		formData := url.Values{}
		formData.Set("client_id", c.ClientID)
		formData.Set("client_secret", c.ClientSecret)
		formData.Set("code", code)
		formData.Set("redirect_uri", redirectURI)
		formData.Set("grant_type", "authorization_code")

		if codeVerifier != "" {
			formData.Set("code_verifier", codeVerifier)
		}

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

func (c *OAuth2Connection) Scopes() []string {
	return ParseScopeString(c.Scope)
}

func (c *OAuth2Connection) Client(ctx context.Context) (*http.Client, error) {
	return c.ClientWithOptions(ctx)
}

// ClientWithOptions returns an HTTP client with the given options.
// Use this to add logging and tracing to the HTTP client.
//
// SSRF protection is always enabled: the underlying connector URL
// (for example a 1Password SCIM bridge URL) is customer-supplied,
// so dials to private, loopback, or other reserved address ranges
// are refused. Hardcoded provider hosts on public IPs are
// unaffected.
func (c *OAuth2Connection) ClientWithOptions(ctx context.Context, opts ...httpclient.Option) (*http.Client, error) {
	opts = append(opts, httpclient.WithSSRFProtection())
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

	// All HTTP traffic on this path (token refresh + API calls)
	// must reject private/loopback/reserved peer IPs because the
	// configured TokenURL or API host can be customer-influenced.
	opts = append(opts, httpclient.WithSSRFProtection())

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

	// TokenURL is stored from customer-supplied connector settings;
	// reject dials to private/loopback/reserved peer IPs.
	opts = append(opts, httpclient.WithSSRFProtection())

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
	// tokenType from the provider's OAuth response is not always a valid HTTP
	// auth scheme (Slack returns "bot" / "user", some providers send an empty
	// string), so we always send "Bearer" -- the only scheme any connector in
	// this codebase actually needs.
	req2.Header.Set("Authorization", "Bearer "+t.token)

	return t.underlying.RoundTrip(req2)
}
