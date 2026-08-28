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
	actual, ok := engine.state.operation(committed.ID)
	require.True(t, ok)
	assert.Equal(t, committed, actual)
	assert.Equal(t, engine.columns.currentHeads(), engine.state.Heads())

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
	assert.Equal(t, engine.columns.currentHeads(), engine.state.Heads())
}

func TestColumnarAuthority_CanonicalRowsWinOverOverlayRows(
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
	assert.Equal(t, "x", actual.Value.String)
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

func TestColumnarAuthority_RepairsStalePersistentRowLookup(t *testing.T) {
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

	actual, ok := engine.columns.operation(expected.ID)
	require.True(t, ok)
	assert.Equal(t, expected, actual)
}
