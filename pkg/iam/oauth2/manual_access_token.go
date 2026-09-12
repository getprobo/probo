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

	"go.probo.inc/probo/pkg/crypto/rand"
)

const (
	base62Alphabet            = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	manualTokenRandomLength   = 32
	manualTokenChecksumLength = 6
	manualAPITokenPrefix      = "prb_a1_"
	secretScanningTokenType   = "probo_api_token"
)

func newManualAccessToken() string {
	randomValue := rand.MustStringFromAlphabet(base62Alphabet, manualTokenRandomLength)

	return manualAPITokenPrefix + randomValue + encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue)))
}

func isValidManualAccessToken(token string) bool {
	if !strings.HasPrefix(token, manualAPITokenPrefix) {
		return false
	}

	body := strings.TrimPrefix(token, manualAPITokenPrefix)
	if len(body) != manualTokenRandomLength+manualTokenChecksumLength {
		return false
	}

	randomValue := body[:manualTokenRandomLength]
	checksum := body[manualTokenRandomLength:]

	for _, value := range randomValue + checksum {
		if !strings.ContainsRune(base62Alphabet, value) {
			return false
		}
	}

	expected := encodeBase62Checksum(crc32.ChecksumIEEE([]byte(randomValue)))

	return checksum == expected
}

func encodeBase62Checksum(value uint32) string {
	encoded := [manualTokenChecksumLength]byte{}

	for i := len(encoded) - 1; i >= 0; i-- {
		encoded[i] = base62Alphabet[value%uint32(len(base62Alphabet))]
		value /= uint32(len(base62Alphabet))
	}

	return string(encoded[:])
}
