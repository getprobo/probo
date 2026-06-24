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

package bootstrap

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOAuth2SigningKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateOAuth2SigningKey()
	require.NoError(t, err)

	block, _ := pem.Decode([]byte(key))
	require.NotNil(t, block, "private key should be valid PEM")
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)

	parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	assert.Equal(t, oauth2SigningKeyBits, parsed.N.BitLen())
}

func TestGenerateOAuth2SigningKey_Unique(t *testing.T) {
	t.Parallel()

	key1, err := GenerateOAuth2SigningKey()
	require.NoError(t, err)

	key2, err := GenerateOAuth2SigningKey()
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2)
}
