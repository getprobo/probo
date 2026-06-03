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

type DatumOrderField string

const (
	DatumOrderFieldCreatedAt          DatumOrderField = "CREATED_AT"
	DatumOrderFieldName               DatumOrderField = "NAME"
	DatumOrderFieldDataClassification DatumOrderField = "DATA_CLASSIFICATION"
)

var (
	_ page.OrderField          = DatumOrderField("")
	_ fmt.Stringer             = DatumOrderField("")
	_ encoding.TextMarshaler   = DatumOrderField("")
	_ encoding.TextUnmarshaler = (*DatumOrderField)(nil)
)

func DatumOrderFields() []DatumOrderField {
	return []DatumOrderField{
		DatumOrderFieldCreatedAt,
		DatumOrderFieldName,
		DatumOrderFieldDataClassification,
	}
}

func (v DatumOrderField) IsValid() bool {
	switch v {
	case
		DatumOrderFieldCreatedAt,
		DatumOrderFieldName,
		DatumOrderFieldDataClassification:
		return true
	}

	return false
}

func (v DatumOrderField) String() string {
	return string(v)
}

func (v DatumOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DatumOrderField) UnmarshalText(text []byte) error {
	val := DatumOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DatumOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p DatumOrderField) Column() string {
	return string(p)
}
