// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type UserAuthMethod string

const (
	UserAuthMethodPassword UserAuthMethod = "PASSWORD"
	UserAuthMethodSAML     UserAuthMethod = "SAML"
)

var (
	_ fmt.Stringer             = UserAuthMethod("")
	_ encoding.TextMarshaler   = UserAuthMethod("")
	_ encoding.TextUnmarshaler = (*UserAuthMethod)(nil)
)

func UserAuthMethods() []UserAuthMethod {
	return []UserAuthMethod{
		UserAuthMethodPassword,
		UserAuthMethodSAML,
	}
}

func (v UserAuthMethod) IsValid() bool {
	switch v {
	case
		UserAuthMethodPassword,
		UserAuthMethodSAML:
		return true
	}

	return false
}

func (v UserAuthMethod) String() string {
	return string(v)
}

func (v UserAuthMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *UserAuthMethod) UnmarshalText(text []byte) error {
	val := UserAuthMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid UserAuthMethod value: %q", string(text))
	}

	*v = val

	return nil
}
