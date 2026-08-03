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
		countryCode    *coredata.CountryCode
		wantRegulation Regulation
		wantSource     RegulationSource
	}{
		{
			name:           "unresolved geolocation defaults to GDPR",
			countryCode:    nil,
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDefault,
		},
		{
			name:           "identified country with no known regulation resolves to none as detected",
			countryCode:    new(coredata.CountryCodeAQ),
			wantRegulation: RegulationNone,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "EU country resolves to GDPR as detected",
			countryCode:    new(coredata.CountryCodeFR),
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "US resolves to CCPA as detected",
			countryCode:    new(coredata.CountryCodeUS),
			wantRegulation: RegulationCCPA,
			wantSource:     RegulationSourceDetected,
		},
		{
			name:           "UK resolves to UK GDPR as detected",
			countryCode:    new(coredata.CountryCodeGB),
			wantRegulation: RegulationUKGDPR,
			wantSource:     RegulationSourceDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regulation, source := ResolveRegulation(tt.countryCode)
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
		{RegulationPIPEDA, PresentationOptOut},
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
		require.Equal(t, StateBanner, layout.ReopenState)
		require.True(t, layout.DefaultNonNecessaryGranted)
		require.False(t, layout.Buttons.Customize)
		require.Equal(t, SettingsLinkCCPAPrivacyChoices, layout.SettingsLink)
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
