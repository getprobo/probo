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

package oauth2

import (
	"hash/crc32"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

func TestNewManualAccessToken_Format(t *testing.T) {
	t.Parallel()

	token := newManualAccessToken()

	assert.Len(t, token, len(manualAPITokenPrefix)+manualTokenRandomLength+manualTokenChecksumLength)
	assert.True(t, strings.HasPrefix(token, manualAPITokenPrefix))
	assert.True(t, isValidManualAccessToken(token))
}

func TestService_SecretScanningTokenTypeIsDeploymentIndependent(t *testing.T) {
	t.Parallel()

	first := NewService(nil, nil, "https://first.example.com", log.NewLogger())
	second := NewService(nil, nil, "https://second.example.com", log.NewLogger())

	assert.Equal(t, secretScanningTokenType, first.SecretScanningTokenType())
	assert.Equal(t, first.SecretScanningTokenType(), second.SecretScanningTokenType())
}

func TestIsValidManualAccessToken_RejectsAlteredToken(t *testing.T) {
	t.Parallel()

	token := newManualAccessToken()
	last := token[len(token)-1]
	replacement := byte('0')
	if last == replacement {
		replacement = '1'
	}

	altered := token[:len(token)-1] + string(replacement)

	assert.False(t, isValidManualAccessToken(altered))
	assert.False(t, isValidManualAccessToken(token+"0"))
	assert.False(
		t,
		isValidManualAccessToken("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
	)
}

func TestEncodeBase62Checksum_GitHubReferenceValue(t *testing.T) {
	t.Parallel()

	randomValue := "IqIMNOZH6zOwIEB4T9A2g4EHMy8Ji4"
	require.Len(t, randomValue, 30)

	assert.Equal(t, "2q4HA5", encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue))))
}
