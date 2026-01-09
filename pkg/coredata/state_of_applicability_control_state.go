// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type StateOfApplicabilityControlState string

const (
	StateOfApplicabilityControlStateExcluded       StateOfApplicabilityControlState = "EXCLUDED"
	StateOfApplicabilityControlStateImplemented    StateOfApplicabilityControlState = "IMPLEMENTED"
	StateOfApplicabilityControlStateNotImplemented StateOfApplicabilityControlState = "NOT_IMPLEMENTED"
)

func (s StateOfApplicabilityControlState) String() string {
	return string(s)
}

func (s StateOfApplicabilityControlState) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *StateOfApplicabilityControlState) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case StateOfApplicabilityControlStateExcluded.String():
		*s = StateOfApplicabilityControlStateExcluded
	case StateOfApplicabilityControlStateImplemented.String():
		*s = StateOfApplicabilityControlStateImplemented
	case StateOfApplicabilityControlStateNotImplemented.String():
		*s = StateOfApplicabilityControlStateNotImplemented
	default:
		return fmt.Errorf("invalid StateOfApplicabilityControlState value: %q", val)
	}

	return nil
}

func (s *StateOfApplicabilityControlState) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot scan nil into StateOfApplicabilityControlState")
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type for StateOfApplicabilityControlState: %T", value)
	}

	return s.UnmarshalText([]byte(str))
}

func (s StateOfApplicabilityControlState) Value() (driver.Value, error) {
	return string(s), nil
}
