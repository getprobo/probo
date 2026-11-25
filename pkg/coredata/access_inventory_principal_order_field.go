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
	AccessInventoryPrincipalOrderField string
)

const (
	AccessInventoryPrincipalOrderFieldCreatedAt AccessInventoryPrincipalOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AccessInventoryPrincipalOrderField("")
	_ fmt.Stringer             = AccessInventoryPrincipalOrderField("")
	_ encoding.TextMarshaler   = AccessInventoryPrincipalOrderField("")
	_ encoding.TextUnmarshaler = (*AccessInventoryPrincipalOrderField)(nil)
)

func AccessInventoryPrincipalOrderFields() []AccessInventoryPrincipalOrderField {
	return []AccessInventoryPrincipalOrderField{
		AccessInventoryPrincipalOrderFieldCreatedAt,
	}
}

func (v AccessInventoryPrincipalOrderField) IsValid() bool {
	return isValidOrderField(v, AccessInventoryPrincipalOrderFields())
}

func (v AccessInventoryPrincipalOrderField) String() string {
	return string(v)
}

func (v AccessInventoryPrincipalOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessInventoryPrincipalOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, AccessInventoryPrincipalOrderFields())
}

func (p AccessInventoryPrincipalOrderField) Column() string {
	switch p {
	case AccessInventoryPrincipalOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
