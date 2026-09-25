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

package probodconfig_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/probodconfig"
	"sigs.k8s.io/yaml"
)

func rsaPEM(t *testing.T, pkcs8 bool) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	if pkcs8 {
		der, err := x509.MarshalPKCS8PrivateKey(key)
		require.NoError(t, err)

		block = &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	}

	return string(pem.EncodeToMemory(block))
}

func ecPEM(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
}

func TestParsePrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pemData string
		wantErr bool
		wantSet bool
	}{
		{name: "rsa", pemData: rsaPEM(t, false), wantSet: true},
		{name: "ec", pemData: ecPEM(t), wantSet: true},
		{name: "empty is the zero value", pemData: ""},
		{name: "malformed", pemData: "not a pem block", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			privateKey, err := probodconfig.ParsePrivateKey(tt.pemData)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSet, !privateKey.IsZero())
			assert.Equal(t, tt.pemData, privateKey.PEM())
		})
	}
}

func TestParseRSAPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pemData string
		wantErr bool
		wantSet bool
	}{
		{name: "pkcs1", pemData: rsaPEM(t, false), wantSet: true},
		{name: "pkcs8", pemData: rsaPEM(t, true), wantSet: true},
		{name: "ec is rejected", pemData: ecPEM(t), wantErr: true},
		{name: "empty is the zero value", pemData: ""},
		{name: "malformed", pemData: "-----BEGIN RSA PRIVATE KEY-----", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			privateKey, err := probodconfig.ParseRSAPrivateKey(tt.pemData)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSet, !privateKey.IsZero())
			assert.Equal(t, tt.pemData, privateKey.PEM())
			assert.Equal(t, tt.wantSet, privateKey.PrivateKey() != nil)
		})
	}
}

// bootstrap writes the configuration file that probod reads back, so a key has
// to survive YAML unchanged: secrets are derived from the PEM as configured.
func TestPrivateKeys_YAMLRoundTrip(t *testing.T) {
	t.Parallel()

	signingKeyPEM := rsaPEM(t, false)
	accountKeyPEM := ecPEM(t)

	signingKey, err := probodconfig.ParseRSAPrivateKey(signingKeyPEM)
	require.NoError(t, err)

	accountKey, err := probodconfig.ParsePrivateKey(accountKeyPEM)
	require.NoError(t, err)

	data, err := yaml.Marshal(
		probodconfig.CustomDomainsConfig{
			ACME: probodconfig.ACMEConfig{AccountKey: accountKey},
		},
	)
	require.NoError(t, err)

	var customDomains probodconfig.CustomDomainsConfig

	require.NoError(t, yaml.Unmarshal(data, &customDomains))
	assert.Equal(t, accountKeyPEM, customDomains.ACME.AccountKey.PEM())

	data, err = yaml.Marshal(
		probodconfig.OAuth2SigningKeyConfig{PrivateKey: signingKey, KID: "default", Active: true},
	)
	require.NoError(t, err)

	var keyCfg probodconfig.OAuth2SigningKeyConfig

	require.NoError(t, yaml.Unmarshal(data, &keyCfg))
	assert.Equal(t, signingKeyPEM, keyCfg.PrivateKey.PEM())
	assert.Equal(t, signingKey.PrivateKey(), keyCfg.PrivateKey.PrivateKey())
}

// An unset optional key is left out of the file, so probod never reads back an
// empty account-key it would have to special-case.
func TestPrivateKey_OmittedWhenUnset(t *testing.T) {
	t.Parallel()

	data, err := yaml.Marshal(
		probodconfig.ACMEConfig{Email: "ops@example.com"},
	)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "account-key")
}
