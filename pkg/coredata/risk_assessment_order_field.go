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
