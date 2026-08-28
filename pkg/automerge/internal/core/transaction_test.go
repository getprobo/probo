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

func TestEngineRollback_PreservesSurvivingHandles(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)

	stable, err := engine.PutObject(0, "stable", "map")
	require.NoError(t, err)
	_, err = engine.Commit("stable", time.Time{})
	require.NoError(t, err)

	transient, err := engine.PutObject(0, "transient", "map")
	require.NoError(t, err)

	cancelled, err := engine.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)

	value, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarString, String: "yes"})
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(stable, "after", value))

	_, err = engine.PutObject(transient, "invalid", "map")
	require.Error(t, err)

	replacement, err := engine.PutObject(0, "replacement", "map")
	require.NoError(t, err)
	assert.Greater(t, replacement, transient)
}

func TestEngineRollback_PreservesIsolationStates(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)

	value, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarString, String: "seed"})
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", value))
	hash, err := engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	require.NoError(t, engine.Isolate([][32]byte{hash}))
	pinned := engine.state
	full := engine.fullState

	pending, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "pending"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", pending))

	cancelled, err := engine.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)
	assert.Same(t, pinned, engine.state)
	assert.Same(t, full, engine.fullState)
	assert.True(t, engine.isolationActive)
}

func TestEngineEmptyCommit_UpdatesCanonicalColumnsWhileIsolated(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "seed"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", value))
	hash, err := engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	require.NoError(t, engine.Isolate([][32]byte{hash}))
	emptyHash, err := engine.EmptyCommit("isolated", time.Time{})
	require.NoError(t, err)
	assert.True(t, engine.fullState.hasChange(opset.ChangeHash(emptyHash)))
	assert.Len(t, engine.columns.changes, 2)

	require.NoError(t, engine.Integrate())
	data, err := engine.Save(true, false)
	require.NoError(t, err)
	loaded, err := LoadEngine(data)
	require.NoError(t, err)
	assert.True(t, loaded.state.hasChange(opset.ChangeHash(emptyHash)))
}

func TestEngineRollback_InvalidatesCreatedSequenceCaches(t *testing.T) {
	t.Parallel()

	for _, objectType := range []string{"list", "text"} {
		t.Run(
			objectType,
			func(t *testing.T) {
				t.Parallel()

				engine, err := NewEngine()
				require.NoError(t, err)

				handle, err := engine.PutObject(0, "sequence", objectType)
				require.NoError(t, err)

				object := engine.objects[handle].OpID
				engine.state.sequence(object)
				engine.state.insertOrder(object)
				engine.state.sequenceValues(object)
				elements := engine.state.sequenceElements(object)
				engine.state.sequenceOffsets(object, elements)
				engine.state.insertOrderPositions(object)

				assert.Contains(t, engine.state.sequenceCache, object)
				assert.Contains(t, engine.state.insertOrderCache, object)
				assert.Contains(t, engine.state.insertOrderPositionCache, object)
				assert.Contains(t, engine.state.sequenceValuesCache, object)
				assert.Contains(t, engine.state.sequenceElementsCache, object)
				assert.Contains(t, engine.state.sequenceOffsetCache, object)

				cancelled, err := engine.Rollback()
				require.NoError(t, err)
				assert.Equal(t, uint64(1), cancelled)

				assert.NotContains(t, engine.state.sequenceCache, object)
				assert.NotContains(t, engine.state.insertOrderCache, object)
				assert.NotContains(t, engine.state.insertOrderPositionCache, object)
				assert.NotContains(t, engine.state.sequenceValuesCache, object)
				assert.NotContains(t, engine.state.sequenceElementsCache, object)
				assert.NotContains(t, engine.state.sequenceOffsetCache, object)
			},
		)
	}
}

func TestStateUndoPending_PreservesRemainingSuperseder(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	predecessor := opset.Operation{
		ID:     opset.OpID{Actor: actor, Counter: 1},
		Action: opset.ActionSet,
	}
	first := opset.Operation{
		ID:           opset.OpID{Actor: actor, Counter: 2},
		Action:       opset.ActionSet,
		Predecessors: []opset.OpID{predecessor.ID},
	}
	second := opset.Operation{
		ID:           opset.OpID{Actor: actor, Counter: 3},
		Action:       opset.ActionSet,
		Predecessors: []opset.OpID{predecessor.ID},
	}

	state := NewState()
	state.operations[predecessor.ID] = predecessor
	require.NoError(t, state.applyPending([]opset.Operation{first, second}))
	require.True(t, state.isSuperseded(predecessor.ID))

	state.undoPending([]opset.Operation{second})
	assert.True(t, state.isSuperseded(predecessor.ID))

	state.undoPending([]opset.Operation{first})
	assert.False(t, state.isSuperseded(predecessor.ID))
}

func TestStateUndoPending_ClearsPendingSupersessionChain(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	base := opset.Operation{
		ID:     opset.OpID{Actor: actor, Counter: 1},
		Action: opset.ActionSet,
	}
	first := opset.Operation{
		ID:           opset.OpID{Actor: actor, Counter: 2},
		Action:       opset.ActionSet,
		Predecessors: []opset.OpID{base.ID},
	}
	second := opset.Operation{
		ID:           opset.OpID{Actor: actor, Counter: 3},
		Action:       opset.ActionSet,
		Predecessors: []opset.OpID{first.ID},
	}

	state := NewState()
	state.operations[base.ID] = base
	require.NoError(t, state.applyPending([]opset.Operation{first, second}))

	state.undoPending([]opset.Operation{first, second})

	assert.False(t, state.isSuperseded(base.ID))
	assert.False(t, state.isSuperseded(first.ID))
	assert.Contains(t, state.operations, base.ID)
	assert.NotContains(t, state.operations, first.ID)
	assert.NotContains(t, state.operations, second.ID)
}

func TestStateUndoPending_DoesNotSupersedeCounterOnIncrement(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	counter := opset.Operation{
		ID:     opset.OpID{Actor: actor, Counter: 1},
		Action: opset.ActionSet,
		Value:  &opset.Scalar{Type: opset.ScalarCounter},
	}
	increment := opset.Operation{
		ID:           opset.OpID{Actor: actor, Counter: 2},
		Action:       opset.ActionIncrement,
		Predecessors: []opset.OpID{counter.ID},
	}

	state := NewState()
	state.operations[counter.ID] = counter
	require.NoError(t, state.applyPending([]opset.Operation{increment}))
	assert.False(t, state.isSuperseded(counter.ID))

	state.undoPending([]opset.Operation{increment})
	assert.False(t, state.isSuperseded(counter.ID))
	assert.Contains(t, state.operations, counter.ID)
}
