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

type OIDCProvider string

const (
	OIDCProviderGoogle    OIDCProvider = "GOOGLE"
	OIDCProviderMicrosoft OIDCProvider = "MICROSOFT"
)

var (
	_ fmt.Stringer             = OIDCProvider("")
	_ encoding.TextMarshaler   = OIDCProvider("")
	_ encoding.TextUnmarshaler = (*OIDCProvider)(nil)
)

func OIDCProviders() []OIDCProvider {
	return []OIDCProvider{
		OIDCProviderGoogle,
		OIDCProviderMicrosoft,
	}
}

func (v OIDCProvider) IsValid() bool {
	switch v {
	case
		OIDCProviderGoogle,
		OIDCProviderMicrosoft:
		return true
	}

	return false
}

func (v OIDCProvider) String() string {
	return string(v)
}

func (v OIDCProvider) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OIDCProvider) UnmarshalText(text []byte) error {
	val := OIDCProvider(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OIDCProvider value: %q", string(text))
	}

	*v = val

	return nil
}
