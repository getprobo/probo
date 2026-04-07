// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package rand

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// HexString returns a hex-encoded cryptographically random string.
// The output is 2*byteLen characters long.
func HexString(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cannot generate random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// MustHexString is like HexString but panics if the system entropy source is
// unavailable.
func MustHexString(byteLen int) string {
	s, err := HexString(byteLen)
	if err != nil {
		panic("rand: crypto/rand is unavailable: " + err.Error())
	}

	return s
}

// StringFromAlphabet returns a random string of length n, where each character
// is drawn uniformly from alphabet using crypto/rand.
func StringFromAlphabet(alphabet string, n int) (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	buf := make([]byte, n)

	for i := range buf {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("cannot generate random bytes: %w", err)
		}

		buf[i] = alphabet[idx.Int64()]
	}

	return string(buf), nil
}

// MustStringFromAlphabet is like StringFromAlphabet but panics if the system
// entropy source is unavailable.
func MustStringFromAlphabet(alphabet string, n int) string {
	s, err := StringFromAlphabet(alphabet, n)
	if err != nil {
		panic("rand: crypto/rand is unavailable: " + err.Error())
	}

	return s
}
