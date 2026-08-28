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

func TestLoadEngine_BuildsCanonicalColumnsBeforeCompatibilityState(t *testing.T) {
	t.Parallel()

	engine, first := isolationFixture(t)
	data, err := engine.Save(true, false)
	require.NoError(t, err)

	loaded, err := LoadEngine(data)
	require.NoError(t, err)

	row, ok := loaded.columns.changeRows[opset.ChangeHash(first)]
	require.True(t, ok)
	change, ok := loaded.state.change(opset.ChangeHash(first))
	require.True(t, ok)
	require.Equal(t, loaded.columns.changes[row].Hash, change.Hash)
	require.Equal(t, loaded.columns.changes[row].Raw, change.Raw)
	assert.NotEmpty(t, change.Raw)

	value, err := loaded.Hydrate()
	require.NoError(t, err)
	assert.NotNil(t, value)
}

func TestEngineFork_SharesColumnsAndClonesQueryIndexes(t *testing.T) {
	t.Parallel()

	engine, _ := isolationFixture(t)
	fork, err := engine.Fork([]byte{0xf0})
	require.NoError(t, err)

	assert.NotSame(t, engine.state, fork.state)
	assert.Same(t, engine.columns.snapshot, fork.columns.snapshot)
	assert.Same(t, fork.columns, fork.state.columns)
	require.NotEmpty(t, engine.columns.changes)
	forkChange, ok := fork.lookupChange(*engine.columns.changes[0].Hash)
	require.True(t, ok)
	assert.Equal(t, &engine.columns.changes[0], forkChange)

	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "fork-only"},
	)
	require.NoError(t, err)
	require.NoError(t, fork.PutScalar(0, "fork", value))
	_, err = fork.Commit("fork", time.Time{})
	require.NoError(t, err)

	sourceKeys, err := engine.Keys(0)
	require.NoError(t, err)
	assert.NotContains(t, sourceKeys, "fork")
}

func TestIsolationView_MatchesReplayAcrossLoadedSnapshot(t *testing.T) {
	t.Parallel()

	engine, first := isolationFixture(t)
	data, err := engine.Save(true, false)
	require.NoError(t, err)
	loaded, err := LoadEngine(data)
	require.NoError(t, err)

	heads := []opset.ChangeHash{opset.ChangeHash(first)}
	replayed, ok := loaded.state.at(heads)
	require.True(t, ok)
	filtered, ok := newIsolationView(loaded.state, loaded.columns, heads)
	require.True(t, ok)

	replayedEngine := &Engine{
		state:   replayed,
		objects: map[uint32]opset.ObjectID{0: opset.RootObject()},
	}
	filteredEngine := &Engine{
		state:   filtered,
		objects: map[uint32]opset.ObjectID{0: opset.RootObject()},
	}
	replayedValue, err := replayedEngine.Hydrate()
	require.NoError(t, err)
	filteredValue, err := filteredEngine.Hydrate()
	require.NoError(t, err)

	assert.Equal(t, replayedValue, filteredValue)
	assert.Equal(t, replayed.Heads(), filtered.Heads())
	assert.Equal(t, replayed.actorSequence, filtered.actorSequence)
	assert.Equal(t, replayed.superseded, filtered.superseded)
}

func TestIsolationView_PreservesIntegrateAndMergedHistory(t *testing.T) {
	t.Parallel()

	engine, first := isolationFixture(t)
	peer, err := engine.Fork([]byte{0xe0})
	require.NoError(t, err)
	peerValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "peer"},
	)
	require.NoError(t, err)
	require.NoError(t, peer.PutScalar(0, "peer", peerValue))
	_, err = peer.Commit("peer", time.Time{})
	require.NoError(t, err)
	changes, _, err := peer.ChangesSince(nil)
	require.NoError(t, err)

	require.NoError(t, engine.Isolate([][32]byte{first}))
	isolatedValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "isolated"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "isolated", isolatedValue))
	_, err = engine.Commit("isolated", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.ApplyChanges(changes))
	require.NoError(t, engine.Integrate())

	keys, err := engine.Keys(0)
	require.NoError(t, err)
	assert.Contains(t, keys, "peer")
	assert.Contains(t, keys, "isolated")
}

func isolationFixture(t *testing.T) (*Engine, [32]byte) {
	t.Helper()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "first"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", value))
	first, err := engine.Commit("first", time.Time{})
	require.NoError(t, err)

	text, err := engine.PutObject(0, "text", "text")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "second"))
	_, err = engine.Commit("second", time.Time{})
	require.NoError(t, err)

	replacement, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "third"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", replacement))
	_, err = engine.Commit("third", time.Time{})
	require.NoError(t, err)

	return engine, first
}
