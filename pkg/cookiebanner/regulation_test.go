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

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestResolveRegulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		location       *coredata.IPLocationBlock
		wantRegulation Regulation
		wantSource     RegulationSource
	}{
		{
			name:           "unresolved geolocation defaults to GDPR",
			location:       nil,
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDefault,
		},
		{
			name:           "identified country with no known regulation resolves to none as detected",
			location:       &coredata.IPLocationBlock{CountryCode: coredata.CountryCodeAQ},
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "EU country resolves to GDPR as detected",
			location:       &coredata.IPLocationBlock{CountryCode: coredata.CountryCodeFR},
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "California resolves to CCPA as detected",
			location: &coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-CA")),
			},
			wantRegulation: RegulationCCPA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "Texas resolves to TDPSA as detected",
			location: &coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-TX")),
			},
			wantRegulation: RegulationTDPSA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "unregulated US state resolves to none as detected",
			location: &coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-NY")),
			},
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "US without subdivision falls back to CCPA as detected",
			location:       &coredata.IPLocationBlock{CountryCode: coredata.CountryCodeUS},
			wantRegulation: RegulationCCPA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "UK resolves to UK GDPR as detected",
			location:       &coredata.IPLocationBlock{CountryCode: coredata.CountryCodeGB},
			wantRegulation: RegulationUKGDPR,
			wantSource:     RegulationSourceDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regulation, source := ResolveRegulation(tt.location)
			require.Equal(t, tt.wantRegulation, regulation)
			require.Equal(t, tt.wantSource, source)
		})
	}
}

func TestPresentationForRegulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		regulation Regulation
		want       Presentation
	}{
		{RegulationGDPR, PresentationOptIn},
		{RegulationUKGDPR, PresentationOptIn},
		{RegulationFADP, PresentationOptIn},
		{RegulationPOPIA, PresentationOptIn},
		{RegulationCCPA, PresentationOptOut},
		{RegulationTDPSA, PresentationOptOut},
		{RegulationVCDPA, PresentationOptOut},
		{RegulationPIPEDA, PresentationOptOut},
		{RegulationPIPAAB, PresentationOptOut},
		{RegulationPIPABC, PresentationOptOut},
		{RegulationPIPACA, PresentationOptOut},
		{RegulationLaw25, PresentationOptIn},
		{RegulationLGPD, PresentationOptIn},
		{RegulationAPPI, PresentationOptOut},
		{RegulationNone, PresentationOptOut},
		{RegulationLFPDPPP, PresentationNotice},
	}

	for _, tt := range tests {
		t.Run(string(tt.regulation), func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, PresentationForRegulation(tt.regulation))
		})
	}
}

func TestLayoutForRegulation(t *testing.T) {
	t.Parallel()

	t.Run("opt-in regulation", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationGDPR)
		require.Equal(t, PresentationOptIn, layout.Presentation)
		require.Equal(t, StateBanner, layout.InitialState)
		require.Equal(t, StatePanel, layout.ReopenState)
		require.False(t, layout.DefaultNonNecessaryGranted)
		require.True(t, layout.Buttons.Customize)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("ccpa gets the statutory settings link", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationCCPA)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, StateHidden, layout.InitialState)
		require.Equal(t, StatePrivacyChoices, layout.ReopenState)
		require.True(t, layout.DefaultNonNecessaryGranted)
		require.False(t, layout.Buttons.Customize)
		require.Equal(t, SettingsLinkCCPAPrivacyChoices, layout.SettingsLink)
	})

	t.Run("other US state privacy law uses generic opt-out chrome", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationTDPSA)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, StateHidden, layout.InitialState)
		require.Equal(t, StateBanner, layout.ReopenState)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("Canadian opt-out regulation keeps the default settings link", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationPIPEDA)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, StateHidden, layout.InitialState)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("LGPD opts Brazil into the consent flow", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationLGPD)
		require.Equal(t, PresentationOptIn, layout.Presentation)
		require.Equal(t, StateBanner, layout.InitialState)
		require.Equal(t, StatePanel, layout.ReopenState)
		require.False(t, layout.DefaultNonNecessaryGranted)
	})

	t.Run("Japan stays hidden behind the settings link", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationAPPI)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, StateHidden, layout.InitialState)
		require.Equal(t, StateBanner, layout.ReopenState)
		require.True(t, layout.DefaultNonNecessaryGranted)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("no cookie-consent law shows nothing on first load", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationNone)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, StateHidden, layout.InitialState)
		require.Equal(t, StateBanner, layout.ReopenState)
		require.True(t, layout.DefaultNonNecessaryGranted)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("Mexico is the only notice regulation", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationLFPDPPP)
		require.Equal(t, PresentationNotice, layout.Presentation)
		require.Equal(t, StateBanner, layout.InitialState)
		require.Equal(t, StateBanner, layout.ReopenState)
		require.True(t, layout.DefaultNonNecessaryGranted)
		require.True(t, layout.Buttons.AcceptAll)
		require.False(t, layout.Buttons.RejectAll)
		require.False(t, layout.Buttons.Customize)
		require.False(t, layout.Buttons.Save)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})
}

// TestLayoutOpensOnlyWhereRequired pins the compliance invariant the mapping
// exists to enforce: the banner may only open on first load where a
// jurisdiction demands a proactive disclosure. Everywhere else the settings
// link the integrator places in the footer is the entire visible surface.
func TestLayoutOpensOnlyWhereRequired(t *testing.T) {
	t.Parallel()

	for _, regulation := range coredata.Regulations() {
		t.Run(string(regulation), func(t *testing.T) {
			t.Parallel()

			layout := LayoutForRegulation(regulation)

			switch layout.Presentation {
			case PresentationOptIn, PresentationNotice:
				require.Equal(t, StateBanner, layout.InitialState)
			default:
				require.Equal(t, StateHidden, layout.InitialState)
			}
		})
	}
}

// TestConsentModeMatchesPresentation guards against the two halves of the
// policy drifting apart: a banner that asks for consent while the server
// records an opt-out mode would fire trackers the visitor never allowed.
func TestConsentModeMatchesPresentation(t *testing.T) {
	t.Parallel()

	for _, regulation := range coredata.Regulations() {
		t.Run(string(regulation), func(t *testing.T) {
			t.Parallel()

			want := ConsentModeOptOut
			if PresentationForRegulation(regulation) == PresentationOptIn {
				want = ConsentModeOptIn
			}

			require.Equal(t, want, ConsentModeForRegulation(regulation))
			require.Equal(
				t,
				want == ConsentModeOptOut,
				LayoutForRegulation(regulation).DefaultNonNecessaryGranted,
			)
		})
	}
}

func TestRegulationForLocationUSPrivacyStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		subdivision coredata.SubdivisionCode
		want        Regulation
	}{
		{"US-CA", RegulationCCPA},
		{"US-VA", RegulationVCDPA},
		{"US-CO", RegulationCPA},
		{"US-CT", RegulationCTDPA},
		{"US-UT", RegulationUCPA},
		{"US-TX", RegulationTDPSA},
		{"US-OR", RegulationOCPA},
		{"US-MT", RegulationMTCDPA},
		{"US-FL", RegulationFDBR},
		{"US-IA", RegulationIAICDPA},
		{"US-DE", RegulationDEPDPA},
		{"US-NH", RegulationNHPA},
		{"US-NE", RegulationNENDPA},
		{"US-NJ", RegulationNJDPA},
		{"US-TN", RegulationTIPA},
		{"US-MN", RegulationMNDPA},
		{"US-MD", RegulationMODPA},
		{"US-IN", RegulationINCDPA},
		{"US-KY", RegulationKCDPA},
		{"US-RI", RegulationRIDTPPA},
	}

	for _, tt := range tests {
		t.Run(tt.subdivision.String(), func(t *testing.T) {
			t.Parallel()

			subdivision := tt.subdivision
			location := coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: &subdivision,
			}
			require.Equal(t, tt.want, RegulationForLocation(location))
			require.Equal(t, ConsentModeOptOut, ConsentModeForRegulation(tt.want))
		})
	}
}

func TestRegulationForLocationUSUnregulatedStates(t *testing.T) {
	t.Parallel()

	for _, subdivision := range []coredata.SubdivisionCode{"US-NY", "US-WA", "US-IL"} {
		t.Run(subdivision.String(), func(t *testing.T) {
			t.Parallel()

			code := subdivision
			location := coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: &code,
			}
			require.Equal(t, RegulationNone, RegulationForLocation(location))
		})
	}
}

func TestApplyUSStatePrivacyBannerTexts(t *testing.T) {
	t.Parallel()

	t.Run("non-California US state uses US opt-out copy", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationTDPSA,
			Texts: map[string]string{
				"banner_description_opt_out":    "generic",
				"banner_description_us_opt_out": "us specific",
			},
		}
		applyUSStatePrivacyBannerTexts(config)
		require.Equal(t, "us specific", config.Texts["banner_description_opt_out"])
	})

	t.Run("California keeps generic opt-out copy", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationCCPA,
			Texts: map[string]string{
				"banner_description_opt_out":    "generic",
				"banner_description_us_opt_out": "us specific",
			},
		}
		applyUSStatePrivacyBannerTexts(config)
		require.Equal(t, "generic", config.Texts["banner_description_opt_out"])
	})
}

func TestRegulationForLocationCanada(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		subdivision *coredata.SubdivisionCode
		want        Regulation
		wantMode    string
	}{
		{
			name: "Quebec uses Law 25",
			subdivision: new(
				coredata.SubdivisionCode("CA-QC"),
			),
			want:     RegulationLaw25,
			wantMode: ConsentModeOptIn,
		},
		{
			name: "Alberta uses provincial PIPA",
			subdivision: new(
				coredata.SubdivisionCode("CA-AB"),
			),
			want:     RegulationPIPAAB,
			wantMode: ConsentModeOptOut,
		},
		{
			name: "British Columbia uses provincial PIPA",
			subdivision: new(
				coredata.SubdivisionCode("CA-BC"),
			),
			want:     RegulationPIPABC,
			wantMode: ConsentModeOptOut,
		},
		{
			name: "Ontario uses PIPEDA",
			subdivision: new(
				coredata.SubdivisionCode("CA-ON"),
			),
			want:     RegulationPIPEDA,
			wantMode: ConsentModeOptOut,
		},
		{
			name:     "missing subdivision uses PIPEDA",
			want:     RegulationPIPEDA,
			wantMode: ConsentModeOptOut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			location := coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeCA,
				SubdivisionCode: tt.subdivision,
			}
			regulation := RegulationForLocation(location)
			require.Equal(t, tt.want, regulation)
			require.Equal(t, tt.wantMode, ConsentModeForRegulation(regulation))
		})
	}

	for subdivision, want := range caPrivacyRegulationBySubdivision {
		t.Run("all mapped provinces/"+subdivision.String(), func(t *testing.T) {
			t.Parallel()

			code := subdivision
			location := coredata.IPLocationBlock{
				CountryCode:     coredata.CountryCodeCA,
				SubdivisionCode: &code,
			}
			require.Equal(t, want, RegulationForLocation(location))
		})
	}
}

func TestApplyCanadianPrivacyBannerTexts(t *testing.T) {
	t.Parallel()

	t.Run("PIPEDA uses Canadian opt-out copy", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationPIPEDA,
			Texts: map[string]string{
				"banner_description_opt_out":    "generic",
				"banner_description_ca_opt_out": "ca specific",
			},
		}
		applyCanadianPrivacyBannerTexts(config)
		require.Equal(t, "ca specific", config.Texts["banner_description_opt_out"])
	})

	t.Run("Quebec keeps generic opt-out copy", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationLaw25,
			Texts: map[string]string{
				"banner_description_opt_out":    "generic",
				"banner_description_ca_opt_out": "ca specific",
			},
		}
		applyCanadianPrivacyBannerTexts(config)
		require.Equal(t, "generic", config.Texts["banner_description_opt_out"])
	})
}

func TestApplyGenericOptOutBannerTexts(t *testing.T) {
	t.Parallel()

	t.Run("Japan drops the California opt-out wording", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationAPPI,
			Texts: map[string]string{
				"button_opt_out":         "Do Not Sell or Share My Personal Information",
				"button_opt_out_generic": "Reject non-essential cookies",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Reject non-essential cookies", config.Texts["button_opt_out"])
	})

	t.Run("US state privacy law keeps the statutory wording", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationTDPSA,
			Texts: map[string]string{
				"button_opt_out":         "Do Not Sell or Share My Personal Information",
				"button_opt_out_generic": "Reject non-essential cookies",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Do Not Sell or Share My Personal Information", config.Texts["button_opt_out"])
	})

	t.Run("Mexico drops the wording its legacy chrome would show", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationLFPDPPP,
			Texts: map[string]string{
				"button_opt_out":         "Do Not Sell or Share My Personal Information",
				"button_opt_out_generic": "Reject non-essential cookies",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Reject non-essential cookies", config.Texts["button_opt_out"])
	})

	t.Run("banner predating the neutral label falls back to reject all", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationNone,
			Texts: map[string]string{
				"button_opt_out":    "Do Not Sell or Share My Personal Information",
				"button_reject_all": "Reject all",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Reject all", config.Texts["button_opt_out"])
	})

	t.Run("no replacement available leaves the texts untouched", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationNone,
			Texts: map[string]string{
				"button_opt_out": "Do Not Sell or Share My Personal Information",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Do Not Sell or Share My Personal Information", config.Texts["button_opt_out"])
	})

	t.Run("integrator copy survives", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationPIPEDA,
			Texts: map[string]string{
				"button_opt_out":         "Withdraw my consent",
				"button_opt_out_generic": "Reject non-essential cookies",
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(t, "Withdraw my consent", config.Texts["button_opt_out"])
	})

	t.Run("localized default is recognized as shipped copy", func(t *testing.T) {
		t.Parallel()

		config := &BannerConfig{
			Regulation: RegulationAPPI,
			Texts: map[string]string{
				"button_opt_out":         defaultUIStringsByLanguage["ja"]["button_opt_out"],
				"button_opt_out_generic": defaultUIStringsByLanguage["ja"]["button_opt_out_generic"],
			},
		}
		applyGenericOptOutBannerTexts(config)
		require.Equal(
			t,
			defaultUIStringsByLanguage["ja"]["button_opt_out_generic"],
			config.Texts["button_opt_out"],
		)
	})
}

// TestDefaultUIStringsHaveGenericOptOutLabel keeps the neutral label in step
// with the supported languages: without it, or with it copied from the
// California wording, a visitor outside the US sees "Do Not Sell or Share".
func TestDefaultUIStringsHaveGenericOptOutLabel(t *testing.T) {
	t.Parallel()

	for language, uiStrings := range defaultUIStringsByLanguage {
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, uiStrings["button_opt_out_generic"])
			require.NotEqual(
				t,
				uiStrings["button_opt_out"],
				uiStrings["button_opt_out_generic"],
				"the neutral label must differ from the California wording it replaces",
			)
		})
	}
}
