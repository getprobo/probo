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

func TestRSAPublicKeyFromJWK(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)
	jwk := jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1")

	t.Run(
		"round trips rsa public key",
		func(t *testing.T) {
			t.Parallel()

			pubKey, err := jose.RSAPublicKeyFromJWK(jwk)
			require.NoError(t, err)
			assert.Equal(t, key.N, pubKey.N)
			assert.Equal(t, key.E, pubKey.E)
		},
	)

	t.Run(
		"rejects unsupported key type",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.RSAPublicKeyFromJWK(jose.JWK{KeyType: "EC"})
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects invalid rsa exponent",
		func(t *testing.T) {
			t.Parallel()

			invalidExponent := jwk
			invalidExponent.E = base64.RawURLEncoding.EncodeToString(
				new(big.Int).Lsh(big.NewInt(1), 128).Bytes(),
			)

			_, err := jose.RSAPublicKeyFromJWK(invalidExponent)
			require.Error(t, err)
		},
	)
}

func TestPublicKeyFromJWKS(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)
	jwks := &jose.JWKS{
		Keys: []jose.JWK{
			jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1"),
			jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-2"),
		},
	}

	t.Run(
		"finds matching key",
		func(t *testing.T) {
			t.Parallel()

			pubKey, err := jose.PublicKeyFromJWKS(jwks, "kid-2")
			require.NoError(t, err)
			assert.Equal(t, key.E, pubKey.E)
		},
	)

	t.Run(
		"errors when kid is missing",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.PublicKeyFromJWKS(jwks, "missing")
			require.Error(t, err)
		},
	)

	t.Run(
		"errors when jwks is nil",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.PublicKeyFromJWKS(nil, "kid-1")
			require.Error(t, err)
		},
	)
}

func TestVerifyJWT(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)

	t.Run(
		"verifies signed jwt",
		func(t *testing.T) {
			t.Parallel()

			claims := map[string]string{"sub": "test"}

			token, err := jose.SignJWT(key, "kid-1", claims)
			require.NoError(t, err)

			payload, err := jose.VerifyJWT(token, &key.PublicKey)
			require.NoError(t, err)

			var decoded map[string]string

			err = json.Unmarshal(payload, &decoded)
			require.NoError(t, err)
			assert.Equal(t, "test", decoded["sub"])
		},
	)

	t.Run(
		"rejects malformed jwt",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.VerifyJWT("", &key.PublicKey)
			require.Error(t, err)

			_, err = jose.VerifyJWT("only-one-part", &key.PublicKey)
			require.Error(t, err)

			_, err = jose.VerifyJWT("two.parts", &key.PublicKey)
			require.Error(t, err)

			_, err = jose.VerifyJWT("too.many.parts.here", &key.PublicKey)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects tampered payload",
		func(t *testing.T) {
			t.Parallel()

			token, err := jose.SignJWT(key, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"evil"}`))
			tampered := strings.Join(parts, ".")

			_, err = jose.VerifyJWT(tampered, &key.PublicKey)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects tampered header",
		func(t *testing.T) {
			t.Parallel()

			token, err := jose.SignJWT(key, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			parts := strings.Split(token, ".")
			parts[0] = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT","kid":"kid-1"}`))
			tampered := strings.Join(parts, ".")

			_, err = jose.VerifyJWT(tampered, &key.PublicKey)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects unsupported algorithm",
		func(t *testing.T) {
			t.Parallel()

			cases := []struct {
				name string
				alg  string
			}{
				{name: "hs256", alg: "HS256"},
				{name: "none", alg: "none"},
				{name: "lowercase rs256", alg: "rs256"},
			}

			for _, tc := range cases {
				t.Run(
					tc.name,
					func(t *testing.T) {
						t.Parallel()

						header := base64.RawURLEncoding.EncodeToString(
							[]byte(`{"alg":"` + tc.alg + `","typ":"JWT","kid":"kid-1"}`),
						)
						payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"test"}`))
						token := header + "." + payload + ".c2ln"

						_, err := jose.VerifyJWT(token, &key.PublicKey)
						require.Error(t, err)
					},
				)
			}
		},
	)

	t.Run(
		"rejects signature from different key",
		func(t *testing.T) {
			t.Parallel()

			otherKey := testRSAKey(t)

			token, err := jose.SignJWT(otherKey, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			_, err = jose.VerifyJWT(token, &key.PublicKey)
			require.Error(t, err)
		},
	)
}

func TestVerifyJWTWithJWKS(t *testing.T) {
	t.Parallel()

	key := testRSAKey(t)
	jwks := &jose.JWKS{
		Keys: []jose.JWK{
			jose.RSAPublicKeyToJWK(&key.PublicKey, "kid-1"),
		},
	}

	t.Run(
		"verifies signed jwt",
		func(t *testing.T) {
			t.Parallel()

			token, err := jose.SignJWT(key, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			payload, err := jose.VerifyJWTWithJWKS(token, jwks)
			require.NoError(t, err)

			var decoded map[string]string

			err = json.Unmarshal(payload, &decoded)
			require.NoError(t, err)
			assert.Equal(t, "test", decoded["sub"])
		},
	)

	t.Run(
		"rejects missing key id",
		func(t *testing.T) {
			t.Parallel()

			header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
			payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"test"}`))
			token := header + "." + payload + ".c2ln"

			_, err := jose.VerifyJWTWithJWKS(token, jwks)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects unknown key id",
		func(t *testing.T) {
			t.Parallel()

			token, err := jose.SignJWT(key, "unknown-kid", map[string]string{"sub": "test"})
			require.NoError(t, err)

			_, err = jose.VerifyJWTWithJWKS(token, jwks)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects token signed with key outside jwks",
		func(t *testing.T) {
			t.Parallel()

			otherKey := testRSAKey(t)

			token, err := jose.SignJWT(otherKey, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			_, err = jose.VerifyJWTWithJWKS(token, jwks)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects nil jwks",
		func(t *testing.T) {
			t.Parallel()

			token, err := jose.SignJWT(key, "kid-1", map[string]string{"sub": "test"})
			require.NoError(t, err)

			_, err = jose.VerifyJWTWithJWKS(token, nil)
			require.Error(t, err)
		},
	)
}
