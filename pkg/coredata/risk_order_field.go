// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
