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

package vetting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestParseCountryLocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw      string
		expected coredata.CountryCode
	}{
		{raw: "US", expected: coredata.CountryCodeUS},
		{raw: "usa", expected: coredata.CountryCodeUS},
		{raw: "United States", expected: coredata.CountryCodeUS},
		{raw: "Seattle, Washington, USA", expected: coredata.CountryCodeUS},
		{raw: "Global presence", expected: coredata.CountryCodeGlobal},
		{raw: "EU", expected: coredata.CountryCodeEU},
		{raw: "Germany", expected: coredata.CountryCodeDE},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			code, ok := parseCountryLocation(tt.raw)
			assert.True(t, ok)
			assert.Equal(t, tt.expected, code)
		})
	}
}

func TestCountriesFromInfo(t *testing.T) {
	t.Parallel()

	countries := countriesFromInfo(ThirdPartyInfo{
		HeadquarterAddress: "Seattle, Washington, USA",
		DataLocations:      []string{"Germany", "EU"},
	})

	assert.Equal(
		t,
		coredata.CountryCodes{
			coredata.CountryCodeDE,
			coredata.CountryCodeEU,
			coredata.CountryCodeUS,
		},
		countries,
	)
}

func TestParseOptionalCountryCodes(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		coredata.CountryCodes{coredata.CountryCodeFR},
		parseOptionalCountryCodes("France"),
	)
}
