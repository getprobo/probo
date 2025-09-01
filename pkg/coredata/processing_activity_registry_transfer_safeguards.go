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

type ProcessingActivityRegistryTransferSafeguards string

const (
	ProcessingActivityRegistryTransferSafeguardsStandardContractualClauses ProcessingActivityRegistryTransferSafeguards = "STANDARD_CONTRACTUAL_CLAUSES"
	ProcessingActivityRegistryTransferSafeguardsBindingCorporateRules      ProcessingActivityRegistryTransferSafeguards = "BINDING_CORPORATE_RULES"
	ProcessingActivityRegistryTransferSafeguardsAdequacyDecision           ProcessingActivityRegistryTransferSafeguards = "ADEQUACY_DECISION"
	ProcessingActivityRegistryTransferSafeguardsDerogations                ProcessingActivityRegistryTransferSafeguards = "DEROGATIONS"
	ProcessingActivityRegistryTransferSafeguardsCodesOfConduct             ProcessingActivityRegistryTransferSafeguards = "CODES_OF_CONDUCT"
	ProcessingActivityRegistryTransferSafeguardsCertificationMechanisms    ProcessingActivityRegistryTransferSafeguards = "CERTIFICATION_MECHANISMS"
)

func (p ProcessingActivityRegistryTransferSafeguards) String() string {
	return string(p)
}

func (p *ProcessingActivityRegistryTransferSafeguards) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityRegistryTransferSafeguards: %T", value)
	}

	switch s {
	case "STANDARD_CONTRACTUAL_CLAUSES":
		*p = ProcessingActivityRegistryTransferSafeguardsStandardContractualClauses
	case "BINDING_CORPORATE_RULES":
		*p = ProcessingActivityRegistryTransferSafeguardsBindingCorporateRules
	case "ADEQUACY_DECISION":
		*p = ProcessingActivityRegistryTransferSafeguardsAdequacyDecision
	case "DEROGATIONS":
		*p = ProcessingActivityRegistryTransferSafeguardsDerogations
	case "CODES_OF_CONDUCT":
		*p = ProcessingActivityRegistryTransferSafeguardsCodesOfConduct
	case "CERTIFICATION_MECHANISMS":
		*p = ProcessingActivityRegistryTransferSafeguardsCertificationMechanisms
	default:
		return fmt.Errorf("invalid ProcessingActivityRegistryTransferSafeguards value: %q", s)
	}
	return nil
}

func (p ProcessingActivityRegistryTransferSafeguards) Value() (driver.Value, error) {
	return p.String(), nil
}
