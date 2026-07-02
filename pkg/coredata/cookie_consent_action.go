// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
