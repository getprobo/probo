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
