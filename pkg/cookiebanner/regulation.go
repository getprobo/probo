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

package cookiebanner

import "go.probo.inc/probo/pkg/coredata"

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

const (
	ConsentModeOptIn  = "OPT_IN"
	ConsentModeOptOut = "OPT_OUT"
)

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
// When the regulation is unknown or RegulationNone, it returns an empty
// string so the caller can fall back to the banner's configured default.
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
		return ""
	}
}
