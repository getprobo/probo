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
	MeasureOrderField string
)

const (
	MeasureOrderFieldCreatedAt MeasureOrderField = "CREATED_AT"
	MeasureOrderFieldName      MeasureOrderField = "NAME"
)

var (
	_ page.OrderField          = MeasureOrderField("")
	_ fmt.Stringer             = MeasureOrderField("")
	_ encoding.TextMarshaler   = MeasureOrderField("")
	_ encoding.TextUnmarshaler = (*MeasureOrderField)(nil)
)

func MeasureOrderFields() []MeasureOrderField {
	return []MeasureOrderField{
		MeasureOrderFieldCreatedAt,
		MeasureOrderFieldName,
	}
}

func (v MeasureOrderField) IsValid() bool {
	switch v {
	case
		MeasureOrderFieldCreatedAt,
		MeasureOrderFieldName:
		return true
	}

	return false
}

func (v MeasureOrderField) String() string {
	return string(v)
}

func (v MeasureOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MeasureOrderField) UnmarshalText(text []byte) error {
	val := MeasureOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MeasureOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p MeasureOrderField) Column() string {
	return string(p)
}
