// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type CookieSource string

const (
	CookieSourceScript      CookieSource = "SCRIPT"
	CookieSourcePreExisting CookieSource = "PRE_EXISTING"
)

func CookieSources() []CookieSource {
	return []CookieSource{
		CookieSourceScript,
		CookieSourcePreExisting,
	}
}

func (s CookieSource) String() string {
	return string(s)
}

func (s *CookieSource) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for CookieSource: %T", value)
	}

	switch CookieSource(v) {
	case CookieSourceScript:
		*s = CookieSourceScript
	case CookieSourcePreExisting:
		*s = CookieSourcePreExisting
	default:
		return fmt.Errorf("invalid CookieSource value: %q", v)
	}
	return nil
}

func (s CookieSource) Value() (driver.Value, error) {
	switch s {
	case CookieSourceScript,
		CookieSourcePreExisting:
		return string(s), nil
	default:
		return nil, fmt.Errorf("invalid CookieSource: %s", s)
	}
}
