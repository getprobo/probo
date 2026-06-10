// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
			name:           "country with no known regulation defaults to GDPR",
			countryCode:    new(coredata.CountryCodeAQ),
			wantRegulation: RegulationGDPR,
			wantSource:     RegulationSourceDefault,
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
