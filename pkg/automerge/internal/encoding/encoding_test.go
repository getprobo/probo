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

package encoding

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestULEBRoundTrip(t *testing.T) {
	t.Parallel()

	for _, value := range []uint64{0, 1, 0x7f, 0x80, 0x3fff, 0x4000, math.MaxUint64} {
		encoded := AppendULEB(nil, value)
		decoded, err := NewReader(encoded).ULEB()
		require.NoError(t, err)
		assert.Equal(t, value, decoded)
	}
}

func TestReaderBounds(t *testing.T) {
	t.Parallel()

	reader := NewReader([]byte{1, 2})
	_, err := reader.Bytes(3)
	require.Error(t, err)
	assert.Equal(t, 0, reader.Offset())
}

func TestLengthPrefixedRoundTrip(t *testing.T) {
	t.Parallel()

	encoded := AppendLengthPrefixed(nil, []byte("value"))
	decoded, err := DecodeLengthPrefixed(NewReader(encoded))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), decoded)
}

func FuzzReader(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{0x80, 0x01})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01})

	f.Fuzz(func(t *testing.T, data []byte) {
		reader := NewReader(data)

		length, err := reader.ULEB()
		if err != nil {
			return
		}

		_, _ = reader.Bytes(length)
	})
}
