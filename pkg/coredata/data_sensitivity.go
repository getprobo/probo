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
	"encoding/json"
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

func DataSensitivities() []DataSensitivity {
	return []DataSensitivity{
		DataSensitivityNone,
		DataSensitivityLow,
		DataSensitivityMedium,
		DataSensitivityHigh,
		DataSensitivityCritical,
	}
}

func (i DataSensitivity) String() string {
	return string(i)
}

func (i *DataSensitivity) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		switch v {
		case "NONE":
			*i = DataSensitivityNone
		case "LOW":
			*i = DataSensitivityLow
		case "MEDIUM":
			*i = DataSensitivityMedium
		case "HIGH":
			*i = DataSensitivityHigh
		case "CRITICAL":
			*i = DataSensitivityCritical
		default:
			return fmt.Errorf("invalid DataSensitivity value: %q", v)
		}
	default:
		return fmt.Errorf("unsupported type for DataSensitivity: %T", value)
	}
	return nil
}

func (i DataSensitivity) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i DataSensitivity) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *DataSensitivity) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "NONE":
		*i = DataSensitivityNone
	case "LOW":
		*i = DataSensitivityLow
	case "MEDIUM":
		*i = DataSensitivityMedium
	case "HIGH":
		*i = DataSensitivityHigh
	case "CRITICAL":
		*i = DataSensitivityCritical
	default:
		return fmt.Errorf("invalid DataSensitivity value: %q", s)
	}
	return nil
}

func (i *DataSensitivity) UnmarshalText(text []byte) error {
	var s string
	if err := json.Unmarshal(text, &s); err != nil {
		return err
	}

	switch s {
	case "NONE":
		*i = DataSensitivityNone
	case "LOW":
		*i = DataSensitivityLow
	case "MEDIUM":
		*i = DataSensitivityMedium
	case "HIGH":
		*i = DataSensitivityHigh
	case "CRITICAL":
		*i = DataSensitivityCritical
	default:
		return fmt.Errorf("invalid DataSensitivity value: %q", s)
	}
	return nil
}
