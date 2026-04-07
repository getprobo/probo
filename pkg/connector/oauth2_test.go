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

package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestBuildTokenRequest_PostForm(t *testing.T) {
	t.Parallel()

	t.Run("empty token endpoint auth", func(t *testing.T) {
		t.Parallel()

		connector := &OAuth2Connector{
			ClientID:          "my-client-id",
			ClientSecret:      "my-client-secret",
			TokenURL:          "https://provider.example.com/oauth/token",
			TokenEndpointAuth: "",
		}

		req, err := connector.buildTokenRequest(
			context.Background(),
			"test-code",
			"https://example.com/callback",
		)
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://provider.example.com/oauth/token", req.URL.String())
		assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", req.Header.Get("Content-Type"))
		assert.Empty(t, req.Header.Get("Authorization"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		formValues, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "my-client-id", formValues.Get("client_id"))
		assert.Equal(t, "my-client-secret", formValues.Get("client_secret"))
		assert.Equal(t, "test-code", formValues.Get("code"))
		assert.Equal(t, "https://example.com/callback", formValues.Get("redirect_uri"))
		assert.Equal(t, "authorization_code", formValues.Get("grant_type"))
	})

	t.Run("explicit post-form token endpoint auth", func(t *testing.T) {
		t.Parallel()

		connector := &OAuth2Connector{
			ClientID:          "my-client-id",
			ClientSecret:      "my-client-secret",
			TokenURL:          "https://provider.example.com/oauth/token",
			TokenEndpointAuth: "post-form",
		}

		req, err := connector.buildTokenRequest(
			context.Background(),
			"test-code",
			"https://example.com/callback",
		)
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Empty(t, req.Header.Get("Authorization"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		formValues, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "my-client-id", formValues.Get("client_id"))
		assert.Equal(t, "my-client-secret", formValues.Get("client_secret"))
		assert.Equal(t, "test-code", formValues.Get("code"))
		assert.Equal(t, "https://example.com/callback", formValues.Get("redirect_uri"))
		assert.Equal(t, "authorization_code", formValues.Get("grant_type"))
	})
}

func TestBuildTokenRequest_BasicForm(t *testing.T) {
	t.Parallel()

	connector := &OAuth2Connector{
		ClientID:          "my-client-id",
		ClientSecret:      "my-client-secret",
		TokenURL:          "https://provider.example.com/oauth/token",
		TokenEndpointAuth: "basic-form",
	}

	req, err := connector.buildTokenRequest(
		context.Background(),
		"test-code",
		"https://example.com/callback",
	)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "https://provider.example.com/oauth/token", req.URL.String())
	assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", req.Header.Get("Content-Type"))

	// Verify Basic auth header
	authHeader := req.Header.Get("Authorization")
	require.NotEmpty(t, authHeader)

	expectedCredentials := base64.StdEncoding.EncodeToString([]byte("my-client-id:my-client-secret"))
	assert.Equal(t, "Basic "+expectedCredentials, authHeader)

	// Verify body does NOT contain client credentials
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	formValues, err := url.ParseQuery(string(body))
	require.NoError(t, err)

	assert.Empty(t, formValues.Get("client_id"))
	assert.Empty(t, formValues.Get("client_secret"))
	assert.Equal(t, "test-code", formValues.Get("code"))
	assert.Equal(t, "https://example.com/callback", formValues.Get("redirect_uri"))
	assert.Equal(t, "authorization_code", formValues.Get("grant_type"))
}

func TestBuildTokenRequest_BasicJSON(t *testing.T) {
	t.Parallel()

	connector := &OAuth2Connector{
		ClientID:          "my-client-id",
		ClientSecret:      "my-client-secret",
		TokenURL:          "https://provider.example.com/oauth/token",
		TokenEndpointAuth: "basic-json",
	}

	req, err := connector.buildTokenRequest(
		context.Background(),
		"test-code",
		"https://example.com/callback",
	)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "https://provider.example.com/oauth/token", req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

	// Verify Basic auth header
	authHeader := req.Header.Get("Authorization")
	require.NotEmpty(t, authHeader)

	expectedCredentials := base64.StdEncoding.EncodeToString([]byte("my-client-id:my-client-secret"))
	assert.Equal(t, "Basic "+expectedCredentials, authHeader)

	// Verify body is valid JSON
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	var jsonBody map[string]string
	err = json.Unmarshal(body, &jsonBody)
	require.NoError(t, err)

	assert.Equal(t, "test-code", jsonBody["code"])
	assert.Equal(t, "https://example.com/callback", jsonBody["redirect_uri"])
	assert.Equal(t, "authorization_code", jsonBody["grant_type"])

	// JSON body must NOT contain client credentials
	_, hasClientID := jsonBody["client_id"]
	_, hasClientSecret := jsonBody["client_secret"]
	assert.False(t, hasClientID, "JSON body should not contain client_id")
	assert.False(t, hasClientSecret, "JSON body should not contain client_secret")
}

func TestClientCredentialsClient(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify Basic auth header is present
		authHeader := r.Header.Get("Authorization")
		assert.NotEmpty(t, authHeader)

		decoded, err := base64.StdEncoding.DecodeString(authHeader[len("Basic "):])
		require.NoError(t, err)
		assert.Equal(t, "cc-client-id:cc-client-secret", string(decoded))

		// Verify form body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		formValues, err := url.ParseQuery(string(body))
		require.NoError(t, err)
		assert.Equal(t, "client_credentials", formValues.Get("grant_type"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "test-token", "expires_in": 3600, "token_type": "Bearer"}`))
	}))
	defer server.Close()

	beforeRequest := time.Now()

	conn := &OAuth2Connection{
		GrantType:    OAuth2GrantTypeClientCredentials,
		ClientID:     "cc-client-id",
		ClientSecret: "cc-client-secret",
		TokenURL:     server.URL,
	}

	client, err := conn.clientCredentialsClient(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, "test-token", conn.AccessToken)
	assert.Equal(t, "Bearer", conn.TokenType)

	// ExpiresAt should be approximately now + 1 hour
	expectedExpiry := beforeRequest.Add(1 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, conn.ExpiresAt, 5*time.Second)
}

func TestClientCredentialsClient_ReusesValidToken(t *testing.T) {
	t.Parallel()

	conn := &OAuth2Connection{
		GrantType:   OAuth2GrantTypeClientCredentials,
		AccessToken: "existing-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	// No test server -- calling clientCredentialsClient should not make any HTTP request
	// because the token is still valid.
	client, err := conn.clientCredentialsClient(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, "existing-token", conn.AccessToken)
}

func TestInitiateWithState_Scopes(t *testing.T) {
	t.Parallel()

	t.Run("scopes are joined and set on auth URL", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURI:  "https://example.com/cb",
			AuthURL:      "https://provider.example.com/authorize",
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{Scopes: []string{"read:user", "write:user"}},
			nil,
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.Equal(t, "read:user write:user", parsed.Query().Get("scope"))
	})

	t.Run("empty scopes omits scope parameter", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURI:  "https://example.com/cb",
			AuthURL:      "https://provider.example.com/authorize",
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{},
			nil,
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.False(t, parsed.Query().Has("scope"), "scope param should be absent when no scopes provided")
	})
}
