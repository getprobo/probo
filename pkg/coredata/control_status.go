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

type ControlStatus string

const (
	ControlStatusIncluded ControlStatus = "INCLUDED"
	ControlStatusExcluded ControlStatus = "EXCLUDED"
)

func ControlStatuses() []ControlStatus {
	return []ControlStatus{
		ControlStatusIncluded,
		ControlStatusExcluded,
	}
}

func (cs ControlStatus) String() string {
	return string(cs)
}

func (cs *ControlStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		switch v {
		case "INCLUDED":
			*cs = ControlStatusIncluded
		case "EXCLUDED":
			*cs = ControlStatusExcluded
		default:
			return fmt.Errorf("invalid ControlStatus value: %q", v)
		}
	default:
		return fmt.Errorf("unsupported type for ControlStatus: %T", value)
	}
	return nil
}

func (cs ControlStatus) Value() (driver.Value, error) {
	return cs.String(), nil
}
