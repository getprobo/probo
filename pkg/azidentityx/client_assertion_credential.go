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

// Package azidentityx acquires Entra access tokens with a client assertion.
// The token request matches azure-sdk-for-go/sdk/azidentity, and the scope is
// only the scopes the caller asked for.
package azidentityx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/uuid"
)

const (
	clientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	tokenContentType    = "application/x-www-form-urlencoded; charset=utf-8"

	// tokenRefreshSkew matches MSAL: a cached access token is reused until five
	// minutes before it expires.
	tokenRefreshSkew = 5 * time.Minute

	// tokenRequestTimeout matches MSAL: a token call with no deadline of its
	// own is bounded so it cannot wait forever.
	tokenRequestTimeout = 30 * time.Second

	maxTokenResponseBytes = 1 << 20
)

type (
	// ClientAssertionCredential exchanges a caller-minted client assertion for
	// an Entra access token. The request matches azidentity's
	// ClientAssertionCredential, except the scope is only what the caller
	// asked for: MSAL would also append openid, offline_access, and profile.
	ClientAssertionCredential struct {
		tenantID   string
		clientID   string
		authority  string
		httpClient *http.Client
		assertion  func(context.Context) (string, error)
		now        func() time.Time

		mu    sync.Mutex
		cache map[string]cachedAccessToken
	}

	cachedAccessToken struct {
		token     azcore.AccessToken
		refreshAt time.Time
	}

	// TokenEndpointError is a non-success response from the Entra token
	// endpoint. The body stays inside this package; callers read error codes
	// through ErrorCodes.
	TokenEndpointError struct {
		statusCode int
		body       []byte
	}

	tokenEndpointResponse struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   tokenSeconds    `json:"expires_in"`
		ExpiresOn   json.RawMessage `json:"expires_on"`
		RefreshIn   tokenSeconds    `json:"refresh_in"`
		Error       string          `json:"error"`
		Description string          `json:"error_description"`
		ErrorCodes  []int           `json:"error_codes"`
	}

	// tokenSeconds accepts the number or numeric string MSAL accepts for
	// expires_in and refresh_in.
	tokenSeconds int64
)

var _ azcore.TokenCredential = (*ClientAssertionCredential)(nil)

// NewClientAssertionCredential returns a credential that exchanges assertion
// for Entra access tokens. assertion may run while the credential holds its
// lock, and must not call GetToken on the same credential.
func NewClientAssertionCredential(
	httpClient *http.Client,
	authority string,
	tenantID string,
	clientID string,
	assertion func(context.Context) (string, error),
) (*ClientAssertionCredential, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("cannot create azure token credential: http client is required")
	}

	if authority == "" {
		return nil, fmt.Errorf("cannot create azure token credential: authority is required")
	}

	if assertion == nil {
		return nil, fmt.Errorf("cannot create azure token credential: assertion callback is required")
	}

	return &ClientAssertionCredential{
		tenantID:   tenantID,
		clientID:   clientID,
		authority:  authority,
		httpClient: httpClient,
		assertion:  assertion,
		now:        time.Now,
		cache:      map[string]cachedAccessToken{},
	}, nil
}

// NewTokenEndpointError builds the error GetToken returns when the token
// endpoint rejects the request. body is the raw JSON response.
func NewTokenEndpointError(statusCode int, body []byte) *TokenEndpointError {
	return &TokenEndpointError{
		statusCode: statusCode,
		body:       body,
	}
}

func (c *ClientAssertionCredential) GetToken(
	ctx context.Context,
	options policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	scope, err := scopeKey(options.Scopes)
	if err != nil {
		return azcore.AccessToken{}, err
	}

	// One exchange at a time, as ClientAssertionCredential does, so two
	// callers for the same scope share one token request.
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, ok := c.cache[scope]; ok && c.now().Before(cached.refreshAt) {
		return cached.token, nil
	}

	assertion, err := c.assertion(ctx)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot mint azure client assertion: %w", err)
	}

	token, err := c.exchange(ctx, scope, assertion)
	if err != nil {
		return azcore.AccessToken{}, err
	}

	c.cache[scope] = cachedAccessToken{
		token:     token,
		refreshAt: tokenRefreshAt(token),
	}

	return token, nil
}

// tokenRefreshAt is the earlier of Entra's refresh_in and five minutes
// before the access token expires.
func tokenRefreshAt(token azcore.AccessToken) time.Time {
	refreshAt := token.ExpiresOn.Add(-tokenRefreshSkew)
	if !token.RefreshOn.IsZero() && token.RefreshOn.Before(refreshAt) {
		return token.RefreshOn
	}

	return refreshAt
}

func (c *ClientAssertionCredential) exchange(
	ctx context.Context,
	scope string,
	assertion string,
) (azcore.AccessToken, error) {
	endpoint, err := url.JoinPath(c.authority, c.tenantID, "oauth2", "v2.0", "token")
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot build azure token endpoint: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("scope", scope)
	form.Set("client_assertion_type", clientAssertionType)
	form.Set("client_assertion", assertion)
	form.Set("client_info", "1")

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, tokenRequestTimeout)
		defer cancel()
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot create azure token request: %w", err)
	}

	request.Header.Set("Content-Type", tokenContentType)
	request.Header.Set("client-request-id", uuid.NewString())
	request.Header.Set("Return-Client-Request-Id", "false")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot reach azure token endpoint: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxTokenResponseBytes+1))
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot read azure token response: %w", err)
	}

	if len(body) > maxTokenResponseBytes {
		return azcore.AccessToken{}, fmt.Errorf("cannot read azure token response: response exceeds %d bytes", maxTokenResponseBytes)
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return azcore.AccessToken{}, NewTokenEndpointError(response.StatusCode, body)
	}

	var payload tokenEndpointResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return azcore.AccessToken{}, fmt.Errorf("cannot decode azure token response: %w", err)
	}

	if payload.AccessToken == "" {
		return azcore.AccessToken{}, fmt.Errorf("cannot acquire azure access token: response has no access token")
	}

	expiresOn, err := tokenExpiresOn(c.now(), payload)
	if err != nil {
		return azcore.AccessToken{}, err
	}

	token := azcore.AccessToken{
		Token:     payload.AccessToken,
		ExpiresOn: expiresOn,
	}
	if payload.RefreshIn > 0 {
		token.RefreshOn = c.now().Add(time.Duration(payload.RefreshIn) * time.Second).UTC()
	}

	return token, nil
}

func (s *tokenSeconds) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	value := strings.Trim(string(data), `"`)
	if value == "" {
		return nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse token lifetime %s: %w", value, err)
	}

	*s = tokenSeconds(parsed)

	return nil
}

// tokenExpiresOn follows MSAL: a present expires_on wins, otherwise
// expires_in is a number of seconds from now.
func tokenExpiresOn(now time.Time, payload tokenEndpointResponse) (time.Time, error) {
	expiresOn, ok, err := parseExpiresOn(payload.ExpiresOn)
	if err != nil {
		return time.Time{}, err
	}

	if ok {
		return expiresOn.UTC(), nil
	}

	if payload.ExpiresIn > 0 {
		return now.Add(time.Duration(payload.ExpiresIn) * time.Second).UTC(), nil
	}

	return time.Time{}, fmt.Errorf("cannot acquire azure access token: response has no expiry")
}

func parseExpiresOn(raw json.RawMessage) (time.Time, bool, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}, false, nil
	}

	var unix float64
	if err := json.Unmarshal(raw, &unix); err == nil {
		return time.Unix(int64(unix), 0), true, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return time.Time{}, false, fmt.Errorf("cannot acquire azure access token: invalid expires_on")
	}

	if text == "" {
		return time.Time{}, false, nil
	}

	if unixSeconds, err := strconv.ParseInt(text, 10, 64); err == nil {
		return time.Unix(unixSeconds, 0), true, nil
	}

	for _, format := range []string{
		"01/02/2006 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
	} {
		parsed, err := time.Parse(format, text)
		if err == nil {
			return parsed, true, nil
		}
	}

	return time.Time{}, false, fmt.Errorf("cannot acquire azure access token: invalid expires_on")
}

func (e *TokenEndpointError) Error() string {
	var payload tokenEndpointResponse
	if json.Unmarshal(e.body, &payload) == nil && payload.Error != "" {
		if payload.Description != "" {
			return fmt.Sprintf(
				"azure token endpoint returned HTTP %d: %s: %s",
				e.statusCode,
				payload.Error,
				payload.Description,
			)
		}

		return fmt.Sprintf("azure token endpoint returned HTTP %d: %s", e.statusCode, payload.Error)
	}

	return fmt.Sprintf("azure token endpoint returned HTTP %d", e.statusCode)
}

// ErrorCodes returns the Entra error_codes from the response body.
func (e *TokenEndpointError) ErrorCodes() []int {
	var payload tokenEndpointResponse
	if json.Unmarshal(e.body, &payload) != nil {
		return nil
	}

	return payload.ErrorCodes
}

func scopeKey(scopes []string) (string, error) {
	parts := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}

		parts = append(parts, scope)
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("cannot acquire azure access token: scope is required")
	}

	return strings.Join(parts, " "), nil
}
