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

type RiskAssessmentOrderField string

const (
	RiskAssessmentOrderFieldCreatedAt RiskAssessmentOrderField = "CREATED_AT"
	RiskAssessmentOrderFieldName      RiskAssessmentOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentOrderField("")
	_ fmt.Stringer             = RiskAssessmentOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentOrderField)(nil)
)

func RiskAssessmentOrderFields() []RiskAssessmentOrderField {
	return []RiskAssessmentOrderField{
		RiskAssessmentOrderFieldCreatedAt,
		RiskAssessmentOrderFieldName,
	}
}

func (v RiskAssessmentOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentOrderFieldCreatedAt,
		RiskAssessmentOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentOrderField) Column() string {
	return string(p)
}
