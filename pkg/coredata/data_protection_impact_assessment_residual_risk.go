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
)

type DataProtectionImpactAssessmentResidualRisk string

const (
	DataProtectionImpactAssessmentResidualRiskLow    DataProtectionImpactAssessmentResidualRisk = "LOW"
	DataProtectionImpactAssessmentResidualRiskMedium DataProtectionImpactAssessmentResidualRisk = "MEDIUM"
	DataProtectionImpactAssessmentResidualRiskHigh   DataProtectionImpactAssessmentResidualRisk = "HIGH"
)

var (
	_ fmt.Stringer             = DataProtectionImpactAssessmentResidualRisk("")
	_ encoding.TextMarshaler   = DataProtectionImpactAssessmentResidualRisk("")
	_ encoding.TextUnmarshaler = (*DataProtectionImpactAssessmentResidualRisk)(nil)
)

func DataProtectionImpactAssessmentResidualRisks() []DataProtectionImpactAssessmentResidualRisk {
	return []DataProtectionImpactAssessmentResidualRisk{
		DataProtectionImpactAssessmentResidualRiskLow,
		DataProtectionImpactAssessmentResidualRiskMedium,
		DataProtectionImpactAssessmentResidualRiskHigh,
	}
}

func (v DataProtectionImpactAssessmentResidualRisk) IsValid() bool {
	switch v {
	case
		DataProtectionImpactAssessmentResidualRiskLow,
		DataProtectionImpactAssessmentResidualRiskMedium,
		DataProtectionImpactAssessmentResidualRiskHigh:
		return true
	}

	return false
}

func (v DataProtectionImpactAssessmentResidualRisk) String() string {
	return string(v)
}

func (v DataProtectionImpactAssessmentResidualRisk) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DataProtectionImpactAssessmentResidualRisk) UnmarshalText(text []byte) error {
	val := DataProtectionImpactAssessmentResidualRisk(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DataProtectionImpactAssessmentResidualRisk value: %q", string(text))
	}

	*v = val

	return nil
}
