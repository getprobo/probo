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

type OAuth2ClientOrderField string

const (
	OAuth2ClientOrderFieldCreatedAt OAuth2ClientOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = OAuth2ClientOrderField("")
	_ fmt.Stringer             = OAuth2ClientOrderField("")
	_ encoding.TextMarshaler   = OAuth2ClientOrderField("")
	_ encoding.TextUnmarshaler = (*OAuth2ClientOrderField)(nil)
)

func OAuth2ClientOrderFields() []OAuth2ClientOrderField {
	return []OAuth2ClientOrderField{
		OAuth2ClientOrderFieldCreatedAt,
	}
}

func (v OAuth2ClientOrderField) IsValid() bool {
	switch v {
	case
		OAuth2ClientOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v OAuth2ClientOrderField) String() string {
	return string(v)
}

func (v OAuth2ClientOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2ClientOrderField) UnmarshalText(text []byte) error {
	val := OAuth2ClientOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2ClientOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f OAuth2ClientOrderField) Column() string {
	switch f {
	case OAuth2ClientOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
