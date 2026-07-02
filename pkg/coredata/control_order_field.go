// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	ControlOrderField string
)

const (
	ControlOrderFieldCreatedAt    ControlOrderField = "CREATED_AT"
	ControlOrderFieldSectionTitle ControlOrderField = "SECTION_TITLE"
)

var (
	_ page.OrderField          = ControlOrderField("")
	_ fmt.Stringer             = ControlOrderField("")
	_ encoding.TextMarshaler   = ControlOrderField("")
	_ encoding.TextUnmarshaler = (*ControlOrderField)(nil)
)

func ControlOrderFields() []ControlOrderField {
	return []ControlOrderField{
		ControlOrderFieldCreatedAt,
		ControlOrderFieldSectionTitle,
	}
}

func (v ControlOrderField) IsValid() bool {
	switch v {
	case
		ControlOrderFieldCreatedAt,
		ControlOrderFieldSectionTitle:
		return true
	}

	return false
}

func (v ControlOrderField) String() string {
	return string(v)
}

func (v ControlOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ControlOrderField) UnmarshalText(text []byte) error {
	val := ControlOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ControlOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ControlOrderField) Column() string {
	switch p {
	case ControlOrderFieldCreatedAt:
		return "created_at"
	case ControlOrderFieldSectionTitle:
		return "section_title_sort_key(section_title)"
	default:
		return string(p)
	}
}
