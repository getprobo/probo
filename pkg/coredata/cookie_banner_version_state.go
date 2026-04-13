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

type CookieBannerVersionState string

const (
	CookieBannerVersionStateDraft     CookieBannerVersionState = "DRAFT"
	CookieBannerVersionStatePublished CookieBannerVersionState = "PUBLISHED"
)

func CookieBannerVersionStates() []CookieBannerVersionState {
	return []CookieBannerVersionState{
		CookieBannerVersionStateDraft,
		CookieBannerVersionStatePublished,
	}
}

func (s CookieBannerVersionState) String() string {
	return string(s)
}

func (s *CookieBannerVersionState) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for CookieBannerVersionState: %T", value)
	}

	switch CookieBannerVersionState(v) {
	case CookieBannerVersionStateDraft:
		*s = CookieBannerVersionStateDraft
	case CookieBannerVersionStatePublished:
		*s = CookieBannerVersionStatePublished
	default:
		return fmt.Errorf("invalid CookieBannerVersionState value: %q", v)
	}
	return nil
}

func (s CookieBannerVersionState) Value() (driver.Value, error) {
	switch s {
	case CookieBannerVersionStateDraft,
		CookieBannerVersionStatePublished:
		return string(s), nil
	default:
		return nil, fmt.Errorf("invalid CookieBannerVersionState: %s", s)
	}
}
