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
)

type CookieBannerState string

const (
	CookieBannerStateActive   CookieBannerState = "ACTIVE"
	CookieBannerStateInactive CookieBannerState = "INACTIVE"
)

var (
	_ fmt.Stringer             = CookieBannerState("")
	_ encoding.TextMarshaler   = CookieBannerState("")
	_ encoding.TextUnmarshaler = (*CookieBannerState)(nil)
)

func CookieBannerStates() []CookieBannerState {
	return []CookieBannerState{
		CookieBannerStateActive,
		CookieBannerStateInactive,
	}
}

func (v CookieBannerState) IsValid() bool {
	switch v {
	case
		CookieBannerStateActive,
		CookieBannerStateInactive:
		return true
	}

	return false
}

func (v CookieBannerState) String() string {
	return string(v)
}

func (v CookieBannerState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerState) UnmarshalText(text []byte) error {
	val := CookieBannerState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieBannerState value: %q", string(text))
	}

	*v = val

	return nil
}
