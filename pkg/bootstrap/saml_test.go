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

package bootstrap

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSAMLCertificate(t *testing.T) {
	cert, key, err := GenerateSAMLCertificate()
	require.NoError(t, err)

	certBlock, _ := pem.Decode([]byte(cert))
	require.NotNil(t, certBlock, "certificate should be valid PEM")
	assert.Equal(t, "CERTIFICATE", certBlock.Type)

	keyBlock, _ := pem.Decode([]byte(key))
	require.NotNil(t, keyBlock, "private key should be valid PEM")
	assert.Equal(t, "RSA PRIVATE KEY", keyBlock.Type)

	parsedCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	assert.Equal(t, "probo-saml", parsedCert.Subject.CommonName)
	assert.Equal(t, []string{"Probo"}, parsedCert.Subject.Organization)
	assert.Equal(t, []string{"US"}, parsedCert.Subject.Country)

	assert.True(t, parsedCert.NotBefore.Before(time.Now().Add(time.Minute)))
	assert.True(t, parsedCert.NotAfter.After(time.Now().AddDate(9, 0, 0)))
	assert.True(t, parsedCert.NotAfter.Before(time.Now().AddDate(11, 0, 0)))
}

func TestGenerateSAMLCertificate_UniqueSerials(t *testing.T) {
	cert1, _, err := GenerateSAMLCertificate()
	require.NoError(t, err)

	cert2, _, err := GenerateSAMLCertificate()
	require.NoError(t, err)

	block1, _ := pem.Decode([]byte(cert1))
	block2, _ := pem.Decode([]byte(cert2))

	parsed1, err := x509.ParseCertificate(block1.Bytes)
	require.NoError(t, err)

	parsed2, err := x509.ParseCertificate(block2.Bytes)
	require.NoError(t, err)

	assert.NotEqual(t, parsed1.SerialNumber, parsed2.SerialNumber)
}
