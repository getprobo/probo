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

type RiskAssessmentThreatOrderField string

const (
	RiskAssessmentThreatOrderFieldCreatedAt RiskAssessmentThreatOrderField = "CREATED_AT"
	RiskAssessmentThreatOrderFieldName      RiskAssessmentThreatOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentThreatOrderField("")
	_ fmt.Stringer             = RiskAssessmentThreatOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentThreatOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentThreatOrderField)(nil)
)

func RiskAssessmentThreatOrderFields() []RiskAssessmentThreatOrderField {
	return []RiskAssessmentThreatOrderField{
		RiskAssessmentThreatOrderFieldCreatedAt,
		RiskAssessmentThreatOrderFieldName,
	}
}

func (v RiskAssessmentThreatOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentThreatOrderFieldCreatedAt,
		RiskAssessmentThreatOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentThreatOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentThreatOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentThreatOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentThreatOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentThreatOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentThreatOrderField) Column() string { return string(p) }
