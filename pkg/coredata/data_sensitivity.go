// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
