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
	FrameworkOrderField string
)

const (
	FrameworkOrderFieldCreatedAt FrameworkOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = FrameworkOrderField("")
	_ fmt.Stringer             = FrameworkOrderField("")
	_ encoding.TextMarshaler   = FrameworkOrderField("")
	_ encoding.TextUnmarshaler = (*FrameworkOrderField)(nil)
)

func FrameworkOrderFields() []FrameworkOrderField {
	return []FrameworkOrderField{
		FrameworkOrderFieldCreatedAt,
	}
}

func (v FrameworkOrderField) IsValid() bool {
	switch v {
	case
		FrameworkOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v FrameworkOrderField) String() string {
	return string(v)
}

func (v FrameworkOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *FrameworkOrderField) UnmarshalText(text []byte) error {
	val := FrameworkOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid FrameworkOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p FrameworkOrderField) Column() string {
	return string(p)
}
