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
		location       *coredata.IPLocation
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
			location:       &coredata.IPLocation{CountryCode: coredata.CountryCodeAQ},
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "EU country resolves to GDPR as detected",
			location:       &coredata.IPLocation{CountryCode: coredata.CountryCodeFR},
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "California resolves to CCPA as detected",
			location: &coredata.IPLocation{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-CA")),
			},
			wantRegulation: RegulationCCPA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "Texas resolves to TDPSA as detected",
			location: &coredata.IPLocation{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-TX")),
			},
			wantRegulation: RegulationTDPSA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name: "unregulated US state resolves to none as detected",
			location: &coredata.IPLocation{
				CountryCode:     coredata.CountryCodeUS,
				SubdivisionCode: new(coredata.SubdivisionCode("US-NY")),
			},
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "US without subdivision resolves to none as detected",
			location:       &coredata.IPLocation{CountryCode: coredata.CountryCodeUS},
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "UK resolves to UK GDPR as detected",
			location:       &coredata.IPLocation{CountryCode: coredata.CountryCodeGB},
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
		{RegulationPIPACA, PresentationOptOut},
		{RegulationLaw25, PresentationOptIn},
		{RegulationLGPD, PresentationOptOut},
		{RegulationAPPI, PresentationNotice},
		{RegulationLFPDPPP, PresentationNotice},
		{RegulationNone, PresentationNotice},
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

	t.Run("other opt-out regulation keeps the default settings link", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationLGPD)
		require.Equal(t, PresentationOptOut, layout.Presentation)
		require.Equal(t, SettingsLinkDefault, layout.SettingsLink)
	})

	t.Run("notice regulation", func(t *testing.T) {
		t.Parallel()

		layout := LayoutForRegulation(RegulationNone)
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
			location := coredata.IPLocation{
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
			location := coredata.IPLocation{
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
			name: "Alberta uses Canadian PIPA",
			subdivision: new(
				coredata.SubdivisionCode("CA-AB"),
			),
			want:     RegulationPIPACA,
			wantMode: ConsentModeOptOut,
		},
		{
			name: "British Columbia uses Canadian PIPA",
			subdivision: new(
				coredata.SubdivisionCode("CA-BC"),
			),
			want:     RegulationPIPACA,
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

			location := coredata.IPLocation{
				CountryCode:     coredata.CountryCodeCA,
				SubdivisionCode: tt.subdivision,
			}
			regulation := RegulationForLocation(location)
			require.Equal(t, tt.want, regulation)
			require.Equal(t, tt.wantMode, ConsentModeForRegulation(regulation))
		})
	}
}
