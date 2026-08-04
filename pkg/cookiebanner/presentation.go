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

// Presentation is the UI presentation policy of the banner. It is a separate
// concern from the consent mode: the consent mode (OPT_IN / OPT_OUT) is the
// legal tracker-firing model, while the presentation decides how the banner
// looks and behaves (which buttons, initial visibility, wording variant).
//
// A regulation maps to exactly one presentation; several regulations can share
// the same presentation.
type Presentation string

const (
	// PresentationOptIn blocks non-necessary trackers until the visitor gives
	// explicit consent and offers accept / reject / customize choices.
	PresentationOptIn Presentation = "OPT_IN"

	// PresentationOptOut lets trackers fire immediately but must offer the
	// visitor a way to opt out (e.g. CCPA "Do Not Sell or Share").
	PresentationOptOut Presentation = "OPT_OUT"

	// PresentationNotice is a purely informational banner (implied consent):
	// trackers fire immediately and the visitor only acknowledges the notice.
	PresentationNotice Presentation = "NOTICE"
)

// SettingsLinkStyle selects how the reopen ("settings") link renders.
type SettingsLinkStyle string

const (
	// SettingsLinkDefault renders the generic "Cookie settings" affordance.
	SettingsLinkDefault SettingsLinkStyle = "default"

	// SettingsLinkCCPAPrivacyChoices renders the statutory CCPA "Your Privacy
	// Choices" label and opt-out icon (11 CCR § 7015).
	SettingsLinkCCPAPrivacyChoices SettingsLinkStyle = "ccpa_privacy_choices"
)

// Banner states shared with the client. The client also has a transient
// "loading" state that never originates from the server.
const (
	StateBanner         = "banner"
	StateHidden         = "hidden"
	StatePanel          = "panel"
	StatePrivacyChoices = "privacy_choices"
)

// LayoutButtons declares which action buttons the banner renders. The client
// hides any button set to false rather than inferring visibility from empty
// text values.
type LayoutButtons struct {
	AcceptAll bool `json:"accept_all"`
	RejectAll bool `json:"reject_all"`
	Customize bool `json:"customize"`
	Save      bool `json:"save"`
}

// Layout is the explicit, structured rendering policy the client consumes. It
// replaces the implicit signalling (hardcoded regulation checks, blanked text
// keys) the client previously relied on. The wording the client renders is
// selected from Presentation; there is no separate text-variant axis.
type Layout struct {
	Presentation               Presentation      `json:"presentation"`
	InitialState               string            `json:"initial_state"`
	ReopenState                string            `json:"reopen_state"`
	DefaultNonNecessaryGranted bool              `json:"default_non_necessary_granted"`
	Buttons                    LayoutButtons     `json:"buttons"`
	SettingsLink               SettingsLinkStyle `json:"settings_link"`
}

// PresentationForRegulation maps a regulation to its banner presentation.
//
// Opt-out regimes with a statutory notice model (Japan's APPI, Mexico's
// LFPDPPP) and jurisdictions with no cookie-consent law (RegulationNone) use
// the informational notice presentation; the remaining opt-out regimes (CCPA,
// PIPEDA, Canadian PIPA, LGPD) use the opt-out presentation; everything else
// defaults to the strict opt-in presentation.
func PresentationForRegulation(r Regulation) Presentation {
	switch r {
	case RegulationCCPA,
		RegulationPIPEDA,
		RegulationPIPACA,
		RegulationLGPD:
		return PresentationOptOut

	case RegulationAPPI,
		RegulationLFPDPPP,
		RegulationNone:
		return PresentationNotice

	default:
		return PresentationOptIn
	}
}

// SettingsLinkForRegulation returns the reopen-link style for a regulation.
// Only CCPA mandates the statutory "Your Privacy Choices" affordance.
func SettingsLinkForRegulation(r Regulation) SettingsLinkStyle {
	if r == RegulationCCPA {
		return SettingsLinkCCPAPrivacyChoices
	}

	return SettingsLinkDefault
}

// LayoutForRegulation resolves the full rendering policy for a regulation.
func LayoutForRegulation(r Regulation) Layout {
	layout := layoutForPresentation(PresentationForRegulation(r))
	layout.SettingsLink = SettingsLinkForRegulation(r)

	// CCPA's Alternative Opt-out Link (11 CCR § 7015) must open a surface that
	// describes sale/sharing opt-out and the right to limit sensitive PI — not
	// the compact opt-out banner used by other OPT_OUT regimes.
	if r == RegulationCCPA {
		layout.ReopenState = StatePrivacyChoices
	}

	return layout
}

func layoutForPresentation(p Presentation) Layout {
	switch p {
	case PresentationOptOut:
		return Layout{
			Presentation:               PresentationOptOut,
			InitialState:               StateHidden,
			ReopenState:                StateBanner,
			DefaultNonNecessaryGranted: true,
			Buttons: LayoutButtons{
				AcceptAll: true,
				RejectAll: true,
				Customize: false,
				Save:      true,
			},
			SettingsLink: SettingsLinkDefault,
		}

	case PresentationNotice:
		return Layout{
			Presentation:               PresentationNotice,
			InitialState:               StateBanner,
			ReopenState:                StateBanner,
			DefaultNonNecessaryGranted: true,
			Buttons: LayoutButtons{
				AcceptAll: true,
				RejectAll: false,
				Customize: false,
				Save:      false,
			},
			SettingsLink: SettingsLinkDefault,
		}

	default:
		return Layout{
			Presentation:               PresentationOptIn,
			InitialState:               StateBanner,
			ReopenState:                StatePanel,
			DefaultNonNecessaryGranted: false,
			Buttons: LayoutButtons{
				AcceptAll: true,
				RejectAll: true,
				Customize: true,
				Save:      true,
			},
			SettingsLink: SettingsLinkDefault,
		}
	}
}
