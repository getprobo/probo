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

package storage

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expandRuns turns a run-held column back into one entry per item.
func expandRuns[T any](runs []rleRun[T]) []optional[T] {
	cursor := runCursor[T]{runs: runs}
	values := make([]optional[T], 0)

	for _, run := range runs {
		for range run.count {
			values = append(values, cursor.next())
		}
	}

	return values
}

func randomOptionalColumn[T any](
	random *rand.Rand,
	length int,
	value func() T,
) []optional[T] {
	values := make([]optional[T], 0, length)

	for len(values) < length {
		run := 1 + random.IntN(6)

		item := optional[T]{}
		if random.IntN(4) != 0 {
			item = some(value())
		}

		// Mixing repeated, null and one-off items exercises every run kind
		// the encoder can emit.
		if random.IntN(3) == 0 {
			run = 1
		}

		for range run {
			values = append(values, item)
		}
	}

	return values[:length]
}

func TestDecodeRLERuns_MatchesExpandedDecode(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(41, 43))

	for range 300 {
		length := random.IntN(60)

		t.Run(
			"unsigned",
			func(t *testing.T) {
				values := randomOptionalColumn(random, length, func() uint64 {
					return uint64(random.IntN(5))
				})
				data := encodeRLE(values, appendULEB)

				want, err := decodeULEBColumn(data)
				require.NoError(t, err)
				runs, total, err := decodeULEBRuns(data, &decodeBudget{})
				require.NoError(t, err)

				assert.Equal(t, uint64(len(want)), total)
				assert.Equal(t, want, expandRuns(runs))
			},
		)

		t.Run(
			"strings",
			func(t *testing.T) {
				values := randomOptionalColumn(random, length, func() string {
					return string(rune('a' + random.IntN(4)))
				})
				data := encodeStrings(values)

				want, err := decodeStringColumn(data)
				require.NoError(t, err)
				runs, total, err := decodeStringRuns(data, &decodeBudget{})
				require.NoError(t, err)

				assert.Equal(t, uint64(len(want)), total)
				assert.Equal(t, want, expandRuns(runs))
			},
		)

		t.Run(
			"delta",
			func(t *testing.T) {
				values := randomOptionalColumn(random, length, func() int64 {
					return int64(random.IntN(1000))
				})
				data := encodeDelta(values)

				want, err := decodeDeltaColumnWithBudget(data, nil)
				require.NoError(t, err)
				runs, total, err := decodeDeltaRuns(data, &decodeBudget{})
				require.NoError(t, err)
				require.Equal(t, uint64(len(want)), total)

				cursor := deltaCursor{deltas: runCursor[int64]{runs: runs}}
				got := make([]optional[uint64], 0, len(want))

				for range want {
					got = append(got, cursor.next())
				}

				assert.Equal(t, want, got)
			},
		)
	}
}

func TestDecodeRLERuns_ReportsExpandedDecodeErrors(t *testing.T) {
	t.Parallel()

	columns := map[string][]byte{
		"truncated run length":   {0x80},
		"empty null run":         {0x00, 0x00},
		"truncated null run":     {0x00},
		"truncated repeat value": {0x02},
		"truncated literal":      {0x7e, 0x01},
		"literal overflow": appendLEB(
			nil,
			math.MinInt64,
		),
		"too many items": appendULEB(appendLEB(nil, maxDecodedItems+1), 1),
	}

	for name, data := range columns {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				_, want := decodeULEBColumn(data)
				require.Error(t, want)

				_, _, err := decodeULEBRuns(data, &decodeBudget{})
				require.Error(t, err)
				assert.Equal(t, want.Error(), err.Error())
			},
		)
	}

	t.Run(
		"delta overflow names the same item",
		func(t *testing.T) {
			t.Parallel()

			data := encodeRLE(
				[]optional[int64]{{}, some(int64(5)), some(int64(-9))},
				appendLEB,
			)

			_, want := decodeDeltaColumnWithBudget(data, nil)
			require.Error(t, want)

			_, _, err := decodeDeltaRuns(data, &decodeBudget{})
			require.Error(t, err)
			assert.Equal(t, want.Error(), err.Error())
		},
	)
}

func TestRequireRunItems(t *testing.T) {
	t.Parallel()

	runs := []rleRun[uint64]{
		{value: 1, count: 2, valid: true},
		{count: 3},
		{value: 2, count: 1, valid: true},
	}
	values := expandRuns(runs)

	for _, expected := range []int{5, 6} {
		for _, nullable := range []bool{true, false} {
			want := requireItems("probe", values, expected, nullable)
			got := requireRunItems("probe", runs, 6, expected, nullable)

			if want == nil {
				assert.NoError(t, got)

				continue
			}

			require.Error(t, got)
			assert.Equal(t, want.Error(), got.Error())
		}
	}
}

func TestRunCursor_YieldsNullsOnceSpent(t *testing.T) {
	t.Parallel()

	cursor := runCursor[uint64]{runs: []rleRun[uint64]{{value: 7, count: 1, valid: true}}}

	assert.Equal(t, some(uint64(7)), cursor.next())
	assert.Equal(t, optional[uint64]{}, cursor.next())

	absent := deltaCursor{}
	assert.Equal(t, optional[uint64]{}, absent.next())
}
