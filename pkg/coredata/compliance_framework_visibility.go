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

type ComplianceFrameworkVisibility string

const (
	ComplianceFrameworkVisibilityNone   ComplianceFrameworkVisibility = "NONE"
	ComplianceFrameworkVisibilityPublic ComplianceFrameworkVisibility = "PUBLIC"
)

var (
	_ fmt.Stringer             = ComplianceFrameworkVisibility("")
	_ encoding.TextMarshaler   = ComplianceFrameworkVisibility("")
	_ encoding.TextUnmarshaler = (*ComplianceFrameworkVisibility)(nil)
)

func ComplianceFrameworkVisibilities() []ComplianceFrameworkVisibility {
	return []ComplianceFrameworkVisibility{
		ComplianceFrameworkVisibilityNone,
		ComplianceFrameworkVisibilityPublic,
	}
}

func (v ComplianceFrameworkVisibility) IsValid() bool {
	switch v {
	case
		ComplianceFrameworkVisibilityNone,
		ComplianceFrameworkVisibilityPublic:
		return true
	}

	return false
}

func (v ComplianceFrameworkVisibility) String() string {
	return string(v)
}

func (v ComplianceFrameworkVisibility) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ComplianceFrameworkVisibility) UnmarshalText(text []byte) error {
	val := ComplianceFrameworkVisibility(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ComplianceFrameworkVisibility value: %q", string(text))
	}

	*v = val

	return nil
}
