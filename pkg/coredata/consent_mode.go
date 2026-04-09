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
	"database/sql/driver"
	"fmt"
)

type ConsentMode string

const (
	ConsentModeOptIn  ConsentMode = "OPT_IN"
	ConsentModeOptOut ConsentMode = "OPT_OUT"
)

func ConsentModes() []ConsentMode {
	return []ConsentMode{
		ConsentModeOptIn,
		ConsentModeOptOut,
	}
}

func (m ConsentMode) String() string {
	return string(m)
}

func (m *ConsentMode) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for ConsentMode: %T", value)
	}

	switch ConsentMode(v) {
	case ConsentModeOptIn:
		*m = ConsentModeOptIn
	case ConsentModeOptOut:
		*m = ConsentModeOptOut
	default:
		return fmt.Errorf("invalid ConsentMode value: %q", v)
	}

	return nil
}

func (m ConsentMode) Value() (driver.Value, error) {
	switch m {
	case ConsentModeOptIn,
		ConsentModeOptOut:
		return string(m), nil
	default:
		return nil, fmt.Errorf("invalid ConsentMode: %s", m)
	}
}
