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
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

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
				1*time.Hour,
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
				1*time.Hour,
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
				1*time.Hour,
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
				1*time.Hour,
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
				1*time.Hour,
			)

			require.NotNil(t, claims.EmailVerified)
			assert.False(t, *claims.EmailVerified)
		},
	)
}
