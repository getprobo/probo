// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
