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

type RiskAssessmentBoundaryOrderField string

const (
	RiskAssessmentBoundaryOrderFieldCreatedAt RiskAssessmentBoundaryOrderField = "CREATED_AT"
	RiskAssessmentBoundaryOrderFieldName      RiskAssessmentBoundaryOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentBoundaryOrderField("")
	_ fmt.Stringer             = RiskAssessmentBoundaryOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentBoundaryOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentBoundaryOrderField)(nil)
)

func RiskAssessmentBoundaryOrderFields() []RiskAssessmentBoundaryOrderField {
	return []RiskAssessmentBoundaryOrderField{
		RiskAssessmentBoundaryOrderFieldCreatedAt,
		RiskAssessmentBoundaryOrderFieldName,
	}
}

func (v RiskAssessmentBoundaryOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentBoundaryOrderFieldCreatedAt,
		RiskAssessmentBoundaryOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentBoundaryOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentBoundaryOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentBoundaryOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentBoundaryOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentBoundaryOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentBoundaryOrderField) Column() string { return string(p) }
