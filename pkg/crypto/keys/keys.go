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
