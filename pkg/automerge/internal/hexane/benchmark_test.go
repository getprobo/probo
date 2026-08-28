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

package hexane_test

import (
	"testing"

	"go.probo.inc/probo/pkg/automerge/internal/hexane"
)

var (
	benchmarkColumn *hexane.Column[uint64]
	benchmarkDelta  *hexane.DeltaColumn
	benchmarkRaw    *hexane.RawColumn
	benchmarkPrefix uint64
)

func BenchmarkColumn_LocalizedCOWSet(b *testing.B) {
	benchmarkLocalizedColumn(b, "1K rows", 1<<10)
	benchmarkLocalizedColumn(b, "1M rows", 1<<20)
}

func benchmarkLocalizedColumn(b *testing.B, name string, rows int) {
	b.Helper()
	b.Run(
		name,
		func(b *testing.B) {
			values := make([]hexane.Value[uint64], rows)
			for i := range values {
				values[i] = hexane.Some(uint64(i))
			}

			base := hexane.NewColumnFromValues(hexane.Uint64Codec(), values...)

			b.ReportAllocs()
			b.ReportMetric(float64(rows), "rows")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				cloned := base.Clone()
				cloned.Set(rows/2, hexane.Some(uint64(i)))
				benchmarkColumn = cloned
			}
		},
	)
}

func BenchmarkColumn_LocalizedCOWSplice(b *testing.B) {
	benchmarkLocalizedSplice(b, "1K rows", 1<<10)
	benchmarkLocalizedSplice(b, "1M rows", 1<<20)
}

func benchmarkLocalizedSplice(b *testing.B, name string, rows int) {
	b.Helper()
	b.Run(
		name,
		func(b *testing.B) {
			values := make([]hexane.Value[uint64], rows)
			for i := range values {
				values[i] = hexane.Some(uint64(i))
			}

			base := hexane.NewColumnFromValues(hexane.Uint64Codec(), values...)
			replacement := []hexane.Value[uint64]{
				hexane.Some(uint64(7)),
				hexane.Some(uint64(8)),
				hexane.Some(uint64(9)),
			}

			b.ReportAllocs()
			b.ReportMetric(float64(rows), "rows")
			b.ResetTimer()

			for b.Loop() {
				cloned := base.Clone()
				cloned.Splice(rows/2, len(replacement), replacement...)
				benchmarkColumn = cloned
			}
		},
	)
}

func BenchmarkColumn_MultiSplice(b *testing.B) {
	const rows = 1 << 20

	values := make([]hexane.Value[uint64], rows)
	for i := range values {
		values[i] = hexane.Some(uint64(i))
	}

	base := hexane.NewColumnFromValues(hexane.Uint64Codec(), values...)
	inserted := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		hexane.Some(uint64(7)),
		hexane.Some(uint64(8)),
	)

	splices := make([]hexane.ColumnSplice[uint64], 64)
	for i := range splices {
		splices[i] = hexane.ColumnSplice[uint64]{
			Index:       (i + 1) * (rows / (len(splices) + 1)),
			DeleteCount: 2,
			Inserted:    inserted,
		}
	}

	b.Run("Sequential", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			column := base.Clone()

			for i := len(splices) - 1; i >= 0; i-- {
				splice := splices[i]
				column.Splice(
					splice.Index,
					splice.DeleteCount,
					splice.Inserted.Values()...,
				)
			}

			benchmarkColumn = column
		}
	})
	b.Run("Batch", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			column := base.Clone()
			if err := column.BatchSplice(splices); err != nil {
				b.Fatal(err)
			}

			benchmarkColumn = column
		}
	})
}

func BenchmarkDeltaColumn_LocalizedCOWSet(b *testing.B) {
	benchmarkLocalizedDelta(b, "1K rows", 1<<10)
	benchmarkLocalizedDelta(b, "1M rows", 1<<20)
}

func benchmarkLocalizedDelta(b *testing.B, name string, rows int) {
	b.Helper()
	b.Run(
		name,
		func(b *testing.B) {
			values := make([]hexane.Value[int64], rows)
			for i := range values {
				values[i] = hexane.Some(int64(i))
			}

			base := hexane.NewDeltaColumnFromValues(values...)

			b.ReportAllocs()
			b.ReportMetric(float64(rows), "rows")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				cloned := base.Clone()
				cloned.Set(rows/2, hexane.Some(int64(i)))
				benchmarkDelta = cloned
			}
		},
	)
}

func BenchmarkPrefixColumn_Query(b *testing.B) {
	values := make([]uint64, 1<<20)
	for i := range values {
		values[i] = uint64(i & 7)
	}

	column := hexane.NewPrefixColumnFromValues(values...)

	b.ReportAllocs()
	b.ReportMetric(float64(len(values)), "rows")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkPrefix = column.Prefix((i % (len(values) - 1)) + 1)
	}
}

func BenchmarkRawColumn_LocalizedCOWSet(b *testing.B) {
	benchmarkLocalizedRaw(b, "4KiB", 4<<10)
	benchmarkLocalizedRaw(b, "4MiB", 4<<20)
}

func benchmarkLocalizedRaw(b *testing.B, name string, size int) {
	b.Helper()
	b.Run(
		name,
		func(b *testing.B) {
			base := hexane.NewRawColumnFromBytes(make([]byte, size))

			b.ReportAllocs()
			b.ReportMetric(float64(size), "bytes")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				cloned := base.Clone()
				cloned.Set(size/2, byte(i))
				benchmarkRaw = cloned
			}
		},
	)
}
