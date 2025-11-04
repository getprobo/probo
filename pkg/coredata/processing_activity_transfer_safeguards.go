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

type ProcessingActivityTransferSafeguard string

const (
	ProcessingActivityTransferSafeguardStandardContractualClauses ProcessingActivityTransferSafeguard = "STANDARD_CONTRACTUAL_CLAUSES"
	ProcessingActivityTransferSafeguardBindingCorporateRules      ProcessingActivityTransferSafeguard = "BINDING_CORPORATE_RULES"
	ProcessingActivityTransferSafeguardAdequacyDecision           ProcessingActivityTransferSafeguard = "ADEQUACY_DECISION"
	ProcessingActivityTransferSafeguardDerogations                ProcessingActivityTransferSafeguard = "DEROGATIONS"
	ProcessingActivityTransferSafeguardCodesOfConduct             ProcessingActivityTransferSafeguard = "CODES_OF_CONDUCT"
	ProcessingActivityTransferSafeguardCertificationMechanisms    ProcessingActivityTransferSafeguard = "CERTIFICATION_MECHANISMS"
)

func ProcessingActivityTransferSafeguards() []ProcessingActivityTransferSafeguard {
	return []ProcessingActivityTransferSafeguard{
		ProcessingActivityTransferSafeguardStandardContractualClauses,
		ProcessingActivityTransferSafeguardBindingCorporateRules,
		ProcessingActivityTransferSafeguardAdequacyDecision,
		ProcessingActivityTransferSafeguardDerogations,
		ProcessingActivityTransferSafeguardCodesOfConduct,
		ProcessingActivityTransferSafeguardCertificationMechanisms,
	}
}

func (p ProcessingActivityTransferSafeguard) String() string {
	return string(p)
}

func (p *ProcessingActivityTransferSafeguard) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityTransferSafeguard: %T", value)
	}

	switch s {
	case "STANDARD_CONTRACTUAL_CLAUSES":
		*p = ProcessingActivityTransferSafeguardStandardContractualClauses
	case "BINDING_CORPORATE_RULES":
		*p = ProcessingActivityTransferSafeguardBindingCorporateRules
	case "ADEQUACY_DECISION":
		*p = ProcessingActivityTransferSafeguardAdequacyDecision
	case "DEROGATIONS":
		*p = ProcessingActivityTransferSafeguardDerogations
	case "CODES_OF_CONDUCT":
		*p = ProcessingActivityTransferSafeguardCodesOfConduct
	case "CERTIFICATION_MECHANISMS":
		*p = ProcessingActivityTransferSafeguardCertificationMechanisms
	default:
		return fmt.Errorf("invalid ProcessingActivityTransferSafeguard value: %q", s)
	}
	return nil
}

func (p ProcessingActivityTransferSafeguard) Value() (driver.Value, error) {
	return p.String(), nil
}
