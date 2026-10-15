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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestIsolate_KeepsNextOpAfterRemoteMerge(t *testing.T) {
	t.Parallel()

	t.Run("apply changes", func(t *testing.T) {
		t.Parallel()

		engine, fullMax := isolateAndImportRemote(t, false)
		require.Greater(t, isolatedWriteCounter(t, engine), fullMax)
	})

	t.Run("merge", func(t *testing.T) {
		t.Parallel()

		engine, fullMax := isolateAndImportRemote(t, true)
		require.Greater(t, isolatedWriteCounter(t, engine), fullMax)
	})

	t.Run("pending counters stay ahead of a smaller merge", func(t *testing.T) {
		t.Parallel()

		engine, hash := committedScalar(t, "base", "base")
		require.NoError(t, engine.PutScalar(0, "second", mustScalar(t, "second")))
		_, err := engine.Commit("second", time.Time{})
		require.NoError(t, err)
		require.NoError(t, engine.PutScalar(0, "third", mustScalar(t, "third")))
		_, err = engine.Commit("third", time.Time{})
		require.NoError(t, err)
		heads := mustHeads(t, engine)

		peer, err := engine.Fork([]byte("peer"))
		require.NoError(t, err)
		require.NoError(t, peer.PutScalar(0, "peer", mustScalar(t, "peer")))
		_, err = peer.Commit("peer", time.Time{})
		require.NoError(t, err)
		changes, _, err := peer.ChangesSince(heads)
		require.NoError(t, err)
		require.NotEmpty(t, changes)

		require.NoError(t, engine.Isolate([][32]byte{hash}))
		require.NoError(t, engine.PutScalar(0, "pending-a", mustScalar(t, "a")))
		require.NoError(t, engine.PutScalar(0, "pending-b", mustScalar(t, "b")))
		require.Len(t, engine.pending, 2)
		lastPending := engine.pending[1].ID.Counter

		require.NoError(t, engine.ApplyChanges(changes))
		fullMax := engine.fullState.maxOpGlobal()
		require.Less(t, fullMax, lastPending)

		counter := isolatedWriteCounter(t, engine)
		require.Greater(t, counter, lastPending)
		require.Greater(t, counter, fullMax)
	})
}

func committedScalar(t *testing.T, actor string, key string) (*Engine, [32]byte) {
	t.Helper()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte(actor)))
	require.NoError(t, engine.PutScalar(0, key, mustScalar(t, key)))
	hash, err := engine.Commit(key, time.Time{})
	require.NoError(t, err)

	return engine, hash
}

func isolateAndImportRemote(t *testing.T, useMerge bool) (*Engine, uint64) {
	t.Helper()

	engine, hash := committedScalar(t, "base", "base")
	peer, err := engine.Fork([]byte("peer"))
	require.NoError(t, err)
	require.NoError(t, peer.PutScalar(0, "peer", mustScalar(t, "peer")))
	require.NoError(t, peer.PutScalar(0, "peer-again", mustScalar(t, "peer-again")))
	_, err = peer.Commit("peer", time.Time{})
	require.NoError(t, err)
	changes, _, err := peer.ChangesSince([][32]byte{hash})
	require.NoError(t, err)
	require.NotEmpty(t, changes)

	require.NoError(t, engine.Isolate([][32]byte{hash}))
	pinnedMax := engine.state.maxOpGlobal()

	if useMerge {
		merged := make([]byte, 0)
		for _, change := range changes {
			merged = append(merged, change...)
		}

		_, err = engine.Merge(merged)
		require.NoError(t, err)
	} else {
		require.NoError(t, engine.ApplyChanges(changes))
	}

	require.True(t, engine.isolationActive)
	fullMax := engine.fullState.maxOpGlobal()
	require.Greater(t, fullMax, pinnedMax)
	require.Greater(t, fullMax, engine.state.maxOpGlobal())

	return engine, fullMax
}

func isolatedWriteCounter(t *testing.T, engine *Engine) uint64 {
	t.Helper()

	require.NoError(t, engine.PutScalar(0, "isolated", mustScalar(t, "isolated")))
	require.NotEmpty(t, engine.pending)

	return engine.pending[len(engine.pending)-1].ID.Counter
}

func mustScalar(t *testing.T, value string) []byte {
	t.Helper()

	encoded, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: value},
	)
	require.NoError(t, err)

	return encoded
}
