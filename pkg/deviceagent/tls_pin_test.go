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

package deviceagent

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifySPKIPins(t *testing.T) {
	t.Parallel()

	ca, caCert := mustTestCA(t, "Amazon")
	leaf := mustLeafCert(t, ca, USConsoleHost)
	other := mustLeafCert(t, ca, USConsoleHost)
	pin := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
	chain := [][]*x509.Certificate{{leaf, caCert}}

	t.Run(
		"pinned leaf in verified chain",
		func(t *testing.T) {
			t.Parallel()

			err := verifySPKIPins(chain, [][32]byte{pin})
			require.NoError(t, err)
		},
	)

	t.Run(
		"unpinned leaf",
		func(t *testing.T) {
			t.Parallel()

			otherPin := sha256.Sum256(other.RawSubjectPublicKeyInfo)
			err := verifySPKIPins(chain, [][32]byte{otherPin})
			require.ErrorIs(t, err, ErrProboCloudPinMismatch)
		},
	)

	t.Run(
		"empty verified chains",
		func(t *testing.T) {
			t.Parallel()

			err := verifyProboCloudSPKI(nil, nil)
			require.ErrorIs(t, err, ErrProboCloudPinMismatch)
		},
	)
}

type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func mustTestCA(t *testing.T, org string) (*testCA, *x509.Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{org}, CommonName: org + " Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return &testCA{cert: cert, key: key}, cert
}

func mustLeafCert(t *testing.T, ca *testCA, dnsName string) *x509.Certificate {
	t.Helper()

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: dnsName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{dnsName},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &leafKey.PublicKey, ca.key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return cert
}
