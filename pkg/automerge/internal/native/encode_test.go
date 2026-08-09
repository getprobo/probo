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

// TestChangeEncodingExpandedRoundTrip reproduces the upstream
// test_change_encoding_expanded_change_round_trip: a change decoded from its
// canonical bytes re-encodes to exactly those bytes.
func TestChangeEncodingExpandedRoundTrip(t *testing.T) {
	t.Parallel()

	changeBytes := []byte{
		0x85, 0x6f, 0x4a, 0x83, // magic
		0xb2, 0x98, 0x9e, 0xa9, // checksum
		1, 61, 0, 2, 0x12, 0x34, // chunk type: change, length, deps, actor '1234'
		1, 1, 252, 250, 220, 255, 5, // seq, startOp, time
		14, 73, 110, 105, 116, 105, 97, 108, 105, 122, 97, 116, 105, 111, 110, // "Initialization"
		0, 6, // actor list, column count
		0x15, 3, 0x34, 1, 0x42, 2, // keyStr, insert, action
		0x56, 2, 0x57, 1, 0x70, 2, // valLen, valRaw, predNum
		0x7f, 1, 0x78, // keyStr: 'x'
		1,       // insert: false
		0x7f, 1, // action: set
		0x7f, 19, // valLen: 1 byte of type uint
		1,       // valRaw: 1
		0x7f, 0, // predNum: 0
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, // 10 trailing bytes inside the chunk
	}

	document, consumed, err := DecodeIncremental(changeBytes)
	require.NoError(t, err)
	require.Equal(t, len(changeBytes), consumed)
	require.Len(t, document.Changes, 1)

	encoded, err := EncodeChange(&document.Changes[0])
	require.NoError(t, err)
	assert.Equal(t, changeBytes, encoded)
}

func TestEncodeRLE_CanonicalRuns(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		values []optional[uint64]
		want   []byte
	}{
		"all null": {
			values: []optional[uint64]{{}, {}},
			want:   nil,
		},
		"null then value": {
			values: []optional[uint64]{{}, {}, some(uint64(3))},
			want:   []byte{0, 2, 0x7f, 3},
		},
		"repeated values": {
			values: []optional[uint64]{some(uint64(0)), some(uint64(0))},
			want:   []byte{2, 0},
		},
		"literal then repeated then literal": {
			values: []optional[uint64]{
				some(uint64(1)),
				some(uint64(2)),
				some(uint64(2)),
				some(uint64(3)),
			},
			want: []byte{0x7f, 1, 2, 2, 0x7f, 3},
		},
		"literal values": {
			values: []optional[uint64]{
				some(uint64(1)),
				some(uint64(2)),
				some(uint64(3)),
			},
			want: []byte{0x7d, 1, 2, 3},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, encodeRLE(test.values, appendULEB))
		})
	}
}
