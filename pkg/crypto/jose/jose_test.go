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

package jose_test

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/crypto/jose"
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

func TestRSAPublicKeyToJWK(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"sets fixed RSA signature fields",
		func(t *testing.T) {
			t.Parallel()

			jwk := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1")

			assert.Equal(t, "RSA", jwk.KeyType)
			assert.Equal(t, "sig", jwk.Use)
			assert.Equal(t, "RS256", jwk.Algorithm)
			assert.Equal(t, "kid-1", jwk.KeyID)
		},
	)

	t.Run(
		"encodes modulus correctly",
		func(t *testing.T) {
			t.Parallel()

			jwk := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1")

			nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
			require.NoError(t, err)

			n := new(big.Int).SetBytes(nBytes)
			assert.Equal(t, key.N, n)
		},
	)

	t.Run(
		"encodes exponent correctly",
		func(t *testing.T) {
			t.Parallel()

			jwk := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1")

			eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
			require.NoError(t, err)

			e := new(big.Int).SetBytes(eBytes)
			assert.Equal(t, int64(key.E), e.Int64())
		},
	)

	t.Run(
		"different key IDs produce different JWKs",
		func(t *testing.T) {
			t.Parallel()

			jwk1 := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-a")
			jwk2 := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-b")

			assert.Equal(t, "kid-a", jwk1.KeyID)
			assert.Equal(t, "kid-b", jwk2.KeyID)
			assert.Equal(t, jwk1.N, jwk2.N)
		},
	)
}

func TestSignJWT(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"produces valid three-part JWT",
		func(t *testing.T) {
			t.Parallel()

			claims := map[string]string{"sub": "test"}

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3)
		},
	)

	t.Run(
		"header contains correct fields",
		func(t *testing.T) {
			t.Parallel()

			claims := map[string]string{"sub": "test"}

			token, err := jose.SignJWT(key, "my-kid", claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)

			var header jose.JWTHeader
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

			claims := map[string]any{
				"iss": "https://issuer.example.com",
				"sub": "sub-123",
				"aud": "aud-456",
			}

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)

			var decoded map[string]any
			err = json.Unmarshal(claimsJSON, &decoded)
			require.NoError(t, err)

			assert.Equal(t, "https://issuer.example.com", decoded["iss"])
			assert.Equal(t, "sub-123", decoded["sub"])
			assert.Equal(t, "aud-456", decoded["aud"])
		},
	)

	t.Run(
		"signature is verifiable",
		func(t *testing.T) {
			t.Parallel()

			claims := map[string]string{"sub": "test"}

			token, err := jose.SignJWT(key, "kid-1", claims)
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
}

func TestJWK_JSON(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"marshals to expected JSON field names",
		func(t *testing.T) {
			t.Parallel()

			jwk := jose.RSAPublicKeyToJWK(&key.PublicKey, "test-kid")

			data, err := json.Marshal(jwk)
			require.NoError(t, err)

			var raw map[string]string
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			assert.Equal(t, "RSA", raw["kty"])
			assert.Equal(t, "sig", raw["use"])
			assert.Equal(t, "RS256", raw["alg"])
			assert.Equal(t, "test-kid", raw["kid"])
			assert.NotEmpty(t, raw["n"])
			assert.NotEmpty(t, raw["e"])
		},
	)
}

func TestJWKS_JSON(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"marshals keys array",
		func(t *testing.T) {
			t.Parallel()

			jwks := jose.JWKS{
				Keys: []jose.JWK{
					jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1"),
					jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-2"),
				},
			}

			data, err := json.Marshal(jwks)
			require.NoError(t, err)

			var raw struct {
				Keys []json.RawMessage `json:"keys"`
			}
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			assert.Len(t, raw.Keys, 2)
		},
	)
}

func TestJWTHeader_JSON(t *testing.T) {
	t.Parallel()

	t.Run(
		"marshals to expected JSON field names",
		func(t *testing.T) {
			t.Parallel()

			header := jose.JWTHeader{
				Algorithm: "RS256",
				Type:      "JWT",
				KeyID:     "my-kid",
			}

			data, err := json.Marshal(header)
			require.NoError(t, err)

			var raw map[string]string
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			assert.Equal(t, "RS256", raw["alg"])
			assert.Equal(t, "JWT", raw["typ"])
			assert.Equal(t, "my-kid", raw["kid"])
		},
	)
}
