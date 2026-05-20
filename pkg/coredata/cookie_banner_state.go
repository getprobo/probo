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
