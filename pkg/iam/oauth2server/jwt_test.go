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

package oauth2server_test

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(
		strings.NewReader(strings.Repeat("deterministic-seed-for-test!!!!!", 100)),
		2048,
	)
	require.NoError(t, err)

	return key
}

func TestSignIDToken(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"produces valid three-part JWT",
		func(t *testing.T) {
			t.Parallel()

			claims := &oauth2server.IDTokenClaims{
				Issuer:    "https://issuer.example.com",
				Subject:   "sub-123",
				Audience:  "aud-456",
				ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
				AuthTime:  time.Now().Unix(),
			}

			token, err := oauth2server.SignIDToken(&oauth2server.SigningKey{PrivateKey: key, KID: "kid-1"}, claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			require.Len(t, parts, 3)
		},
	)

	t.Run(
		"header contains correct fields",
		func(t *testing.T) {
			t.Parallel()

			claims := &oauth2server.IDTokenClaims{
				Issuer:   "https://issuer.example.com",
				Subject:  "sub-123",
				Audience: "aud-456",
			}

			token, err := oauth2server.SignIDToken(&oauth2server.SigningKey{PrivateKey: key, KID: "my-kid"}, claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)

			var header oauth2server.JWTHeader
			err = json.Unmarshal(headerJSON, &header)
			require.NoError(t, err)

			assert.Equal(t, "RS256", header.Algorithm)
			assert.Equal(t, "JWT", header.Type)
			assert.Equal(t, "my-kid", header.KeyID)
		},
	)

	t.Run(
		"claims are correctly encoded",
		func(t *testing.T) {
			t.Parallel()

			now := time.Now()

			claims := &oauth2server.IDTokenClaims{
				Issuer:    "https://issuer.example.com",
				Subject:   "sub-123",
				Audience:  "aud-456",
				ExpiresAt: now.Add(1 * time.Hour).Unix(),
				IssuedAt:  now.Unix(),
				AuthTime:  now.Unix(),
				Nonce:     "test-nonce",
			}

			token, err := oauth2server.SignIDToken(&oauth2server.SigningKey{PrivateKey: key, KID: "kid-1"}, claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var decoded oauth2server.IDTokenClaims
			err = json.Unmarshal(claimsJSON, &decoded)
			require.NoError(t, err)

			assert.Equal(t, "https://issuer.example.com", decoded.Issuer)
			assert.Equal(t, "sub-123", decoded.Subject)
			assert.Equal(t, "aud-456", decoded.Audience)
			assert.Equal(t, now.Add(1*time.Hour).Unix(), decoded.ExpiresAt)
			assert.Equal(t, now.Unix(), decoded.IssuedAt)
			assert.Equal(t, now.Unix(), decoded.AuthTime)
			assert.Equal(t, "test-nonce", decoded.Nonce)
		},
	)

	t.Run(
		"signature is verifiable",
		func(t *testing.T) {
			t.Parallel()

			claims := &oauth2server.IDTokenClaims{
				Issuer:   "https://issuer.example.com",
				Subject:  "sub-123",
				Audience: "aud-456",
			}

			token, err := oauth2server.SignIDToken(&oauth2server.SigningKey{PrivateKey: key, KID: "kid-1"}, claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			signingInput := parts[0] + "." + parts[1]
			signature, err := base64.RawURLEncoding.DecodeString(parts[2])
			require.NoError(t, err)

			h := sha256.Sum256([]byte(signingInput))
			err = rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, h[:], signature)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"omits empty optional claims",
		func(t *testing.T) {
			t.Parallel()

			claims := &oauth2server.IDTokenClaims{
				Issuer:   "https://issuer.example.com",
				Subject:  "sub-123",
				Audience: "aud-456",
			}

			token, err := oauth2server.SignIDToken(&oauth2server.SigningKey{PrivateKey: key, KID: "kid-1"}, claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var raw map[string]json.RawMessage
			err = json.Unmarshal(claimsJSON, &raw)
			require.NoError(t, err)

			assert.NotContains(t, raw, "nonce")
			assert.NotContains(t, raw, "at_hash")
			assert.NotContains(t, raw, "email")
			assert.NotContains(t, raw, "email_verified")
			assert.NotContains(t, raw, "name")
		},
	)
}

func TestComputeAtHash(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns left half of sha256 base64url encoded",
		func(t *testing.T) {
			t.Parallel()

			accessToken := "ya29.test-access-token"
			h := sha256.Sum256([]byte(accessToken))
			expected := base64.RawURLEncoding.EncodeToString(h[:16])

			result := oauth2server.ComputeAtHash(accessToken)
			assert.Equal(t, expected, result)
		},
	)

	t.Run(
		"different tokens produce different hashes",
		func(t *testing.T) {
			t.Parallel()

			hash1 := oauth2server.ComputeAtHash("token-a")
			hash2 := oauth2server.ComputeAtHash("token-b")
			assert.NotEqual(t, hash1, hash2)
		},
	)

	t.Run(
		"empty token",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ComputeAtHash("")
			assert.NotEmpty(t, result)
		},
	)

	t.Run(
		"deterministic",
		func(t *testing.T) {
			t.Parallel()

			hash1 := oauth2server.ComputeAtHash("same-token")
			hash2 := oauth2server.ComputeAtHash("same-token")
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

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID},
				"",
				"",
				"user@example.com",
				true,
				"John Doe",
			)

			assert.Equal(t, "https://issuer.example.com", claims.Issuer)
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

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID},
				"test-nonce",
				"",
				"",
				false,
				"",
			)

			assert.Equal(t, "test-nonce", claims.Nonce)
		},
	)

	t.Run(
		"computes at_hash when access token provided",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID},
				"",
				"access-token-123",
				"",
				false,
				"",
			)

			expected := oauth2server.ComputeAtHash("access-token-123")
			assert.Equal(t, expected, claims.AtHash)
		},
	)

	t.Run(
		"includes email claims with email scope",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID, coredata.OAuth2ScopeEmail},
				"",
				"",
				"user@example.com",
				true,
				"",
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

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID, coredata.OAuth2ScopeProfile},
				"",
				"",
				"",
				false,
				"Jane Doe",
			)

			assert.Equal(t, "Jane Doe", claims.Name)
		},
	)

	t.Run(
		"includes all claims with all scopes",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{
					coredata.OAuth2ScopeOpenID,
					coredata.OAuth2ScopeEmail,
					coredata.OAuth2ScopeProfile,
				},
				"nonce-val",
				"access-token",
				"user@example.com",
				false,
				"John Doe",
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
		"sets expiration to one hour from now",
		func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID},
				"",
				"",
				"",
				false,
				"",
			)
			after := time.Now()

			assert.GreaterOrEqual(t, claims.ExpiresAt, before.Add(1*time.Hour).Unix())
			assert.LessOrEqual(t, claims.ExpiresAt, after.Add(1*time.Hour).Unix())
			assert.GreaterOrEqual(t, claims.IssuedAt, before.Unix())
			assert.LessOrEqual(t, claims.IssuedAt, after.Unix())
		},
	)

	t.Run(
		"email not verified",
		func(t *testing.T) {
			t.Parallel()

			claims := oauth2server.NewIDTokenClaims(
				"https://issuer.example.com",
				identityID,
				clientID,
				authTime,
				coredata.OAuth2Scopes{coredata.OAuth2ScopeOpenID, coredata.OAuth2ScopeEmail},
				"",
				"",
				"user@example.com",
				false,
				"",
			)

			require.NotNil(t, claims.EmailVerified)
			assert.False(t, *claims.EmailVerified)
		},
	)
}

func TestPublicJWKS(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"returns single key in set",
		func(t *testing.T) {
			t.Parallel()

			jwks := oauth2server.PublicJWKS([]oauth2server.SigningKey{
				{PrivateKey: key, KID: "kid-1"},
			})
			require.Len(t, jwks.Keys, 1)
		},
	)

	t.Run(
		"key fields are correct",
		func(t *testing.T) {
			t.Parallel()

			jwks := oauth2server.PublicJWKS([]oauth2server.SigningKey{
				{PrivateKey: key, KID: "kid-1"},
			})
			jwk := jwks.Keys[0]

			assert.Equal(t, "RSA", jwk.KeyType)
			assert.Equal(t, "sig", jwk.Use)
			assert.Equal(t, "RS256", jwk.Algorithm)
			assert.Equal(t, "kid-1", jwk.KeyID)
		},
	)

	t.Run(
		"modulus encodes correctly",
		func(t *testing.T) {
			t.Parallel()

			jwks := oauth2server.PublicJWKS([]oauth2server.SigningKey{
				{PrivateKey: key, KID: "kid-1"},
			})
			jwk := jwks.Keys[0]

			nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
			require.NoError(t, err)

			n := new(big.Int).SetBytes(nBytes)
			assert.Equal(t, key.N, n)
		},
	)

	t.Run(
		"exponent encodes correctly",
		func(t *testing.T) {
			t.Parallel()

			jwks := oauth2server.PublicJWKS([]oauth2server.SigningKey{
				{PrivateKey: key, KID: "kid-1"},
			})
			jwk := jwks.Keys[0]

			eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
			require.NoError(t, err)

			e := new(big.Int).SetBytes(eBytes)
			assert.Equal(t, int64(key.E), e.Int64())
		},
	)

	t.Run(
		"multiple keys are all included",
		func(t *testing.T) {
			t.Parallel()

			key2 := testRSAKey(t)
			jwks := oauth2server.PublicJWKS([]oauth2server.SigningKey{
				{PrivateKey: key, KID: "kid-a"},
				{PrivateKey: key2, KID: "kid-b"},
			})

			require.Len(t, jwks.Keys, 2)
			assert.Equal(t, "kid-a", jwks.Keys[0].KeyID)
			assert.Equal(t, "kid-b", jwks.Keys[1].KeyID)
		},
	)
}
