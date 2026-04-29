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

type CookiePatternMatchType string

const (
	CookiePatternMatchTypeExact  CookiePatternMatchType = "EXACT"
	CookiePatternMatchTypePrefix CookiePatternMatchType = "PREFIX"
)

func CookiePatternMatchTypes() []CookiePatternMatchType {
	return []CookiePatternMatchType{
		CookiePatternMatchTypeExact,
		CookiePatternMatchTypePrefix,
	}
}

func (m CookiePatternMatchType) String() string {
	return string(m)
}

func (m *CookiePatternMatchType) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for CookiePatternMatchType: %T", value)
	}

	switch CookiePatternMatchType(v) {
	case CookiePatternMatchTypeExact:
		*m = CookiePatternMatchTypeExact
	case CookiePatternMatchTypePrefix:
		*m = CookiePatternMatchTypePrefix
	default:
		return fmt.Errorf("invalid CookiePatternMatchType value: %q", v)
	}
	return nil
}

func (m CookiePatternMatchType) Value() (driver.Value, error) {
	switch m {
	case CookiePatternMatchTypeExact,
		CookiePatternMatchTypePrefix:
		return string(m), nil
	default:
		return nil, fmt.Errorf("invalid CookiePatternMatchType: %s", m)
	}
}
