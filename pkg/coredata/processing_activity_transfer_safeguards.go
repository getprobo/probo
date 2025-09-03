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

type ProcessingActivityTransferSafeguards string

const (
	ProcessingActivityTransferSafeguardsStandardContractualClauses ProcessingActivityTransferSafeguards = "STANDARD_CONTRACTUAL_CLAUSES"
	ProcessingActivityTransferSafeguardsBindingCorporateRules      ProcessingActivityTransferSafeguards = "BINDING_CORPORATE_RULES"
	ProcessingActivityTransferSafeguardsAdequacyDecision           ProcessingActivityTransferSafeguards = "ADEQUACY_DECISION"
	ProcessingActivityTransferSafeguardsDerogations                ProcessingActivityTransferSafeguards = "DEROGATIONS"
	ProcessingActivityTransferSafeguardsCodesOfConduct             ProcessingActivityTransferSafeguards = "CODES_OF_CONDUCT"
	ProcessingActivityTransferSafeguardsCertificationMechanisms    ProcessingActivityTransferSafeguards = "CERTIFICATION_MECHANISMS"
)

func (p ProcessingActivityTransferSafeguards) String() string {
	return string(p)
}

func (p *ProcessingActivityTransferSafeguards) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityTransferSafeguards: %T", value)
	}

	switch s {
	case "STANDARD_CONTRACTUAL_CLAUSES":
		*p = ProcessingActivityTransferSafeguardsStandardContractualClauses
	case "BINDING_CORPORATE_RULES":
		*p = ProcessingActivityTransferSafeguardsBindingCorporateRules
	case "ADEQUACY_DECISION":
		*p = ProcessingActivityTransferSafeguardsAdequacyDecision
	case "DEROGATIONS":
		*p = ProcessingActivityTransferSafeguardsDerogations
	case "CODES_OF_CONDUCT":
		*p = ProcessingActivityTransferSafeguardsCodesOfConduct
	case "CERTIFICATION_MECHANISMS":
		*p = ProcessingActivityTransferSafeguardsCertificationMechanisms
	default:
		return fmt.Errorf("invalid ProcessingActivityTransferSafeguards value: %q", s)
	}
	return nil
}

func (p ProcessingActivityTransferSafeguards) Value() (driver.Value, error) {
	return p.String(), nil
}
