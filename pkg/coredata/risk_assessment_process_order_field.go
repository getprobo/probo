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

type RiskAssessmentProcessOrderField string

const (
	RiskAssessmentProcessOrderFieldCreatedAt RiskAssessmentProcessOrderField = "CREATED_AT"
	RiskAssessmentProcessOrderFieldName      RiskAssessmentProcessOrderField = "NAME"
)

var (
	_ page.OrderField          = RiskAssessmentProcessOrderField("")
	_ fmt.Stringer             = RiskAssessmentProcessOrderField("")
	_ encoding.TextMarshaler   = RiskAssessmentProcessOrderField("")
	_ encoding.TextUnmarshaler = (*RiskAssessmentProcessOrderField)(nil)
)

func RiskAssessmentProcessOrderFields() []RiskAssessmentProcessOrderField {
	return []RiskAssessmentProcessOrderField{
		RiskAssessmentProcessOrderFieldCreatedAt,
		RiskAssessmentProcessOrderFieldName,
	}
}

func (v RiskAssessmentProcessOrderField) IsValid() bool {
	switch v {
	case
		RiskAssessmentProcessOrderFieldCreatedAt,
		RiskAssessmentProcessOrderFieldName:
		return true
	}

	return false
}

func (v RiskAssessmentProcessOrderField) String() string {
	return string(v)
}

func (v RiskAssessmentProcessOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RiskAssessmentProcessOrderField) UnmarshalText(text []byte) error {
	val := RiskAssessmentProcessOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RiskAssessmentProcessOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p RiskAssessmentProcessOrderField) Column() string { return string(p) }
