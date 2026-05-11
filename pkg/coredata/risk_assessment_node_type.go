// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
	"fmt"
)

type RiskAssessmentNodeType string

const (
	RiskAssessmentNodeTypeEntity   RiskAssessmentNodeType = "ENTITY"
	RiskAssessmentNodeTypeBoundary RiskAssessmentNodeType = "BOUNDARY"
	RiskAssessmentNodeTypeAsset    RiskAssessmentNodeType = "ASSET"
	RiskAssessmentNodeTypeData     RiskAssessmentNodeType = "DATA"
)

func RiskAssessmentNodeTypes() []RiskAssessmentNodeType {
	return []RiskAssessmentNodeType{
		RiskAssessmentNodeTypeEntity,
		RiskAssessmentNodeTypeBoundary,
		RiskAssessmentNodeTypeAsset,
		RiskAssessmentNodeTypeData,
	}
}

func (t RiskAssessmentNodeType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *RiskAssessmentNodeType) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case RiskAssessmentNodeTypeEntity.String():
		*t = RiskAssessmentNodeTypeEntity
	case RiskAssessmentNodeTypeBoundary.String():
		*t = RiskAssessmentNodeTypeBoundary
	case RiskAssessmentNodeTypeAsset.String():
		*t = RiskAssessmentNodeTypeAsset
	case RiskAssessmentNodeTypeData.String():
		*t = RiskAssessmentNodeTypeData
	default:
		return fmt.Errorf("invalid RiskAssessmentNodeType value: %q", val)
	}

	return nil
}

func (t RiskAssessmentNodeType) String() string {
	return string(t)
}

func (t *RiskAssessmentNodeType) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for RiskAssessmentNodeType, expected string got %T", value)
	}

	return t.UnmarshalText([]byte(val))
}

func (t RiskAssessmentNodeType) Value() (driver.Value, error) {
	return t.String(), nil
}
