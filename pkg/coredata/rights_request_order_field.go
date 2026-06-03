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

type RightsRequestOrderField string

const (
	RightsRequestOrderFieldCreatedAt RightsRequestOrderField = "CREATED_AT"
	RightsRequestOrderFieldDeadline  RightsRequestOrderField = "DEADLINE"
	RightsRequestOrderFieldState     RightsRequestOrderField = "STATE"
	RightsRequestOrderFieldType      RightsRequestOrderField = "TYPE"
)

var (
	_ page.OrderField          = RightsRequestOrderField("")
	_ fmt.Stringer             = RightsRequestOrderField("")
	_ encoding.TextMarshaler   = RightsRequestOrderField("")
	_ encoding.TextUnmarshaler = (*RightsRequestOrderField)(nil)
)

func RightsRequestOrderFields() []RightsRequestOrderField {
	return []RightsRequestOrderField{
		RightsRequestOrderFieldCreatedAt,
		RightsRequestOrderFieldDeadline,
		RightsRequestOrderFieldState,
		RightsRequestOrderFieldType,
	}
}

func (v RightsRequestOrderField) IsValid() bool {
	switch v {
	case
		RightsRequestOrderFieldCreatedAt,
		RightsRequestOrderFieldDeadline,
		RightsRequestOrderFieldState,
		RightsRequestOrderFieldType:
		return true
	}

	return false
}

func (v RightsRequestOrderField) String() string {
	return string(v)
}

func (v RightsRequestOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RightsRequestOrderField) UnmarshalText(text []byte) error {
	val := RightsRequestOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RightsRequestOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RightsRequestOrderField) Column() string {
	return string(p)
}
