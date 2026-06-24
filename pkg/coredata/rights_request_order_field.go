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
