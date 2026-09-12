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
	"go.probo.inc/probo/pkg/uri"
)

func TestNewManualAccessToken_CloudFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseURL   uri.URI
		prefix    string
		tokenType string
	}{
		{
			name:      "US cloud",
			baseURL:   "https://us.probo.com",
			prefix:    cloudUSAPITokenPrefix,
			tokenType: cloudUSSecretScanningType,
		},
		{
			name:      "EU cloud",
			baseURL:   "https://eu.probo.com",
			prefix:    cloudEUAPITokenPrefix,
			tokenType: cloudEUSecretScanningType,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				token := newManualAccessToken(tt.baseURL)
				prefix, tokenType := cloudManualAccessTokenFormat(tt.baseURL)

				assert.Equal(t, tt.prefix, prefix)
				assert.Equal(t, tt.tokenType, tokenType)
				assert.Len(t, token, len(tt.prefix)+cloudTokenRandomLength+cloudTokenChecksumLength)
				assert.True(t, strings.HasPrefix(token, tt.prefix))
				assert.True(t, isValidManualAccessToken(tt.baseURL, token))
			},
		)
	}
}

func TestNewManualAccessToken_SelfHostedLegacyFormat(t *testing.T) {
	t.Parallel()

	baseURL := uri.URI("https://probo.example.com")
	token := newManualAccessToken(baseURL)
	prefix, tokenType := cloudManualAccessTokenFormat(baseURL)

	assert.Empty(t, prefix)
	assert.Empty(t, tokenType)
	assert.Len(t, token, tokenByteLength*2)
	assert.Regexp(t, "^[0-9a-f]{64}$", token)
	assert.False(t, isValidManualAccessToken(baseURL, token))
}

func TestIsValidManualAccessToken_RejectsAlteredToken(t *testing.T) {
	t.Parallel()

	baseURL := uri.URI("https://us.probo.com")
	token := newManualAccessToken(baseURL)
	last := token[len(token)-1]
	replacement := byte('0')
	if last == replacement {
		replacement = '1'
	}

	altered := token[:len(token)-1] + string(replacement)

	assert.False(t, isValidManualAccessToken(baseURL, altered))
	assert.False(t, isValidManualAccessToken("https://eu.probo.com", token))
	assert.False(t, isValidManualAccessToken(baseURL, token+"0"))
}

func TestEncodeBase62Checksum_GitHubReferenceValue(t *testing.T) {
	t.Parallel()

	randomValue := "IqIMNOZH6zOwIEB4T9A2g4EHMy8Ji4"
	require.Len(t, randomValue, 30)

	assert.Equal(t, "2q4HA5", encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue))))
}
