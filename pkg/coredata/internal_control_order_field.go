// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type (
	InternalControlOrderField string
)

const (
	InternalControlOrderFieldCreatedAt InternalControlOrderField = "CREATED_AT"
	InternalControlOrderFieldName      InternalControlOrderField = "NAME"
	InternalControlOrderFieldNextDue   InternalControlOrderField = "NEXT_DUE"

	// Cursor keys always store empty due dates as the latest instant, so
	// ascending order needs no mapping. Descending order maps them to the
	// earliest instant, on both the column and the cursor value, so empty
	// due dates sort last in both directions.
	internalControlNextDueEmpty    = "9999-12-31 23:59:59+00"
	internalControlNextDueEarliest = "0001-01-01 00:00:00+00"
	internalControlNextDueExpr     = "COALESCE(LEAST(next_evidence_due, next_test_due), '" + internalControlNextDueEmpty + "'::timestamptz)"
	internalControlNextDueDescExpr = "COALESCE(LEAST(next_evidence_due, next_test_due), '" + internalControlNextDueEarliest + "'::timestamptz)"
)

var (
	_ page.OrderField            = InternalControlOrderField("")
	_ page.DirectionalOrderField = InternalControlOrderField("")
	_ fmt.Stringer               = InternalControlOrderField("")
	_ encoding.TextMarshaler     = InternalControlOrderField("")
	_ encoding.TextUnmarshaler   = (*InternalControlOrderField)(nil)
)

func InternalControlOrderFields() []InternalControlOrderField {
	return []InternalControlOrderField{
		InternalControlOrderFieldCreatedAt,
		InternalControlOrderFieldName,
		InternalControlOrderFieldNextDue,
	}
}

func (v InternalControlOrderField) IsValid() bool {
	return isValidOrderField(v, InternalControlOrderFields())
}

func (v InternalControlOrderField) String() string {
	return string(v)
}

func (v InternalControlOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InternalControlOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, InternalControlOrderFields())
}

func (p InternalControlOrderField) Column() string {
	if p == InternalControlOrderFieldNextDue {
		return internalControlNextDueExpr
	}

	return string(p)
}

func (p InternalControlOrderField) DirectionalColumn(direction page.OrderDirection) string {
	if p == InternalControlOrderFieldNextDue && direction == page.OrderDirectionDesc {
		return internalControlNextDueDescExpr
	}

	return p.Column()
}

func (p InternalControlOrderField) DirectionalCursorValue(direction page.OrderDirection, placeholder string) string {
	if p == InternalControlOrderFieldNextDue && direction == page.OrderDirectionDesc {
		return "COALESCE(NULLIF(" + placeholder + "::timestamptz, '" + internalControlNextDueEmpty + "'::timestamptz), '" + internalControlNextDueEarliest + "'::timestamptz)"
	}

	return placeholder
}
