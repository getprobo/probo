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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooleanDecoder_AlternatingRuns(t *testing.T) {
	t.Parallel()

	var data []byte
	data = appendULEB128(data, 2)
	data = appendULEB128(data, 3)
	data = appendULEB128(data, 1)
	decoder := newBooleanDecoder(data)

	expected := []bool{false, false, true, true, true, false}
	for _, value := range expected {
		actual, err := decoder.next()
		require.NoError(t, err)
		assert.Equal(t, value, actual)
	}
}

func TestDeltaDecoder_AccumulatesNonNullValues(t *testing.T) {
	t.Parallel()

	var data []byte
	data = appendSLEB128(data, -3)
	data = appendSLEB128(data, 10)
	data = appendSLEB128(data, -3)
	data = appendSLEB128(data, 8)
	decoder := newDeltaDecoder(data)

	first, null, err := decoder.next()
	require.NoError(t, err)
	assert.False(t, null)
	assert.Equal(t, int64(10), first)

	second, null, err := decoder.next()
	require.NoError(t, err)
	assert.False(t, null)
	assert.Equal(t, int64(7), second)

	third, null, err := decoder.next()
	require.NoError(t, err)
	assert.False(t, null)
	assert.Equal(t, int64(15), third)
}
