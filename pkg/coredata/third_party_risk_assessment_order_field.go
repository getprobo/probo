// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	ThirdPartyRiskAssessmentOrderField string
)

const (
	ThirdPartyRiskAssessmentOrderFieldCreatedAt ThirdPartyRiskAssessmentOrderField = "CREATED_AT"
	ThirdPartyRiskAssessmentOrderFieldExpiresAt ThirdPartyRiskAssessmentOrderField = "EXPIRES_AT"
)

var (
	_ page.OrderField          = ThirdPartyRiskAssessmentOrderField("")
	_ fmt.Stringer             = ThirdPartyRiskAssessmentOrderField("")
	_ encoding.TextMarshaler   = ThirdPartyRiskAssessmentOrderField("")
	_ encoding.TextUnmarshaler = (*ThirdPartyRiskAssessmentOrderField)(nil)
)

func ThirdPartyRiskAssessmentOrderFields() []ThirdPartyRiskAssessmentOrderField {
	return []ThirdPartyRiskAssessmentOrderField{
		ThirdPartyRiskAssessmentOrderFieldCreatedAt,
		ThirdPartyRiskAssessmentOrderFieldExpiresAt,
	}
}

func (v ThirdPartyRiskAssessmentOrderField) IsValid() bool {
	switch v {
	case
		ThirdPartyRiskAssessmentOrderFieldCreatedAt,
		ThirdPartyRiskAssessmentOrderFieldExpiresAt:
		return true
	}

	return false
}

func (v ThirdPartyRiskAssessmentOrderField) String() string {
	return string(v)
}

func (v ThirdPartyRiskAssessmentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ThirdPartyRiskAssessmentOrderField) UnmarshalText(text []byte) error {
	val := ThirdPartyRiskAssessmentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ThirdPartyRiskAssessmentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ThirdPartyRiskAssessmentOrderField) Column() string {
	return string(p)
}
