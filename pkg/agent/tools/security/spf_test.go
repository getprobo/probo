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

func TestParseSPFPolicy(t *testing.T) {
	t.Parallel()

	t.Run(
		"detects hard fail",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "fail", parseSPFPolicy("v=spf1 include:_spf.google.com -all"))
		},
	)

	t.Run(
		"detects soft fail",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "softfail", parseSPFPolicy("v=spf1 include:spf.example.com ~all"))
		},
	)

	t.Run(
		"detects neutral",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "neutral", parseSPFPolicy("v=spf1 ?all"))
		},
	)

	t.Run(
		"detects pass all",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "pass", parseSPFPolicy("v=spf1 +all"))
		},
	)

	t.Run(
		"returns empty for no all qualifier",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "", parseSPFPolicy("v=spf1 include:_spf.google.com"))
		},
	)
}
