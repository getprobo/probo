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

type CookieBannerOrderField string

const (
	CookieBannerOrderFieldCreatedAt CookieBannerOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = CookieBannerOrderField("")
	_ fmt.Stringer             = CookieBannerOrderField("")
	_ encoding.TextMarshaler   = CookieBannerOrderField("")
	_ encoding.TextUnmarshaler = (*CookieBannerOrderField)(nil)
)

func CookieBannerOrderFields() []CookieBannerOrderField {
	return []CookieBannerOrderField{
		CookieBannerOrderFieldCreatedAt,
	}
}

func (v CookieBannerOrderField) IsValid() bool {
	switch v {
	case
		CookieBannerOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v CookieBannerOrderField) String() string {
	return string(v)
}

func (v CookieBannerOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerOrderField) UnmarshalText(text []byte) error {
	val := CookieBannerOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieBannerOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p CookieBannerOrderField) Column() string {
	switch p {
	case CookieBannerOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
