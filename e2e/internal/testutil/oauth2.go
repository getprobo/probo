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

package testutil

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	OAuth2TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}

	OAuth2ErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description,omitempty"`
	}

	OAuth2RegisterResponse struct {
		ClientID                string   `json:"client_id"`
		ClientSecret            string   `json:"client_secret,omitempty"`
		ClientName              string   `json:"client_name"`
		Visibility              string   `json:"visibility"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		Scopes                  string   `json:"scopes"`
	}

	OAuth2IntrospectResponse struct {
		Active    bool   `json:"active"`
		Scope     string `json:"scope,omitempty"`
		ClientID  string `json:"client_id,omitempty"`
		Sub       string `json:"sub,omitempty"`
		Exp       int64  `json:"exp,omitempty"`
		Iat       int64  `json:"iat,omitempty"`
		TokenType string `json:"token_type,omitempty"`
	}

	OAuth2DeviceAuthResponse struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}

	OAuth2DiscoveryResponse struct {
		Issuer                            string   `json:"issuer"`
		AuthorizationEndpoint             string   `json:"authorization_endpoint"`
		TokenEndpoint                     string   `json:"token_endpoint"`
		UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
		JwksURI                           string   `json:"jwks_uri"`
		RegistrationEndpoint              string   `json:"registration_endpoint"`
		IntrospectionEndpoint             string   `json:"introspection_endpoint"`
		RevocationEndpoint                string   `json:"revocation_endpoint"`
		DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
		ScopesSupported                   []string `json:"scopes_supported"`
		ResponseTypesSupported            []string `json:"response_types_supported"`
		GrantTypesSupported               []string `json:"grant_types_supported"`
		TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
		SubjectTypesSupported             []string `json:"subject_types_supported"`
		IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
		CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
		ClaimsSupported                   []string `json:"claims_supported"`
	}

	OAuth2JWKSResponse struct {
		Keys []map[string]any `json:"keys"`
	}

	OAuth2UserInfoResponse struct {
		Sub           string `json:"sub"`
		Email         string `json:"email,omitempty"`
		EmailVerified bool   `json:"email_verified,omitempty"`
		Name          string `json:"name,omitempty"`
	}

	OAuth2HTTPResponse struct {
		StatusCode int
		Header     http.Header
		Body       []byte
	}
)

func oauth2BaseURL(c *Client) string {
	return c.BaseURL() + "/api/connect/v1/oauth2"
}

func postForm(
	httpClient *http.Client,
	url string,
	values url.Values,
) (*OAuth2HTTPResponse, error) {
	resp, err := httpClient.PostForm(url, values)
	if err != nil {
		return nil, fmt.Errorf("cannot post form: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

func postJSON(
	httpClient *http.Client,
	url string,
	payload any,
) (*OAuth2HTTPResponse, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal payload: %w", err)
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("cannot post json: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

func getJSON(
	httpClient *http.Client,
	url string,
	headers map[string]string,
) (*OAuth2HTTPResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

func postFormWithBasicAuth(
	httpClient *http.Client,
	rawURL string,
	values url.Values,
	username, password string,
) (*OAuth2HTTPResponse, error) {
	req, err := http.NewRequest(
		"POST",
		rawURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(username, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// OAuth2Discovery fetches the OpenID Connect discovery document.
func OAuth2Discovery(c *Client) (*OAuth2DiscoveryResponse, *OAuth2HTTPResponse, error) {
	raw, err := getJSON(c.HTTPClient(), c.BaseURL()+"/.well-known/openid-configuration", nil)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2DiscoveryResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode discovery response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2JWKS fetches the JSON Web Key Set.
func OAuth2JWKS(c *Client) (*OAuth2JWKSResponse, *OAuth2HTTPResponse, error) {
	raw, err := getJSON(c.HTTPClient(), oauth2BaseURL(c)+"/jwks", nil)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2JWKSResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode jwks response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2RegisterClient registers a new OAuth2 client via dynamic registration.
func OAuth2RegisterClient(
	c *Client,
	input map[string]any,
) (*OAuth2RegisterResponse, *OAuth2HTTPResponse, error) {
	raw, err := postJSON(c.HTTPClient(), oauth2BaseURL(c)+"/register", input)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusCreated {
		return nil, raw, nil
	}

	var result OAuth2RegisterResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode register response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2Authorize performs a GET to the authorize endpoint and returns the
// raw HTTP response without following redirects.
func OAuth2Authorize(
	c *Client,
	params url.Values,
) (*OAuth2HTTPResponse, error) {
	noRedirectClient := &http.Client{
		Jar:     c.HTTPClient().Jar,
		Timeout: c.HTTPClient().Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	reqURL := oauth2BaseURL(c) + "/authorize?" + params.Encode()
	resp, err := noRedirectClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get authorize: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// OAuth2AuthorizeCodeFromRedirect extracts the authorization code from the
// Location header of a 302 response.
func OAuth2AuthorizeCodeFromRedirect(resp *OAuth2HTTPResponse) (string, error) {
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("no Location header in redirect response (status=%d body=%s)", resp.StatusCode, string(resp.Body))
	}

	u, err := url.Parse(loc)
	if err != nil {
		return "", fmt.Errorf("cannot parse redirect url: %w", err)
	}

	code := u.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("no code in redirect url: %s", loc)
	}

	return code, nil
}

// OAuth2ConsentApprove submits the consent form with action=approve.
func OAuth2ConsentApprove(c *Client, consentID string) (*OAuth2HTTPResponse, error) {
	noRedirectClient := &http.Client{
		Jar:     c.HTTPClient().Jar,
		Timeout: c.HTTPClient().Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	values := url.Values{
		"consent_id": {consentID},
		"action":     {"approve"},
	}

	req, err := http.NewRequest(
		"POST",
		oauth2BaseURL(c)+"/authorize",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot post consent: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// OAuth2ConsentDeny submits the consent form with action=deny.
func OAuth2ConsentDeny(c *Client, consentID string) (*OAuth2HTTPResponse, error) {
	noRedirectClient := &http.Client{
		Jar:     c.HTTPClient().Jar,
		Timeout: c.HTTPClient().Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	values := url.Values{
		"consent_id": {consentID},
		"action":     {"deny"},
	}

	req, err := http.NewRequest(
		"POST",
		oauth2BaseURL(c)+"/authorize",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot post consent deny: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// OAuth2TokenWithCode exchanges an authorization code for tokens.
func OAuth2TokenWithCode(
	c *Client,
	clientID, clientSecret, code, redirectURI, codeVerifier string,
) (*OAuth2TokenResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}

	if codeVerifier != "" {
		values.Set("code_verifier", codeVerifier)
	}

	raw, err := postFormWithBasicAuth(
		c.HTTPClient(),
		oauth2BaseURL(c)+"/token",
		values,
		clientID,
		clientSecret,
	)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2TokenResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode token response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2TokenWithCodePostAuth exchanges an authorization code for tokens
// using client_secret_post authentication (credentials in POST body).
func OAuth2TokenWithCodePostAuth(
	c *Client,
	clientID, clientSecret, code, redirectURI, codeVerifier string,
) (*OAuth2TokenResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	if codeVerifier != "" {
		values.Set("code_verifier", codeVerifier)
	}

	raw, err := postForm(c.HTTPClient(), oauth2BaseURL(c)+"/token", values)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2TokenResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode token response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2TokenWithRefreshToken refreshes tokens using a refresh token.
func OAuth2TokenWithRefreshToken(
	c *Client,
	clientID, clientSecret, refreshToken string,
) (*OAuth2TokenResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	raw, err := postFormWithBasicAuth(
		c.HTTPClient(),
		oauth2BaseURL(c)+"/token",
		values,
		clientID,
		clientSecret,
	)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2TokenResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode token response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2TokenWithDeviceCode polls the token endpoint for device code grant.
func OAuth2TokenWithDeviceCode(
	c *Client,
	clientID, deviceCode string,
) (*OAuth2TokenResponse, *OAuth2ErrorResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_id":   {clientID},
		"device_code": {deviceCode},
	}

	raw, err := postForm(c.HTTPClient(), oauth2BaseURL(c)+"/token", values)
	if err != nil {
		return nil, nil, nil, err
	}

	if raw.StatusCode == http.StatusOK {
		var result OAuth2TokenResponse
		if err := json.Unmarshal(raw.Body, &result); err != nil {
			return nil, nil, raw, fmt.Errorf("cannot decode token response: %w", err)
		}
		return &result, nil, raw, nil
	}

	var errResp OAuth2ErrorResponse
	if err := json.Unmarshal(raw.Body, &errResp); err != nil {
		return nil, nil, raw, nil
	}

	return nil, &errResp, raw, nil
}

// OAuth2TokenRaw posts arbitrary form values to the token endpoint.
func OAuth2TokenRaw(
	c *Client,
	values url.Values,
) (*OAuth2HTTPResponse, error) {
	raw, err := postForm(c.HTTPClient(), oauth2BaseURL(c)+"/token", values)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// OAuth2TokenRawWithBasicAuth posts form values to the token endpoint with
// HTTP Basic authentication.
func OAuth2TokenRawWithBasicAuth(
	c *Client,
	values url.Values,
	username, password string,
) (*OAuth2HTTPResponse, error) {
	raw, err := postFormWithBasicAuth(
		c.HTTPClient(),
		oauth2BaseURL(c)+"/token",
		values,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// OAuth2DeviceAuth starts the device authorization flow.
func OAuth2DeviceAuth(
	c *Client,
	clientID, scope string,
) (*OAuth2DeviceAuthResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"client_id": {clientID},
	}

	if scope != "" {
		values.Set("scope", scope)
	}

	raw, err := postForm(c.HTTPClient(), oauth2BaseURL(c)+"/device", values)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2DeviceAuthResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode device auth response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2DeviceVerify submits the device verification form as an authenticated
// user.
func OAuth2DeviceVerify(c *Client, userCode string) (*OAuth2HTTPResponse, error) {
	values := url.Values{
		"user_code": {userCode},
	}

	noRedirectClient := &http.Client{
		Jar:     c.HTTPClient().Jar,
		Timeout: c.HTTPClient().Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(
		"POST",
		oauth2BaseURL(c)+"/device/verify",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot post device verify: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return &OAuth2HTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// OAuth2UserInfo fetches the UserInfo endpoint with a Bearer token.
func OAuth2UserInfo(
	c *Client,
	accessToken string,
) (*OAuth2UserInfoResponse, *OAuth2HTTPResponse, error) {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}

	raw, err := getJSON(c.HTTPClient(), oauth2BaseURL(c)+"/userinfo", headers)
	if err != nil {
		return nil, nil, err
	}

	if raw.StatusCode != http.StatusOK {
		return nil, raw, nil
	}

	var result OAuth2UserInfoResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode userinfo response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2UserInfoRaw fetches the UserInfo endpoint with custom query params
// and no Authorization header (for testing that query/body tokens are rejected).
func OAuth2UserInfoRaw(
	c *Client,
	queryParams url.Values,
) (*OAuth2HTTPResponse, error) {
	reqURL := oauth2BaseURL(c) + "/userinfo"
	if len(queryParams) > 0 {
		reqURL += "?" + queryParams.Encode()
	}

	raw, err := getJSON(c.HTTPClient(), reqURL, nil)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// OAuth2Introspect introspects a token using client credentials.
func OAuth2Introspect(
	c *Client,
	clientID, clientSecret, token string,
) (*OAuth2IntrospectResponse, *OAuth2HTTPResponse, error) {
	values := url.Values{
		"token": {token},
	}

	raw, err := postFormWithBasicAuth(
		c.HTTPClient(),
		oauth2BaseURL(c)+"/introspect",
		values,
		clientID,
		clientSecret,
	)
	if err != nil {
		return nil, nil, err
	}

	var result OAuth2IntrospectResponse
	if err := json.Unmarshal(raw.Body, &result); err != nil {
		return nil, raw, fmt.Errorf("cannot decode introspect response: %w", err)
	}

	return &result, raw, nil
}

// OAuth2Revoke revokes a token using client credentials.
func OAuth2Revoke(
	c *Client,
	clientID, clientSecret, token string,
) (*OAuth2HTTPResponse, error) {
	values := url.Values{
		"token": {token},
	}

	raw, err := postFormWithBasicAuth(
		c.HTTPClient(),
		oauth2BaseURL(c)+"/revoke",
		values,
		clientID,
		clientSecret,
	)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// PKCE helpers

// GeneratePKCE generates a code_verifier and code_challenge (S256) pair.
func GeneratePKCE() (verifier, challenge string) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	verifier = string(b)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge
}

// ExtractConsentID extracts the consent_id from a consent HTML page.
func ExtractConsentID(body []byte) (string, error) {
	s := string(body)

	needle := `name="consent_id" value="`
	idx := strings.Index(s, needle)
	if idx == -1 {
		return "", fmt.Errorf("consent_id not found in page")
	}

	start := idx + len(needle)
	end := strings.Index(s[start:], `"`)
	if end == -1 {
		return "", fmt.Errorf("malformed consent_id value")
	}

	return s[start : start+end], nil
}

// OAuth2PerformAuthorizationCodeFlow performs the full authorization code flow
// and returns the token response. This is a convenience function for tests that
// need tokens but are not testing the authorization flow itself.
func OAuth2PerformAuthorizationCodeFlow(
	t testing.TB,
	c *Client,
	clientID, clientSecret, redirectURI string,
) *OAuth2TokenResponse {
	t.Helper()

	verifier, challenge := GeneratePKCE()

	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":         {redirectURI},
		"response_type":        {"code"},
		"scope":                {"openid email profile"},
		"state":                {"test-state"},
		"code_challenge":       {challenge},
		"code_challenge_method": {"S256"},
	}

	authResp, err := OAuth2Authorize(c, params)
	require.NoError(t, err)

	var code string
	if authResp.StatusCode == http.StatusOK {
		consentID, err := ExtractConsentID(authResp.Body)
		require.NoError(t, err)

		consentResp, err := OAuth2ConsentApprove(c, consentID)
		require.NoError(t, err)
		require.Equal(t, http.StatusFound, consentResp.StatusCode)

		code, err = OAuth2AuthorizeCodeFromRedirect(consentResp)
		require.NoError(t, err)
	} else {
		require.Equal(t, http.StatusFound, authResp.StatusCode)
		code, err = OAuth2AuthorizeCodeFromRedirect(authResp)
		require.NoError(t, err)
	}

	tokenResp, raw, err := OAuth2TokenWithCode(
		c,
		clientID,
		clientSecret,
		code,
		redirectURI,
		verifier,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, raw.StatusCode, "token exchange failed: %s", string(raw.Body))
	require.NotNil(t, tokenResp)

	return tokenResp
}
