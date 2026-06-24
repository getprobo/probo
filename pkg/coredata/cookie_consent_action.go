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

type CookieConsentAction string

const (
	CookieConsentActionAcceptAll CookieConsentAction = "ACCEPT_ALL"
	CookieConsentActionRejectAll CookieConsentAction = "REJECT_ALL"
	CookieConsentActionCustomize CookieConsentAction = "CUSTOMIZE"
	// Global Privacy Control
	CookieConsentActionGPC CookieConsentAction = "GPC"
)

var (
	_ fmt.Stringer             = CookieConsentAction("")
	_ encoding.TextMarshaler   = CookieConsentAction("")
	_ encoding.TextUnmarshaler = (*CookieConsentAction)(nil)
)

func CookieConsentActions() []CookieConsentAction {
	return []CookieConsentAction{
		CookieConsentActionAcceptAll,
		CookieConsentActionRejectAll,
		CookieConsentActionCustomize,
		CookieConsentActionGPC,
	}
}

func (v CookieConsentAction) IsValid() bool {
	switch v {
	case
		CookieConsentActionAcceptAll,
		CookieConsentActionRejectAll,
		CookieConsentActionCustomize,
		CookieConsentActionGPC:
		return true
	}

	return false
}

func (v CookieConsentAction) String() string {
	return string(v)
}

func (v CookieConsentAction) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieConsentAction) UnmarshalText(text []byte) error {
	val := CookieConsentAction(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieConsentAction value: %q", string(text))
	}

	*v = val

	return nil
}
