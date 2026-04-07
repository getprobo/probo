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

package console_test

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// ---------------------------------------------------------------------------
// 1. Discovery and JWKS
// ---------------------------------------------------------------------------

func TestOAuth2_Discovery(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	discovery, raw, err := testutil.OAuth2Discovery(owner)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, raw.StatusCode)
	require.NotNil(t, discovery)

	assert.NotEmpty(t, discovery.Issuer)
	assert.Contains(t, discovery.AuthorizationEndpoint, "/oauth2/authorize")
	assert.Contains(t, discovery.TokenEndpoint, "/oauth2/token")
	assert.Contains(t, discovery.UserinfoEndpoint, "/oauth2/userinfo")
	assert.Contains(t, discovery.JwksURI, "/oauth2/jwks")
	assert.Contains(t, discovery.RegistrationEndpoint, "/oauth2/register")
	assert.Contains(t, discovery.IntrospectionEndpoint, "/oauth2/introspect")
	assert.Contains(t, discovery.RevocationEndpoint, "/oauth2/revoke")
	assert.Contains(t, discovery.DeviceAuthorizationEndpoint, "/oauth2/device")

	assert.Contains(t, discovery.GrantTypesSupported, "authorization_code")
	assert.Contains(t, discovery.GrantTypesSupported, "refresh_token")
	assert.Contains(t, discovery.GrantTypesSupported, "urn:ietf:params:oauth:grant-type:device_code")

	assert.Contains(t, discovery.ScopesSupported, "openid")
	assert.Contains(t, discovery.ScopesSupported, "profile")
	assert.Contains(t, discovery.ScopesSupported, "email")

	assert.Contains(t, discovery.ResponseTypesSupported, "code")
	assert.Contains(t, discovery.CodeChallengeMethodsSupported, "S256")

	assert.Contains(t, discovery.TokenEndpointAuthMethodsSupported, "client_secret_basic")
	assert.Contains(t, discovery.TokenEndpointAuthMethodsSupported, "client_secret_post")
	assert.Contains(t, discovery.TokenEndpointAuthMethodsSupported, "none")

	assert.Contains(t, discovery.SubjectTypesSupported, "public")
	assert.Contains(t, discovery.IDTokenSigningAlgValuesSupported, "RS256")

	assert.Contains(t, discovery.ClaimsSupported, "iss")
	assert.Contains(t, discovery.ClaimsSupported, "sub")
	assert.Contains(t, discovery.ClaimsSupported, "aud")
	assert.Contains(t, discovery.ClaimsSupported, "exp")
	assert.Contains(t, discovery.ClaimsSupported, "iat")
	assert.Contains(t, discovery.ClaimsSupported, "auth_time")
	assert.Contains(t, discovery.ClaimsSupported, "nonce")
	assert.Contains(t, discovery.ClaimsSupported, "at_hash")
	assert.Contains(t, discovery.ClaimsSupported, "email")
	assert.Contains(t, discovery.ClaimsSupported, "email_verified")
	assert.Contains(t, discovery.ClaimsSupported, "name")
}

func TestOAuth2_JWKS(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	jwks, raw, err := testutil.OAuth2JWKS(owner)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, raw.StatusCode)
	require.NotNil(t, jwks)

	require.NotEmpty(t, jwks.Keys, "JWKS must contain at least one key")

	key := jwks.Keys[0]
	assert.Equal(t, "RSA", key["kty"])
	assert.NotEmpty(t, key["kid"])
	assert.NotEmpty(t, key["n"])
	assert.NotEmpty(t, key["e"])
	assert.Equal(t, "sig", key["use"])
	assert.Equal(t, "RS256", key["alg"])
}

// ---------------------------------------------------------------------------
// 2. Dynamic Client Registration
// ---------------------------------------------------------------------------

func TestOAuth2_RegisterClient(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"happy path with confidential client",
		func(t *testing.T) {
			t.Parallel()

			result := factory.CreateOAuth2Client(owner, nil)
			assert.NotEmpty(t, result.ClientID)
			assert.NotEmpty(t, result.ClientSecret)
		},
	)

	t.Run(
		"public client has no secret",
		func(t *testing.T) {
			t.Parallel()

			result := factory.CreatePublicOAuth2Client(owner, nil)
			assert.NotEmpty(t, result.ClientID)
			assert.Empty(t, result.ClientSecret)
		},
	)

	t.Run(
		"invalid redirect URI for private client",
		func(t *testing.T) {
			t.Parallel()

			_, raw, err := testutil.OAuth2RegisterClient(owner, map[string]any{
				"organization_id": owner.GetOrganizationID().String(),
				"client_name":     factory.SafeName("Bad Redirect"),
				"visibility":      "private",
				"redirect_uris":   []string{"http://evil.example.com/callback"},
				"grant_types":     []string{"authorization_code"},
				"response_types":  []string{"code"},
				"scopes":          "openid",
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, raw.StatusCode)
		},
	)

	t.Run(
		"non-member cannot register",
		func(t *testing.T) {
			t.Parallel()

			otherOwner := testutil.NewClient(t, testutil.RoleOwner)

			_, raw, err := testutil.OAuth2RegisterClient(otherOwner, map[string]any{
				"organization_id": owner.GetOrganizationID().String(),
				"client_name":     factory.SafeName("Foreign Client"),
				"visibility":      "private",
				"redirect_uris":   []string{"http://localhost:9999/callback"},
				"grant_types":     []string{"authorization_code"},
				"response_types":  []string{"code"},
				"scopes":          "openid",
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 3. Authorization Code Flow (with PKCE)
// ---------------------------------------------------------------------------

func TestOAuth2_AuthorizationCodeFlow(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"full happy path with PKCE",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokenResp := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.NotEmpty(t, tokenResp.RefreshToken)
			assert.NotEmpty(t, tokenResp.IDToken)
			assert.Equal(t, "Bearer", tokenResp.TokenType)
			assert.Greater(t, tokenResp.ExpiresIn, int64(0))
			assert.Contains(t, tokenResp.Scope, "openid")
		},
	)

	t.Run(
		"consent deny returns access_denied",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()
			_ = verifier

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"deny-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, authResp.StatusCode, "expected consent page")

			consentID, err := testutil.ExtractConsentID(authResp.Body)
			require.NoError(t, err)

			denyResp, err := testutil.OAuth2ConsentDeny(owner, consentID)
			require.NoError(t, err)
			require.Equal(t, http.StatusFound, denyResp.StatusCode)

			loc := denyResp.Header.Get("Location")
			assert.Contains(t, loc, "error=access_denied")
		},
	)

	t.Run(
		"invalid scope",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid super_admin"},
				"state":                {"scope-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode)
			}
		},
	)

	t.Run(
		"bad code verifier fails token exchange",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"pkce-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			_, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				"wrong-verifier-that-does-not-match-the-challenge-at-all",
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode)
		},
	)

	t.Run(
		"code reuse fails",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"reuse-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			tokenResp, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)
			require.NotNil(t, tokenResp)

			_, raw2, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw2.StatusCode, "second exchange should fail")
		},
	)
}

// ---------------------------------------------------------------------------
// 4. Refresh Token Flow
// ---------------------------------------------------------------------------

func TestOAuth2_RefreshToken(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"token rotation",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			firstTokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			refreshResp, raw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				firstTokens.RefreshToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode, "refresh failed: %s", string(raw.Body))
			require.NotNil(t, refreshResp)

			assert.NotEqual(t, firstTokens.AccessToken, refreshResp.AccessToken)
			assert.NotEqual(t, firstTokens.RefreshToken, refreshResp.RefreshToken)
			assert.NotEmpty(t, refreshResp.IDToken, "should include id_token for openid scope")
			assert.Equal(t, "Bearer", refreshResp.TokenType)
			assert.Greater(t, refreshResp.ExpiresIn, int64(0))
		},
	)

	t.Run(
		"replay detection revokes all tokens",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			firstTokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			secondTokens, raw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				firstTokens.RefreshToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)
			require.NotNil(t, secondTokens)

			_, replayRaw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				firstTokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, replayRaw.StatusCode, "replayed refresh token should fail")

			_, newRaw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				secondTokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.NotEqual(
				t,
				http.StatusOK,
				newRaw.StatusCode,
				"new refresh token should also be revoked after replay detection",
			)
		},
	)

	t.Run(
		"cross-client refresh token theft rejected",
		func(t *testing.T) {
			t.Parallel()

			clientA := factory.CreateOAuth2Client(owner, nil)
			clientB := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				clientA.ClientID,
				clientA.ClientSecret,
				redirectURI,
			)

			_, raw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				clientB.ClientID,
				clientB.ClientSecret,
				tokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode,
				"client B must not be able to use client A's refresh token")
		},
	)

	t.Run(
		"invalid refresh token",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			_, raw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				"totally-invalid-refresh-token",
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 5. Device Code Flow
// ---------------------------------------------------------------------------

func TestOAuth2_DeviceCodeFlow(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"full happy path",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)

			deviceResp, raw, err := testutil.OAuth2DeviceAuth(
				owner,
				client.ClientID,
				"openid email profile",
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode, "device auth failed: %s", string(raw.Body))
			require.NotNil(t, deviceResp)

			assert.NotEmpty(t, deviceResp.DeviceCode)
			assert.NotEmpty(t, deviceResp.UserCode)
			assert.NotEmpty(t, deviceResp.VerificationURI)
			assert.Greater(t, deviceResp.ExpiresIn, 0)
			assert.Greater(t, deviceResp.Interval, 0)

			_, errResp, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp)
			assert.Equal(t, "authorization_pending", errResp.Code)

			userCode := deviceResp.UserCode
			verifyResp, err := testutil.OAuth2DeviceVerify(owner, userCode)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, verifyResp.StatusCode, "device verify failed: %s", string(verifyResp.Body))

			time.Sleep(time.Duration(deviceResp.Interval+1) * time.Second)

			tokenResp, _, pollRaw, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, pollRaw.StatusCode, "device token poll failed: %s", string(pollRaw.Body))
			require.NotNil(t, tokenResp)

			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.Equal(t, "Bearer", tokenResp.TokenType)
			assert.Greater(t, tokenResp.ExpiresIn, int64(0))
		},
	)

	t.Run(
		"invalid client",
		func(t *testing.T) {
			t.Parallel()

			_, raw, err := testutil.OAuth2DeviceAuth(
				owner,
				"nonexistent-client-id",
				"openid",
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode)
		},
	)

	t.Run(
		"slow down polling",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)

			deviceResp, _, err := testutil.OAuth2DeviceAuth(
				owner,
				client.ClientID,
				"openid",
			)
			require.NoError(t, err)
			require.NotNil(t, deviceResp)

			_, errResp1, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp1)
			assert.Equal(t, "authorization_pending", errResp1.Code)

			_, errResp2, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp2)
			assert.Equal(t, "slow_down", errResp2.Code)
		},
	)
}

// ---------------------------------------------------------------------------
// 6. Token Introspection
// ---------------------------------------------------------------------------

func TestOAuth2_Introspect(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"active token",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			introspect, raw, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)

			assert.True(t, introspect.Active)
			assert.Equal(t, "Bearer", introspect.TokenType)
			assert.NotEmpty(t, introspect.Sub)
			assert.Greater(t, introspect.Exp, int64(0))
			assert.NotEmpty(t, introspect.Scope, "introspection should return scope")
			assert.Equal(t, client.ClientID, introspect.ClientID, "introspection should return client_id")
		},
	)

	t.Run(
		"revoked token is inactive",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			revokeRaw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, revokeRaw.StatusCode)

			introspect, _, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.False(t, introspect.Active)
		},
	)

	t.Run(
		"wrong client gets inactive",
		func(t *testing.T) {
			t.Parallel()

			clientA := factory.CreateOAuth2Client(owner, nil)
			clientB := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				clientA.ClientID,
				clientA.ClientSecret,
				redirectURI,
			)

			introspect, _, err := testutil.OAuth2Introspect(
				owner,
				clientB.ClientID,
				clientB.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.False(t, introspect.Active)
		},
	)

	t.Run(
		"bad client auth",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			_, raw, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				"wrong-secret",
				"some-token",
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 7. Token Revocation
// ---------------------------------------------------------------------------

func TestOAuth2_Revoke(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"revoke access token then introspect is inactive",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			raw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, raw.StatusCode)

			introspect, _, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.False(t, introspect.Active)
		},
	)

	t.Run(
		"revoke refresh token then refresh fails",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			raw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, raw.StatusCode)

			_, refreshRaw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, refreshRaw.StatusCode)
		},
	)

	t.Run(
		"unknown token returns 200 per RFC 7009",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			raw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				client.ClientSecret,
				"unknown-token-that-does-not-exist",
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, raw.StatusCode)
		},
	)

	t.Run(
		"bad client auth",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			raw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				"wrong-secret",
				"some-token",
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 8. UserInfo
// ---------------------------------------------------------------------------

func TestOAuth2_UserInfo(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"returns claims for openid email profile scopes",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			userInfo, raw, err := testutil.OAuth2UserInfo(owner, tokens.AccessToken)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode, "userinfo failed: %s", string(raw.Body))
			require.NotNil(t, userInfo)

			assert.NotEmpty(t, userInfo.Sub)
			assert.NotEmpty(t, userInfo.Email)
			assert.NotEmpty(t, userInfo.Name)
		},
	)

	t.Run(
		"revoked access token returns 401",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			revokeRaw, err := testutil.OAuth2Revoke(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, revokeRaw.StatusCode)

			_, raw, err := testutil.OAuth2UserInfo(owner, tokens.AccessToken)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode,
				"revoked access token must be rejected by userinfo")
		},
	)

	t.Run(
		"no bearer token returns 401",
		func(t *testing.T) {
			t.Parallel()

			_, raw, err := testutil.OAuth2UserInfo(owner, "")
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)

	t.Run(
		"invalid bearer token returns 401",
		func(t *testing.T) {
			t.Parallel()

			_, raw, err := testutil.OAuth2UserInfo(owner, "invalid-access-token")
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 9. Token Endpoint Errors
// ---------------------------------------------------------------------------

func TestOAuth2_Token_Errors(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"unsupported grant type",
		func(t *testing.T) {
			t.Parallel()

			raw, err := testutil.OAuth2TokenRaw(owner, url.Values{
				"grant_type": {"password"},
				"username":   {"user"},
				"password":   {"pass"},
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, raw.StatusCode)

			var errResp testutil.OAuth2ErrorResponse
			require.NoError(t, json.Unmarshal(raw.Body, &errResp))
			assert.Equal(t, "unsupported_grant_type", errResp.Code)
		},
	)

	t.Run(
		"missing grant type",
		func(t *testing.T) {
			t.Parallel()

			raw, err := testutil.OAuth2TokenRaw(owner, url.Values{})
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, raw.StatusCode)
		},
	)

	t.Run(
		"bad client auth on authorization code grant",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			raw, err := testutil.OAuth2TokenRawWithBasicAuth(
				owner,
				url.Values{
					"grant_type": {"authorization_code"},
					"code":       {"fake-code"},
				},
				client.ClientID,
				"wrong-secret",
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)

	t.Run(
		"slow down device code polling",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)

			deviceResp, _, err := testutil.OAuth2DeviceAuth(
				owner,
				client.ClientID,
				"openid",
			)
			require.NoError(t, err)
			require.NotNil(t, deviceResp)

			_, errResp1, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp1)

			_, errResp2, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp2)
			assert.Equal(t, "slow_down", errResp2.Code)

			time.Sleep(time.Duration(deviceResp.Interval+1) * time.Second)

			_, errResp3, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp3)
			assert.Equal(t, "authorization_pending", errResp3.Code, "after waiting, should not get slow_down")
		},
	)
}

// ---------------------------------------------------------------------------
// 10. Security
// ---------------------------------------------------------------------------

func TestOAuth2_Security(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"public client requires PKCE",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			params := url.Values{
				"client_id":      {client.ClientID},
				"redirect_uri":   {redirectURI},
				"response_type":  {"code"},
				"scope":          {"openid"},
				"state":          {"no-pkce"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=", "public client without code_challenge should be rejected")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode)
			}
		},
	)

	t.Run(
		"only S256 code challenge method accepted",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"plain-method"},
				"code_challenge":       {"some-challenge-value"},
				"code_challenge_method": {"plain"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=", "plain code_challenge_method should be rejected")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode)
			}
		},
	)

	t.Run(
		"redirect URI mismatch at token exchange",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"redirect-mismatch"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			_, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				"http://localhost:9999/WRONG-callback",
				verifier,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode, "mismatched redirect_uri must fail")
		},
	)

	t.Run(
		"state parameter roundtrip",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			state := "csrf-protection-nonce-abc123"

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {state},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var redirectLoc string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)
				require.Equal(t, http.StatusFound, consentResp.StatusCode)

				redirectLoc = consentResp.Header.Get("Location")
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				redirectLoc = authResp.Header.Get("Location")
			}

			u, err := url.Parse(redirectLoc)
			require.NoError(t, err)
			assert.Equal(t, state, u.Query().Get("state"), "state must be returned unchanged")
			assert.NotEmpty(t, u.Query().Get("code"), "code must be present")
		},
	)

	t.Run(
		"scope escalation rejected",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, factory.Attrs{
				"scopes": "openid",
			})
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid email profile"},
				"state":                {"scope-esc"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=", "requesting scopes beyond registration must fail")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode,
					"should not show consent page for disallowed scopes")
			}
		},
	)

	t.Run(
		"cross-client code exchange rejected",
		func(t *testing.T) {
			t.Parallel()

			clientA := factory.CreateOAuth2Client(owner, nil)
			clientB := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {clientA.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"cross-client"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			_, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				clientB.ClientID,
				clientB.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode,
				"code issued to client A must not be exchangeable by client B")
		},
	)

	t.Run(
		"unregistered redirect URI rejected at authorize",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {"http://localhost:9999/evil"},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"open-redirect"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, authResp.StatusCode,
				"unregistered redirect_uri must not show consent page")
			assert.NotEqual(t, http.StatusFound, authResp.StatusCode,
				"invalid redirect_uri must not redirect (must return JSON error)")

			var errResp testutil.OAuth2ErrorResponse
			if json.Unmarshal(authResp.Body, &errResp) == nil {
				assert.Equal(t, "invalid_redirect_uri", errResp.Code,
					"should return invalid_redirect_uri error code")
			}
		},
	)

	t.Run(
		"private client rejects non-member at authorize",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			_, challenge := testutil.GeneratePKCE()

			otherOwner := testutil.NewClient(t, testutil.RoleOwner)

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {"http://localhost:9999/callback"},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"non-member"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(otherOwner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=",
					"non-member must be rejected when authorizing against a private client")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode,
					"non-member must not see the consent page for a private client")
			}
		},
	)

	t.Run(
		"consent skipped on re-authorization with same scopes",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge1 := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid email profile"},
				"state":                {"consent-first"},
				"code_challenge":       {challenge1},
				"code_challenge_method": {"S256"},
			}

			firstResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, firstResp.StatusCode, "first request should show consent page")

			consentID, err := testutil.ExtractConsentID(firstResp.Body)
			require.NoError(t, err)

			_, err = testutil.OAuth2ConsentApprove(owner, consentID)
			require.NoError(t, err)

			_, challenge2 := testutil.GeneratePKCE()
			params.Set("state", "consent-second")
			params.Set("code_challenge", challenge2)

			secondResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			assert.Equal(t, http.StatusFound, secondResp.StatusCode,
				"second authorize with same scopes should skip consent and redirect with code")

			code, err := testutil.OAuth2AuthorizeCodeFromRedirect(secondResp)
			require.NoError(t, err)
			assert.NotEmpty(t, code)
		},
	)

	t.Run(
		"ID token contains nonce from authorize request",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			nonce := "test-nonce-value-abc123"

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"nonce-test"},
				"nonce":                {nonce},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			tokenResp, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)
			require.NotEmpty(t, tokenResp.IDToken)

			parts := strings.SplitN(tokenResp.IDToken, ".", 3)
			require.Len(t, parts, 3)

			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var claims struct {
				Nonce string `json:"nonce"`
			}
			require.NoError(t, json.Unmarshal(claimsJSON, &claims))
			assert.Equal(t, nonce, claims.Nonce,
				"ID token must contain the nonce from the authorize request")
		},
	)

	t.Run(
		"ID token contains valid at_hash claim",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			require.NotEmpty(t, tokens.IDToken)

			parts := strings.SplitN(tokens.IDToken, ".", 3)
			require.Len(t, parts, 3)

			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var claims struct {
				AtHash string `json:"at_hash"`
			}
			require.NoError(t, json.Unmarshal(claimsJSON, &claims))
			require.NotEmpty(t, claims.AtHash, "at_hash must be present in ID token")

			h := sha256.Sum256([]byte(tokens.AccessToken))
			expectedAtHash := base64.RawURLEncoding.EncodeToString(h[:16])
			assert.Equal(t, expectedAtHash, claims.AtHash,
				"at_hash must be the left half of SHA-256 of the access token, base64url-encoded")
		},
	)

	t.Run(
		"ID token signature verifiable with JWKS",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			require.NotEmpty(t, tokens.IDToken, "expected id_token")

			// Parse JWT parts.
			parts := strings.SplitN(tokens.IDToken, ".", 3)
			require.Len(t, parts, 3, "JWT must have 3 parts")

			// Decode and verify header.
			headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)

			var header struct {
				Alg string `json:"alg"`
				Typ string `json:"typ"`
				Kid string `json:"kid"`
			}
			require.NoError(t, json.Unmarshal(headerJSON, &header))
			assert.Equal(t, "RS256", header.Alg)
			assert.Equal(t, "JWT", header.Typ)
			assert.NotEmpty(t, header.Kid)

			// Decode claims and verify standard fields.
			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var claims struct {
				Iss      string `json:"iss"`
				Sub      string `json:"sub"`
				Aud      string `json:"aud"`
				Exp      int64  `json:"exp"`
				Iat      int64  `json:"iat"`
				AuthTime int64  `json:"auth_time"`
			}
			require.NoError(t, json.Unmarshal(claimsJSON, &claims))
			assert.NotEmpty(t, claims.Iss)
			assert.NotEmpty(t, claims.Sub)
			assert.Equal(t, client.ClientID, claims.Aud)
			assert.Greater(t, claims.Exp, time.Now().Unix(), "token must not be expired")
			assert.LessOrEqual(t, claims.Iat, time.Now().Unix())
			assert.Greater(t, claims.AuthTime, int64(0), "auth_time must be set")

			// Fetch JWKS and find the matching key.
			jwks, _, err := testutil.OAuth2JWKS(owner)
			require.NoError(t, err)
			require.NotNil(t, jwks)

			var matchingKey map[string]any
			for _, k := range jwks.Keys {
				if kid, ok := k["kid"].(string); ok && kid == header.Kid {
					matchingKey = k
					break
				}
			}
			require.NotNil(t, matchingKey, "JWKS must contain key matching kid=%s", header.Kid)

			// Reconstruct the RSA public key from JWK.
			nB64, ok := matchingKey["n"].(string)
			require.True(t, ok)
			eB64, ok := matchingKey["e"].(string)
			require.True(t, ok)

			nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
			require.NoError(t, err)
			eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
			require.NoError(t, err)

			pubKey := &rsa.PublicKey{
				N: new(big.Int).SetBytes(nBytes),
				E: int(new(big.Int).SetBytes(eBytes).Int64()),
			}

			// Verify RS256 signature.
			signingInput := parts[0] + "." + parts[1]
			sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
			require.NoError(t, err)

			digest := sha256.Sum256([]byte(signingInput))
			err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest[:], sigBytes)
			assert.NoError(t, err, "ID token signature must be verifiable with JWKS public key")
		},
	)

	t.Run(
		"consent approval with different user rejected",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"csrf-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, authResp.StatusCode, "expected consent page")

			consentID, err := testutil.ExtractConsentID(authResp.Body)
			require.NoError(t, err)

			otherOwner := testutil.NewClient(t, testutil.RoleOwner)

			approveResp, err := testutil.OAuth2ConsentApprove(otherOwner, consentID)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusFound, approveResp.StatusCode,
				"consent approval by a different user must be rejected")
		},
	)

	t.Run(
		"cross-client revocation does not revoke the token",
		func(t *testing.T) {
			t.Parallel()

			clientA := factory.CreateOAuth2Client(owner, nil)
			clientB := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				clientA.ClientID,
				clientA.ClientSecret,
				redirectURI,
			)

			revokeRaw, err := testutil.OAuth2Revoke(
				owner,
				clientB.ClientID,
				clientB.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, revokeRaw.StatusCode,
				"cross-client revoke should return 200 per RFC 7009")

			introspect, _, err := testutil.OAuth2Introspect(
				owner,
				clientA.ClientID,
				clientA.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			assert.True(t, introspect.Active,
				"token must still be active when revoked by a different client")
		},
	)

	t.Run(
		"bearer token in query string rejected by userinfo",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			raw, err := testutil.OAuth2UserInfoRaw(owner, url.Values{
				"access_token": {tokens.AccessToken},
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode,
				"bearer token in query string must not authenticate")
		},
	)

	t.Run(
		"confidential client can complete flow without PKCE",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			params := url.Values{
				"client_id":      {client.ClientID},
				"redirect_uri":   {redirectURI},
				"response_type":  {"code"},
				"scope":          {"openid"},
				"state":          {"no-pkce-confidential"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			tokenResp, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				"",
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode,
				"confidential client should succeed without PKCE: %s", string(raw.Body))
			assert.NotEmpty(t, tokenResp.AccessToken)
		},
	)
}

// ---------------------------------------------------------------------------
// 11. Client Secret Post Authentication
// ---------------------------------------------------------------------------

func TestOAuth2_ClientSecretPost(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"full auth code flow with client_secret_post",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, factory.Attrs{
				"token_endpoint_auth_method": "client_secret_post",
			})
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"post-auth"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			tokenResp, raw, err := testutil.OAuth2TokenWithCodePostAuth(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode,
				"client_secret_post exchange failed: %s", string(raw.Body))
			require.NotNil(t, tokenResp)

			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.Equal(t, "Bearer", tokenResp.TokenType)
			assert.Greater(t, tokenResp.ExpiresIn, int64(0))
		},
	)

	t.Run(
		"wrong secret via client_secret_post returns 401",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, factory.Attrs{
				"token_endpoint_auth_method": "client_secret_post",
			})

			raw, err := testutil.OAuth2TokenRaw(owner, url.Values{
				"grant_type":    {"authorization_code"},
				"code":          {"fake-code"},
				"client_id":     {client.ClientID},
				"client_secret": {"wrong-secret"},
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode)
		},
	)
}

// ---------------------------------------------------------------------------
// 12. Public Client Authorization Code Flow
// ---------------------------------------------------------------------------

func TestOAuth2_PublicClientAuthCodeFlow(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"full happy path with PKCE and no secret",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid email profile"},
				"state":                {"public-pkce"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			tokenResp, raw, err := testutil.OAuth2TokenWithCodePostAuth(
				owner,
				client.ClientID,
				"",
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode,
				"public client token exchange failed: %s", string(raw.Body))
			require.NotNil(t, tokenResp)

			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.Equal(t, "Bearer", tokenResp.TokenType)
			assert.Greater(t, tokenResp.ExpiresIn, int64(0))
			assert.Contains(t, tokenResp.Scope, "openid")
		},
	)
}

// ---------------------------------------------------------------------------
// 13. Registration Edge Cases
// ---------------------------------------------------------------------------

func TestOAuth2_RegisterClient_EdgeCases(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"public client with http redirect URI rejected",
		func(t *testing.T) {
			t.Parallel()

			_, raw, err := testutil.OAuth2RegisterClient(owner, map[string]any{
				"organization_id":            owner.GetOrganizationID().String(),
				"client_name":               factory.SafeName("Public HTTP"),
				"visibility":                "public",
				"redirect_uris":             []string{"http://example.com/callback"},
				"grant_types":               []string{"authorization_code"},
				"response_types":            []string{"code"},
				"token_endpoint_auth_method": "none",
				"scopes":                    "openid",
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, raw.StatusCode,
				"public clients must require https redirect URIs")
		},
	)

	t.Run(
		"public client with https redirect URI accepted",
		func(t *testing.T) {
			t.Parallel()

			resp, raw, err := testutil.OAuth2RegisterClient(owner, map[string]any{
				"organization_id":            owner.GetOrganizationID().String(),
				"client_name":               factory.SafeName("Public HTTPS"),
				"visibility":                "public",
				"redirect_uris":             []string{"https://example.com/callback"},
				"grant_types":               []string{"authorization_code"},
				"response_types":            []string{"code"},
				"token_endpoint_auth_method": "none",
				"scopes":                    "openid",
			})
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, raw.StatusCode, "body: %s", string(raw.Body))
			assert.NotEmpty(t, resp.ClientID)
		},
	)
}

// ---------------------------------------------------------------------------
// 14. ID Token Claims
// ---------------------------------------------------------------------------

func TestOAuth2_IDTokenClaims(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"ID token from auth code flow omits email and name",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			require.NotEmpty(t, tokens.IDToken)

			parts := strings.SplitN(tokens.IDToken, ".", 3)
			require.Len(t, parts, 3)

			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var claims map[string]any
			require.NoError(t, json.Unmarshal(claimsJSON, &claims))

			assert.NotEmpty(t, claims["iss"])
			assert.NotEmpty(t, claims["sub"])
			assert.NotEmpty(t, claims["aud"])
			assert.NotEmpty(t, claims["exp"])
			assert.NotEmpty(t, claims["iat"])
			assert.NotEmpty(t, claims["auth_time"])

			// The current implementation passes empty strings to NewIDTokenClaims
			// in ExchangeAuthorizationCode, so email/name are omitted from the
			// ID token even when the scopes include email and profile. UserInfo
			// is the canonical source for these claims.
			_, hasEmail := claims["email"]
			_, hasName := claims["name"]
			assert.False(t, hasEmail, "ID token from auth code should not contain email (use UserInfo)")
			assert.False(t, hasName, "ID token from auth code should not contain name (use UserInfo)")
		},
	)

	t.Run(
		"ID token from refresh omits nonce",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()
			nonce := "test-refresh-nonce"

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"refresh-nonce"},
				"nonce":                {nonce},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			firstTokens, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)

			refreshResp, refreshRaw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				firstTokens.RefreshToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, refreshRaw.StatusCode)
			require.NotEmpty(t, refreshResp.IDToken)

			parts := strings.SplitN(refreshResp.IDToken, ".", 3)
			require.Len(t, parts, 3)

			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var claims map[string]any
			require.NoError(t, json.Unmarshal(claimsJSON, &claims))

			nonceVal, hasNonce := claims["nonce"]
			if hasNonce {
				assert.Empty(t, nonceVal,
					"ID token from refresh must not contain the original nonce")
			}

			assert.NotEmpty(t, claims["auth_time"],
				"refresh ID token must contain auth_time")
		},
	)
}

// ---------------------------------------------------------------------------
// 15. Cache-Control Headers
// ---------------------------------------------------------------------------

func TestOAuth2_CacheHeaders(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"token endpoint response has no-store cache header",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"cache-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			_, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)

			cacheControl := raw.Header.Get("Cache-Control")
			assert.Contains(t, cacheControl, "no-store",
				"token response must include Cache-Control: no-store per RFC 6749 section 5.1")
		},
	)
}

// ---------------------------------------------------------------------------
// 16. Device Flow Edge Cases
// ---------------------------------------------------------------------------

func TestOAuth2_DeviceCodeFlow_EdgeCases(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"invalid user code on device verify",
		func(t *testing.T) {
			t.Parallel()

			verifyResp, err := testutil.OAuth2DeviceVerify(owner, "ZZZZ-ZZZZ")
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, verifyResp.StatusCode,
				"device verify should return an HTML page even for invalid codes")
			assert.Contains(t, strings.ToLower(string(verifyResp.Body)), "failed",
				"response should indicate verification failure")
		},
	)

	t.Run(
		"scope exceeding client registration rejected",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, factory.Attrs{
				"scopes": "openid",
			})

			_, raw, err := testutil.OAuth2DeviceAuth(
				owner,
				client.ClientID,
				"openid email profile",
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode,
				"device auth with scope exceeding registration must be rejected")
		},
	)
}

// ---------------------------------------------------------------------------
// 17. Authorize Endpoint Edge Cases
// ---------------------------------------------------------------------------

func TestOAuth2_Authorize_EdgeCases(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"unsupported response_type rejected",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"token"},
				"scope":                {"openid"},
				"state":                {"implicit-attempt"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			if authResp.StatusCode == http.StatusFound {
				loc := authResp.Header.Get("Location")
				assert.Contains(t, loc, "error=",
					"response_type=token must be rejected")
			} else {
				assert.NotEqual(t, http.StatusOK, authResp.StatusCode,
					"response_type=token must not show consent page")
			}
		},
	)

	t.Run(
		"already approved consent cannot be resubmitted",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			_, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"double-approve"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, authResp.StatusCode, "expected consent page")

			consentID, err := testutil.ExtractConsentID(authResp.Body)
			require.NoError(t, err)

			firstApproval, err := testutil.OAuth2ConsentApprove(owner, consentID)
			require.NoError(t, err)
			require.Equal(t, http.StatusFound, firstApproval.StatusCode,
				"first consent approval should redirect with code")

			secondApproval, err := testutil.OAuth2ConsentApprove(owner, consentID)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusFound, secondApproval.StatusCode,
				"second consent approval must be rejected")
		},
	)
}

// ---------------------------------------------------------------------------
// 18. Introspect Edge Cases
// ---------------------------------------------------------------------------

func TestOAuth2_Introspect_EdgeCases(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"unknown token returns inactive with valid client auth",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			introspect, raw, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				"completely-unknown-token-that-was-never-issued",
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode,
				"introspect of unknown token must return 200 per RFC 7662")
			assert.False(t, introspect.Active,
				"unknown token must be reported as inactive")
		},
	)

	t.Run(
		"empty token returns error",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)

			_, raw, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				"",
			)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, raw.StatusCode,
				"empty token should be rejected by input validation")
		},
	)
}

// ---------------------------------------------------------------------------
// 19. Token Expiry (requires short durations in e2e config)
// ---------------------------------------------------------------------------

func TestOAuth2_Expiry(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run(
		"expired authorization code rejected at token exchange",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"
			verifier, challenge := testutil.GeneratePKCE()

			params := url.Values{
				"client_id":             {client.ClientID},
				"redirect_uri":         {redirectURI},
				"response_type":        {"code"},
				"scope":                {"openid"},
				"state":                {"expiry-test"},
				"code_challenge":       {challenge},
				"code_challenge_method": {"S256"},
			}

			authResp, err := testutil.OAuth2Authorize(owner, params)
			require.NoError(t, err)

			var code string
			if authResp.StatusCode == http.StatusOK {
				consentID, err := testutil.ExtractConsentID(authResp.Body)
				require.NoError(t, err)

				consentResp, err := testutil.OAuth2ConsentApprove(owner, consentID)
				require.NoError(t, err)

				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(consentResp)
				require.NoError(t, err)
			} else {
				require.Equal(t, http.StatusFound, authResp.StatusCode)
				code, err = testutil.OAuth2AuthorizeCodeFromRedirect(authResp)
				require.NoError(t, err)
			}

			// e2e config sets authorization-code-duration to 5s
			time.Sleep(6 * time.Second)

			_, raw, err := testutil.OAuth2TokenWithCode(
				owner,
				client.ClientID,
				client.ClientSecret,
				code,
				redirectURI,
				verifier,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode,
				"expired authorization code must be rejected")
		},
	)

	t.Run(
		"expired access token rejected by userinfo",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			// e2e config sets access-token-duration to 10s
			time.Sleep(11 * time.Second)

			_, raw, err := testutil.OAuth2UserInfo(owner, tokens.AccessToken)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, raw.StatusCode,
				"expired access token must be rejected by userinfo")
		},
	)

	t.Run(
		"expired access token introspects as inactive",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			// e2e config sets access-token-duration to 10s
			time.Sleep(11 * time.Second)

			introspect, raw, err := testutil.OAuth2Introspect(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.AccessToken,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)
			assert.False(t, introspect.Active,
				"expired access token must introspect as inactive")
		},
	)

	t.Run(
		"expired refresh token rejected",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreateOAuth2Client(owner, nil)
			redirectURI := "http://localhost:9999/callback"

			tokens := testutil.OAuth2PerformAuthorizationCodeFlow(
				t,
				owner,
				client.ClientID,
				client.ClientSecret,
				redirectURI,
			)

			// e2e config sets refresh-token-duration to 10s
			time.Sleep(11 * time.Second)

			_, raw, err := testutil.OAuth2TokenWithRefreshToken(
				owner,
				client.ClientID,
				client.ClientSecret,
				tokens.RefreshToken,
			)
			require.NoError(t, err)
			assert.NotEqual(t, http.StatusOK, raw.StatusCode,
				"expired refresh token must be rejected")
		},
	)

	t.Run(
		"expired device code rejected at poll",
		func(t *testing.T) {
			t.Parallel()

			client := factory.CreatePublicOAuth2Client(owner, nil)

			deviceResp, raw, err := testutil.OAuth2DeviceAuth(
				owner,
				client.ClientID,
				"openid",
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, raw.StatusCode)
			require.NotNil(t, deviceResp)

			// e2e config sets device-code-duration to 15s
			time.Sleep(16 * time.Second)

			_, errResp, _, err := testutil.OAuth2TokenWithDeviceCode(
				owner,
				client.ClientID,
				deviceResp.DeviceCode,
			)
			require.NoError(t, err)
			require.NotNil(t, errResp)
			assert.Equal(t, "expired_token", errResp.Code,
				"expired device code must return expired_token error")
		},
	)
}
