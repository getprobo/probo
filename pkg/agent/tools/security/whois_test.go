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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWhoisField(t *testing.T) {
	t.Parallel()

	t.Run(
		"extracts known field",
		func(t *testing.T) {
			t.Parallel()

			raw := "refer: whois.verisign-grs.com\nstatus: ACTIVE\n"
			assert.Equal(t, "whois.verisign-grs.com", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"returns first match",
		func(t *testing.T) {
			t.Parallel()

			raw := "refer: first.example.com\nrefer: second.example.com\n"
			assert.Equal(t, "first.example.com", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"handles missing field",
		func(t *testing.T) {
			t.Parallel()

			raw := "status: ACTIVE\ncreated: 2020-01-01\n"
			assert.Equal(t, "", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"handles empty input",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "", parseWhoisField("", "refer"))
		},
	)

	t.Run(
		"case insensitive field matching",
		func(t *testing.T) {
			t.Parallel()

			raw := "Refer: whois.example.com\n"
			assert.Equal(t, "whois.example.com", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"case insensitive field name argument",
		func(t *testing.T) {
			t.Parallel()

			raw := "refer: whois.example.com\n"
			assert.Equal(t, "whois.example.com", parseWhoisField(raw, "REFER"))
		},
	)

	t.Run(
		"skips comment lines",
		func(t *testing.T) {
			t.Parallel()

			raw := "% This is a comment\n# Another comment\nrefer: whois.example.com\n"
			assert.Equal(t, "whois.example.com", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"skips lines without colon",
		func(t *testing.T) {
			t.Parallel()

			raw := "no colon here\nrefer: whois.example.com\n"
			assert.Equal(t, "whois.example.com", parseWhoisField(raw, "refer"))
		},
	)

	t.Run(
		"trims whitespace around key and value",
		func(t *testing.T) {
			t.Parallel()

			raw := "  refer :  whois.example.com  \n"
			assert.Equal(t, "whois.example.com", parseWhoisField(raw, "refer"))
		},
	)
}

func TestParseWhoisResponse(t *testing.T) {
	t.Parallel()

	t.Run(
		"parses full realistic response",
		func(t *testing.T) {
			t.Parallel()

			raw := `Domain Name: EXAMPLE.COM
Registrar: Example Registrar, Inc.
Sponsoring Registrar: Another Registrar
Creation Date: 2005-03-15T00:00:00Z
Registry Expiry Date: 2030-03-15T00:00:00Z
Updated Date: 2024-01-10T12:00:00Z
Registrant Organization: Example Corp
Registrant Country: US
Name Server: ns1.example.com
Name Server: ns2.example.com
`
			result := parseWhoisResponse(raw)

			assert.Equal(t, "Example Registrar, Inc.", result.Registrar)
			assert.Equal(t, "2005-03-15T00:00:00Z", result.CreationDate)
			assert.Equal(t, "2030-03-15T00:00:00Z", result.ExpiryDate)
			assert.Equal(t, "2024-01-10T12:00:00Z", result.UpdatedDate)
			assert.Equal(t, "Example Corp", result.RegistrantOrg)
			assert.Equal(t, "US", result.RegistrantCC)
			require.Len(t, result.NameServers, 2)
			assert.Equal(t, "ns1.example.com", result.NameServers[0])
			assert.Equal(t, "ns2.example.com", result.NameServers[1])
		},
	)

	t.Run(
		"uses first value for duplicate fields",
		func(t *testing.T) {
			t.Parallel()

			raw := `Registrar: First Registrar
Registrar: Second Registrar
Creation Date: 2005-01-01
Creation Date: 2010-01-01
`
			result := parseWhoisResponse(raw)

			assert.Equal(t, "First Registrar", result.Registrar)
			assert.Equal(t, "2005-01-01", result.CreationDate)
		},
	)

	t.Run(
		"accumulates all name servers",
		func(t *testing.T) {
			t.Parallel()

			raw := `Name Server: NS1.EXAMPLE.COM
Name Server: NS2.EXAMPLE.COM
Name Server: NS3.EXAMPLE.COM
`
			result := parseWhoisResponse(raw)

			require.Len(t, result.NameServers, 3)
			assert.Equal(t, "ns1.example.com", result.NameServers[0])
			assert.Equal(t, "ns2.example.com", result.NameServers[1])
			assert.Equal(t, "ns3.example.com", result.NameServers[2])
		},
	)

	t.Run(
		"maps alternative field names",
		func(t *testing.T) {
			t.Parallel()

			raw := `Registrar Name: Alt Registrar
Created: 2010-06-01
Paid-Till: 2030-06-01
Last Modified: 2024-06-01
Registrant Organisation: Alt Org
nserver: ns1.alt.com
`
			result := parseWhoisResponse(raw)

			assert.Equal(t, "Alt Registrar", result.Registrar)
			assert.Equal(t, "2010-06-01", result.CreationDate)
			assert.Equal(t, "2030-06-01", result.ExpiryDate)
			assert.Equal(t, "2024-06-01", result.UpdatedDate)
			assert.Equal(t, "Alt Org", result.RegistrantOrg)
			require.Len(t, result.NameServers, 1)
			assert.Equal(t, "ns1.alt.com", result.NameServers[0])
		},
	)

	t.Run(
		"empty input returns zero value",
		func(t *testing.T) {
			t.Parallel()

			result := parseWhoisResponse("")

			assert.Equal(t, "", result.Registrar)
			assert.Equal(t, "", result.CreationDate)
			assert.Equal(t, "", result.ExpiryDate)
			assert.Equal(t, "", result.UpdatedDate)
			assert.Equal(t, "", result.RegistrantOrg)
			assert.Equal(t, "", result.RegistrantCC)
			assert.Nil(t, result.NameServers)
		},
	)

	t.Run(
		"skips comment and blank lines",
		func(t *testing.T) {
			t.Parallel()

			raw := `% WHOIS server comment
# Another comment

Registrar: Good Registrar

Creation Date: 2020-01-01
`
			result := parseWhoisResponse(raw)

			assert.Equal(t, "Good Registrar", result.Registrar)
			assert.Equal(t, "2020-01-01", result.CreationDate)
		},
	)

	t.Run(
		"skips lines with empty values",
		func(t *testing.T) {
			t.Parallel()

			raw := `Registrar:
Registrar: Actual Registrar
`
			result := parseWhoisResponse(raw)

			assert.Equal(t, "Actual Registrar", result.Registrar)
		},
	)

	t.Run(
		"handles extra whitespace around keys and values",
		func(t *testing.T) {
			t.Parallel()

			raw := "  Registrar :  Spaced Registrar  \n  Creation Date :  2023-05-01  \n"
			result := parseWhoisResponse(raw)

			assert.Equal(t, "Spaced Registrar", result.Registrar)
			assert.Equal(t, "2023-05-01", result.CreationDate)
		},
	)
}
