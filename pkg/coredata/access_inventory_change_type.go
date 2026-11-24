// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type (
	AccessInventoryChangeType       string
	AccessInventoryChangeTypeValues []AccessInventoryChangeType
)

const (
	AccessInventoryChangeTypeAppeared         AccessInventoryChangeType = "APPEARED"
	AccessInventoryChangeTypeDisappeared      AccessInventoryChangeType = "DISAPPEARED"
	AccessInventoryChangeTypeAttributeChanged AccessInventoryChangeType = "ATTRIBUTE_CHANGED"
)

var (
	_ fmt.Stringer             = AccessInventoryChangeType("")
	_ encoding.TextMarshaler   = AccessInventoryChangeType("")
	_ encoding.TextUnmarshaler = (*AccessInventoryChangeType)(nil)
	_ driver.Valuer            = AccessInventoryChangeTypeValues(nil)
)

func AccessInventoryChangeTypes() []AccessInventoryChangeType {
	return []AccessInventoryChangeType{
		AccessInventoryChangeTypeAppeared,
		AccessInventoryChangeTypeDisappeared,
		AccessInventoryChangeTypeAttributeChanged,
	}
}

func (v AccessInventoryChangeType) IsValid() bool {
	switch v {
	case
		AccessInventoryChangeTypeAppeared,
		AccessInventoryChangeTypeDisappeared,
		AccessInventoryChangeTypeAttributeChanged:
		return true
	}

	return false
}

func (v AccessInventoryChangeType) String() string {
	return string(v)
}

func (v AccessInventoryChangeType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessInventoryChangeType) UnmarshalText(text []byte) error {
	val := AccessInventoryChangeType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessInventoryChangeType value: %q", string(text))
	}

	*v = val

	return nil
}

func (values AccessInventoryChangeTypeValues) Value() (driver.Value, error) {
	if len(values) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")

	for i, value := range values {
		if i > 0 {
			result.WriteString(",")
		}

		fmt.Fprintf(&result, "%q", value.String())
	}

	result.WriteString("}")

	return result.String(), nil
}
