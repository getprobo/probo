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
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	"go.probo.inc/probo/pkg/coredata"
)

func ValidateCodeChallenge(verifier, challenge string, method coredata.OAuth2CodeChallengeMethod) bool {
	if verifier == "" || challenge == "" {
		return false
	}

	switch method {
	case coredata.OAuth2CodeChallengeMethodS256:
		return validateS256(verifier, challenge)
	default:
		return false
	}
}

func validateS256(verifier, challenge string) bool {
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])

	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
