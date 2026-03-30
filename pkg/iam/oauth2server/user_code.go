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

package oauth2server

import (
	"crypto/rand"
	"math/big"

	"go.probo.inc/probo/pkg/coredata"
)

// userCodeAlphabet excludes ambiguous characters: 0/O, 1/I/L.
const userCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// GenerateUserCode generates a raw 8-character user code for the device flow.
// Characters are drawn from an unambiguous alphabet (no 0/O/1/I/L).
func GenerateUserCode() (coredata.OAuth2UserCode, error) {
	code := make([]byte, 8)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(userCodeAlphabet))))
		if err != nil {
			return "", err
		}

		code[i] = userCodeAlphabet[n.Int64()]
	}

	return coredata.OAuth2UserCode(code), nil
}
