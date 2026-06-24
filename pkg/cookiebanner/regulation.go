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

package cookiebanner

import "go.probo.inc/probo/pkg/coredata"

type Regulation = coredata.Regulation

const (
	RegulationNone    = coredata.RegulationNone
	RegulationGDPR    = coredata.RegulationGDPR
	RegulationUKGDPR  = coredata.RegulationUKGDPR
	RegulationFADP    = coredata.RegulationFADP
	RegulationCCPA    = coredata.RegulationCCPA
	RegulationPIPEDA  = coredata.RegulationPIPEDA
	RegulationLGPD    = coredata.RegulationLGPD
	RegulationLFPDPPP = coredata.RegulationLFPDPPP
	RegulationPOPIA   = coredata.RegulationPOPIA
	RegulationPDPA    = coredata.RegulationPDPA
	RegulationPIPL    = coredata.RegulationPIPL
	RegulationPIPA    = coredata.RegulationPIPA
	RegulationAPPI    = coredata.RegulationAPPI
	RegulationDPDP    = coredata.RegulationDPDP
	RegulationPDPL    = coredata.RegulationPDPL
)

type RegulationSource = coredata.RegulationSource

const (
	RegulationSourceDetected = coredata.RegulationSourceDetected
	RegulationSourceDefault  = coredata.RegulationSourceDefault
)

const (
	ConsentModeOptIn  = "OPT_IN"
	ConsentModeOptOut = "OPT_OUT"
)

// ResolveRegulation returns the regulation to apply for a visitor along
// with its source. It defaults to GDPR when geolocation is unresolved
// (cc is nil) or when the resolved country maps to no known regulation,
// ensuring the strictest opt-in consent model applies by default.
func ResolveRegulation(cc *coredata.CountryCode) (Regulation, RegulationSource) {
	if cc != nil {
		if reg := RegulationForCountry(*cc); reg != RegulationNone {
			return reg, RegulationSourceDetected
		}
	}

	return RegulationGDPR, RegulationSourceDefault
}

// RegulationForCountry maps a country code to the applicable privacy
// regulation. For countries with no known cookie-consent regulation it
// returns RegulationNone.
//
// US states (CCPA/CPRA, CPA, VCDPA, UCPA) and Canadian provinces
// (PIPEDA, Law 25) are collapsed to the country level because IP
// geolocation only resolves to a country code.
func RegulationForCountry(cc coredata.CountryCode) Regulation {
	switch cc {
	// EU 27 member states
	case
		coredata.CountryCodeAT, // Austria
		coredata.CountryCodeBE, // Belgium
		coredata.CountryCodeBG, // Bulgaria
		coredata.CountryCodeHR, // Croatia
		coredata.CountryCodeCY, // Cyprus
		coredata.CountryCodeCZ, // Czechia
		coredata.CountryCodeDK, // Denmark
		coredata.CountryCodeEE, // Estonia
		coredata.CountryCodeFI, // Finland
		coredata.CountryCodeFR, // France
		coredata.CountryCodeDE, // Germany
		coredata.CountryCodeGR, // Greece
		coredata.CountryCodeHU, // Hungary
		coredata.CountryCodeIE, // Ireland
		coredata.CountryCodeIT, // Italy
		coredata.CountryCodeLV, // Latvia
		coredata.CountryCodeLT, // Lithuania
		coredata.CountryCodeLU, // Luxembourg
		coredata.CountryCodeMT, // Malta
		coredata.CountryCodeNL, // Netherlands
		coredata.CountryCodePL, // Poland
		coredata.CountryCodePT, // Portugal
		coredata.CountryCodeRO, // Romania
		coredata.CountryCodeSK, // Slovakia
		coredata.CountryCodeSI, // Slovenia
		coredata.CountryCodeES, // Spain
		coredata.CountryCodeSE, // Sweden
		// EEA (non-EU)
		coredata.CountryCodeIS, // Iceland
		coredata.CountryCodeLI, // Liechtenstein
		coredata.CountryCodeNO: // Norway
		return RegulationGDPR

	case coredata.CountryCodeGB:
		return RegulationUKGDPR

	case coredata.CountryCodeCH:
		return RegulationFADP

	case coredata.CountryCodeUS:
		return RegulationCCPA

	case coredata.CountryCodeCA:
		return RegulationPIPEDA

	case coredata.CountryCodeBR:
		return RegulationLGPD

	case coredata.CountryCodeMX:
		return RegulationLFPDPPP

	case coredata.CountryCodeZA:
		return RegulationPOPIA

	case coredata.CountryCodeTH:
		return RegulationPDPA

	case coredata.CountryCodeCN:
		return RegulationPIPL

	case coredata.CountryCodeKR:
		return RegulationPIPA

	case coredata.CountryCodeJP:
		return RegulationAPPI

	case coredata.CountryCodeIN:
		return RegulationDPDP

	case coredata.CountryCodeSA:
		return RegulationPDPL

	default:
		return RegulationNone
	}
}

// ConsentModeForRegulation returns the consent model implied by a
// regulation. OPT_IN means non-necessary cookies must be blocked until
// the visitor gives explicit consent; OPT_OUT means cookies may fire
// immediately but the visitor must be offered a way to opt out.
//
// When the regulation is unknown or RegulationNone, it defaults to
// OPT_OUT (cookies may fire immediately, visitor can opt out).
func ConsentModeForRegulation(r Regulation) string {
	switch r {
	case RegulationGDPR,
		RegulationUKGDPR,
		RegulationFADP,
		RegulationPOPIA,
		RegulationPDPA,
		RegulationPIPL,
		RegulationPIPA,
		RegulationDPDP,
		RegulationPDPL:
		return ConsentModeOptIn

	case RegulationCCPA,
		RegulationPIPEDA,
		RegulationLGPD,
		RegulationLFPDPPP,
		RegulationAPPI:
		return ConsentModeOptOut

	default:
		return ConsentModeOptOut
	}
}
