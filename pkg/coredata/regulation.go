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
	"encoding"
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

var (
	_ fmt.Stringer             = Regulation("")
	_ encoding.TextMarshaler   = Regulation("")
	_ encoding.TextUnmarshaler = (*Regulation)(nil)
)

func Regulations() []Regulation {
	return []Regulation{
		RegulationNone,
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

func (v Regulation) IsValid() bool {
	switch v {
	case
		RegulationNone,
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
		RegulationPDPL:
		return true
	}

	return false
}

func (v Regulation) String() string {
	return string(v)
}

func (v Regulation) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Regulation) UnmarshalText(text []byte) error {
	val := Regulation(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid Regulation value: %q", string(text))
	}

	*v = val

	return nil
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
