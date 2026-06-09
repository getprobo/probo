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

	"go.probo.inc/probo/pkg/page"
)

type RiskAssessmentScenarioOrderField string

const (
	RiskAssessmentScenarioOrderFieldCreatedAt RiskAssessmentScenarioOrderField = "CREATED_AT"
	RiskAssessmentScenarioOrderFieldName      RiskAssessmentScenarioOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentScenarioOrderField("")
	_ fmt.Stringer             = RiskAssessmentScenarioOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentScenarioOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentScenarioOrderField)(nil)
)

func RiskAssessmentScenarioOrderFields() []RiskAssessmentScenarioOrderField {
	return []RiskAssessmentScenarioOrderField{
		RiskAssessmentScenarioOrderFieldCreatedAt,
		RiskAssessmentScenarioOrderFieldName,
	}
}

func (v RiskAssessmentScenarioOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentScenarioOrderFieldCreatedAt,
		RiskAssessmentScenarioOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentScenarioOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentScenarioOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentScenarioOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentScenarioOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentScenarioOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentScenarioOrderField) Column() string { return string(p) }
