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

type CustomDomainOrderField string

const (
	CustomDomainOrderFieldCreatedAt CustomDomainOrderField = "CREATED_AT"
	CustomDomainOrderFieldDomain    CustomDomainOrderField = "DOMAIN"
	CustomDomainOrderFieldUpdatedAt CustomDomainOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = CustomDomainOrderField("")
	_ fmt.Stringer             = CustomDomainOrderField("")
	_ encoding.TextMarshaler   = CustomDomainOrderField("")
	_ encoding.TextUnmarshaler = (*CustomDomainOrderField)(nil)
)

func CustomDomainOrderFields() []CustomDomainOrderField {
	return []CustomDomainOrderField{
		CustomDomainOrderFieldCreatedAt,
		CustomDomainOrderFieldDomain,
		CustomDomainOrderFieldUpdatedAt,
	}
}

func (v CustomDomainOrderField) IsValid() bool {
	switch v {
	case
		CustomDomainOrderFieldCreatedAt,
		CustomDomainOrderFieldDomain,
		CustomDomainOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v CustomDomainOrderField) String() string {
	return string(v)
}

func (v CustomDomainOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CustomDomainOrderField) UnmarshalText(text []byte) error {
	val := CustomDomainOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CustomDomainOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f CustomDomainOrderField) Column() string {
	switch f {
	case CustomDomainOrderFieldCreatedAt:
		return "created_at"
	case CustomDomainOrderFieldDomain:
		return "domain"
	case CustomDomainOrderFieldUpdatedAt:
		return "updated_at"
	default:
		panic(fmt.Sprintf("unsupported order by: %s", f))
	}
}
