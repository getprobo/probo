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

	"go.probo.inc/probo/pkg/page"
)

type RiskAnalysisBoundaryOrderField string

const (
	RiskAnalysisBoundaryOrderFieldCreatedAt RiskAnalysisBoundaryOrderField = "CREATED_AT"
	RiskAnalysisBoundaryOrderFieldName      RiskAnalysisBoundaryOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAnalysisBoundaryOrderField("")
	_ fmt.Stringer             = RiskAnalysisBoundaryOrderField("")
	_ encoding.TextMarshaler   = RiskAnalysisBoundaryOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAnalysisBoundaryOrderField)(nil)
)

func RiskAnalysisBoundaryOrderFields() []RiskAnalysisBoundaryOrderField {
	return []RiskAnalysisBoundaryOrderField{
		RiskAnalysisBoundaryOrderFieldCreatedAt,
		RiskAnalysisBoundaryOrderFieldName,
	}
}

func (v RiskAnalysisBoundaryOrderField) IsValid() bool {
	switch v {
	case
		RiskAnalysisBoundaryOrderFieldCreatedAt,
		RiskAnalysisBoundaryOrderFieldName:
		return true
	}

	return false
}

func (v RiskAnalysisBoundaryOrderField) String() string {
	return string(v)
}

func (v RiskAnalysisBoundaryOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAnalysisBoundaryOrderField) UnmarshalText(text []byte) error {
	val := RiskAnalysisBoundaryOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAnalysisBoundaryOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAnalysisBoundaryOrderField) Column() string { return string(p) }
