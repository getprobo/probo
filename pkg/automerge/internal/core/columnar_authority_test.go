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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestColumnarAuthority_CommitAndPendingOverlay(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "committed"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", value))
	_, err = engine.Commit("base", time.Time{})
	require.NoError(t, err)

	require.Same(t, engine.columns, engine.state.columns)
	require.Len(t, engine.columns.operations, 1)
	committed := engine.columns.operations[0]
	assert.NotContains(t, engine.state.operations, committed.ID)
	assert.Empty(t, engine.state.changes)
	actual, ok := engine.state.operation(committed.ID)
	require.True(t, ok)
	assert.Equal(t, committed, actual)
	assert.Equal(t, engine.columns.heads, engine.state.Heads())

	pendingValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "pending"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", pendingValue))
	require.Len(t, engine.pending, 1)
	pending := engine.pending[0]
	assert.Contains(t, engine.state.operations, pending.ID)
	actual, ok = engine.state.operation(pending.ID)
	require.True(t, ok)
	assert.Equal(t, pending, actual)

	cancelled, err := engine.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)

	_, ok = engine.state.operation(pending.ID)
	assert.False(t, ok)
	assert.Equal(t, engine.columns.heads, engine.state.Heads())
}

func TestColumnarAuthority_PendingOverlayWinsOverCanonicalRows(
	t *testing.T,
) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "x"))
	_, err = engine.Commit("text", time.Time{})
	require.NoError(t, err)

	var retained opset.Operation

	for _, operation := range engine.columns.operations {
		if operation.Insert && operation.Value != nil {
			retained = operation
			break
		}
	}

	require.NotEqual(t, opset.OpID{}, retained.ID)
	require.NotContains(t, engine.state.operations, retained.ID)

	shadow := retained
	shadow.Value = &opset.Scalar{Type: opset.ScalarString, String: "shadow"}
	engine.state.operations[retained.ID] = shadow

	actual, ok := engine.state.operation(retained.ID)
	require.True(t, ok)
	assert.Equal(t, "shadow", actual.Value.String)
	assert.Equal(t, "shadow", engine.state.operations[retained.ID].Value.String)
}

func TestColumnarAuthority_StandaloneStateUsesMapFallback(t *testing.T) {
	t.Parallel()

	state := NewState()
	property := "key"
	operation := opset.Operation{
		ID: opset.OpID{
			Actor:   opset.ActorID("actor"),
			Counter: 1,
		},
		Object: opset.RootObject(),
		Key:    opset.Key{Property: &property},
		Action: opset.ActionSet,
		Value:  &opset.Scalar{Type: opset.ScalarString, String: "value"},
	}
	state.operations[operation.ID] = operation
	state.operationIDs[operation.ID] = struct{}{}

	actual, ok := state.operation(operation.ID)
	require.True(t, ok)
	assert.Equal(t, operation, actual)
	assert.Nil(t, state.columns)
	assert.Equal(t, 1, state.operationCount())
}

func TestColumnarAuthority_LoadedAndAppliedRowsUseColumns(t *testing.T) {
	t.Parallel()

	source, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "source"},
	)
	require.NoError(t, err)
	require.NoError(t, source.PutScalar(0, "key", value))
	_, err = source.Commit("source", time.Time{})
	require.NoError(t, err)

	saved, err := source.Save(true, false)
	require.NoError(t, err)
	loaded, err := LoadEngine(saved)
	require.NoError(t, err)
	requireCanonicalRowsOwnedByColumns(t, loaded)

	target, err := NewEngine()
	require.NoError(t, err)
	changes, _, err := source.ChangesSince(nil)
	require.NoError(t, err)
	require.NoError(t, target.ApplyChanges(changes))
	requireCanonicalRowsOwnedByColumns(t, target)
}

func TestColumnarAuthority_CommittedDeleteUsesColumnarSemanticIndex(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "value"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", value))
	_, err = engine.Commit("put", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.DeleteMap(0, "key"))
	require.Len(t, engine.pending, 1)
	deleted := engine.pending[0]
	_, err = engine.Commit("delete", time.Time{})
	require.NoError(t, err)

	assert.Empty(t, engine.state.operations)
	assert.Empty(t, engine.state.changes)
	assert.Contains(t, engine.columns.nonCanonicalOperations, deleted.ID)
	actual, ok := engine.lookupOperation(deleted.ID)
	require.True(t, ok)
	assert.Equal(t, deleted, actual)
}

func TestColumnarAuthority_ChangeOverlayWinsOverColumns(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	_, err = engine.EmptyCommit("base", time.Time{})
	require.NoError(t, err)
	require.Len(t, engine.columns.changes, 1)
	hash := *engine.columns.changes[0].Hash
	change := &opset.Change{Hash: &hash, Message: "overlay"}
	engine.state.changes[hash] = change

	actual, ok := engine.lookupChange(hash)
	require.True(t, ok)
	assert.Same(t, change, actual)
}

func TestColumnarAuthority_RejectsStalePersistentRowLookup(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "value"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", value))
	_, err = engine.Commit("value", time.Time{})
	require.NoError(t, err)

	require.Len(t, engine.columns.operations, 1)
	expected := engine.columns.operations[0]
	engine.columns.operationRows = &operationRowIndex{
		rows: map[opset.OpID]int{expected.ID: 99},
	}

	_, ok := engine.columns.operation(expected.ID)
	assert.False(t, ok)
}

func requireCanonicalRowsOwnedByColumns(t *testing.T, engine *Engine) {
	t.Helper()

	assert.Empty(t, engine.state.operations)
	assert.Empty(t, engine.state.changes)

	for _, operation := range engine.columns.operations {
		assert.NotContains(t, engine.state.operations, operation.ID)
		actual, ok := engine.lookupOperation(operation.ID)
		require.True(t, ok)
		assert.Equal(t, operation, actual)
	}

	for i := range engine.columns.changes {
		change := &engine.columns.changes[i]
		require.NotNil(t, change.Hash)
		actual, ok := engine.lookupChange(*change.Hash)
		require.True(t, ok)
		assert.Same(t, change, actual)
	}
}
