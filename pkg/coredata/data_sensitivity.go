// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type DataSensitivity string

const (
	DataSensitivityNone     DataSensitivity = "NONE"
	DataSensitivityLow      DataSensitivity = "LOW"
	DataSensitivityMedium   DataSensitivity = "MEDIUM"
	DataSensitivityHigh     DataSensitivity = "HIGH"
	DataSensitivityCritical DataSensitivity = "CRITICAL"
)

var (
	_ fmt.Stringer             = DataSensitivity("")
	_ encoding.TextMarshaler   = DataSensitivity("")
	_ encoding.TextUnmarshaler = (*DataSensitivity)(nil)
)

func DataSensitivities() []DataSensitivity {
	return []DataSensitivity{
		DataSensitivityNone,
		DataSensitivityLow,
		DataSensitivityMedium,
		DataSensitivityHigh,
		DataSensitivityCritical,
	}
}

func (v DataSensitivity) IsValid() bool {
	switch v {
	case
		DataSensitivityNone,
		DataSensitivityLow,
		DataSensitivityMedium,
		DataSensitivityHigh,
		DataSensitivityCritical:
		return true
	}

	return false
}

func (v DataSensitivity) String() string {
	return string(v)
}

func (v DataSensitivity) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DataSensitivity) UnmarshalText(text []byte) error {
	val := DataSensitivity(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DataSensitivity value: %q", string(text))
	}

	*v = val

	return nil
}
