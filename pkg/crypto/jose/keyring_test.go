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
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/crypto/jose"
)

// keyRingTestKeys generates the distinct RSA keys the key ring tests share. Key
// generation dominates the runtime here, so they are generated once and reused.
var keyRingTestKeys = sync.OnceValue(
	func() []*rsa.PrivateKey {
		keys := make([]*rsa.PrivateKey, 3)

		for i := range keys {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				panic(err)
			}

			keys[i] = key
		}

		return keys
	},
)

// weakRSAKey returns a key whose modulus is below the floor jose enforces.
// NewKeyRing reads only the public half, so nothing is generated.
func weakRSAKey() *rsa.PrivateKey {
	return &rsa.PrivateKey{
		N: new(big.Int).Lsh(big.NewInt(1), 1023),
		E: 65537,
	}
}

func testKeyRing(t testing.TB, active ...bool) *jose.KeyRing {
	t.Helper()

	keys := keyRingTestKeys()
	require.LessOrEqual(t, len(active), len(keys))

	signingKeys := make([]jose.SigningKey, 0, len(active))

	for i, isActive := range active {
		signingKeys = append(
			signingKeys,
			jose.SigningKey{
				PrivateKey: keys[i],
				KID:        string(rune('a' + i)),
				Active:     isActive,
			},
		)
	}

	keyRing, err := jose.NewKeyRing(signingKeys)
	require.NoError(t, err)

	return keyRing
}

func decodeKID(t testing.TB, token string) string {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)

	header := jose.JWTHeader{}
	require.NoError(t, json.Unmarshal(raw, &header))

	return header.KeyID
}

func TestNewKeyRing_Errors(t *testing.T) {
	t.Parallel()

	keys := keyRingTestKeys()

	t.Run(
		"no signing key at all",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing(nil)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "at least one signing key must be active")
		},
	)

	t.Run(
		"no active signing key",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing(
				[]jose.SigningKey{
					{PrivateKey: keys[0], KID: "a", Active: false},
					{PrivateKey: keys[1], KID: "b", Active: false},
				},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "at least one signing key must be active")
		},
	)

	t.Run(
		"missing private key",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing([]jose.SigningKey{{KID: "a", Active: true}})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "has no private key")
		},
	)

	t.Run(
		"missing key id",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing([]jose.SigningKey{{PrivateKey: keys[0], Active: true}})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "has no key id")
		},
	)

	t.Run(
		"duplicate key id",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing(
				[]jose.SigningKey{
					{PrivateKey: keys[0], KID: "a", Active: true},
					{PrivateKey: keys[1], KID: "a", Active: true},
				},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), `signing key "a" is duplicated`)
		},
	)

	t.Run(
		"duplicate key id on a retired key",
		func(t *testing.T) {
			t.Parallel()

			// Both entries are published, and a verifier stops at the first
			// match, so the collision breaks verification even though only one
			// of the two ever signs.
			_, err := jose.NewKeyRing(
				[]jose.SigningKey{
					{PrivateKey: keys[0], KID: "a", Active: false},
					{PrivateKey: keys[1], KID: "a", Active: true},
				},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), `signing key "a" is duplicated`)
		},
	)

	t.Run(
		"active signing key below the modulus floor",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing(
				[]jose.SigningKey{{PrivateKey: weakRSAKey(), KID: "a", Active: true}},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "modulus is 1024 bits")
			assert.Contains(t, err.Error(), "minimum is 2048")
		},
	)

	t.Run(
		"retired signing key below the modulus floor",
		func(t *testing.T) {
			t.Parallel()

			_, err := jose.NewKeyRing(
				[]jose.SigningKey{
					{PrivateKey: keys[0], KID: "a", Active: true},
					{PrivateKey: weakRSAKey(), KID: "b", Active: false},
				},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "modulus is 1024 bits")
		},
	)
}

func TestNewKeyRing_IgnoresLaterMutationOfTheInputSlice(t *testing.T) {
	t.Parallel()

	keys := keyRingTestKeys()

	signingKeys := []jose.SigningKey{{PrivateKey: keys[0], KID: "a", Active: true}}

	keyRing, err := jose.NewKeyRing(signingKeys)
	require.NoError(t, err)

	// Nothing validated the replacement, so the ring must keep the key it
	// accepted rather than read the caller's slice at signing time.
	signingKeys[0] = jose.SigningKey{PrivateKey: weakRSAKey(), KID: "evil", Active: true}

	token, err := keyRing.Sign(map[string]string{"sub": "test"})
	require.NoError(t, err)

	assert.Equal(t, "a", decodeKID(t, token))

	jwks := keyRing.JWKS()
	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "a", jwks.Keys[0].KeyID)
}

func TestKeyRing_Sign(t *testing.T) {
	t.Parallel()

	t.Run(
		"signs with active keys only",
		func(t *testing.T) {
			t.Parallel()

			keyRing := testKeyRing(t, false, true, true)

			seen := map[string]bool{}

			for range 10 {
				token, err := keyRing.Sign(map[string]string{"sub": "test"})
				require.NoError(t, err)

				seen[decodeKID(t, token)] = true
			}

			assert.False(t, seen["a"], "inactive key must never sign")
			assert.True(t, seen["b"])
			assert.True(t, seen["c"])
		},
	)

	t.Run(
		"produces a token the published key set verifies",
		func(t *testing.T) {
			t.Parallel()

			keyRing := testKeyRing(t, true)

			token, err := keyRing.Sign(map[string]string{"sub": "test"})
			require.NoError(t, err)

			payload, err := jose.VerifyJWTWithJWKS(token, keyRing.JWKS())
			require.NoError(t, err)

			claims := map[string]string{}
			require.NoError(t, json.Unmarshal(payload, &claims))
			assert.Equal(t, "test", claims["sub"])
		},
	)

	t.Run(
		"reports unmarshalable claims",
		func(t *testing.T) {
			t.Parallel()

			keyRing := testKeyRing(t, true)

			_, err := keyRing.Sign(func() {})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot marshal jwt claims")
		},
	)
}

func TestKeyRing_JWKS(t *testing.T) {
	t.Parallel()

	// A retired key stays published, which is what lets a token already minted
	// with it keep verifying while a verifier holds a cached key set.
	keyRing := testKeyRing(t, false, true)

	jwks := keyRing.JWKS()
	require.Len(t, jwks.Keys, 2)

	published := map[string]string{}
	for _, key := range jwks.Keys {
		published[key.KeyID] = key.Algorithm
	}

	assert.Equal(t, map[string]string{"a": "RS256", "b": "RS256"}, published)
}
