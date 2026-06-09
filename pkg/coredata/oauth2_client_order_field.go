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

type OAuth2ClientOrderField string

const (
	OAuth2ClientOrderFieldCreatedAt OAuth2ClientOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = OAuth2ClientOrderField("")
	_ fmt.Stringer             = OAuth2ClientOrderField("")
	_ encoding.TextMarshaler   = OAuth2ClientOrderField("")
	_ encoding.TextUnmarshaler = (*OAuth2ClientOrderField)(nil)
)

func OAuth2ClientOrderFields() []OAuth2ClientOrderField {
	return []OAuth2ClientOrderField{
		OAuth2ClientOrderFieldCreatedAt,
	}
}

func (v OAuth2ClientOrderField) IsValid() bool {
	switch v {
	case
		OAuth2ClientOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v OAuth2ClientOrderField) String() string {
	return string(v)
}

func (v OAuth2ClientOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ClientOrderField) UnmarshalText(text []byte) error {
	val := OAuth2ClientOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ClientOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f OAuth2ClientOrderField) Column() string {
	switch f {
	case OAuth2ClientOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
