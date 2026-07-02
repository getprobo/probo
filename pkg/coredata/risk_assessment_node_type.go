// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
)

type RiskAssessmentNodeType string

const (
	RiskAssessmentNodeTypeEntity RiskAssessmentNodeType = "ENTITY"
	RiskAssessmentNodeTypeAsset  RiskAssessmentNodeType = "ASSET"
	RiskAssessmentNodeTypeData   RiskAssessmentNodeType = "DATA"
)

var (
	_ fmt.Stringer             = RiskAssessmentNodeType("")
	_ encoding.TextMarshaler   = RiskAssessmentNodeType("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentNodeType)(nil)
)

func RiskAssessmentNodeTypes() []RiskAssessmentNodeType {
	return []RiskAssessmentNodeType{
		RiskAssessmentNodeTypeEntity,
		RiskAssessmentNodeTypeAsset,
		RiskAssessmentNodeTypeData,
	}
}

func (v RiskAssessmentNodeType) IsValid() bool {
	switch v {
	case
		RiskAssessmentNodeTypeEntity,
		RiskAssessmentNodeTypeAsset,
		RiskAssessmentNodeTypeData:
		return true
	}

	return false
}

func (v RiskAssessmentNodeType) String() string {
	return string(v)
}

func (v RiskAssessmentNodeType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentNodeType) UnmarshalText(text []byte) error {
	val := RiskAssessmentNodeType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentNodeType value: %q", string(text))
	}

	*v = val

	return nil
}
