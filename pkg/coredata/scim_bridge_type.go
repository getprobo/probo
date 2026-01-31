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

type SCIMBridgeType string

const (
	SCIMBridgeTypeGoogleWorkspace SCIMBridgeType = "GOOGLE_WORKSPACE"
)

func (t SCIMBridgeType) String() string {
	return string(t)
}

func (t *SCIMBridgeType) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type for SCIMBridgeType: %T", value)
	}

	switch str {
	case "GOOGLE_WORKSPACE":
		*t = SCIMBridgeTypeGoogleWorkspace
	default:
		return fmt.Errorf("invalid SCIMBridgeType value: %q", str)
	}
	return nil
}

func (t SCIMBridgeType) Value() (driver.Value, error) {
	return t.String(), nil
}
