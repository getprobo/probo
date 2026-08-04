// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	RegulationVCDPA   Regulation = "VCDPA"
	RegulationCPA     Regulation = "CPA"
	RegulationCTDPA   Regulation = "CTDPA"
	RegulationUCPA    Regulation = "UCPA"
	RegulationTDPSA   Regulation = "TDPSA"
	RegulationOCPA    Regulation = "OCPA"
	RegulationMTCDPA  Regulation = "MTCDPA"
	RegulationFDBR    Regulation = "FDBR"
	RegulationIAICDPA Regulation = "IA_ICDPA"
	RegulationDEPDPA  Regulation = "DE_DPDPA"
	RegulationNHPA    Regulation = "NHPA"
	RegulationNENDPA  Regulation = "NE_NDPA"
	RegulationNJDPA   Regulation = "NJDPA"
	RegulationTIPA    Regulation = "TIPA"
	RegulationMNDPA   Regulation = "MNDPA"
	RegulationMODPA   Regulation = "MODPA"
	RegulationINCDPA  Regulation = "IN_ICDPA"
	RegulationKCDPA   Regulation = "KCDPA"
	RegulationRIDTPPA Regulation = "RIDTPPA"
	RegulationPIPEDA  Regulation = "PIPEDA"
	RegulationPIPACA  Regulation = "PIPA_CA"
	RegulationLaw25   Regulation = "LAW_25"
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
		RegulationVCDPA,
		RegulationCPA,
		RegulationCTDPA,
		RegulationUCPA,
		RegulationTDPSA,
		RegulationOCPA,
		RegulationMTCDPA,
		RegulationFDBR,
		RegulationIAICDPA,
		RegulationDEPDPA,
		RegulationNHPA,
		RegulationNENDPA,
		RegulationNJDPA,
		RegulationTIPA,
		RegulationMNDPA,
		RegulationMODPA,
		RegulationINCDPA,
		RegulationKCDPA,
		RegulationRIDTPPA,
		RegulationPIPEDA,
		RegulationPIPACA,
		RegulationLaw25,
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
		RegulationVCDPA,
		RegulationCPA,
		RegulationCTDPA,
		RegulationUCPA,
		RegulationTDPSA,
		RegulationOCPA,
		RegulationMTCDPA,
		RegulationFDBR,
		RegulationIAICDPA,
		RegulationDEPDPA,
		RegulationNHPA,
		RegulationNENDPA,
		RegulationNJDPA,
		RegulationTIPA,
		RegulationMNDPA,
		RegulationMODPA,
		RegulationINCDPA,
		RegulationKCDPA,
		RegulationRIDTPPA,
		RegulationPIPEDA,
		RegulationPIPACA,
		RegulationLaw25,
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
	case RegulationVCDPA:
		return RegulationVCDPA, nil
	case RegulationCPA:
		return RegulationCPA, nil
	case RegulationCTDPA:
		return RegulationCTDPA, nil
	case RegulationUCPA:
		return RegulationUCPA, nil
	case RegulationTDPSA:
		return RegulationTDPSA, nil
	case RegulationOCPA:
		return RegulationOCPA, nil
	case RegulationMTCDPA:
		return RegulationMTCDPA, nil
	case RegulationFDBR:
		return RegulationFDBR, nil
	case RegulationIAICDPA:
		return RegulationIAICDPA, nil
	case RegulationDEPDPA:
		return RegulationDEPDPA, nil
	case RegulationNHPA:
		return RegulationNHPA, nil
	case RegulationNENDPA:
		return RegulationNENDPA, nil
	case RegulationNJDPA:
		return RegulationNJDPA, nil
	case RegulationTIPA:
		return RegulationTIPA, nil
	case RegulationMNDPA:
		return RegulationMNDPA, nil
	case RegulationMODPA:
		return RegulationMODPA, nil
	case RegulationINCDPA:
		return RegulationINCDPA, nil
	case RegulationKCDPA:
		return RegulationKCDPA, nil
	case RegulationRIDTPPA:
		return RegulationRIDTPPA, nil
	case RegulationPIPEDA:
		return RegulationPIPEDA, nil
	case RegulationPIPACA:
		return RegulationPIPACA, nil
	case RegulationLaw25:
		return RegulationLaw25, nil
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
