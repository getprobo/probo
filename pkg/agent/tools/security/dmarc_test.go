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
)

func TestParseDMARCTag(t *testing.T) {
	t.Parallel()

	t.Run(
		"extracts policy tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
			assert.Equal(t, "reject", parseDMARCTag(record, "p"))
		},
	)

	t.Run(
		"extracts rua tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=none; rua=mailto:reports@example.com"
			assert.Equal(t, "mailto:reports@example.com", parseDMARCTag(record, "rua"))
		},
	)

	t.Run(
		"returns empty string for missing tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=quarantine"
			assert.Equal(t, "", parseDMARCTag(record, "ruf"))
		},
	)

	t.Run(
		"extracts pct tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=reject; pct=50; rua=mailto:d@example.com"
			assert.Equal(t, "50", parseDMARCTag(record, "pct"))
		},
	)
}
