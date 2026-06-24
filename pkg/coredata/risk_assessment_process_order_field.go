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

type RiskAssessmentProcessOrderField string

const (
	RiskAssessmentProcessOrderFieldCreatedAt RiskAssessmentProcessOrderField = "CREATED_AT"
	RiskAssessmentProcessOrderFieldName      RiskAssessmentProcessOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentProcessOrderField("")
	_ fmt.Stringer             = RiskAssessmentProcessOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentProcessOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentProcessOrderField)(nil)
)

func RiskAssessmentProcessOrderFields() []RiskAssessmentProcessOrderField {
	return []RiskAssessmentProcessOrderField{
		RiskAssessmentProcessOrderFieldCreatedAt,
		RiskAssessmentProcessOrderFieldName,
	}
}

func (v RiskAssessmentProcessOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentProcessOrderFieldCreatedAt,
		RiskAssessmentProcessOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentProcessOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentProcessOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentProcessOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentProcessOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentProcessOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentProcessOrderField) Column() string { return string(p) }
