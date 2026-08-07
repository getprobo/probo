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

func TestRLEDecoder_RunsLiteralsAndNulls(t *testing.T) {
	t.Parallel()

	var data []byte
	data = appendSLEB128(data, 3)
	data = appendULEB128(data, 7)
	data = appendSLEB128(data, 0)
	data = appendULEB128(data, 2)
	data = appendSLEB128(data, -2)
	data = appendULEB128(data, 4)
	data = appendULEB128(data, 5)
	decoder := decodeRLEUint(data)

	expected := []struct {
		value uint64
		null  bool
	}{
		{value: 7},
		{value: 7},
		{value: 7},
		{null: true},
		{null: true},
		{value: 4},
		{value: 5},
	}
	for _, item := range expected {
		value, null, err := decoder.next()
		require.NoError(t, err)
		assert.Equal(t, item.value, value)
		assert.Equal(t, item.null, null)
	}

	_, null, err := decoder.next()
	require.NoError(t, err)
	assert.True(t, null)
}

func TestRLEDecoder_RejectsZeroNullRun(t *testing.T) {
	t.Parallel()

	data := appendSLEB128(nil, 0)
	data = appendULEB128(data, 0)
	_, _, err := decodeRLEUint(data).next()
	assert.Error(t, err)
}
