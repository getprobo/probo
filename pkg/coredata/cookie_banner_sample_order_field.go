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

type CookieBannerSampleOrderField string

const (
	CookieBannerSampleOrderFieldCreatedAt CookieBannerSampleOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = CookieBannerSampleOrderField("")
	_ fmt.Stringer             = CookieBannerSampleOrderField("")
	_ encoding.TextMarshaler   = CookieBannerSampleOrderField("")
	_ encoding.TextUnmarshaler = (*CookieBannerSampleOrderField)(nil)
)

func CookieBannerSampleOrderFields() []CookieBannerSampleOrderField {
	return []CookieBannerSampleOrderField{
		CookieBannerSampleOrderFieldCreatedAt,
	}
}

func (v CookieBannerSampleOrderField) IsValid() bool {
	return isValidOrderField(v, CookieBannerSampleOrderFields())
}

func (v CookieBannerSampleOrderField) String() string {
	return string(v)
}

func (v CookieBannerSampleOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerSampleOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, CookieBannerSampleOrderFields())
}

func (p CookieBannerSampleOrderField) Column() string {
	switch p {
	case CookieBannerSampleOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
