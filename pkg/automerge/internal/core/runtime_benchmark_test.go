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

package core

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/sync"
)

var (
	benchmarkRuntimeBytes []byte
	benchmarkRuntimeHash  [32]byte
)

func BenchmarkQueryHydration(b *testing.B) {
	data := benchmarkRuntimeDocument(b, 10_000)

	engine, err := LoadEngine(data)
	if err != nil {
		b.Fatal(err)
	}

	value, err := engine.Hydrate()
	if err != nil {
		b.Fatal(err)
	}

	expected := sha256.Sum256(value)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		value, err := engine.Hydrate()
		if err != nil {
			b.Fatal(err)
		}

		benchmarkRuntimeBytes = value
	}

	b.StopTimer()

	if actual := sha256.Sum256(benchmarkRuntimeBytes); actual != expected {
		b.Fatalf("hydration checksum changed: got %x, want %x", actual, expected)
	}
}

func BenchmarkColumnReconcile(b *testing.B) {
	data := benchmarkRuntimeDocument(b, 10_000)

	b.Run("append", func(b *testing.B) {
		benchmarkCommitReconcile(b, data, false)
	})
	b.Run("fallback", func(b *testing.B) {
		benchmarkCommitReconcile(b, data, true)
	})
}

func BenchmarkConcurrentTailReconcile(b *testing.B) {
	for _, size := range []int{1_000, 10_000} {
		b.Run(
			fmt.Sprintf("size=%d", size),
			func(b *testing.B) {
				leftData, rightData := benchmarkConcurrentTailDocuments(b, size)
				b.ReportAllocs()
				b.ReportMetric(float64(len(rightData)), "wire-B/op")
				b.StopTimer()

				targets := make([]*Engine, b.N)
				for index := range b.N {
					left, err := LoadEngine(leftData)
					if err != nil {
						b.Fatal(err)
					}

					targets[index] = left
				}

				runtime.GC()

				previousGC := debug.SetGCPercent(-1)
				before := ReadRuntimeMetrics()

				var memoryBefore runtime.MemStats
				runtime.ReadMemStats(&memoryBefore)
				b.ResetTimer()
				b.StartTimer()

				for _, left := range targets {
					_, err := left.Merge(rightData)
					if err != nil {
						b.Fatal(err)
					}
				}

				b.StopTimer()

				var memoryAfter runtime.MemStats
				runtime.ReadMemStats(&memoryAfter)

				after := ReadRuntimeMetrics()

				debug.SetGCPercent(previousGC)

				measured := RuntimeMetrics{
					GeneralReconciles: after.GeneralReconciles -
						before.GeneralReconciles,
					GlobalOrderFallbacks: after.GlobalOrderFallbacks -
						before.GlobalOrderFallbacks,
					SnapshotReplacements: after.SnapshotReplacements -
						before.SnapshotReplacements,
					FullColumnEncodings: after.FullColumnEncodings -
						before.FullColumnEncodings,
					QueryIndexRescues: after.QueryIndexRescues -
						before.QueryIndexRescues,
					SemanticOperationRows: after.SemanticOperationRows -
						before.SemanticOperationRows,
					DirectPlanningRows: after.DirectPlanningRows -
						before.DirectPlanningRows,
					DirectOperationCopies: after.DirectOperationCopies -
						before.DirectOperationCopies,
				}

				b.ReportMetric(
					float64(memoryAfter.TotalAlloc-memoryBefore.TotalAlloc)/
						float64(b.N),
					"merge-B/op",
				)
				b.ReportMetric(
					float64(memoryAfter.Mallocs-memoryBefore.Mallocs)/
						float64(b.N),
					"merge-allocs/op",
				)

				b.ReportMetric(
					float64(measured.GlobalOrderFallbacks)/float64(b.N),
					"global-order/op",
				)
				b.ReportMetric(
					float64(measured.SnapshotReplacements)/float64(b.N),
					"snapshot-replace/op",
				)
				b.ReportMetric(
					float64(measured.FullColumnEncodings)/float64(b.N),
					"full-column-encode/op",
				)
				b.ReportMetric(
					float64(measured.QueryIndexRescues)/float64(b.N),
					"query-rescue/op",
				)
				b.ReportMetric(
					float64(measured.SemanticOperationRows)/float64(b.N),
					"semantic-row/op",
				)
				b.ReportMetric(
					float64(measured.DirectPlanningRows)/float64(b.N),
					"planned-row/op",
				)
				b.ReportMetric(
					float64(measured.DirectOperationCopies)/float64(b.N),
					"copied-row/op",
				)

				if measured.GeneralReconciles != 0 ||
					measured.GlobalOrderFallbacks != 0 ||
					measured.SnapshotReplacements != 0 ||
					measured.FullColumnEncodings != 0 ||
					measured.QueryIndexRescues != 0 ||
					measured.SemanticOperationRows != 0 ||
					measured.DirectPlanningRows > uint64(b.N) ||
					measured.DirectOperationCopies != 0 {
					b.Fatalf("diverged merge used compatibility work: %+v", measured)
				}
			},
		)
	}
}

func BenchmarkConcurrentTailSyncReceive(b *testing.B) {
	for _, size := range []int{1_000, 10_000} {
		b.Run(
			fmt.Sprintf("size=%d", size),
			func(b *testing.B) {
				leftData, rightData := benchmarkConcurrentTailDocuments(b, size)

				message, err := (sync.Message{
					Version: sync.MessageVersion2,
					Changes: [][]byte{rightData},
				}).Encode()
				if err != nil {
					b.Fatal(err)
				}

				b.ReportMetric(float64(len(message)), "wire-B/op")
				b.StopTimer()
				targets := make([]*Engine, b.N)

				handles := make([]uint32, b.N)
				for index := range b.N {
					target, err := LoadEngine(leftData)
					if err != nil {
						b.Fatal(err)
					}

					handle, err := target.NewSyncState()
					if err != nil {
						b.Fatal(err)
					}

					targets[index] = target
					handles[index] = handle
				}

				runtime.GC()

				previousGC := debug.SetGCPercent(-1)
				before := ReadRuntimeMetrics()

				b.ResetTimer()
				b.StartTimer()

				for index, target := range targets {
					if err := target.ReceiveSyncMessage(
						handles[index],
						message,
					); err != nil {
						b.Fatal(err)
					}
				}

				b.StopTimer()

				after := ReadRuntimeMetrics()

				debug.SetGCPercent(previousGC)
				b.ReportMetric(
					float64(after.DirectPlanningRows-before.DirectPlanningRows)/
						float64(b.N),
					"planned-row/op",
				)
				b.ReportMetric(
					float64(after.DirectOperationCopies-before.DirectOperationCopies)/
						float64(b.N),
					"copied-row/op",
				)

				if after.GeneralReconciles != before.GeneralReconciles ||
					after.GlobalOrderFallbacks != before.GlobalOrderFallbacks ||
					after.SnapshotReplacements != before.SnapshotReplacements ||
					after.FullColumnEncodings != before.FullColumnEncodings ||
					after.QueryIndexRescues != before.QueryIndexRescues ||
					after.SemanticOperationRows != before.SemanticOperationRows ||
					after.DirectPlanningRows-before.DirectPlanningRows >
						uint64(b.N) ||
					after.DirectOperationCopies != before.DirectOperationCopies {
					b.Fatalf(
						"sync receive used compatibility work: before=%+v after=%+v",
						before,
						after,
					)
				}
			},
		)
	}
}

func benchmarkCommitReconcile(b *testing.B, data []byte, fallback bool) {
	b.Helper()
	b.ReportAllocs()
	b.StopTimer()

	var measured RuntimeMetrics

	for range b.N {
		engine, err := LoadEngine(data)
		if err != nil {
			b.Fatal(err)
		}

		text, err := engine.GetText(0, "body")
		if err != nil {
			b.Fatal(err)
		}

		if fallback {
			err = engine.SpliceText(text, 5_000, 1, "z")
		} else {
			err = engine.SpliceText(text, 10_000, 0, "z")
		}

		if err != nil {
			b.Fatal(err)
		}

		before := ReadRuntimeMetrics()

		b.StartTimer()

		hash, err := engine.Commit("benchmark reconcile", time.Unix(0, 0))

		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}

		after := ReadRuntimeMetrics()
		measured.GeneralReconciles += after.GeneralReconciles -
			before.GeneralReconciles
		measured.DirectColumnBatches += after.DirectColumnBatches -
			before.DirectColumnBatches
		measured.GlobalOrderFallbacks += after.GlobalOrderFallbacks -
			before.GlobalOrderFallbacks
		measured.FullColumnEncodings += after.FullColumnEncodings -
			before.FullColumnEncodings
		measured.SnapshotReplacements += after.SnapshotReplacements -
			before.SnapshotReplacements
		benchmarkRuntimeHash = hash
	}

	b.ReportMetric(
		float64(measured.GeneralReconciles)/float64(b.N),
		"general-reconcile/op",
	)
	b.ReportMetric(
		float64(measured.DirectColumnBatches)/float64(b.N),
		"direct-column-batch/op",
	)
	b.ReportMetric(
		float64(measured.GlobalOrderFallbacks)/float64(b.N),
		"global-order/op",
	)
	b.ReportMetric(
		float64(measured.FullColumnEncodings)/float64(b.N),
		"full-column-encode/op",
	)
	b.ReportMetric(
		float64(measured.SnapshotReplacements)/float64(b.N),
		"snapshot-replace/op",
	)

	if measured.GeneralReconciles != 0 ||
		measured.GlobalOrderFallbacks != 0 ||
		measured.SnapshotReplacements != 0 ||
		measured.FullColumnEncodings != 0 {
		b.Fatalf("ordinary commit used compatibility work: %+v", measured)
	}
}

func BenchmarkSyncCodecAndApplication(b *testing.B) {
	data := benchmarkRuntimeDocument(b, 10_000)

	b.Run("generate", func(b *testing.B) {
		b.StopTimer()

		source, err := LoadEngine(data)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()

		for range b.N {
			state, err := source.NewSyncState()
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()

			message, ok, err := source.GenerateSyncMessage(state)

			b.StopTimer()

			if err != nil {
				b.Fatal(err)
			}

			if !ok {
				b.Fatal("initial sync message was suppressed")
			}

			benchmarkRuntimeBytes = message

			if err := source.CloseSyncState(state); err != nil {
				b.Fatal(err)
			}
		}
	})

	source, err := LoadEngine(data)
	if err != nil {
		b.Fatal(err)
	}

	sourceState, err := source.NewSyncState()
	if err != nil {
		b.Fatal(err)
	}

	message, ok, err := source.GenerateSyncMessage(sourceState)
	if err != nil {
		b.Fatal(err)
	}

	if !ok {
		b.Fatal("initial sync message was suppressed")
	}

	b.Run("receive", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()

		for range b.N {
			target, err := NewEngine()
			if err != nil {
				b.Fatal(err)
			}

			state, err := target.NewSyncState()
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()

			err = target.ReceiveSyncMessage(state, message)

			b.StopTimer()

			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTextEdits(b *testing.B) {
	data := benchmarkRuntimeDocument(b, 10_000)

	mark, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarTrue},
	)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("plain", func(b *testing.B) {
		benchmarkTextEdit(b, data, nil)
	})
	b.Run("rich", func(b *testing.B) {
		benchmarkTextEdit(b, data, mark)
	})
}

func BenchmarkTextOverlayCommit(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()

	for range b.N {
		engine, err := NewEngine()
		if err != nil {
			b.Fatal(err)
		}

		text, err := engine.PutText(0, "body")
		if err != nil {
			b.Fatal(err)
		}

		if err := engine.SpliceText(text, 0, 0, string(make([]byte, 1_000))); err != nil {
			b.Fatal(err)
		}

		if _, err := engine.Commit("fixture", time.Time{}); err != nil {
			b.Fatal(err)
		}

		b.StartTimer()

		for index := range 100 {
			if err := engine.SpliceText(
				text,
				uint32(index*10),
				1,
				"z",
			); err != nil {
				b.Fatal(err)
			}
		}

		hash, err := engine.Commit("edits", time.Time{})

		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}

		benchmarkRuntimeHash = hash
	}
}

func benchmarkTextEdit(b *testing.B, data, mark []byte) {
	b.Helper()
	b.ReportAllocs()
	b.StopTimer()

	for range b.N {
		engine, err := LoadEngine(data)
		if err != nil {
			b.Fatal(err)
		}

		text, err := engine.GetText(0, "body")
		if err != nil {
			b.Fatal(err)
		}

		b.StartTimer()

		for index := range 100 {
			position := uint32(index * 100)
			if err := engine.SpliceText(text, position, 1, "z"); err != nil {
				b.Fatal(err)
			}
		}

		if mark != nil {
			if err := engine.MarkText(text, 2_500, 7_500, "bold", mark, "both"); err != nil {
				b.Fatal(err)
			}
		}

		hash, err := engine.Commit("benchmark text edit", time.Unix(0, 0))

		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}

		benchmarkRuntimeHash = hash
	}
}

func benchmarkRuntimeDocument(b *testing.B, size int) []byte {
	b.Helper()

	engine, err := NewEngine()
	if err != nil {
		b.Fatal(err)
	}

	if err := engine.SetActor([]byte("base-benchmark")); err != nil {
		b.Fatal(err)
	}

	text, err := engine.PutText(0, "body")
	if err != nil {
		b.Fatal(err)
	}

	value := make([]byte, size)
	for index := range value {
		value[index] = byte('a' + index%26)
	}

	if err := engine.SpliceText(text, 0, 0, string(value)); err != nil {
		b.Fatal(err)
	}

	if _, err := engine.Commit("benchmark fixture", time.Unix(0, 0)); err != nil {
		b.Fatal(err)
	}

	data, err := engine.Save(true, false)
	if err != nil {
		b.Fatal(err)
	}

	return data
}

func benchmarkConcurrentTailDocuments(b *testing.B, size int) ([]byte, []byte) {
	b.Helper()

	baseData := benchmarkRuntimeDocument(b, size)

	base, err := LoadEngine(baseData)
	if err != nil {
		b.Fatal(err)
	}

	baseHeads, err := base.Heads()
	if err != nil {
		b.Fatal(err)
	}

	left, err := LoadEngine(baseData)
	if err != nil {
		b.Fatal(err)
	}

	right, err := LoadEngine(baseData)
	if err != nil {
		b.Fatal(err)
	}

	if err := left.SetActor([]byte("left-benchmark")); err != nil {
		b.Fatal(err)
	}

	if err := right.SetActor([]byte("right-benchmark")); err != nil {
		b.Fatal(err)
	}

	leftText, err := left.GetText(0, "body")
	if err != nil {
		b.Fatal(err)
	}

	rightText, err := right.GetText(0, "body")
	if err != nil {
		b.Fatal(err)
	}

	for index := range min(size, 1) {
		if err := left.SpliceText(leftText, uint32(size+index), 0, "L"); err != nil {
			b.Fatal(err)
		}
	}

	if _, err := left.Commit("left branch", time.Unix(0, 0)); err != nil {
		b.Fatal(err)
	}

	for index := range min(size, 1) {
		if err := right.SpliceText(rightText, uint32(size+index), 0, "R"); err != nil {
			b.Fatal(err)
		}
	}

	if _, err := right.Commit("right branch", time.Unix(0, 0)); err != nil {
		b.Fatal(err)
	}

	leftData, err := left.Save(true, false)
	if err != nil {
		b.Fatal(err)
	}

	rightChanges, _, err := right.ChangesSince(baseHeads)
	if err != nil {
		b.Fatal(err)
	}

	rightData := make([]byte, 0)
	for _, change := range rightChanges {
		rightData = append(rightData, change...)
	}

	return leftData, rightData
}
