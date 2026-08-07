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

package native

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestULEB128_RoundTrip(t *testing.T) {
	t.Parallel()

	values := []uint64{
		0,
		1,
		127,
		128,
		255,
		16_384,
		math.MaxUint32,
		math.MaxUint64,
	}
	for _, expected := range values {
		encoded := appendULEB128(nil, expected)
		actual, consumed, err := readULEB128(encoded)

		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		assert.Equal(t, len(encoded), consumed)
	}
}

func TestULEB128_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	_, _, err := readULEB128([]byte{0x80})
	assert.Error(t, err)

	_, _, err = readULEB128([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0x02,
	})
	assert.Error(t, err)
}

func TestSLEB128_RoundTrip(t *testing.T) {
	t.Parallel()

	values := []int64{
		math.MinInt64,
		-16_384,
		-128,
		-1,
		0,
		1,
		127,
		128,
		16_384,
		math.MaxInt64,
	}
	for _, expected := range values {
		encoded := appendSLEB128(nil, expected)
		r := newReader(encoded)
		actual, err := r.readSLEB128()

		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		assert.True(t, r.done())
	}
}
