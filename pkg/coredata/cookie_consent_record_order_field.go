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

	"go.probo.inc/probo/pkg/page"
)

type CookieConsentRecordOrderField string

const (
	CookieConsentRecordOrderFieldCreatedAt CookieConsentRecordOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = CookieConsentRecordOrderField("")
	_ fmt.Stringer             = CookieConsentRecordOrderField("")
	_ encoding.TextMarshaler   = CookieConsentRecordOrderField("")
	_ encoding.TextUnmarshaler = (*CookieConsentRecordOrderField)(nil)
)

func CookieConsentRecordOrderFields() []CookieConsentRecordOrderField {
	return []CookieConsentRecordOrderField{
		CookieConsentRecordOrderFieldCreatedAt,
	}
}

func (v CookieConsentRecordOrderField) IsValid() bool {
	switch v {
	case
		CookieConsentRecordOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v CookieConsentRecordOrderField) String() string {
	return string(v)
}

func (v CookieConsentRecordOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieConsentRecordOrderField) UnmarshalText(text []byte) error {
	val := CookieConsentRecordOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieConsentRecordOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p CookieConsentRecordOrderField) Column() string {
	switch p {
	case CookieConsentRecordOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
