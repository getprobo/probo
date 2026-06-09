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

type OAuth2ConsentOrderField string

const (
	OAuth2ConsentOrderFieldCreatedAt OAuth2ConsentOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = OAuth2ConsentOrderField("")
	_ fmt.Stringer             = OAuth2ConsentOrderField("")
	_ encoding.TextMarshaler   = OAuth2ConsentOrderField("")
	_ encoding.TextUnmarshaler = (*OAuth2ConsentOrderField)(nil)
)

func OAuth2ConsentOrderFields() []OAuth2ConsentOrderField {
	return []OAuth2ConsentOrderField{
		OAuth2ConsentOrderFieldCreatedAt,
	}
}

func (v OAuth2ConsentOrderField) IsValid() bool {
	switch v {
	case
		OAuth2ConsentOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v OAuth2ConsentOrderField) String() string {
	return string(v)
}

func (v OAuth2ConsentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ConsentOrderField) UnmarshalText(text []byte) error {
	val := OAuth2ConsentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ConsentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f OAuth2ConsentOrderField) Column() string {
	switch f {
	case OAuth2ConsentOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
