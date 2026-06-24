// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package keys

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

type Type string

const (
	// TypeEC256 represents ECDSA with P-256 curve
	TypeEC256 Type = "EC256"
	// TypeEC384 represents ECDSA with P-384 curve
	TypeEC384 Type = "EC384"
	// TypeRSA2048 represents RSA with 2048-bit key
	TypeRSA2048 Type = "RSA2048"
	// TypeRSA4096 represents RSA with 4096-bit key
	TypeRSA4096 Type = "RSA4096"
)

// Generate creates a new private key of the specified type
func Generate(keyType Type) (crypto.Signer, error) {
	switch keyType {
	case TypeEC256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case TypeEC384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case TypeRSA2048:
		return rsa.GenerateKey(rand.Reader, 2048)
	case TypeRSA4096:
		return rsa.GenerateKey(rand.Reader, 4096)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}
