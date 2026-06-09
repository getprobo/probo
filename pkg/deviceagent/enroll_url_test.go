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

package deviceagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnrollURL(t *testing.T) {
	t.Parallel()

	t.Run(
		"parses host enroll URL",
		func(t *testing.T) {
			t.Parallel()

			serverURL, token, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.console.getprobo.com&token=secret-token",
			)
			require.NoError(t, err)
			assert.Equal(t, "https://us.console.getprobo.com", serverURL)
			assert.Equal(t, "secret-token", token)
		},
	)

	t.Run(
		"parses path enroll URL",
		func(t *testing.T) {
			t.Parallel()

			serverURL, token, err := ParseEnrollURL(
				"probo:///enroll?server=https%3A%2F%2Feu.console.getprobo.com&token=abc123",
			)
			require.NoError(t, err)
			assert.Equal(t, "https://eu.console.getprobo.com", serverURL)
			assert.Equal(t, "abc123", token)
		},
	)

	t.Run(
		"rejects non probo scheme",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"https://example.com/enroll?server=https%3A%2F%2Fus.console.getprobo.com&token=secret-token",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "probo scheme")
		},
	)

	t.Run(
		"rejects missing token",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.console.getprobo.com",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "token is missing")
		},
	)

	t.Run(
		"rejects invalid server URL",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.console.getprobo.com%2Fextra&token=abc123",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "invalid server")
		},
	)
}
