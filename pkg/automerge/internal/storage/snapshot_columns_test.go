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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestSnapshotColumns_MatchesPreparedEncoder(t *testing.T) {
	t.Parallel()

	document, operations := benchmarkSnapshot(1_000)
	columns, err := NewSnapshotColumns(document, operations)
	require.NoError(t, err)

	for _, compress := range []bool{false, true} {
		compress := compress
		t.Run(compressionName(compress), func(t *testing.T) {
			t.Parallel()

			expected, err := EncodePreparedDocument(
				document,
				operations,
				compress,
			)
			require.NoError(t, err)
			actual, err := columns.Encode(document.UnknownColumns, compress)
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	}
}

func TestSnapshotColumns_ReusesDecodedKnownColumns(t *testing.T) {
	t.Parallel()

	document, operations := benchmarkSnapshot(100)
	_, err := EncodeChange(&document.Changes[0])
	require.NoError(t, err)

	document.Heads = []opset.ChangeHash{*document.Changes[0].Hash}
	encoded, err := EncodePreparedDocument(document, operations, false)
	require.NoError(t, err)
	decoded, err := Decode(encoded)
	require.NoError(t, err)

	decodedOperations := make([]opset.Operation, 0, len(decoded.OperationOrder))
	operationsByID := make(map[opset.OpID]opset.Operation)

	for i := range decoded.Changes {
		for _, operation := range decoded.Changes[i].Operations {
			operationsByID[operation.ID] = operation
		}
	}

	for _, identifier := range decoded.OperationOrder {
		decodedOperations = append(decodedOperations, operationsByID[identifier])
	}

	columns, err := NewSnapshotColumns(decoded, decodedOperations)
	require.NoError(t, err)
	actual, err := columns.Encode(decoded.UnknownColumns, false)
	require.NoError(t, err)

	assert.Equal(t, encoded, actual)
}

func TestSnapshotColumns_SpliceRandomizedDifferentialAndClone(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0x53504c494345))
	testActors := testHexaneActors()
	changePointers := randomHexaneChanges(rng, testActors, 8)
	document := &opset.Document{Changes: dereferenceChanges(changePointers)}
	operations := randomHexaneOperations(rng, testActors, 30)
	columns, err := NewSnapshotColumns(document, operations)
	require.NoError(t, err)

	for iteration := range 100 {
		before := columns.Clone()
		beforeBytes, err := before.Encode(nil, false)
		require.NoError(t, err)

		operationIndex := rng.Intn(len(operations) + 1)
		deleteCount := min(rng.Intn(4), len(operations)-operationIndex)
		insertedOperations := randomHexaneOperations(rng, testActors, rng.Intn(4))
		operations = spliceOperations(
			operations,
			operationIndex,
			deleteCount,
			insertedOperations,
		)

		var insertedChanges []*opset.Change

		if iteration%4 == 0 {
			hash := opset.ChangeHash{}
			_, _ = rng.Read(hash[:])
			insertedChanges = []*opset.Change{
				{
					Hash:         new(hash),
					Actor:        testActors[rng.Intn(len(testActors))],
					Sequence:     uint64(len(document.Changes) + 1),
					MaxOp:        uint64(10_000 + iteration),
					Time:         int64(iteration) - 50,
					Message:      randomHexaneString(rng),
					Dependencies: []opset.ChangeHash{*document.Changes[len(document.Changes)-1].Hash},
					Extra:        randomHexaneScalar(rng, iteration),
				},
			}
			document.Changes = append(document.Changes, *insertedChanges[0])
		}

		finalChanges := changePointersFromDocument(document)
		actors := documentActorTable(finalChanges, operations)
		heads, headIndexes, err := documentHeads(finalChanges)
		require.NoError(t, err)

		dependencyIndexes := make(map[opset.ChangeHash]uint64, len(finalChanges))
		for i, change := range finalChanges {
			dependencyIndexes[*change.Hash] = uint64(i)
		}

		err = columns.Splice(
			SnapshotSplice{
				Actors:               actors,
				Heads:                heads,
				HeadIndexes:          headIndexes,
				DependencyIndexes:    dependencyIndexes,
				ChangeIndex:          len(finalChanges) - len(insertedChanges),
				Changes:              insertedChanges,
				OperationIndex:       operationIndex,
				OperationDeleteCount: deleteCount,
				Operations:           insertedOperations,
			},
		)
		require.NoError(t, err)

		expected, err := EncodePreparedDocument(document, operations, false)
		require.NoError(t, err)
		actual, err := columns.Encode(nil, false)
		require.NoError(t, err)
		assert.Equal(t, expected, actual, "iteration %d", iteration)

		unchanged, err := before.Encode(nil, false)
		require.NoError(t, err)
		assert.Equal(t, beforeBytes, unchanged, "clone iteration %d", iteration)
	}
}

func TestSnapshotColumns_BatchedOperationSplicesMatchPreparedEncoder(t *testing.T) {
	t.Parallel()

	document, operations := benchmarkSnapshot(1_000)
	columns, err := NewSnapshotColumns(document, operations)
	require.NoError(t, err)

	next := append([]opset.Operation(nil), operations...)
	splices := make([]SnapshotOperationSplice, 0, 10)

	for index := 900; index >= 0; index -= 100 {
		inserted := next[index]
		inserted.ID = opset.OpID{
			Actor:   opset.ActorID("batch"),
			Counter: uint64(10_000 + index),
		}
		inserted.Successors = nil
		updated := next[index]
		updated.Successors = []opset.OpID{inserted.ID}
		replacement := []opset.Operation{inserted, updated}
		next = spliceOperations(next, index, 1, replacement)
		splices = append(splices, SnapshotOperationSplice{
			Index:       index,
			DeleteCount: 1,
			Operations:  replacement,
		})
	}

	changes := changePointersFromDocument(document)
	actors := documentActorTable(changes, next)
	heads, headIndexes, err := documentHeads(changes)
	require.NoError(t, err)

	dependencyIndexes := make(map[opset.ChangeHash]uint64, len(changes))
	for i, change := range changes {
		dependencyIndexes[*change.Hash] = uint64(i)
	}

	err = columns.Splice(SnapshotSplice{
		Actors:            actors,
		Heads:             heads,
		HeadIndexes:       headIndexes,
		DependencyIndexes: dependencyIndexes,
		ChangeIndex:       len(changes),
		OperationSplices:  splices,
	})
	require.NoError(t, err)

	expected, err := EncodePreparedDocument(document, next, false)
	require.NoError(t, err)
	actual, err := columns.Encode(nil, false)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestSnapshotColumns_LocalizedSpliceAvoidsFullEncoding(t *testing.T) {
	document, operations := benchmarkSnapshot(1_000)
	columns, err := NewSnapshotColumns(document, operations)
	require.NoError(t, err)

	changes := changePointersFromDocument(document)
	actors := documentActorTable(changes, operations)
	heads, headIndexes, err := documentHeads(changes)
	require.NoError(t, err)

	dependencyIndexes := make(map[opset.ChangeHash]uint64, len(changes))
	for i, change := range changes {
		dependencyIndexes[*change.Hash] = uint64(i)
	}

	firstReplacement := operations[500]
	firstReplacement.Successors = append(
		[]opset.OpID(nil),
		firstReplacement.Successors...,
	)
	secondReplacement := operations[700]
	secondReplacement.Successors = append(
		[]opset.OpID(nil),
		secondReplacement.Successors...,
	)

	ResetRuntimeMetrics()

	err = columns.Splice(SnapshotSplice{
		Actors:            actors,
		Heads:             heads,
		HeadIndexes:       headIndexes,
		DependencyIndexes: dependencyIndexes,
		ChangeIndex:       len(changes),
		OperationSplices: []SnapshotOperationSplice{
			{
				Index:       700,
				DeleteCount: 1,
				Operations:  []opset.Operation{secondReplacement},
			},
			{
				Index:       500,
				DeleteCount: 1,
				Operations:  []opset.Operation{firstReplacement},
			},
		},
	})
	require.NoError(t, err)
	assert.Zero(t, FullColumnEncodings())

	expected, err := EncodePreparedDocument(document, operations, false)
	require.NoError(t, err)
	actual, err := columns.Encode(nil, false)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	assert.Zero(t, FullColumnEncodings())
}

func TestSnapshotColumns_SpliceFailureIsTransactional(t *testing.T) {
	t.Parallel()

	document, operations := benchmarkSnapshot(100)
	columns, err := NewSnapshotColumns(document, operations)
	require.NoError(t, err)
	before, err := columns.Encode(nil, false)
	require.NoError(t, err)

	err = columns.Splice(
		SnapshotSplice{
			Actors: documentActorTable(
				changePointersFromDocument(document),
				operations,
			),
			Heads:       document.Heads,
			ChangeIndex: len(document.Changes),
			OperationSplices: []SnapshotOperationSplice{
				{
					Index:       0,
					DeleteCount: 1,
					Operations:  operations[:1],
				},
				{
					Index:       len(operations) + 1,
					DeleteCount: 1,
				},
			},
		},
	)
	require.Error(t, err)

	after, err := columns.Encode(nil, false)
	require.NoError(t, err)
	assert.Equal(t, before, after)

	invalid := operations[0]
	invalid.Key.Property = new(string([]byte{0xff}))
	invalid.Key.Element = nil
	invalid.Key.IsHead = false
	err = columns.Splice(
		SnapshotSplice{
			Actors: documentActorTable(
				changePointersFromDocument(document),
				operations,
			),
			Heads:       document.Heads,
			ChangeIndex: len(document.Changes),
			OperationSplices: []SnapshotOperationSplice{{
				Index:       0,
				DeleteCount: 1,
				Operations:  []opset.Operation{invalid},
			}},
		},
	)
	require.Error(t, err)

	after, err = columns.Encode(nil, false)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func dereferenceChanges(changes []*opset.Change) []opset.Change {
	values := make([]opset.Change, len(changes))
	for i, change := range changes {
		values[i] = *change
	}

	return values
}

func changePointersFromDocument(document *opset.Document) []*opset.Change {
	changes := make([]*opset.Change, len(document.Changes))
	for i := range document.Changes {
		changes[i] = &document.Changes[i]
	}

	return changes
}

func spliceOperations(
	operations []opset.Operation,
	index int,
	deleteCount int,
	inserted []opset.Operation,
) []opset.Operation {
	result := make([]opset.Operation, 0, len(operations)-deleteCount+len(inserted))
	result = append(result, operations[:index]...)
	result = append(result, inserted...)
	result = append(result, operations[index+deleteCount:]...)

	return result
}

func compressionName(compress bool) string {
	if compress {
		return "Compress"
	}

	return "NoCompress"
}
