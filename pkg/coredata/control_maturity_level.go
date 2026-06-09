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

type (
	ControlMaturityLevel string
)

const (
	ControlMaturityLevelNone                  ControlMaturityLevel = "NONE"
	ControlMaturityLevelInitial               ControlMaturityLevel = "INITIAL"
	ControlMaturityLevelManaged               ControlMaturityLevel = "MANAGED"
	ControlMaturityLevelDefined               ControlMaturityLevel = "DEFINED"
	ControlMaturityLevelQuantitativelyManaged ControlMaturityLevel = "QUANTITATIVELY_MANAGED"
	ControlMaturityLevelOptimizing            ControlMaturityLevel = "OPTIMIZING"
)

var (
	_ fmt.Stringer             = ControlMaturityLevel("")
	_ encoding.TextMarshaler   = ControlMaturityLevel("")
	_ encoding.TextUnmarshaler = (*ControlMaturityLevel)(nil)
)

func ControlMaturityLevels() []ControlMaturityLevel {
	return []ControlMaturityLevel{
		ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing,
	}
}

func (v ControlMaturityLevel) IsValid() bool {
	switch v {
	case
		ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing:
		return true
	}

	return false
}

func (v ControlMaturityLevel) String() string {
	return string(v)
}

func (v ControlMaturityLevel) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ControlMaturityLevel) UnmarshalText(text []byte) error {
	val := ControlMaturityLevel(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ControlMaturityLevel value: %q", string(text))
	}

	*v = val

	return nil
}
