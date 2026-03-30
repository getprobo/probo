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

package coredata

import "fmt"

type OAuth2Claim string

const (
	OAuth2ClaimIssuer        OAuth2Claim = "iss"
	OAuth2ClaimSubject       OAuth2Claim = "sub"
	OAuth2ClaimAudience      OAuth2Claim = "aud"
	OAuth2ClaimExpiration    OAuth2Claim = "exp"
	OAuth2ClaimIssuedAt      OAuth2Claim = "iat"
	OAuth2ClaimAuthTime      OAuth2Claim = "auth_time"
	OAuth2ClaimNonce         OAuth2Claim = "nonce"
	OAuth2ClaimAtHash        OAuth2Claim = "at_hash"
	OAuth2ClaimEmail         OAuth2Claim = "email"
	OAuth2ClaimEmailVerified OAuth2Claim = "email_verified"
	OAuth2ClaimName          OAuth2Claim = "name"
)

func (c OAuth2Claim) IsValid() bool {
	switch c {
	case OAuth2ClaimIssuer,
		OAuth2ClaimSubject,
		OAuth2ClaimAudience,
		OAuth2ClaimExpiration,
		OAuth2ClaimIssuedAt,
		OAuth2ClaimAuthTime,
		OAuth2ClaimNonce,
		OAuth2ClaimAtHash,
		OAuth2ClaimEmail,
		OAuth2ClaimEmailVerified,
		OAuth2ClaimName:
		return true
	}

	return false
}

func (c OAuth2Claim) String() string { return string(c) }

func (c *OAuth2Claim) UnmarshalText(text []byte) error {
	*c = OAuth2Claim(text)
	if !c.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2Claim", string(text))
	}
	return nil
}

func (c OAuth2Claim) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
