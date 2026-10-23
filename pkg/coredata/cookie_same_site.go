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
)

type CookieSameSite string

const (
	CookieSameSiteStrict CookieSameSite = "STRICT"
	CookieSameSiteLax    CookieSameSite = "LAX"
	CookieSameSiteNone   CookieSameSite = "NONE"
)

var (
	_ fmt.Stringer             = CookieSameSite("")
	_ encoding.TextMarshaler   = CookieSameSite("")
	_ encoding.TextUnmarshaler = (*CookieSameSite)(nil)
)

func CookieSameSites() []CookieSameSite {
	return []CookieSameSite{
		CookieSameSiteStrict,
		CookieSameSiteLax,
		CookieSameSiteNone,
	}
}

func (v CookieSameSite) IsValid() bool {
	switch v {
	case
		CookieSameSiteStrict,
		CookieSameSiteLax,
		CookieSameSiteNone:
		return true
	}

	return false
}

func (v CookieSameSite) String() string {
	return string(v)
}

func (v CookieSameSite) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieSameSite) UnmarshalText(text []byte) error {
	val := CookieSameSite(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieSameSite value: %q", string(text))
	}

	*v = val

	return nil
}
