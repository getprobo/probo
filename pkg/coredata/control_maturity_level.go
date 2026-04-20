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

func (l ControlMaturityLevel) IsValid() bool {
	switch l {
	case ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing:
		return true
	}
	return false
}

func (l ControlMaturityLevel) String() string {
	return string(l)
}

func (l ControlMaturityLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *ControlMaturityLevel) UnmarshalText(data []byte) error {
	val := ControlMaturityLevel(data)
	if !val.IsValid() {
		return fmt.Errorf("invalid ControlMaturityLevel value: %q", string(data))
	}
	*l = val
	return nil
}

func (l *ControlMaturityLevel) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for ControlMaturityLevel, expected string got %T", value)
	}
	return l.UnmarshalText([]byte(val))
}

func (l ControlMaturityLevel) Value() (driver.Value, error) {
	return l.String(), nil
}
