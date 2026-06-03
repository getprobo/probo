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

type RightsRequestType string

const (
	RightsRequestTypeAccess      RightsRequestType = "ACCESS"
	RightsRequestTypeDeletion    RightsRequestType = "DELETION"
	RightsRequestTypePortability RightsRequestType = "PORTABILITY"
)

var (
	_ fmt.Stringer             = RightsRequestType("")
	_ encoding.TextMarshaler   = RightsRequestType("")
	_ encoding.TextUnmarshaler = (*RightsRequestType)(nil)
)

func RightsRequestTypes() []RightsRequestType {
	return []RightsRequestType{
		RightsRequestTypeAccess,
		RightsRequestTypeDeletion,
		RightsRequestTypePortability,
	}
}

func (v RightsRequestType) IsValid() bool {
	switch v {
	case
		RightsRequestTypeAccess,
		RightsRequestTypeDeletion,
		RightsRequestTypePortability:
		return true
	}

	return false
}

func (v RightsRequestType) String() string {
	return string(v)
}

func (v RightsRequestType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RightsRequestType) UnmarshalText(text []byte) error {
	val := RightsRequestType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RightsRequestType value: %q", string(text))
	}

	*v = val

	return nil
}
