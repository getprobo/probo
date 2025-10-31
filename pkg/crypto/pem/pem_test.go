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

package pem_test

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256" // for crypto.SHA256
	"crypto/x509"
	"crypto/x509/pkix"
	stdpem "encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"go.probo.inc/probo/pkg/crypto/pem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCertificate(t *testing.T) {
	// Generate a test certificate
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "failed to generate private key")

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		DNSNames:     []string{"test.example.com"},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		SubjectKeyId: []byte{1, 2, 3, 4, 5},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err, "failed to create certificate")

	// Test encoding
	pemBytes := pem.EncodeCertificate(certDER)

	// Verify the result is valid PEM
	block, rest := stdpem.Decode(pemBytes)
	require.NotNil(t, block, "failed to decode PEM block")
	assert.Empty(t, rest, "unexpected remaining bytes")
	assert.Equal(t, "CERTIFICATE", block.Type, "incorrect block type")
	assert.True(t, bytes.Equal(block.Bytes, certDER), "certificate DER bytes don't match")

	// Verify we can parse the certificate back
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err, "failed to parse certificate from PEM")
	assert.Equal(t, "test.example.com", cert.Subject.CommonName, "incorrect common name")
}

func TestEncodePrivateKey(t *testing.T) {
	tests := []struct {
		name         string
		generateKey  func() (crypto.Signer, error)
		expectedType string
		parseFunc    func([]byte) (crypto.Signer, error)
	}{
		{
			name: "ECDSA P256",
			generateKey: func() (crypto.Signer, error) {
				return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			},
			expectedType: "EC PRIVATE KEY",
			parseFunc: func(der []byte) (crypto.Signer, error) {
				return x509.ParseECPrivateKey(der)
			},
		},
		{
			name: "ECDSA P384",
			generateKey: func() (crypto.Signer, error) {
				return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			},
			expectedType: "EC PRIVATE KEY",
			parseFunc: func(der []byte) (crypto.Signer, error) {
				return x509.ParseECPrivateKey(der)
			},
		},
		{
			name: "RSA 2048",
			generateKey: func() (crypto.Signer, error) {
				return rsa.GenerateKey(rand.Reader, 2048)
			},
			expectedType: "RSA PRIVATE KEY",
			parseFunc: func(der []byte) (crypto.Signer, error) {
				return x509.ParsePKCS1PrivateKey(der)
			},
		},
		{
			name: "ED25519",
			generateKey: func() (crypto.Signer, error) {
				_, priv, err := ed25519.GenerateKey(rand.Reader)
				return priv, err
			},
			expectedType: "PRIVATE KEY",
			parseFunc: func(der []byte) (crypto.Signer, error) {
				key, err := x509.ParsePKCS8PrivateKey(der)
				if err != nil {
					return nil, err
				}
				signer, ok := key.(crypto.Signer)
				if !ok {
					return nil, errors.New("key is not a crypto.Signer")
				}
				return signer, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate key
			key, err := tt.generateKey()
			require.NoError(t, err, "failed to generate key")

			// Encode to PEM
			pemBytes, err := pem.EncodePrivateKey(key)
			require.NoError(t, err, "failed to encode key")

			// Verify the result is valid PEM
			block, rest := stdpem.Decode(pemBytes)
			require.NotNil(t, block, "failed to decode PEM block")
			assert.Empty(t, rest, "unexpected remaining bytes")
			assert.Equal(t, tt.expectedType, block.Type, "incorrect block type")

			// Verify we can parse the key back
			parsedKey, err := tt.parseFunc(block.Bytes)
			require.NoError(t, err, "failed to parse key from PEM")
			require.NotNil(t, parsedKey, "parsed key is nil")

			// For ECDSA keys, verify the curves match
			if ecKey, ok := key.(*ecdsa.PrivateKey); ok {
				parsedECKey, ok := parsedKey.(*ecdsa.PrivateKey)
				require.True(t, ok, "parsed key is not ECDSA")
				assert.Equal(t, ecKey.Curve.Params().Name, parsedECKey.Curve.Params().Name, "curve mismatch")
			}
		})
	}
}

func TestEncodePrivateKeyUnsupportedType(t *testing.T) {
	// Test with an unsupported key type
	type unsupportedKey struct {
		crypto.Signer
	}

	key := &unsupportedKey{}
	_, err := pem.EncodePrivateKey(key)
	require.Error(t, err, "expected error for unsupported key type")
	assert.Contains(t, err.Error(), "unsupported key type", "error should mention unsupported key type")
}

func TestRoundTrip(t *testing.T) {
	// Test that we can encode and decode keys without loss
	keyTypes := []struct {
		name string
		gen  func() (crypto.Signer, error)
	}{
		{"EC256", func() (crypto.Signer, error) { return ecdsa.GenerateKey(elliptic.P256(), rand.Reader) }},
		{"EC384", func() (crypto.Signer, error) { return ecdsa.GenerateKey(elliptic.P384(), rand.Reader) }},
		{"RSA2048", func() (crypto.Signer, error) { return rsa.GenerateKey(rand.Reader, 2048) }},
		{"ED25519", func() (crypto.Signer, error) {
			_, priv, err := ed25519.GenerateKey(rand.Reader)
			return priv, err
		}},
	}

	for _, kt := range keyTypes {
		t.Run(kt.name, func(t *testing.T) {
			// Generate original key
			originalKey, err := kt.gen()
			require.NoError(t, err, "failed to generate key")

			// Encode to PEM
			pemBytes, err := pem.EncodePrivateKey(originalKey)
			require.NoError(t, err, "failed to encode key")

			// Decode PEM
			block, _ := stdpem.Decode(pemBytes)
			require.NotNil(t, block, "failed to decode PEM block")

			// Parse key based on type
			var parsedKey crypto.Signer
			switch block.Type {
			case "EC PRIVATE KEY":
				parsedKey, err = x509.ParseECPrivateKey(block.Bytes)
			case "RSA PRIVATE KEY":
				parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			case "PRIVATE KEY":
				key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err == nil {
					parsedKey = key.(crypto.Signer)
				}
			default:
				t.Fatalf("unknown key type: %s", block.Type)
			}

			require.NoError(t, err, "failed to parse key")

			// Test that both keys can sign
			// Different key types need different signing approaches
			testData := []byte("test data")
			h := crypto.SHA256.New()
			h.Write(testData)
			hashed := h.Sum(nil)

			var dataToSign []byte
			var hashFunc crypto.Hash

			switch originalKey.(type) {
			case *rsa.PrivateKey:
				dataToSign = hashed
				hashFunc = crypto.SHA256
			case ed25519.PrivateKey:
				dataToSign = testData
				hashFunc = crypto.Hash(0) // ED25519 requires zero hash
			case *ecdsa.PrivateKey:
				// ECDSA requires pre-hashed data
				dataToSign = hashed
				hashFunc = crypto.SHA256
			default:
				dataToSign = hashed
				hashFunc = crypto.SHA256
			}

			_, err = originalKey.Sign(rand.Reader, dataToSign, hashFunc)
			require.NoError(t, err, "original key failed to sign")

			switch parsedKey.(type) {
			case *rsa.PrivateKey:
				dataToSign = hashed
				hashFunc = crypto.SHA256
			case ed25519.PrivateKey:
				dataToSign = testData
				hashFunc = crypto.Hash(0) // ED25519 requires zero hash
			case *ecdsa.PrivateKey:
				// ECDSA requires pre-hashed data
				dataToSign = hashed
				hashFunc = crypto.SHA256
			default:
				dataToSign = hashed
				hashFunc = crypto.SHA256
			}

			_, err = parsedKey.Sign(rand.Reader, dataToSign, hashFunc)
			require.NoError(t, err, "parsed key failed to sign")
		})
	}
}

func TestEncodeCertificateEmptyDER(t *testing.T) {
	// Test with empty DER bytes
	pemBytes := pem.EncodeCertificate([]byte{})

	block, _ := stdpem.Decode(pemBytes)
	require.NotNil(t, block, "should encode empty DER")
	assert.Equal(t, "CERTIFICATE", block.Type)
	assert.Empty(t, block.Bytes)
}

func TestEncodeCertificateLargeDER(t *testing.T) {
	// Test with large DER bytes
	largeDER := make([]byte, 10000)
	for i := range largeDER {
		largeDER[i] = byte(i % 256)
	}

	pemBytes := pem.EncodeCertificate(largeDER)

	block, _ := stdpem.Decode(pemBytes)
	require.NotNil(t, block)
	assert.Equal(t, "CERTIFICATE", block.Type)
	assert.Equal(t, largeDER, block.Bytes)
}

func BenchmarkEncodeCertificate(b *testing.B) {
	// Generate a test certificate DER
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pem.EncodeCertificate(certDER)
	}
}

func BenchmarkEncodePrivateKey(b *testing.B) {
	benchmarks := []struct {
		name string
		key  crypto.Signer
	}{
		{"EC256", mustGenerateKey(func() (crypto.Signer, error) {
			return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		})},
		{"RSA2048", mustGenerateKey(func() (crypto.Signer, error) {
			return rsa.GenerateKey(rand.Reader, 2048)
		})},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := pem.EncodePrivateKey(bm.key)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func mustGenerateKey(gen func() (crypto.Signer, error)) crypto.Signer {
	key, err := gen()
	if err != nil {
		panic(err)
	}
	return key
}
