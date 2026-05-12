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

type CookieConsentMode string

const (
	CookieConsentModeOptIn  CookieConsentMode = "OPT_IN"
	CookieConsentModeOptOut CookieConsentMode = "OPT_OUT"
)

func CookieConsentModes() []CookieConsentMode {
	return []CookieConsentMode{
		CookieConsentModeOptIn,
		CookieConsentModeOptOut,
	}
}

func (m CookieConsentMode) String() string {
	return string(m)
}

func (m *CookieConsentMode) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for CookieConsentMode: %T", value)
	}

	switch CookieConsentMode(v) {
	case CookieConsentModeOptIn:
		*m = CookieConsentModeOptIn
	case CookieConsentModeOptOut:
		*m = CookieConsentModeOptOut
	default:
		return fmt.Errorf("invalid CookieConsentMode value: %q", v)
	}

	return nil
}

func (m CookieConsentMode) Value() (driver.Value, error) {
	switch m {
	case CookieConsentModeOptIn,
		CookieConsentModeOptOut:
		return string(m), nil
	default:
		return nil, fmt.Errorf("invalid CookieConsentMode: %s", m)
	}
}
