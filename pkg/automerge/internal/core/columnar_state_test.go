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
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestColumnarState_TracksCommitRollbackForkAndMerge(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, engine)

	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "one"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", value))
	_, err = engine.Commit("map", time.Time{})
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, engine)

	text, err := engine.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "hello"))
	_, err = engine.Commit("text", time.Time{})
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, engine)

	before, err := engine.columns.snapshot.Encode(engine.unknownColumns, false)
	require.NoError(t, err)
	pending, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "pending"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", pending))
	_, err = engine.Rollback()
	require.NoError(t, err)
	after, err := engine.columns.snapshot.Encode(engine.unknownColumns, false)
	require.NoError(t, err)
	assert.Equal(t, before, after)
	requireColumnarStateEquivalent(t, engine)

	fork, err := engine.Fork([]byte{42})
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, fork)
	forkValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "fork"},
	)
	require.NoError(t, err)
	require.NoError(t, fork.PutScalar(0, "fork", forkValue))
	_, err = fork.Commit("fork", time.Time{})
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, fork)
	requireColumnarStateEquivalent(t, engine)

	changes, _, err := fork.ChangesSince(nil)
	require.NoError(t, err)
	require.NoError(t, engine.ApplyChanges(changes))
	requireColumnarStateEquivalent(t, engine)
}

func TestColumnarState_DivergedTextUsesObjectBatch(t *testing.T) {
	t.Parallel()

	left, err := NewEngine()
	require.NoError(t, err)
	text, err := left.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, left.SpliceText(text, 0, 0, "hello"))
	_, err = left.Commit("base", time.Time{})
	require.NoError(t, err)
	baseHeads, err := left.Heads()
	require.NoError(t, err)

	right, err := left.Fork([]byte{42})
	require.NoError(t, err)
	rightText, objectType, err := right.GetObject(0, "body")
	require.NoError(t, err)
	require.Equal(t, "text", objectType)
	require.NoError(t, left.SpliceText(text, 5, 0, " left"))
	_, err = left.Commit("left", time.Time{})
	require.NoError(t, err)
	require.NoError(t, right.SpliceText(rightText, 5, 0, " right"))
	_, err = right.Commit("right", time.Time{})
	require.NoError(t, err)

	changes, _, err := right.ChangesSince(baseHeads)
	require.NoError(t, err)
	fallbacks := left.columns.globalOrderFallbacks
	require.NoError(t, left.ApplyChanges(changes))
	assert.Equal(t, fallbacks, left.columns.globalOrderFallbacks)
	requireColumnarStateEquivalent(t, left)
}

func requireColumnarStateEquivalent(t *testing.T, engine *Engine) {
	t.Helper()

	state := engine.state
	if engine.isolationActive && engine.fullState != nil {
		state = engine.fullState
	}

	changes, ok := state.allChanges()
	require.True(t, ok)
	require.Len(t, engine.columns.changes, len(changes))
	for i, change := range changes {
		require.NotNil(t, change.Hash)
		require.NotNil(t, engine.columns.changes[i].Hash)
		assert.Equal(t, *change.Hash, *engine.columns.changes[i].Hash)
	}

	order := state.documentOperationOrder()
	require.Len(t, engine.columns.operations, len(order))
	operations := make([]opset.Operation, len(order))
	for i, identifier := range order {
		assert.Equal(t, identifier, engine.columns.operations[i].ID)
		operation, exists := state.operation(identifier)
		require.True(t, exists)
		operation.Predecessors = nil
		operation.Successors = append(
			[]opset.OpID(nil),
			state.successorIndex[identifier]...,
		)
		slices.SortFunc(
			operation.Successors,
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
		operations[i] = operation
		assert.Equal(t, operation.ID, engine.columns.operations[i].ID)
		assert.Equal(
			t,
			operation.Successors,
			engine.columns.operations[i].Successors,
		)
	}

	document := &opset.Document{
		Changes:        make([]opset.Change, len(changes)),
		Heads:          state.Heads(),
		UnknownColumns: cloneRawColumns(engine.unknownColumns),
	}
	for i, change := range changes {
		document.Changes[i] = *change
	}
	expected, err := storage.EncodePreparedDocument(document, operations, false)
	require.NoError(t, err)
	actual, err := engine.columns.snapshot.Encode(
		engine.unknownColumns,
		false,
	)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
