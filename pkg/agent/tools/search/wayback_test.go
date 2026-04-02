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

package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCDXSnapshot(t *testing.T) {
	t.Parallel()

	t.Run(
		"valid JSON array response",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`[["timestamp","original"],["20200115120000","https://example.com/privacy"]]`)

			snap := parseCDXSnapshot(body)

			require.NotNil(t, snap)
			assert.Equal(t, "20200115120000", snap.Timestamp)
			assert.Equal(t, "https://example.com/privacy", snap.URL)
		},
	)

	t.Run(
		"empty array returns nil",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`[]`)

			assert.Nil(t, parseCDXSnapshot(body))
		},
	)

	t.Run(
		"single row header only returns nil",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`[["timestamp","original"]]`)

			assert.Nil(t, parseCDXSnapshot(body))
		},
	)

	t.Run(
		"malformed JSON returns nil",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`not valid json`)

			assert.Nil(t, parseCDXSnapshot(body))
		},
	)

	t.Run(
		"data row with insufficient fields returns nil",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`[["timestamp","original"],["20200115120000"]]`)

			assert.Nil(t, parseCDXSnapshot(body))
		},
	)

	t.Run(
		"empty body returns nil",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, parseCDXSnapshot([]byte{}))
		},
	)

	t.Run(
		"response with extra fields uses first two",
		func(t *testing.T) {
			t.Parallel()

			body := []byte(`[["timestamp","original","extra"],["20210601000000","https://example.com/tos","200"]]`)

			snap := parseCDXSnapshot(body)

			require.NotNil(t, snap)
			assert.Equal(t, "20210601000000", snap.Timestamp)
			assert.Equal(t, "https://example.com/tos", snap.URL)
		},
	)
}
