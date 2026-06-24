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
