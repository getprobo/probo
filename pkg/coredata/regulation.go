// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
	"fmt"
)

type Regulation string

const (
	RegulationNone    Regulation = ""
	RegulationGDPR    Regulation = "GDPR"
	RegulationUKGDPR  Regulation = "UK_GDPR"
	RegulationFADP    Regulation = "FADP"
	RegulationCCPA    Regulation = "CCPA"
	RegulationPIPEDA  Regulation = "PIPEDA"
	RegulationLGPD    Regulation = "LGPD"
	RegulationLFPDPPP Regulation = "LFPDPPP"
	RegulationPOPIA   Regulation = "POPIA"
	RegulationPDPA    Regulation = "PDPA"
	RegulationPIPL    Regulation = "PIPL"
	RegulationPIPA    Regulation = "PIPA"
	RegulationAPPI    Regulation = "APPI"
	RegulationDPDP    Regulation = "DPDP"
	RegulationPDPL    Regulation = "PDPL"
)

func Regulations() []Regulation {
	return []Regulation{
		RegulationGDPR,
		RegulationUKGDPR,
		RegulationFADP,
		RegulationCCPA,
		RegulationPIPEDA,
		RegulationLGPD,
		RegulationLFPDPPP,
		RegulationPOPIA,
		RegulationPDPA,
		RegulationPIPL,
		RegulationPIPA,
		RegulationAPPI,
		RegulationDPDP,
		RegulationPDPL,
	}
}

func ParseRegulation(s string) (Regulation, error) {
	switch Regulation(s) {
	case RegulationNone:
		return RegulationNone, nil
	case RegulationGDPR:
		return RegulationGDPR, nil
	case RegulationUKGDPR:
		return RegulationUKGDPR, nil
	case RegulationFADP:
		return RegulationFADP, nil
	case RegulationCCPA:
		return RegulationCCPA, nil
	case RegulationPIPEDA:
		return RegulationPIPEDA, nil
	case RegulationLGPD:
		return RegulationLGPD, nil
	case RegulationLFPDPPP:
		return RegulationLFPDPPP, nil
	case RegulationPOPIA:
		return RegulationPOPIA, nil
	case RegulationPDPA:
		return RegulationPDPA, nil
	case RegulationPIPL:
		return RegulationPIPL, nil
	case RegulationPIPA:
		return RegulationPIPA, nil
	case RegulationAPPI:
		return RegulationAPPI, nil
	case RegulationDPDP:
		return RegulationDPDP, nil
	case RegulationPDPL:
		return RegulationPDPL, nil
	default:
		return "", fmt.Errorf("invalid Regulation value: %q", s)
	}
}

func (r Regulation) String() string {
	return string(r)
}

func (r *Regulation) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []byte:
		v = string(val)
	default:
		return fmt.Errorf("unsupported type for Regulation: %T", value)
	}

	parsed, err := ParseRegulation(v)
	if err != nil {
		return err
	}

	*r = parsed
	return nil
}

func (r Regulation) Value() (driver.Value, error) {
	if r == RegulationNone {
		return "", nil
	}

	if _, err := ParseRegulation(string(r)); err != nil {
		return nil, fmt.Errorf("invalid Regulation: %s", r)
	}

	return string(r), nil
}

func (r Regulation) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

func (r *Regulation) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal Regulation: %w", err)
	}

	parsed, err := ParseRegulation(s)
	if err != nil {
		return err
	}

	*r = parsed
	return nil
}
