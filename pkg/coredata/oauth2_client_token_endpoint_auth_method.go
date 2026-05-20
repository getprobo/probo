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

import (
	"encoding"
	"fmt"
)

type OAuth2ClientTokenEndpointAuthMethod string

const (
	OAuth2ClientTokenEndpointAuthMethodClientSecretBasic OAuth2ClientTokenEndpointAuthMethod = "client_secret_basic"
	OAuth2ClientTokenEndpointAuthMethodClientSecretPost  OAuth2ClientTokenEndpointAuthMethod = "client_secret_post"
	OAuth2ClientTokenEndpointAuthMethodNone              OAuth2ClientTokenEndpointAuthMethod = "none"
)

var (
	_ fmt.Stringer             = OAuth2ClientTokenEndpointAuthMethod("")
	_ encoding.TextMarshaler   = OAuth2ClientTokenEndpointAuthMethod("")
	_ encoding.TextUnmarshaler = (*OAuth2ClientTokenEndpointAuthMethod)(nil)
)

func OAuth2ClientTokenEndpointAuthMethods() []OAuth2ClientTokenEndpointAuthMethod {
	return []OAuth2ClientTokenEndpointAuthMethod{
		OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
		OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
		OAuth2ClientTokenEndpointAuthMethodNone,
	}
}

func (v OAuth2ClientTokenEndpointAuthMethod) IsValid() bool {
	switch v {
	case
		OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
		OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
		OAuth2ClientTokenEndpointAuthMethodNone:
		return true
	}

	return false
}

func (v OAuth2ClientTokenEndpointAuthMethod) String() string {
	return string(v)
}

func (v OAuth2ClientTokenEndpointAuthMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ClientTokenEndpointAuthMethod) UnmarshalText(text []byte) error {
	val := OAuth2ClientTokenEndpointAuthMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ClientTokenEndpointAuthMethod value: %q", string(text))
	}

	*v = val

	return nil
}
