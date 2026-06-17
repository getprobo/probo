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

	"go.probo.inc/probo/pkg/page"
)

type OAuth2AccessTokenOrderField string

const (
	OAuth2AccessTokenOrderFieldCreatedAt OAuth2AccessTokenOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = OAuth2AccessTokenOrderField("")
	_ fmt.Stringer             = OAuth2AccessTokenOrderField("")
	_ encoding.TextMarshaler   = OAuth2AccessTokenOrderField("")
	_ encoding.TextUnmarshaler = (*OAuth2AccessTokenOrderField)(nil)
)

func OAuth2AccessTokenOrderFields() []OAuth2AccessTokenOrderField {
	return []OAuth2AccessTokenOrderField{
		OAuth2AccessTokenOrderFieldCreatedAt,
	}
}

func (v OAuth2AccessTokenOrderField) IsValid() bool {
	switch v {
	case OAuth2AccessTokenOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v OAuth2AccessTokenOrderField) String() string {
	return string(v)
}

func (v OAuth2AccessTokenOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2AccessTokenOrderField) UnmarshalText(text []byte) error {
	val := OAuth2AccessTokenOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2AccessTokenOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f OAuth2AccessTokenOrderField) Column() string {
	switch f {
	case OAuth2AccessTokenOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
