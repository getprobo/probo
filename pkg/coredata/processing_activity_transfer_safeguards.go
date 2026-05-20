// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type ProcessingActivityTransferSafeguard string

const (
	ProcessingActivityTransferSafeguardStandardContractualClauses ProcessingActivityTransferSafeguard = "STANDARD_CONTRACTUAL_CLAUSES"
	ProcessingActivityTransferSafeguardBindingCorporateRules      ProcessingActivityTransferSafeguard = "BINDING_CORPORATE_RULES"
	ProcessingActivityTransferSafeguardAdequacyDecision           ProcessingActivityTransferSafeguard = "ADEQUACY_DECISION"
	ProcessingActivityTransferSafeguardDerogations                ProcessingActivityTransferSafeguard = "DEROGATIONS"
	ProcessingActivityTransferSafeguardCodesOfConduct             ProcessingActivityTransferSafeguard = "CODES_OF_CONDUCT"
	ProcessingActivityTransferSafeguardCertificationMechanisms    ProcessingActivityTransferSafeguard = "CERTIFICATION_MECHANISMS"
)

var (
	_ fmt.Stringer             = ProcessingActivityTransferSafeguard("")
	_ encoding.TextMarshaler   = ProcessingActivityTransferSafeguard("")
	_ encoding.TextUnmarshaler = (*ProcessingActivityTransferSafeguard)(nil)
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

func (v ProcessingActivityTransferSafeguard) IsValid() bool {
	switch v {
	case
		ProcessingActivityTransferSafeguardStandardContractualClauses,
		ProcessingActivityTransferSafeguardBindingCorporateRules,
		ProcessingActivityTransferSafeguardAdequacyDecision,
		ProcessingActivityTransferSafeguardDerogations,
		ProcessingActivityTransferSafeguardCodesOfConduct,
		ProcessingActivityTransferSafeguardCertificationMechanisms:
		return true
	}

	return false
}

func (v ProcessingActivityTransferSafeguard) String() string {
	return string(v)
}

func (v ProcessingActivityTransferSafeguard) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProcessingActivityTransferSafeguard) UnmarshalText(text []byte) error {
	val := ProcessingActivityTransferSafeguard(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProcessingActivityTransferSafeguard value: %q", string(text))
	}

	*v = val

	return nil
}
