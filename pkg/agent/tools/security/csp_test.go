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

func TestParseCSPDirectives(t *testing.T) {
	t.Parallel()

	t.Run(
		"parses multiple directives",
		func(t *testing.T) {
			t.Parallel()

			raw := "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'unsafe-inline'"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 3)
			assert.Equal(t, "default-src", directives[0].Name)
			assert.Equal(t, []string{"'self'"}, directives[0].Values)
			assert.Equal(t, "script-src", directives[1].Name)
			assert.Equal(t, []string{"'self'", "https://cdn.example.com"}, directives[1].Values)
			assert.Equal(t, "style-src", directives[2].Name)
			assert.Equal(t, []string{"'unsafe-inline'"}, directives[2].Values)
		},
	)

	t.Run(
		"handles empty string",
		func(t *testing.T) {
			t.Parallel()

			directives := parseCSPDirectives("")
			assert.Empty(t, directives)
		},
	)

	t.Run(
		"handles directive without values",
		func(t *testing.T) {
			t.Parallel()

			raw := "upgrade-insecure-requests"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 1)
			assert.Equal(t, "upgrade-insecure-requests", directives[0].Name)
			assert.Empty(t, directives[0].Values)
		},
	)

	t.Run(
		"ignores trailing semicolons",
		func(t *testing.T) {
			t.Parallel()

			raw := "default-src 'self';"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 1)
			assert.Equal(t, "default-src", directives[0].Name)
		},
	)
}
