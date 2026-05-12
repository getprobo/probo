// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
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

func CookieConsentActions() []CookieConsentAction {
	return []CookieConsentAction{
		CookieConsentActionAcceptAll,
		CookieConsentActionRejectAll,
		CookieConsentActionCustomize,
		CookieConsentActionGPC,
	}
}

func (a CookieConsentAction) String() string {
	return string(a)
}

func (a *CookieConsentAction) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for CookieConsentAction: %T", value)
	}

	switch CookieConsentAction(v) {
	case CookieConsentActionAcceptAll:
		*a = CookieConsentActionAcceptAll
	case CookieConsentActionRejectAll:
		*a = CookieConsentActionRejectAll
	case CookieConsentActionCustomize:
		*a = CookieConsentActionCustomize
	case CookieConsentActionGPC:
		*a = CookieConsentActionGPC
	default:
		return fmt.Errorf("invalid CookieConsentAction value: %q", v)
	}

	return nil
}

func (a CookieConsentAction) Value() (driver.Value, error) {
	switch a {
	case CookieConsentActionAcceptAll,
		CookieConsentActionRejectAll,
		CookieConsentActionCustomize,
		CookieConsentActionGPC:
		return string(a), nil
	default:
		return nil, fmt.Errorf("invalid CookieConsentAction: %s", a)
	}
}
