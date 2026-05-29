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
	"fmt"
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
			"",
			connector.TokenURL,
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
			"",
			connector.TokenURL,
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
		"",
		connector.TokenURL,
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
		"",
		connector.TokenURL,
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

// TestInitiateWithState_PKCE verifies that connectors with RequiresPKCE=true
// embed the S256 challenge in the authorization URL (RFC 7636 §4.3) and
// persist a random nonce (not the verifier) in the signed state token, so
// CompleteWithState can re-derive the verifier and replay it on the token
// exchange without ever exposing it in the state parameter.
func TestInitiateWithState_PKCE(t *testing.T) {
	t.Parallel()

	t.Run("authorize URL carries S256 code_challenge when PKCE is required", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURI:  "https://example.com/cb",
			AuthURL:      "https://provider.example.com/authorize",
			RequiresPKCE: true,
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

		challenge := parsed.Query().Get("code_challenge")
		require.NotEmpty(t, challenge, "code_challenge must be present when RequiresPKCE=true")
		assert.Equal(t, "S256", parsed.Query().Get("code_challenge_method"))

		// Only a nonce is persisted in the state token; the verifier is
		// derived server-side from the state salt + nonce and must never
		// appear in the (signed-but-unencrypted) state. Re-deriving it
		// must reproduce the published challenge.
		stateToken := parsed.Query().Get("state")
		require.NotEmpty(t, stateToken)

		payload, err := DecodeOAuth2StatePayload(stateToken)
		require.NoError(t, err)
		require.NotEmpty(t, payload.Data.PKCENonce, "nonce must be persisted in state token")

		verifier := derivePKCEVerifier("secret", payload.Data.PKCENonce)
		require.NotContains(t, stateToken, verifier,
			"the derived verifier must never appear in the state token")
		assert.Equal(t, challenge, pkceChallenge(verifier),
			"code_challenge must equal base64url(sha256(derived verifier))")
	})

	t.Run("authorize URL omits PKCE params when PKCE is not required", func(t *testing.T) {
		t.Parallel()

		c := &OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURI:  "https://example.com/cb",
			AuthURL:      "https://provider.example.com/authorize",
			RequiresPKCE: false,
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
		assert.False(t, parsed.Query().Has("code_challenge"))
		assert.False(t, parsed.Query().Has("code_challenge_method"))
	})

	t.Run("token POST replays code_verifier from state on PKCE flow", func(t *testing.T) {
		t.Parallel()

		var capturedVerifier string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			form, err := url.ParseQuery(string(body))
			assert.NoError(t, err)

			capturedVerifier = form.Get("code_verifier")

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
			RequiresPKCE: true,
			HTTPClient:   httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
		}

		// Initiate to mint a state token that embeds a fresh PKCE verifier.
		orgID := gid.New(gid.NewTenantID(), 0)
		authURL, err := c.InitiateWithState(
			context.Background(),
			OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
			InitiateOptions{Scopes: []string{"read:user"}},
		)
		require.NoError(t, err)

		parsed, err := url.Parse(authURL)
		require.NoError(t, err)

		stateToken := parsed.Query().Get("state")
		require.NotEmpty(t, stateToken)

		payload, err := DecodeOAuth2StatePayload(stateToken)
		require.NoError(t, err)

		require.NotEmpty(t, payload.Data.PKCENonce)
		expectedVerifier := derivePKCEVerifier("secret", payload.Data.PKCENonce)

		// Drive Complete with that same state token + an arbitrary code.
		req := httptest.NewRequest(
			http.MethodGet,
			"https://example.com/cb?code=the-code&state="+stateToken,
			nil,
		)

		_, _, err = c.CompleteWithState(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, expectedVerifier, capturedVerifier,
			"token POST body must carry the verifier persisted in the state token")
	})
}

func TestInitiateWithState_PerSiteAuthURL(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		ClientID:            "cid",
		ClientSecret:        "secret",
		RedirectURI:         "https://probo.example/cb",
		RequiresPKCE:        true,
		BuildAuthURLForSite: DatadogAuthorizeURL,
	}

	got, err := c.InitiateWithState(context.Background(),
		OAuth2State{OrganizationID: "org", Provider: DatadogProvider},
		InitiateOptions{Scopes: []string{"user_access_read"}, Site: "US3"},
	)
	require.NoError(t, err)

	u, err := url.Parse(got)
	require.NoError(t, err)
	assert.Equal(t, "us3.datadoghq.com", u.Host)
	assert.Equal(t, "/oauth2/v1/authorize", u.Path)
	assert.NotEmpty(t, u.Query().Get("code_challenge"))
}

func TestInitiateWithState_MissingSiteForMultiSite(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		ClientID:            "cid",
		ClientSecret:        "secret",
		BuildAuthURLForSite: DatadogAuthorizeURL,
	}

	_, err := c.InitiateWithState(context.Background(),
		OAuth2State{OrganizationID: "org", Provider: DatadogProvider},
		InitiateOptions{Site: ""},
	)
	require.Error(t, err)
}

func TestInitiateWithState_InvalidSiteRejected(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		ClientID:            "cid",
		ClientSecret:        "secret",
		BuildAuthURLForSite: DatadogAuthorizeURL,
	}

	_, err := c.InitiateWithState(context.Background(),
		OAuth2State{OrganizationID: "org", Provider: DatadogProvider},
		InitiateOptions{Site: "BOGUS"},
	)
	require.Error(t, err)
}

func TestCompleteWithState_PerDomainTokenURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"at","refresh_token":"rt","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	// Build a token-URL closure that targets the httptest server,
	// mirroring DatadogTokenURL's shape (validate then build).
	c := &OAuth2Connector{
		ClientID:     "cid",
		ClientSecret: "secret",
		RedirectURI:  "https://probo.example/cb",
		RequiresPKCE: true,
		HTTPClient:   httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
		BuildTokenURLForDomain: func(domain string) (string, error) {
			if domain != "us3.datadoghq.com" {
				return "", fmt.Errorf("unknown domain")
			}
			return srv.URL + "/oauth2/v1/token", nil
		},
	}

	state, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL,
		OAuth2State{OrganizationID: validOrgGID(t), Provider: DatadogProvider})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet,
		"https://probo.example/cb?code=abc&state="+state+"&domain=us3.datadoghq.com", nil)

	conn, _, err := c.CompleteWithState(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "/oauth2/v1/token", gotPath)

	oc, ok := conn.(*OAuth2Connection)
	require.True(t, ok)
	assert.Equal(t, srv.URL+"/oauth2/v1/token", oc.TokenURL)
}

func TestCompleteWithState_MissingDomainForMultiSite(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		ClientID:               "cid",
		ClientSecret:           "secret",
		RedirectURI:            "https://probo.example/cb",
		HTTPClient:             httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
		BuildTokenURLForDomain: func(string) (string, error) { return "", fmt.Errorf("unused") },
	}

	state, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL,
		OAuth2State{OrganizationID: validOrgGID(t), Provider: DatadogProvider})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet,
		"https://probo.example/cb?code=abc&state="+state, nil)

	_, _, err = c.CompleteWithState(context.Background(), req)
	require.Error(t, err)
}

// TestCompleteWithState_InvalidDomainRejected exercises the SSRF guard: a
// tampered callback `domain` must fail the flow (the closure validates against
// the fixed allow-list) before credentials are POSTed anywhere.
func TestCompleteWithState_InvalidDomainRejected(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		ClientID:               "cid",
		ClientSecret:           "secret",
		RedirectURI:            "https://probo.example/cb",
		HTTPClient:             httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
		BuildTokenURLForDomain: DatadogTokenURL,
	}

	state, err := statelesstoken.NewToken(c.ClientSecret, OAuth2TokenType, OAuth2TokenTTL,
		OAuth2State{OrganizationID: validOrgGID(t), Provider: DatadogProvider})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet,
		"https://probo.example/cb?code=abc&state="+state+"&domain=evil.example.com", nil)

	_, _, err = c.CompleteWithState(context.Background(), req)
	require.Error(t, err)
}

func TestRefreshableClient_PrefersConnectionTokenURL(t *testing.T) {
	t.Parallel()

	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	conn := &OAuth2Connection{
		AccessToken:  "old",
		RefreshToken: "rt",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(-time.Hour),
		TokenURL:     srv.URL, // per-connection (Datadog-style)
	}

	// cfg.TokenURL is empty (multi-site providers carry no static token URL).
	_, err := conn.RefreshableClient(context.Background(), OAuth2RefreshConfig{
		ClientID: "cid", ClientSecret: "secret",
	}, httpclient.WithSSRFAllowLoopback())
	require.NoError(t, err)

	u, _ := url.Parse(srv.URL)
	assert.Equal(t, u.Host, gotHost)
	assert.Equal(t, "new", conn.AccessToken)
}

func validOrgGID(t *testing.T) string {
	t.Helper()
	return gid.New(gid.NewTenantID(), 0).String()
}

// TestGeneratePKCENonce exercises the nonce generator: each call must
// return a fresh value, encoded as RFC 4648 §5 base64url-without-padding
// (32 bytes yields 43 chars). The nonce seeds derivePKCEVerifier, so a
// predictable or short nonce would weaken PKCE.
func TestGeneratePKCENonce(t *testing.T) {
	t.Parallel()

	v1, err := generatePKCENonce()
	require.NoError(t, err)
	v2, err := generatePKCENonce()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(v1), 43, "nonce must be at least 43 base64url chars")
	assert.LessOrEqual(t, len(v1), 128, "nonce must be at most 128 chars")
	assert.NotEqual(t, v1, v2, "nonce must be unpredictable across calls")

	// Charset: base64url unreserved (RFC 4648 §5) — A-Z a-z 0-9 - _.
	for _, c := range v1 {
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_':
		default:
			t.Errorf("nonce contains non-base64url character %q", c)
		}
	}
}

// TestStateSalt verifies the OAuth2 state / PKCE salt selection: a public
// client's StateSigningKey takes precedence, and a confidential client
// falls back to its ClientSecret.
func TestStateSalt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "secret", (&OAuth2Connector{ClientSecret: "secret"}).stateSalt())
	assert.Equal(t, "server-key", (&OAuth2Connector{StateSigningKey: "server-key"}).stateSalt())
	assert.Equal(t, "server-key",
		(&OAuth2Connector{ClientSecret: "secret", StateSigningKey: "server-key"}).stateSalt(),
		"StateSigningKey must win when both are present")
	assert.Empty(t, (&OAuth2Connector{}).stateSalt(),
		"both empty yields empty salt (InitiateWithState/CompleteWithState reject this)")
}

// TestDeriveConnectorStateKey verifies the connector state-key derivation is
// deterministic, hides the raw secret, is sensitive to the secret, and is
// domain-separated from the PKCE verifier derived from the same secret.
func TestDeriveConnectorStateKey(t *testing.T) {
	t.Parallel()

	k1 := DeriveConnectorStateKey("server-secret")

	assert.NotEmpty(t, k1)
	assert.Equal(t, k1, DeriveConnectorStateKey("server-secret"), "derivation must be deterministic")
	assert.NotEqual(t, "server-secret", k1, "must not echo the raw secret")
	assert.NotEqual(t, k1, DeriveConnectorStateKey("other-secret"), "different secrets must yield different keys")
	assert.NotEqual(t, k1, derivePKCEVerifier("server-secret", "nonce"),
		"state key must be domain-separated from the PKCE verifier")
}

// TestInitiateWithState_RejectsEmptySalt confirms a connector with neither a
// StateSigningKey nor a ClientSecret cannot mint a state token (an empty HMAC
// key would make the token forgeable).
func TestInitiateWithState_RejectsEmptySalt(t *testing.T) {
	t.Parallel()

	c := &OAuth2Connector{
		RedirectURI: "https://example.com/cb",
		AuthURL:     "https://provider.example.com/authorize",
	}

	orgID := gid.New(gid.NewTenantID(), 0)

	_, err := c.InitiateWithState(
		context.Background(),
		OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
		InitiateOptions{Scopes: []string{"read:user"}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no state signing key or client secret")
}

// TestCompleteWithState_PublicClientCIMD exercises the public-client (CIMD)
// flow end to end: there is no client_secret, the state token is signed with
// the server-side StateSigningKey (so validation still succeeds), and the
// token POST carries client_id + the PKCE code_verifier but NEVER a
// client_secret.
func TestCompleteWithState_PublicClientCIMD(t *testing.T) {
	t.Parallel()

	var (
		hadSecretField   bool
		capturedClientID string
		capturedVerifier string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		form, err := url.ParseQuery(string(body))
		assert.NoError(t, err)

		_, hadSecretField = form["client_secret"]
		capturedClientID = form.Get("client_id")
		capturedVerifier = form.Get("code_verifier")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"live-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	c := &OAuth2Connector{
		ClientID:          "https://probo.example.com/api/console/v1/connectors/oauth-client-metadata",
		ClientSecret:      "", // public client: no secret
		StateSigningKey:   "server-side-signing-key",
		RedirectURI:       "https://example.com/cb",
		AuthURL:           "https://provider.example.com/authorize",
		TokenURL:          server.URL,
		TokenEndpointAuth: "none",
		RequiresPKCE:      true,
		HTTPClient:        httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
	}

	orgID := gid.New(gid.NewTenantID(), 0)
	authURL, err := c.InitiateWithState(
		context.Background(),
		OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
		InitiateOptions{Scopes: []string{"organization_member:read"}},
	)
	require.NoError(t, err)

	parsed, err := url.Parse(authURL)
	require.NoError(t, err)
	require.NotEmpty(t, parsed.Query().Get("code_challenge"), "public client must use PKCE")

	stateToken := parsed.Query().Get("state")
	require.NotEmpty(t, stateToken)

	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.com/cb?code=the-code&state="+stateToken,
		nil,
	)

	_, _, err = c.CompleteWithState(context.Background(), req)
	require.NoError(t, err, "state signed with StateSigningKey must validate")

	assert.False(t, hadSecretField, "public-client token POST must NOT include client_secret")
	assert.Equal(t, c.ClientID, capturedClientID)
	assert.NotEmpty(t, capturedVerifier, "public-client token POST must carry the PKCE code_verifier")
}

// TestRefreshableClient_PublicClientOmitsSecret confirms that refreshing a
// public-client (CIMD) token sends client_id but NO client_secret — the
// provider advertises token_endpoint_auth_method "none" and would reject an
// (empty) secret. This guards the token-refresh path used when an access
// token expires mid-campaign.
func TestRefreshableClient_PublicClientOmitsSecret(t *testing.T) {
	t.Parallel()

	var (
		hadSecret        bool
		capturedClientID string
		capturedGrant    string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		_, hadSecret = r.Form["client_secret"]
		capturedClientID = r.Form.Get("client_id")
		capturedGrant = r.Form.Get("grant_type")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	conn := &OAuth2Connection{
		AccessToken:  "stale",
		RefreshToken: "refresh-tok",
		ExpiresAt:    time.Now().Add(-time.Hour), // expired → force a refresh
		TokenType:    "Bearer",
	}

	cfg := OAuth2RefreshConfig{
		ClientID:          "https://probo.example.com/api/console/v1/connectors/oauth-client-metadata",
		ClientSecret:      "", // public client
		TokenURL:          server.URL,
		TokenEndpointAuth: "none",
	}

	_, err := conn.RefreshableClient(context.Background(), cfg, httpclient.WithSSRFAllowLoopback())
	require.NoError(t, err)

	assert.Equal(t, "refreshed", conn.AccessToken, "refresh must update the access token")
	assert.Equal(t, "refresh_token", capturedGrant)
	assert.Equal(t, cfg.ClientID, capturedClientID)
	assert.False(t, hadSecret, "public-client refresh must NOT send client_secret")
}

// TestCompleteWithState_PKCEMismatch confirms that a token endpoint
// rejecting a stale or mismatched code_verifier (the standard PKCE
// failure path) surfaces as an error from CompleteWithState rather
// than being silently swallowed.
func TestCompleteWithState_PKCEMismatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The provider is supposed to validate the verifier; emulate a
		// reject so we can observe the failure path.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"invalid_grant"}`))
	}))
	defer server.Close()

	c := &OAuth2Connector{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/cb",
		AuthURL:      "https://provider.example.com/authorize",
		TokenURL:     server.URL,
		RequiresPKCE: true,
		HTTPClient:   httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
	}

	orgID := gid.New(gid.NewTenantID(), 0)
	authURL, err := c.InitiateWithState(
		context.Background(),
		OAuth2State{OrganizationID: orgID.String(), Provider: "TEST"},
		InitiateOptions{Scopes: []string{"read"}},
	)
	require.NoError(t, err)
	parsed, err := url.Parse(authURL)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.com/cb?code=the-code&state="+parsed.Query().Get("state"),
		nil,
	)
	_, _, err = c.CompleteWithState(context.Background(), req)
	require.Error(t, err, "PKCE rejection from token endpoint must propagate")
}
