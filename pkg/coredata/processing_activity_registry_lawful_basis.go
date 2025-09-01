// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type ProcessingActivityRegistryLawfulBasis string

const (
	ProcessingActivityRegistryLawfulBasisLegitimateInterest   ProcessingActivityRegistryLawfulBasis = "LEGITIMATE_INTEREST"
	ProcessingActivityRegistryLawfulBasisConsent              ProcessingActivityRegistryLawfulBasis = "CONSENT"
	ProcessingActivityRegistryLawfulBasisContractualNecessity ProcessingActivityRegistryLawfulBasis = "CONTRACTUAL_NECESSITY"
	ProcessingActivityRegistryLawfulBasisLegalObligation      ProcessingActivityRegistryLawfulBasis = "LEGAL_OBLIGATION"
	ProcessingActivityRegistryLawfulBasisVitalInterests       ProcessingActivityRegistryLawfulBasis = "VITAL_INTERESTS"
	ProcessingActivityRegistryLawfulBasisPublicTask           ProcessingActivityRegistryLawfulBasis = "PUBLIC_TASK"
)

func (p ProcessingActivityRegistryLawfulBasis) String() string {
	return string(p)
}

func (p *ProcessingActivityRegistryLawfulBasis) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityRegistryLawfulBasis: %T", value)
	}

	switch s {
	case "LEGITIMATE_INTEREST":
		*p = ProcessingActivityRegistryLawfulBasisLegitimateInterest
	case "CONSENT":
		*p = ProcessingActivityRegistryLawfulBasisConsent
	case "CONTRACTUAL_NECESSITY":
		*p = ProcessingActivityRegistryLawfulBasisContractualNecessity
	case "LEGAL_OBLIGATION":
		*p = ProcessingActivityRegistryLawfulBasisLegalObligation
	case "VITAL_INTERESTS":
		*p = ProcessingActivityRegistryLawfulBasisVitalInterests
	case "PUBLIC_TASK":
		*p = ProcessingActivityRegistryLawfulBasisPublicTask
	default:
		return fmt.Errorf("invalid ProcessingActivityRegistryLawfulBasis value: %q", s)
	}
	return nil
}

func (p ProcessingActivityRegistryLawfulBasis) Value() (driver.Value, error) {
	return p.String(), nil
}
