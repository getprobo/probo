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

type BusinessFunctionClassification string

const (
	BusinessFunctionClassificationCritical  BusinessFunctionClassification = "CRITICAL"
	BusinessFunctionClassificationImportant BusinessFunctionClassification = "IMPORTANT"
	BusinessFunctionClassificationSecondary BusinessFunctionClassification = "SECONDARY"
	BusinessFunctionClassificationStandard  BusinessFunctionClassification = "STANDARD"
)

var (
	_ fmt.Stringer             = BusinessFunctionClassification("")
	_ encoding.TextMarshaler   = BusinessFunctionClassification("")
	_ encoding.TextUnmarshaler = (*BusinessFunctionClassification)(nil)
)

func BusinessFunctionClassifications() []BusinessFunctionClassification {
	return []BusinessFunctionClassification{
		BusinessFunctionClassificationCritical,
		BusinessFunctionClassificationImportant,
		BusinessFunctionClassificationSecondary,
		BusinessFunctionClassificationStandard,
	}
}

func (v BusinessFunctionClassification) IsValid() bool {
	switch v {
	case
		BusinessFunctionClassificationCritical,
		BusinessFunctionClassificationImportant,
		BusinessFunctionClassificationSecondary,
		BusinessFunctionClassificationStandard:
		return true
	}

	return false
}

func (v BusinessFunctionClassification) String() string {
	return string(v)
}

func (v BusinessFunctionClassification) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *BusinessFunctionClassification) UnmarshalText(text []byte) error {
	val := BusinessFunctionClassification(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid BusinessFunctionClassification value: %q", string(text))
	}

	*v = val

	return nil
}
