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

type RiskAssessmentNodeOrderField string

const (
	RiskAssessmentNodeOrderFieldCreatedAt RiskAssessmentNodeOrderField = "CREATED_AT"
	RiskAssessmentNodeOrderFieldName      RiskAssessmentNodeOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentNodeOrderField("")
	_ fmt.Stringer             = RiskAssessmentNodeOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentNodeOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentNodeOrderField)(nil)
)

func RiskAssessmentNodeOrderFields() []RiskAssessmentNodeOrderField {
	return []RiskAssessmentNodeOrderField{
		RiskAssessmentNodeOrderFieldCreatedAt,
		RiskAssessmentNodeOrderFieldName,
	}
}

func (v RiskAssessmentNodeOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentNodeOrderFieldCreatedAt,
		RiskAssessmentNodeOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentNodeOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentNodeOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentNodeOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentNodeOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentNodeOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentNodeOrderField) Column() string { return string(p) }
