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
	"encoding"
	"fmt"
	"strings"
)

type (
	ProfileState       string
	ProfileStateValues []ProfileState
)

const (
	ProfileStatePending     ProfileState = "PENDING"
	ProfileStateActive      ProfileState = "ACTIVE"
	ProfileStateDeactivated ProfileState = "DEACTIVATED"
)

var (
	_ fmt.Stringer             = ProfileState("")
	_ encoding.TextMarshaler   = ProfileState("")
	_ encoding.TextUnmarshaler = (*ProfileState)(nil)
)

func ProfileStates() []ProfileState {
	return []ProfileState{
		ProfileStatePending,
		ProfileStateActive,
		ProfileStateDeactivated,
	}
}

func (v ProfileState) IsValid() bool {
	switch v {
	case
		ProfileStatePending,
		ProfileStateActive,
		ProfileStateDeactivated:
		return true
	}

	return false
}

func (v ProfileState) String() string {
	return string(v)
}

func (v ProfileState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProfileState) UnmarshalText(text []byte) error {
	val := ProfileState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProfileState value: %q", string(text))
	}

	*v = val

	return nil
}

func (states ProfileStateValues) Value() (driver.Value, error) {
	if len(states) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")

	for i, state := range states {
		if i > 0 {
			result.WriteString(",")
		}

		fmt.Fprintf(&result, "%q", state.String())
	}

	result.WriteString("}")

	return result.String(), nil
}
