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

type (
	SessionOrderField string
)

const (
	SessionOrderFieldCreatedAt SessionOrderField = "CREATED_AT"
	SessionOrderFieldExpiredAt SessionOrderField = "EXPIRED_AT"
	SessionOrderFieldUpdatedAt SessionOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = SessionOrderField("")
	_ fmt.Stringer             = SessionOrderField("")
	_ encoding.TextMarshaler   = SessionOrderField("")
	_ encoding.TextUnmarshaler = (*SessionOrderField)(nil)
)

func SessionOrderFields() []SessionOrderField {
	return []SessionOrderField{
		SessionOrderFieldCreatedAt,
		SessionOrderFieldExpiredAt,
		SessionOrderFieldUpdatedAt,
	}
}

func (v SessionOrderField) IsValid() bool {
	switch v {
	case
		SessionOrderFieldCreatedAt,
		SessionOrderFieldExpiredAt,
		SessionOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v SessionOrderField) String() string {
	return string(v)
}

func (v SessionOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SessionOrderField) UnmarshalText(text []byte) error {
	val := SessionOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SessionOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SessionOrderField) Column() string {
	switch p {
	case SessionOrderFieldCreatedAt:
		return "created_at"
	case SessionOrderFieldExpiredAt:
		return "expired_at"
	case SessionOrderFieldUpdatedAt:
		return "updated_at"
	}

	return string(p)
}
