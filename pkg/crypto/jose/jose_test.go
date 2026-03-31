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
	"crypto/rsa"
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
