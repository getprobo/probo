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
