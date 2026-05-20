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
	"encoding"
	"fmt"
)

type OAuth2SubjectType string

const (
	OAuth2SubjectTypePublic OAuth2SubjectType = "public"
)

var (
	_ fmt.Stringer             = OAuth2SubjectType("")
	_ encoding.TextMarshaler   = OAuth2SubjectType("")
	_ encoding.TextUnmarshaler = (*OAuth2SubjectType)(nil)
)

func OAuth2SubjectTypes() []OAuth2SubjectType {
	return []OAuth2SubjectType{
		OAuth2SubjectTypePublic,
	}
}

func (v OAuth2SubjectType) IsValid() bool {
	switch v {
	case
		OAuth2SubjectTypePublic:
		return true
	}

	return false
}

func (v OAuth2SubjectType) String() string {
	return string(v)
}

func (v OAuth2SubjectType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2SubjectType) UnmarshalText(text []byte) error {
	val := OAuth2SubjectType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2SubjectType value: %q", string(text))
	}

	*v = val

	return nil
}
