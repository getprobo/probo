// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type CookieBannerVersionState string

const (
	CookieBannerVersionStateDraft     CookieBannerVersionState = "DRAFT"
	CookieBannerVersionStatePublished CookieBannerVersionState = "PUBLISHED"
)

var (
	_ fmt.Stringer             = CookieBannerVersionState("")
	_ encoding.TextMarshaler   = CookieBannerVersionState("")
	_ encoding.TextUnmarshaler = (*CookieBannerVersionState)(nil)
)

func CookieBannerVersionStates() []CookieBannerVersionState {
	return []CookieBannerVersionState{
		CookieBannerVersionStateDraft,
		CookieBannerVersionStatePublished,
	}
}

func (v CookieBannerVersionState) IsValid() bool {
	switch v {
	case
		CookieBannerVersionStateDraft,
		CookieBannerVersionStatePublished:
		return true
	}

	return false
}

func (v CookieBannerVersionState) String() string {
	return string(v)
}

func (v CookieBannerVersionState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieBannerVersionState) UnmarshalText(text []byte) error {
	val := CookieBannerVersionState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieBannerVersionState value: %q", string(text))
	}

	*v = val

	return nil
}
