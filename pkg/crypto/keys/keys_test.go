// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package keys_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"errors"
	"testing"

	"go.probo.inc/probo/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		keyType     keys.Type
		checkFunc   func(t *testing.T, key interface{})
		expectError bool
	}{
		{
			name:    "EC256",
			keyType: keys.TypeEC256,
			checkFunc: func(t *testing.T, key interface{}) {
				ecKey, ok := key.(*ecdsa.PrivateKey)
				require.True(t, ok, "expected *ecdsa.PrivateKey, got %T", key)
				assert.Equal(t, elliptic.P256(), ecKey.Curve, "expected P256 curve")
			},
		},
		{
			name:    "EC384",
			keyType: keys.TypeEC384,
			checkFunc: func(t *testing.T, key interface{}) {
				ecKey, ok := key.(*ecdsa.PrivateKey)
				require.True(t, ok, "expected *ecdsa.PrivateKey, got %T", key)
				assert.Equal(t, elliptic.P384(), ecKey.Curve, "expected P384 curve")
			},
		},
		{
			name:    "RSA2048",
			keyType: keys.TypeRSA2048,
			checkFunc: func(t *testing.T, key interface{}) {
				rsaKey, ok := key.(*rsa.PrivateKey)
				require.True(t, ok, "expected *rsa.PrivateKey, got %T", key)
				bitSize := rsaKey.N.BitLen()
				assert.GreaterOrEqual(t, bitSize, 2047, "RSA key too small")
				assert.LessOrEqual(t, bitSize, 2048, "RSA key too large")
			},
		},
		{
			name:    "RSA4096",
			keyType: keys.TypeRSA4096,
			checkFunc: func(t *testing.T, key interface{}) {
				rsaKey, ok := key.(*rsa.PrivateKey)
				require.True(t, ok, "expected *rsa.PrivateKey, got %T", key)
				bitSize := rsaKey.N.BitLen()
				assert.GreaterOrEqual(t, bitSize, 4095, "RSA key too small")
				assert.LessOrEqual(t, bitSize, 4096, "RSA key too large")
			},
		},
		{
			name:        "Invalid key type",
			keyType:     keys.Type("INVALID"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := keys.Generate(tt.keyType)

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			require.NoError(t, err, "unexpected error")
			require.NotNil(t, key, "expected key, got nil")

			// Verify the key can be used for signing (crypto.Signer has Public() method)
			assert.NotNil(t, key.Public(), "key.Public() returned nil")

			if tt.checkFunc != nil {
				tt.checkFunc(t, key)
			}
		})
	}
}

func TestGenerateConcurrency(t *testing.T) {
	// Test that key generation is safe for concurrent use
	keyTypes := []keys.Type{
		keys.TypeEC256,
		keys.TypeEC384,
		keys.TypeRSA2048,
		keys.TypeRSA4096,
	}

	for _, keyType := range keyTypes {
		t.Run(string(keyType), func(t *testing.T) {
			t.Parallel()

			const numGoroutines = 10
			errorsChan := make(chan error, numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					key, err := keys.Generate(keyType)
					if err != nil {
						errorsChan <- err
						return
					}
					if key == nil {
						errorsChan <- errors.New("generated key is nil")
						return
					}
					errorsChan <- nil
				}()
			}

			for i := 0; i < numGoroutines; i++ {
				err := <-errorsChan
				assert.NoError(t, err, "concurrent generation failed")
			}
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	benchmarks := []struct {
		name    string
		keyType keys.Type
	}{
		{"EC256", keys.TypeEC256},
		{"EC384", keys.TypeEC384},
		{"RSA2048", keys.TypeRSA2048},
		{"RSA4096", keys.TypeRSA4096},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := keys.Generate(bm.keyType)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}