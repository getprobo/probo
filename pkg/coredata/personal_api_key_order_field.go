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
	PersonalAPIKeyOrderField string
)

const (
	PersonalAPIKeyOrderFieldCreatedAt PersonalAPIKeyOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = PersonalAPIKeyOrderField("")
	_ fmt.Stringer             = PersonalAPIKeyOrderField("")
	_ encoding.TextMarshaler   = PersonalAPIKeyOrderField("")
	_ encoding.TextUnmarshaler = (*PersonalAPIKeyOrderField)(nil)
)

func PersonalAPIKeyOrderFields() []PersonalAPIKeyOrderField {
	return []PersonalAPIKeyOrderField{
		PersonalAPIKeyOrderFieldCreatedAt,
	}
}

func (v PersonalAPIKeyOrderField) IsValid() bool {
	switch v {
	case
		PersonalAPIKeyOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v PersonalAPIKeyOrderField) String() string {
	return string(v)
}

func (v PersonalAPIKeyOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *PersonalAPIKeyOrderField) UnmarshalText(text []byte) error {
	val := PersonalAPIKeyOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid PersonalAPIKeyOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p PersonalAPIKeyOrderField) Column() string {
	return string(p)
}
