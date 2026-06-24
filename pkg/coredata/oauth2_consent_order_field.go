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

type OAuth2ConsentOrderField string

const (
	OAuth2ConsentOrderFieldCreatedAt OAuth2ConsentOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = OAuth2ConsentOrderField("")
	_ fmt.Stringer             = OAuth2ConsentOrderField("")
	_ encoding.TextMarshaler   = OAuth2ConsentOrderField("")
	_ encoding.TextUnmarshaler = (*OAuth2ConsentOrderField)(nil)
)

func OAuth2ConsentOrderFields() []OAuth2ConsentOrderField {
	return []OAuth2ConsentOrderField{
		OAuth2ConsentOrderFieldCreatedAt,
	}
}

func (v OAuth2ConsentOrderField) IsValid() bool {
	switch v {
	case
		OAuth2ConsentOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v OAuth2ConsentOrderField) String() string {
	return string(v)
}

func (v OAuth2ConsentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ConsentOrderField) UnmarshalText(text []byte) error {
	val := OAuth2ConsentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ConsentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f OAuth2ConsentOrderField) Column() string {
	switch f {
	case OAuth2ConsentOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
