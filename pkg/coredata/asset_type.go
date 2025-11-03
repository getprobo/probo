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

type (
	AssetType string
)

const (
	AssetTypePhysical AssetType = "PHYSICAL"
	AssetTypeVirtual  AssetType = "VIRTUAL"
)

func AssetTypes() []AssetType {
	return []AssetType{
		AssetTypePhysical,
		AssetTypeVirtual,
	}
}

func (at AssetType) MarshalText() ([]byte, error) {
	return []byte(at.String()), nil
}

func (at *AssetType) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case AssetTypePhysical.String():
		*at = AssetTypePhysical
	case AssetTypeVirtual.String():
		*at = AssetTypeVirtual
	default:
		return fmt.Errorf("invalid AssetType value: %q", val)
	}

	return nil
}

func (at AssetType) String() string {
	return string(at)
}

func (at *AssetType) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for AssetType, expected string got %T", value)
	}

	return at.UnmarshalText([]byte(val))
}

func (at AssetType) Value() (driver.Value, error) {
	return at.String(), nil
}
