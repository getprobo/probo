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

func TestNormalizeServerURL(t *testing.T) {
	t.Parallel()

	t.Run(
		"adds https to bare hostnames",
		func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeServerURL("eu.console.getprobo.com")
			require.NoError(t, err)
			assert.Equal(t, EUConsoleURL, got)
		},
	)

	t.Run(
		"trims trailing slash",
		func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeServerURL("https://us.console.getprobo.com/")
			require.NoError(t, err)
			assert.Equal(t, USConsoleURL, got)
		},
	)

	t.Run(
		"rejects empty input",
		func(t *testing.T) {
			t.Parallel()

			_, err := NormalizeServerURL("   ")
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects paths",
		func(t *testing.T) {
			t.Parallel()

			_, err := NormalizeServerURL("https://probo.example.com/workspace")
			require.Error(t, err)
		},
	)
}
