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

type ServiceAccountOrderField string

const (
	ServiceAccountOrderFieldCreatedAt ServiceAccountOrderField = "CREATED_AT"
	ServiceAccountOrderFieldName      ServiceAccountOrderField = "NAME"
)

var (
	_ page.OrderField          = ServiceAccountOrderField("")
	_ fmt.Stringer             = ServiceAccountOrderField("")
	_ encoding.TextMarshaler   = ServiceAccountOrderField("")
	_ encoding.TextUnmarshaler = (*ServiceAccountOrderField)(nil)
)

func ServiceAccountOrderFields() []ServiceAccountOrderField {
	return []ServiceAccountOrderField{
		ServiceAccountOrderFieldCreatedAt,
		ServiceAccountOrderFieldName,
	}
}

func (v ServiceAccountOrderField) IsValid() bool {
	return isValidOrderField(v, ServiceAccountOrderFields())
}

func (v ServiceAccountOrderField) String() string {
	return string(v)
}

func (v ServiceAccountOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ServiceAccountOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, ServiceAccountOrderFields())
}

func (f ServiceAccountOrderField) Column() string {
	switch f {
	case ServiceAccountOrderFieldCreatedAt:
		return "created_at"
	case ServiceAccountOrderFieldName:
		return "name"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
