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

type (
	RiskOrderField string
)

const (
	RiskOrderFieldCreatedAt         RiskOrderField = "CREATED_AT"
	RiskOrderFieldUpdatedAt         RiskOrderField = "UPDATED_AT"
	RiskOrderFieldName              RiskOrderField = "NAME"
	RiskOrderFieldCategory          RiskOrderField = "CATEGORY"
	RiskOrderFieldTreatment         RiskOrderField = "TREATMENT"
	RiskOrderFieldInherentRiskScore RiskOrderField = "INHERENT_RISK_SCORE"
	RiskOrderFieldResidualRiskScore RiskOrderField = "RESIDUAL_RISK_SCORE"
	RiskOrderFieldOwnerFullName     RiskOrderField = "OWNER_FULL_NAME"
)

var (
	_ page.OrderField          = RiskOrderField("")
	_ fmt.Stringer             = RiskOrderField("")
	_ encoding.TextMarshaler   = RiskOrderField("")
	_ encoding.TextUnmarshaler = (*RiskOrderField)(nil)
)

func RiskOrderFields() []RiskOrderField {
	return []RiskOrderField{
		RiskOrderFieldCreatedAt,
		RiskOrderFieldUpdatedAt,
		RiskOrderFieldName,
		RiskOrderFieldCategory,
		RiskOrderFieldTreatment,
		RiskOrderFieldInherentRiskScore,
		RiskOrderFieldResidualRiskScore,
		RiskOrderFieldOwnerFullName,
	}
}

func (v RiskOrderField) IsValid() bool {
	switch v {
	case
		RiskOrderFieldCreatedAt,
		RiskOrderFieldUpdatedAt,
		RiskOrderFieldName,
		RiskOrderFieldCategory,
		RiskOrderFieldTreatment,
		RiskOrderFieldInherentRiskScore,
		RiskOrderFieldResidualRiskScore,
		RiskOrderFieldOwnerFullName:
		return true
	}

	return false
}

func (v RiskOrderField) String() string {
	return string(v)
}

func (v RiskOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskOrderField) UnmarshalText(text []byte) error {
	val := RiskOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskOrderField) Column() string {
	return string(p)
}
