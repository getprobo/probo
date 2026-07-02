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
