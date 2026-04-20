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
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
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

	// httptest binds to loopback, which the SSRF-protected default
	// transport refuses; relax just for this test.
	client, err := conn.clientCredentialsClient(context.Background(), httpclient.WithSSRFAllowLoopback())
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
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.False(t, parsed.Query().Has("scope"), "scope param should be absent when no scopes provided")
	})

	t.Run("include_granted_scopes set when provider supports and caller requests", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: true,
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{
				Scopes:               []string{"read:user"},
				IncludeGrantedScopes: true,
			},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.Equal(t, "true", parsed.Query().Get("include_granted_scopes"))
	})

	t.Run("include_granted_scopes absent when provider does not support it", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: false,
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{
				Scopes:               []string{"read:user"},
				IncludeGrantedScopes: true,
			},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.False(t, parsed.Query().Has("include_granted_scopes"))
	})

	t.Run("include_granted_scopes absent when caller does not request", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: true,
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{Scopes: []string{"read:user"}},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.False(t, parsed.Query().Has("include_granted_scopes"))
	})

	t.Run("prompt=consent skipped when incremental auth is active", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: true,
			ExtraAuthParams: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{
				Scopes:               []string{"read:user"},
				IncludeGrantedScopes: true,
			},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.Equal(t, "offline", parsed.Query().Get("access_type"))
		assert.False(t, parsed.Query().Has("prompt"), "prompt=consent should be skipped when doing incremental auth on a provider that supports it")
		assert.Equal(t, "true", parsed.Query().Get("include_granted_scopes"))
	})

	t.Run("prompt=consent preserved on first install", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: true,
			ExtraAuthParams: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{
				Scopes:               []string{"read:user"},
				IncludeGrantedScopes: false, // first install, no existing grant
			},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.Equal(t, "offline", parsed.Query().Get("access_type"))
		assert.Equal(t, "consent", parsed.Query().Get("prompt"), "prompt=consent must still fire on first install so Google issues a refresh token")
		assert.False(t, parsed.Query().Has("include_granted_scopes"))
	})

	t.Run("prompt=consent preserved when provider does not support incremental auth", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:                "id",
			ClientSecret:            "secret",
			RedirectURI:             "https://example.com/cb",
			AuthURL:                 "https://provider.example.com/authorize",
			SupportsIncrementalAuth: false,
			ExtraAuthParams: map[string]string{
				"prompt": "consent",
			},
		}

		orgID := gid.New(gid.NewTenantID(), 0)

		u, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{
				Scopes:               []string{"read:user"},
				IncludeGrantedScopes: true, // caller requested, but provider does not support
			},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(u)
		require.NoError(t, err)
		assert.Equal(t, "consent", parsed.Query().Get("prompt"), "prompt=consent must not be skipped for providers that do not support incremental auth")
	})
}

// TestCompleteWithState_ScopeFallback verifies that when the provider's
// token endpoint returns a successful token response that omits the
// `scope` field (which RFC 6749 §5.1 allows when the granted scope is
// identical to the requested scope), CompleteWithState falls back to
// the RequestedScopes carried in the OAuth2State so the persisted
// connection still carries the scope set. This is load-bearing for the
// scope-union logic on subsequent reconnects -- without it we would
// store empty scope and lose the diff.
func TestCompleteWithState_ScopeFallback(t *testing.T) {
	t.Parallel()

	// Fake provider token endpoint: returns a valid token response
	// with NO `scope` field, matching RFC 6749 §5.1 "identical to
	// requested" shape.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"live-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	c := &OAuth2Connector{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
		AuthURL:      "https://provider.example.com/authorize",
		TokenURL:     server.URL,
		// httptest binds to loopback, which the SSRF-protected
		// default client refuses; inject a permissive client.
		HTTPClient: httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
	}

	orgID := gid.New(gid.NewTenantID(), 0)
	stateData := OAuth2State{
		OrganizationID:  orgID.String(),
		Provider:        "TEST",
		RequestedScopes: []string{"read:user", "write:user"},
	}
	stateToken, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL, stateData)
	require.NoError(t, err)

	// Fabricate a callback request with a code + the signed state.
	req := httptest.NewRequest(http.MethodGet, "https://example.com/cb?code=the-code&state="+stateToken, nil)

	conn, returnedState, err := c.CompleteWithState(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NotNil(t, returnedState)

	oauth2Conn, ok := conn.(*OAuth2Connection)
	require.True(t, ok, "expected *OAuth2Connection, got %T", conn)

	assert.Equal(t, "live-token", oauth2Conn.AccessToken)
	// The provider omitted scope, so CompleteWithState must fall back
	// to the RequestedScopes carried in the state token, formatted as
	// a space-separated RFC 6749 §3.3 scope string (sorted).
	assert.Equal(t, "read:user write:user", oauth2Conn.Scope)
	assert.Equal(t, []string{"read:user", "write:user"}, returnedState.RequestedScopes)
}
