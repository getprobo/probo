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
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/uri"
)

const (
	base62Alphabet            = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	cloudTokenRandomLength    = 32
	cloudTokenChecksumLength  = 6
	cloudUSAPITokenPrefix     = "prb_a1u_"
	cloudEUAPITokenPrefix     = "prb_a1e_"
	cloudUSSecretScanningType = "probo_cloud_api_token_us"
	cloudEUSecretScanningType = "probo_cloud_api_token_eu"
	cloudUSHostname           = "us.probo.com"
	cloudEUHostname           = "eu.probo.com"
)

func newManualAccessToken(baseURL uri.URI) string {
	prefix, _ := cloudManualAccessTokenFormat(baseURL)
	if prefix == "" {
		return rand.MustHexString(tokenByteLength)
	}

	randomValue := rand.MustStringFromAlphabet(base62Alphabet, cloudTokenRandomLength)

	return prefix + randomValue + encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue)))
}

func isValidManualAccessToken(baseURL uri.URI, token string) bool {
	prefix, _ := cloudManualAccessTokenFormat(baseURL)
	if prefix == "" || !strings.HasPrefix(token, prefix) {
		return false
	}

	body := strings.TrimPrefix(token, prefix)
	if len(body) != cloudTokenRandomLength+cloudTokenChecksumLength {
		return false
	}

	randomValue := body[:cloudTokenRandomLength]
	checksum := body[cloudTokenRandomLength:]

	for _, value := range randomValue + checksum {
		if !strings.ContainsRune(base62Alphabet, value) {
			return false
		}
	}

	expected := encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue)))

	return checksum == expected
}

func cloudManualAccessTokenFormat(baseURL uri.URI) (string, string) {
	parsed, err := url.Parse(baseURL.String())
	if err != nil {
		return "", ""
	}

	switch strings.ToLower(parsed.Hostname()) {
	case cloudUSHostname:
		return cloudUSAPITokenPrefix, cloudUSSecretScanningType
	case cloudEUHostname:
		return cloudEUAPITokenPrefix, cloudEUSecretScanningType
	default:
		return "", ""
	}
}

func encodeBase62Checksum(value uint32) string {
	encoded := [cloudTokenChecksumLength]byte{}

	for i := len(encoded) - 1; i >= 0; i-- {
		encoded[i] = base62Alphabet[value%uint32(len(base62Alphabet))]
		value /= uint32(len(base62Alphabet))
	}

	return string(encoded[:])
}
