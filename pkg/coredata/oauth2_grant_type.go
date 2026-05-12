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

type (
	OAuth2GrantType  string
	OAuth2GrantTypes []OAuth2GrantType
)

const (
	OAuth2GrantTypeAuthorizationCode OAuth2GrantType = "authorization_code"
	OAuth2GrantTypeRefreshToken      OAuth2GrantType = "refresh_token"
	OAuth2GrantTypeDeviceCode        OAuth2GrantType = "urn:ietf:params:oauth:grant-type:device_code"
)

func (g OAuth2GrantType) IsValid() bool {
	switch g {
	case OAuth2GrantTypeAuthorizationCode,
		OAuth2GrantTypeRefreshToken,
		OAuth2GrantTypeDeviceCode:
		return true
	}

	return false
}

func (g OAuth2GrantType) String() string { return string(g) }

func (g *OAuth2GrantType) UnmarshalText(text []byte) error {
	*g = OAuth2GrantType(text)
	if !g.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2GrantType", string(text))
	}

	return nil
}

func (g OAuth2GrantType) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}
