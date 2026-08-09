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
)

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
