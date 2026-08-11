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
	RegulationVCDPA   = coredata.RegulationVCDPA
	RegulationCPA     = coredata.RegulationCPA
	RegulationCTDPA   = coredata.RegulationCTDPA
	RegulationUCPA    = coredata.RegulationUCPA
	RegulationTDPSA   = coredata.RegulationTDPSA
	RegulationOCPA    = coredata.RegulationOCPA
	RegulationMTCDPA  = coredata.RegulationMTCDPA
	RegulationFDBR    = coredata.RegulationFDBR
	RegulationIAICDPA = coredata.RegulationIAICDPA
	RegulationDEPDPA  = coredata.RegulationDEPDPA
	RegulationNHPA    = coredata.RegulationNHPA
	RegulationNENDPA  = coredata.RegulationNENDPA
	RegulationNJDPA   = coredata.RegulationNJDPA
	RegulationTIPA    = coredata.RegulationTIPA
	RegulationMNDPA   = coredata.RegulationMNDPA
	RegulationMODPA   = coredata.RegulationMODPA
	RegulationINCDPA  = coredata.RegulationINCDPA
	RegulationKCDPA   = coredata.RegulationKCDPA
	RegulationRIDTPPA = coredata.RegulationRIDTPPA
	RegulationPIPEDA  = coredata.RegulationPIPEDA
	RegulationPIPAAB  = coredata.RegulationPIPAAB
	RegulationPIPABC  = coredata.RegulationPIPABC
	RegulationPIPACA  = coredata.RegulationPIPACA
	RegulationLaw25   = coredata.RegulationLaw25
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

// usPrivacyRegulationBySubdivision maps ISO 3166-2 US subdivisions with
// comprehensive state privacy laws to their regulation identifier.
//
// Arkansas is intentionally omitted: the proposed Consumer Data Protection Act
// was never enacted (see arkleg.state.ar.us ISP-2023-003).
//
// Upcoming (signed, not yet effective — do not map until effective date):
//   - US-OK — OCDPA (SB 546), effective 2027-01-01
//   - US-LA — LDPA (SB 386), effective 2027-01-01
//   - US-AL — APDPA (HB 351), effective 2027-05-01 (verify)
var usPrivacyRegulationBySubdivision = map[coredata.SubdivisionCode]Regulation{
	"US-CA": RegulationCCPA,
	"US-VA": RegulationVCDPA,
	"US-CO": RegulationCPA,
	"US-CT": RegulationCTDPA,
	"US-UT": RegulationUCPA,
	"US-TX": RegulationTDPSA,
	"US-OR": RegulationOCPA,
	"US-MT": RegulationMTCDPA,
	"US-FL": RegulationFDBR,
	"US-IA": RegulationIAICDPA,
	"US-DE": RegulationDEPDPA,
	"US-NH": RegulationNHPA,
	"US-NE": RegulationNENDPA,
	"US-NJ": RegulationNJDPA,
	"US-TN": RegulationTIPA,
	"US-MN": RegulationMNDPA,
	"US-MD": RegulationMODPA,
	"US-IN": RegulationINCDPA,
	"US-KY": RegulationKCDPA,
	"US-RI": RegulationRIDTPPA,
}

// caPrivacyRegulationBySubdivision maps ISO 3166-2 Canadian subdivisions to
// their applicable private-sector privacy regulation.
//
// Quebec Law 25 requires opt-in consent for cookies/tracking. Alberta and
// British Columbia have substantially similar provincial PIPA statutes.
// All other provinces and territories fall back to the federal PIPEDA.
var caPrivacyRegulationBySubdivision = map[coredata.SubdivisionCode]Regulation{
	"CA-QC": RegulationLaw25,
	"CA-AB": RegulationPIPAAB,
	"CA-BC": RegulationPIPABC,
	"CA-MB": RegulationPIPEDA,
	"CA-NB": RegulationPIPEDA,
	"CA-NL": RegulationPIPEDA,
	"CA-NS": RegulationPIPEDA,
	"CA-NT": RegulationPIPEDA,
	"CA-NU": RegulationPIPEDA,
	"CA-ON": RegulationPIPEDA,
	"CA-PE": RegulationPIPEDA,
	"CA-SK": RegulationPIPEDA,
	"CA-YT": RegulationPIPEDA,
}

// IsCanadianOptOutPrivacyRegulation reports whether the regulation is a
// Canadian opt-out privacy regime (PIPEDA or provincial PIPA outside Quebec).
func IsCanadianOptOutPrivacyRegulation(r Regulation) bool {
	switch r {
	case RegulationPIPEDA,
		RegulationPIPAAB,
		RegulationPIPABC,
		RegulationPIPACA:
		return true
	}

	return false
}

// IsCaliforniaPrivacyRegulation reports whether the regulation is California
// CCPA/CPRA, which mandates the statutory Privacy Choices affordance (11 CCR § 7015).
func IsCaliforniaPrivacyRegulation(r Regulation) bool {
	return r == RegulationCCPA
}

// IsUSStatePrivacyRegulation reports whether the regulation is a US comprehensive
// state privacy law (including California).
func IsUSStatePrivacyRegulation(r Regulation) bool {
	switch r {
	case RegulationCCPA,
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
		RegulationRIDTPPA:
		return true
	}

	return false
}

// ResolveRegulation returns the regulation to apply for a visitor along
// with its source.
//
// When the country is positively identified it returns that country's
// regulation as detected, including RegulationNone for jurisdictions with no
// cookie-consent law (which the presentation layer maps to a hidden opt-out
// banner reachable only from the settings link). When geolocation is
// unresolved (location is nil) it falls back to GDPR, applying the strictest
// opt-in consent model by default.
func ResolveRegulation(location *coredata.IPLocationBlock) (Regulation, RegulationSource) {
	if location != nil {
		return RegulationForLocation(*location), RegulationSourceDetected
	}

	return RegulationGDPR, RegulationSourceDefault
}

// RegulationForLocation maps a country and optional ISO 3166-2 subdivision to
// the applicable privacy regulation.
//
// A US visitor whose subdivision cannot be resolved falls back to CCPA, the
// strictest US state regime: we cannot rule out California, and the statutory
// "Your Privacy Choices" link it adds is the safe answer for an unknown state.
// A subdivision that resolves but is absent from the map is a different case —
// that state genuinely has no comprehensive privacy law, so it maps to
// RegulationNone.
func RegulationForLocation(location coredata.IPLocationBlock) Regulation {
	if location.CountryCode == coredata.CountryCodeUS {
		if location.SubdivisionCode == nil {
			return RegulationCCPA
		}

		if regulation, ok := usPrivacyRegulationBySubdivision[*location.SubdivisionCode]; ok {
			return regulation
		}

		return RegulationNone
	}

	if location.CountryCode == coredata.CountryCodeCA {
		if location.SubdivisionCode == nil {
			return RegulationPIPEDA
		}

		if regulation, ok := caPrivacyRegulationBySubdivision[*location.SubdivisionCode]; ok {
			return regulation
		}

		return RegulationPIPEDA
	}

	return RegulationForCountry(location.CountryCode)
}

// RegulationForCountry maps a country code to the applicable privacy
// regulation. For countries with no known cookie-consent regulation it
// returns RegulationNone.
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
		RegulationLaw25,
		RegulationPOPIA,
		RegulationPDPA,
		RegulationPIPL,
		RegulationPIPA,
		RegulationDPDP,
		RegulationPDPL,
		RegulationLGPD:
		return ConsentModeOptIn

	case RegulationCCPA,
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
		RegulationPIPAAB,
		RegulationPIPABC,
		RegulationPIPACA,
		RegulationLFPDPPP,
		RegulationAPPI:
		return ConsentModeOptOut

	default:
		return ConsentModeOptOut
	}
}

// applyUSStatePrivacyBannerTexts selects US-specific opt-out copy for
// non-California state privacy regulations.
func applyUSStatePrivacyBannerTexts(config *BannerConfig) {
	if config == nil || config.Texts == nil {
		return
	}

	if !IsUSStatePrivacyRegulation(config.Regulation) || IsCaliforniaPrivacyRegulation(config.Regulation) {
		return
	}

	if description, ok := config.Texts["banner_description_us_opt_out"]; ok && description != "" {
		config.Texts["banner_description_opt_out"] = description
	}
}

// applyCanadianPrivacyBannerTexts selects Canadian-specific opt-out copy for
// PIPEDA and provincial PIPA regulations (excluding Quebec Law 25).
func applyCanadianPrivacyBannerTexts(config *BannerConfig) {
	if config == nil || config.Texts == nil {
		return
	}

	if !IsCanadianOptOutPrivacyRegulation(config.Regulation) {
		return
	}

	if description, ok := config.Texts["banner_description_ca_opt_out"]; ok && description != "" {
		config.Texts["banner_description_opt_out"] = description
	}
}

// applyCCPABannerTexts selects CCPA statutory opt-out copy for California.
// button_opt_out holds the generic CTA; button_opt_out_ccpa holds the
// 11 CCR § 7015 "Do Not Sell or Share…" wording and is folded into the
// canonical key only for California. Privacy Choices keys are stored with a
// _ccpa suffix; unsuffixed aliases are filled so older SDKs that still read
// privacy_choices_* keep working.
func applyCCPABannerTexts(config *BannerConfig) {
	if config == nil || config.Texts == nil {
		return
	}

	if !IsCaliforniaPrivacyRegulation(config.Regulation) {
		return
	}

	if label, ok := config.Texts["button_opt_out_ccpa"]; ok && label != "" {
		config.Texts["button_opt_out"] = label
	}

	for _, key := range []string{
		"privacy_choices_title",
		"privacy_choices_intro",
		"privacy_choices_sale_title",
		"privacy_choices_sale_description",
		"privacy_choices_spi_title",
		"privacy_choices_spi_description",
	} {
		ccpaKey := key + "_ccpa"
		if v := config.Texts[ccpaKey]; v != "" {
			if config.Texts[key] == "" {
				config.Texts[key] = v
			}
		} else if v := config.Texts[key]; v != "" {
			config.Texts[ccpaKey] = v
		}
	}
}
