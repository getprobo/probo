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

package oauth2_test

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/uri"
)

var testIssuer = uri.URI("https://issuer.example.com")

func TestComputeAtHash(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns left half of sha256 base64url encoded",
		func(t *testing.T) {
			t.Parallel()

			accessToken := "ya29.test-access-token"
			h := sha256.Sum256([]byte(accessToken))
			expected := base64.RawURLEncoding.EncodeToString(h[:16])

			result := oauth2.ComputeAtHash(accessToken)
			assert.Equal(t, expected, result)
		},
	)

	t.Run(
		"different tokens produce different hashes",
		func(t *testing.T) {
			t.Parallel()

			hash1 := oauth2.ComputeAtHash("token-a")
			hash2 := oauth2.ComputeAtHash("token-b")
			assert.NotEqual(t, hash1, hash2)
		},
	)

	t.Run(
		"empty token",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2.ComputeAtHash("")
			assert.NotEmpty(t, result)
		},
	)

	t.Run(
		"deterministic",
		func(t *testing.T) {
			t.Parallel()

			hash1 := oauth2.ComputeAtHash("same-token")
			hash2 := oauth2.ComputeAtHash("same-token")
			assert.Equal(t, hash1, hash2)
		},
	)
}

func TestNewIDTokenClaims(t *testing.T) {
	t.Parallel()

	identityID := gid.Nil
	clientID := gid.Nil
	authTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run(
		"basic claims without optional scopes",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"",
				"",
				"user@example.com",
				true,
				"John Doe",
				1*time.Hour,
			)

			assert.Equal(t, testIssuer, claims.Issuer)
			assert.Equal(t, identityID.String(), claims.Subject)
			assert.Equal(t, clientID.String(), claims.Audience)
			assert.Equal(t, authTime.Unix(), claims.AuthTime)
			assert.Empty(t, claims.Nonce)
			assert.Empty(t, claims.AtHash)
			assert.Empty(t, claims.Email)
			assert.Nil(t, claims.EmailVerified)
			assert.Empty(t, claims.Name)
		},
	)

	t.Run(
		"sets nonce when provided",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			assert.Equal(t, "test-nonce", claims.Nonce)
		},
	)

	t.Run(
		"computes at_hash when access token provided",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"",
				"access-token-123",
				"",
				false,
				"",
				1*time.Hour,
			)

			expected := oauth2.ComputeAtHash("access-token-123")
			assert.Equal(t, expected, claims.AtHash)
		},
	)

	t.Run(
		"includes email claims with email scope",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID, oauth2.ScopeEmail},
				"",
				"",
				"user@example.com",
				true,
				"",
				1*time.Hour,
			)

			assert.Equal(t, "user@example.com", claims.Email)
			require.NotNil(t, claims.EmailVerified)
			assert.True(t, *claims.EmailVerified)
		},
	)

	t.Run(
		"includes name with profile scope",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID, oauth2.ScopeProfile},
				"",
				"",
				"",
				false,
				"Jane Doe",
				1*time.Hour,
			)

			assert.Equal(t, "Jane Doe", claims.Name)
		},
	)

	t.Run(
		"includes all claims with all scopes",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{
					oauth2.ScopeOpenID,
					oauth2.ScopeEmail,
					oauth2.ScopeProfile,
				},
				"nonce-val",
				"access-token",
				"user@example.com",
				false,
				"John Doe",
				1*time.Hour,
			)

			assert.Equal(t, "nonce-val", claims.Nonce)
			assert.NotEmpty(t, claims.AtHash)
			assert.Equal(t, "user@example.com", claims.Email)
			require.NotNil(t, claims.EmailVerified)
			assert.False(t, *claims.EmailVerified)
			assert.Equal(t, "John Doe", claims.Name)
		},
	)

	t.Run(
		"sets expiration based on ttl",
		func(t *testing.T) {
			t.Parallel()

			ttl := 2 * time.Hour
			before := time.Now()
			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"",
				"",
				"",
				false,
				"",
				ttl,
			)
			after := time.Now()

			assert.GreaterOrEqual(t, claims.ExpiresAt, before.Add(ttl).Unix())
			assert.LessOrEqual(t, claims.ExpiresAt, after.Add(ttl).Unix())
			assert.GreaterOrEqual(t, claims.IssuedAt, before.Unix())
			assert.LessOrEqual(t, claims.IssuedAt, after.Unix())
		},
	)

	t.Run(
		"email not verified",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID, oauth2.ScopeEmail},
				"",
				"",
				"user@example.com",
				false,
				"",
				1*time.Hour,
			)

			require.NotNil(t, claims.EmailVerified)
			assert.False(t, *claims.EmailVerified)
		},
	)
}

func testSigningKey(t *testing.T) (*rsa.PrivateKey, *jose.JWKS) {
	t.Helper()

	key, err := rsa.GenerateKey(
		strings.NewReader(strings.Repeat("deterministic-seed-for-test!!!!!", 100)),
		2048,
	)
	require.NoError(t, err)

	jwks := &jose.JWKS{
		Keys: []jose.JWK{
			jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1"),
		},
	}

	return key, jwks
}

func TestParseIDTokenIdentity(t *testing.T) {
	t.Parallel()

	identityID := gid.MustParseGID("AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp")
	clientID := gid.MustParseGID("AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp")
	authTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key, jwks := testSigningKey(t)

	parseIDToken := func(
		t *testing.T,
		token string,
		nonce string,
	) (gid.GID, error) {
		t.Helper()

		return oauth2.ParseIDTokenIdentity(
			token,
			jwks,
			nonce,
			testIssuer,
			clientID.String(),
		)
	}

	t.Run(
		"returns identity when signature nonce and expiry are valid",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			got, err := parseIDToken(t, token, "test-nonce")
			require.NoError(t, err)
			assert.Equal(t, identityID, got)
		},
	)

	t.Run(
		"rejects nonce mismatch",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"expected-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "other-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot validate nonce")
		},
	)

	t.Run(
		"rejects issuer mismatch",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				uri.URI("https://evil.example.com"),
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot validate issuer")
		},
	)

	t.Run(
		"rejects audience mismatch",
		func(t *testing.T) {
			t.Parallel()

			otherClientID := gid.New(clientID.TenantID(), coredata.OAuth2ClientEntityType)
			require.NotEqual(t, clientID, otherClientID)

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				otherClientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot validate audience")
		},
	)

	t.Run(
		"rejects expired token",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				-1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "token has expired")
		},
	)

	t.Run(
		"rejects invalid signature",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			parts[2] = base64.RawURLEncoding.EncodeToString([]byte("bad-signature"))
			tampered := strings.Join(parts, ".")

			_, err = parseIDToken(t, tampered, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot verify id token")
		},
	)

	t.Run(
		"rejects empty token",
		func(t *testing.T) {
			t.Parallel()

			_, err := oauth2.ParseIDTokenIdentity(
				"",
				jwks,
				"test-nonce",
				testIssuer,
				clientID.String(),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "missing token")
		},
	)

	t.Run(
		"rejects invalid subject",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)
			claims.Subject = "not-a-gid"

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot parse identity from id token")
		},
	)

	t.Run(
		"rejects token signed with key outside jwks",
		func(t *testing.T) {
			t.Parallel()

			otherKey, err := rsa.GenerateKey(
				strings.NewReader(strings.Repeat("other-deterministic-seed!!!!!!!!", 100)),
				2048,
			)
			require.NoError(t, err)

			claims := oauth2.NewIDTokenClaims(
				testIssuer,
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{oauth2.ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
				1*time.Hour,
			)

			token, err := jose.SignJWT(otherKey, "kid-1", claims)
			require.NoError(t, err)

			_, err = parseIDToken(t, token, "test-nonce")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot verify id token")
		},
	)
}
