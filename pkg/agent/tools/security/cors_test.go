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

func TestSplitTrimmed(t *testing.T) {
	t.Parallel()

	t.Run(
		"splits and trims values",
		func(t *testing.T) {
			t.Parallel()

			result := splitTrimmed("GET, POST, PUT", ",")
			require.Len(t, result, 3)
			assert.Equal(t, "GET", result[0])
			assert.Equal(t, "POST", result[1])
			assert.Equal(t, "PUT", result[2])
		},
	)

	t.Run(
		"returns nil for empty string",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, splitTrimmed("", ","))
		},
	)

	t.Run(
		"skips empty parts",
		func(t *testing.T) {
			t.Parallel()

			result := splitTrimmed("GET,,POST", ",")
			require.Len(t, result, 2)
			assert.Equal(t, "GET", result[0])
			assert.Equal(t, "POST", result[1])
		},
	)

	t.Run(
		"single value",
		func(t *testing.T) {
			t.Parallel()

			result := splitTrimmed("GET", ",")
			require.Len(t, result, 1)
			assert.Equal(t, "GET", result[0])
		},
	)
}
