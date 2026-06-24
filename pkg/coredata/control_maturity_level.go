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
	"encoding"
	"fmt"
)

type (
	ControlMaturityLevel string
)

const (
	ControlMaturityLevelNone                  ControlMaturityLevel = "NONE"
	ControlMaturityLevelInitial               ControlMaturityLevel = "INITIAL"
	ControlMaturityLevelManaged               ControlMaturityLevel = "MANAGED"
	ControlMaturityLevelDefined               ControlMaturityLevel = "DEFINED"
	ControlMaturityLevelQuantitativelyManaged ControlMaturityLevel = "QUANTITATIVELY_MANAGED"
	ControlMaturityLevelOptimizing            ControlMaturityLevel = "OPTIMIZING"
)

var (
	_ fmt.Stringer             = ControlMaturityLevel("")
	_ encoding.TextMarshaler   = ControlMaturityLevel("")
	_ encoding.TextUnmarshaler = (*ControlMaturityLevel)(nil)
)

func ControlMaturityLevels() []ControlMaturityLevel {
	return []ControlMaturityLevel{
		ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing,
	}
}

func (v ControlMaturityLevel) IsValid() bool {
	switch v {
	case
		ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing:
		return true
	}

	return false
}

func (v ControlMaturityLevel) String() string {
	return string(v)
}

func (v ControlMaturityLevel) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ControlMaturityLevel) UnmarshalText(text []byte) error {
	val := ControlMaturityLevel(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ControlMaturityLevel value: %q", string(text))
	}

	*v = val

	return nil
}
