// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	AccessInventoryResourceOrderField string
)

const (
	AccessInventoryResourceOrderFieldCreatedAt AccessInventoryResourceOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AccessInventoryResourceOrderField("")
	_ fmt.Stringer             = AccessInventoryResourceOrderField("")
	_ encoding.TextMarshaler   = AccessInventoryResourceOrderField("")
	_ encoding.TextUnmarshaler = (*AccessInventoryResourceOrderField)(nil)
)

func AccessInventoryResourceOrderFields() []AccessInventoryResourceOrderField {
	return []AccessInventoryResourceOrderField{
		AccessInventoryResourceOrderFieldCreatedAt,
	}
}

func (v AccessInventoryResourceOrderField) IsValid() bool {
	return isValidOrderField(v, AccessInventoryResourceOrderFields())
}

func (v AccessInventoryResourceOrderField) String() string {
	return string(v)
}

func (v AccessInventoryResourceOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessInventoryResourceOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, AccessInventoryResourceOrderFields())
}

func (p AccessInventoryResourceOrderField) Column() string {
	switch p {
	case AccessInventoryResourceOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
