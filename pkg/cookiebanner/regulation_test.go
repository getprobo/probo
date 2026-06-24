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
