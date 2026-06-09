// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import (
	"encoding"
	"fmt"
)

type (
	OAuth2GrantType  string
	OAuth2GrantTypes []OAuth2GrantType
)

const (
	OAuth2GrantTypeAuthorizationCode OAuth2GrantType = "authorization_code"
	OAuth2GrantTypeRefreshToken      OAuth2GrantType = "refresh_token"
	OAuth2GrantTypeDeviceCode        OAuth2GrantType = "urn:ietf:params:oauth:grant-type:device_code"
)

var (
	_ fmt.Stringer             = OAuth2GrantType("")
	_ encoding.TextMarshaler   = OAuth2GrantType("")
	_ encoding.TextUnmarshaler = (*OAuth2GrantType)(nil)
)

func (v OAuth2GrantType) IsValid() bool {
	switch v {
	case
		OAuth2GrantTypeAuthorizationCode,
		OAuth2GrantTypeRefreshToken,
		OAuth2GrantTypeDeviceCode:
		return true
	}

	return false
}

func (v OAuth2GrantType) String() string {
	return string(v)
}

func (v OAuth2GrantType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2GrantType) UnmarshalText(text []byte) error {
	val := OAuth2GrantType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2GrantType value: %q", string(text))
	}

	*v = val

	return nil
}
