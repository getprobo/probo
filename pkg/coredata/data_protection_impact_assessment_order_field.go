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

	"go.probo.inc/probo/pkg/page"
)

type DataProtectionImpactAssessmentOrderField string

const (
	DataProtectionImpactAssessmentOrderFieldCreatedAt DataProtectionImpactAssessmentOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DataProtectionImpactAssessmentOrderField("")
	_ fmt.Stringer             = DataProtectionImpactAssessmentOrderField("")
	_ encoding.TextMarshaler   = DataProtectionImpactAssessmentOrderField("")
	_ encoding.TextUnmarshaler = (*DataProtectionImpactAssessmentOrderField)(nil)
)

func DataProtectionImpactAssessmentOrderFields() []DataProtectionImpactAssessmentOrderField {
	return []DataProtectionImpactAssessmentOrderField{
		DataProtectionImpactAssessmentOrderFieldCreatedAt,
	}
}

func (v DataProtectionImpactAssessmentOrderField) IsValid() bool {
	switch v {
	case
		DataProtectionImpactAssessmentOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v DataProtectionImpactAssessmentOrderField) String() string {
	return string(v)
}

func (v DataProtectionImpactAssessmentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DataProtectionImpactAssessmentOrderField) UnmarshalText(text []byte) error {
	val := DataProtectionImpactAssessmentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DataProtectionImpactAssessmentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p DataProtectionImpactAssessmentOrderField) Column() string {
	return string(p)
}
