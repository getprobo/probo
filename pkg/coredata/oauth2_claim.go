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

package coredata

import (
	"encoding"
	"fmt"
)

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

var (
	_ fmt.Stringer             = OAuth2Claim("")
	_ encoding.TextMarshaler   = OAuth2Claim("")
	_ encoding.TextUnmarshaler = (*OAuth2Claim)(nil)
)

func OAuth2Claims() []OAuth2Claim {
	return []OAuth2Claim{
		OAuth2ClaimIssuer,
		OAuth2ClaimSubject,
		OAuth2ClaimAudience,
		OAuth2ClaimExpiration,
		OAuth2ClaimIssuedAt,
		OAuth2ClaimAuthTime,
		OAuth2ClaimNonce,
		OAuth2ClaimAtHash,
		OAuth2ClaimEmail,
		OAuth2ClaimEmailVerified,
		OAuth2ClaimName,
	}
}

func (v OAuth2Claim) IsValid() bool {
	switch v {
	case
		OAuth2ClaimIssuer,
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

func (v OAuth2Claim) String() string {
	return string(v)
}

func (v OAuth2Claim) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2Claim) UnmarshalText(text []byte) error {
	val := OAuth2Claim(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2Claim value: %q", string(text))
	}

	*v = val

	return nil
}
