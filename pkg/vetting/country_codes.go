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

package vetting

import (
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

var countryAliases = map[string]coredata.CountryCode{
	"global":            coredata.CountryCodeGlobal,
	"global presence":   coredata.CountryCodeGlobal,
	"worldwide":         coredata.CountryCodeGlobal,
	"international":     coredata.CountryCodeGlobal,
	"multiple regions":  coredata.CountryCodeGlobal,
	"eu":                coredata.CountryCodeEU,
	"european union":    coredata.CountryCodeEU,
	"europe":            coredata.CountryCodeEU,
	"united states":     coredata.CountryCodeUS,
	"united states usa": coredata.CountryCodeUS,
	"usa":               coredata.CountryCodeUS,
	"u.s.":              coredata.CountryCodeUS,
	"u.s.a.":            coredata.CountryCodeUS,
	"us":                coredata.CountryCodeUS,
	"united kingdom":    coredata.CountryCodeGB,
	"uk":                coredata.CountryCodeGB,
	"great britain":     coredata.CountryCodeGB,
	"germany":           coredata.CountryCodeDE,
	"france":            coredata.CountryCodeFR,
	"canada":            coredata.CountryCodeCA,
	"australia":         coredata.CountryCodeAU,
	"japan":             coredata.CountryCodeJP,
	"china":             coredata.CountryCodeCN,
	"india":             coredata.CountryCodeIN,
	"ireland":           coredata.CountryCodeIE,
	"netherlands":       coredata.CountryCodeNL,
	"singapore":         coredata.CountryCodeSG,
	"switzerland":       coredata.CountryCodeCH,
	"sweden":            coredata.CountryCodeSE,
	"spain":             coredata.CountryCodeES,
	"italy":             coredata.CountryCodeIT,
	"brazil":            coredata.CountryCodeBR,
	"mexico":            coredata.CountryCodeMX,
	"south korea":       coredata.CountryCodeKR,
	"korea":             coredata.CountryCodeKR,
}

func parseOptionalCountryCodes(raw string) coredata.CountryCodes {
	code, ok := parseCountryLocation(raw)
	if !ok {
		return nil
	}

	return coredata.CountryCodes{code}
}

func countriesFromInfo(info ThirdPartyInfo) coredata.CountryCodes {
	raw := append([]string{}, info.DataLocations...)
	if info.HeadquarterAddress != "" {
		raw = append(raw, info.HeadquarterAddress)
	}

	return parseCountryLocations(raw...)
}

func parseCountryLocations(raw ...string) coredata.CountryCodes {
	seen := make(map[coredata.CountryCode]struct{})
	out := make(coredata.CountryCodes, 0, len(raw))

	for _, value := range raw {
		for _, part := range splitCountryList(value) {
			code, ok := parseCountryLocation(part)
			if !ok {
				continue
			}

			if _, exists := seen[code]; exists {
				continue
			}

			seen[code] = struct{}{}
			out = append(out, code)
		}
	}

	return out
}

func parseCountryLocation(raw string) (coredata.CountryCode, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	code := coredata.CountryCode(strings.ToUpper(raw))
	if code.IsValid() {
		return code, true
	}

	if mapped, ok := countryAliases[normalizeCountryKey(raw)]; ok {
		return mapped, true
	}

	if strings.Contains(raw, ",") {
		parts := strings.Split(raw, ",")
		last := strings.TrimSpace(parts[len(parts)-1])

		if mapped, ok := countryAliases[normalizeCountryKey(last)]; ok {
			return mapped, true
		}

		lastCode := coredata.CountryCode(strings.ToUpper(last))
		if lastCode.IsValid() {
			return lastCode, true
		}
	}

	return "", false
}

func splitCountryList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	for _, sep := range []string{";", "|", "/", " and ", " & "} {
		if strings.Contains(strings.ToLower(raw), sep) {
			parts := strings.Split(raw, sep)
			out := make([]string, 0, len(parts))

			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					out = append(out, part)
				}
			}

			return out
		}
	}

	return []string{raw}
}

func normalizeCountryKey(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.TrimPrefix(raw, "the ")

	return strings.Join(strings.Fields(raw), " ")
}
