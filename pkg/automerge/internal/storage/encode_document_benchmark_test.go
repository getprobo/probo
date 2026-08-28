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
	"testing"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func BenchmarkValidateSnapshotEncodeDomain(b *testing.B) {
	document, _ := benchmarkSnapshot(10_000)

	b.ReportAllocs()

	for b.Loop() {
		if err := validateSnapshotEncodeDomain(document); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodePreparedDocument(b *testing.B) {
	document, operations := benchmarkSnapshot(10_000)

	for _, compress := range []bool{false, true} {
		name := "NoCompress"
		if compress {
			name = "Compress"
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				if _, err := EncodePreparedDocument(
					document,
					operations,
					compress,
				); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSnapshotColumnsEncode(b *testing.B) {
	document, operations := benchmarkSnapshot(10_000)

	columns, err := NewSnapshotColumns(document, operations)
	if err != nil {
		b.Fatal(err)
	}

	for _, compress := range []bool{false, true} {
		name := "NoCompress"
		if compress {
			name = "Compress"
		}

		b.Run(name, func(b *testing.B) {
			result, err := columns.Encode(document.UnknownColumns, compress)
			if err != nil {
				b.Fatal(err)
			}

			b.SetBytes(int64(len(result)))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, err := columns.Encode(
					document.UnknownColumns,
					compress,
				); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSnapshotColumnsLocalizedBatch(b *testing.B) {
	document, operations := benchmarkSnapshot(10_000)

	base, err := NewSnapshotColumns(document, operations)
	if err != nil {
		b.Fatal(err)
	}

	changes := make([]*opset.Change, len(document.Changes))

	dependencyIndexes := make(map[opset.ChangeHash]uint64, len(changes))
	for i := range document.Changes {
		changes[i] = &document.Changes[i]
		dependencyIndexes[*document.Changes[i].Hash] = uint64(i)
	}

	actors := documentActorTable(changes, operations)

	heads, headIndexes, err := documentHeads(changes)
	if err != nil {
		b.Fatal(err)
	}

	replacement := operations[len(operations)/2]
	edit := SnapshotSplice{
		Actors:            actors,
		Heads:             heads,
		HeadIndexes:       headIndexes,
		DependencyIndexes: dependencyIndexes,
		ChangeIndex:       len(changes),
		OperationSplices: []SnapshotOperationSplice{{
			Index:       len(operations) / 2,
			DeleteCount: 1,
			Operations:  []opset.Operation{replacement},
		}},
	}

	ResetRuntimeMetrics()
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		columns := base.Clone()
		if err := columns.Splice(edit); err != nil {
			b.Fatal(err)
		}

		if _, err := columns.Encode(nil, false); err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
	b.ReportMetric(
		float64(FullColumnEncodings())/float64(b.N),
		"full-column-encode/op",
	)
}

func benchmarkSnapshot(size int) (*opset.Document, []opset.Operation) {
	actor, err := opset.NewActorID([]byte{1})
	if err != nil {
		panic(err)
	}

	body := "body"
	makeTextID := opset.OpID{Actor: actor, Counter: 1}
	operations := make([]opset.Operation, 0, size+1)
	operations = append(operations, opset.Operation{
		ID:     makeTextID,
		Object: opset.RootObject(),
		Key:    opset.Key{Property: &body},
		Action: opset.ActionMakeText,
	})

	for i := range size {
		identifier := opset.OpID{Actor: actor, Counter: uint64(i + 2)}

		key := opset.Key{IsHead: i == 0}
		if i > 0 {
			previous := operations[len(operations)-1].ID
			key.Element = &previous
		}

		operations = append(operations, opset.Operation{
			ID:     identifier,
			Object: opset.ObjectID{OpID: makeTextID},
			Key:    key,
			Insert: true,
			Action: opset.ActionSet,
			Value: &opset.Scalar{
				Type:   opset.ScalarString,
				String: "x",
			},
		})
	}

	hash := opset.ChangeHash{1}

	changeOperations := append([]opset.Operation(nil), operations...)
	document := &opset.Document{
		Heads: []opset.ChangeHash{hash},
		Changes: []opset.Change{{
			Hash:       &hash,
			Actor:      actor,
			Sequence:   1,
			StartOp:    1,
			MaxOp:      uint64(size + 1),
			Operations: changeOperations,
		}},
	}

	return document, operations
}
