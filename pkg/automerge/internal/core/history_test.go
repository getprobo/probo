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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestApplyChanges_RemovesQueuedDuplicateActorSequence(t *testing.T) {
	t.Parallel()

	t.Run("fast path", func(t *testing.T) {
		t.Parallel()

		fixture := queuedDuplicateFixture(t)
		before := fixture.engine.state.changeCount()

		err := fixture.engine.ApplyChanges([][]byte{fixture.conflicting})
		require.ErrorContains(
			t,
			err,
			fmt.Sprintf("duplicate actor sequence %d", fixture.sequence),
		)
		require.Equal(t, before, fixture.engine.state.changeCount())
		requireQueued(t, fixture.engine, fixture.unrelated, fixture.orphan)
		require.NotContains(t, fixture.engine.queuedChanges, fixture.successor)
	})

	t.Run("pending fallback", func(t *testing.T) {
		t.Parallel()

		fixture := queuedDuplicateFixture(t)
		value, err := encodeScalarWire(
			opset.Scalar{Type: opset.ScalarString, String: "pending"},
		)
		require.NoError(t, err)
		require.NoError(t, fixture.engine.PutScalar(0, "pending", value))
		require.NotEmpty(t, fixture.engine.pending)

		before := fixture.engine.state.changeCount()
		err = fixture.engine.ApplyChanges([][]byte{fixture.conflicting})
		require.ErrorContains(
			t,
			err,
			fmt.Sprintf(
				"actor sequence is %d, expected %d",
				fixture.sequence,
				fixture.sequence+1,
			),
		)
		require.Equal(t, before, fixture.engine.state.changeCount())
		requireQueued(t, fixture.engine, fixture.unrelated, fixture.orphan)
		require.NotContains(t, fixture.engine.queuedChanges, fixture.successor)
		require.NotContains(t, fixture.engine.queuedChanges, fixture.conflict)
	})
}

func TestApplyChanges_KeepsQueuedBranchOnSequenceGap(t *testing.T) {
	t.Parallel()

	t.Run("fast path", func(t *testing.T) {
		t.Parallel()

		fixture := queuedGapFixture(t)
		err := fixture.engine.ApplyChanges([][]byte{fixture.gap})
		require.ErrorContains(t, err, "actor sequence is 4, expected 2")
		requireQueued(t, fixture.engine, fixture.branch, fixture.later)
	})

	t.Run("pending fallback", func(t *testing.T) {
		t.Parallel()

		fixture := queuedGapFixture(t)
		value, err := encodeScalarWire(
			opset.Scalar{Type: opset.ScalarString, String: "pending"},
		)
		require.NoError(t, err)
		require.NoError(t, fixture.engine.PutScalar(0, "pending", value))
		require.NotEmpty(t, fixture.engine.pending)

		err = fixture.engine.ApplyChanges([][]byte{fixture.gap})
		require.ErrorContains(
			t,
			err,
			"cannot apply merged native change: actor sequence is 4, expected 2",
		)
		requireQueued(t, fixture.engine, fixture.branch, fixture.later, fixture.gapHash)
	})
}

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

type queuedDuplicate struct {
	engine      *Engine
	sequence    uint64
	orphan      opset.ChangeHash
	successor   opset.ChangeHash
	unrelated   opset.ChangeHash
	conflicting []byte
	conflict    opset.ChangeHash
}

func queuedDuplicateFixture(t *testing.T) queuedDuplicate {
	t.Helper()

	engine, real := committedChange(t)
	var missing opset.ChangeHash
	missing[0] = 1

	orphan := real
	orphan.Message = "orphan"
	orphan.Dependencies = []opset.ChangeHash{missing}
	orphan = mustEncodeChange(t, orphan)

	successor := opset.Change{
		Actor:        real.Actor,
		Sequence:     real.Sequence + 1,
		StartOp:      real.MaxOp + 1,
		MaxOp:        real.MaxOp,
		Message:      "successor",
		Dependencies: []opset.ChangeHash{*orphan.Hash},
	}
	successor = mustEncodeChange(t, successor)

	other, err := opset.NewActorID([]byte("other"))
	require.NoError(t, err)
	unrelated := mustEncodeChange(
		t,
		opset.Change{
			Actor:        other,
			Sequence:     1,
			StartOp:      1,
			MaxOp:        0,
			Message:      "unrelated",
			Dependencies: []opset.ChangeHash{missing},
		},
	)

	conflicting := real
	conflicting.Message = "conflict"
	conflicting = mustEncodeChange(t, conflicting)
	require.NotEqual(t, *orphan.Hash, *conflicting.Hash)

	require.NoError(
		t,
		engine.ApplyChanges([][]byte{orphan.Raw, successor.Raw, unrelated.Raw}),
	)
	requireQueued(t, engine, *orphan.Hash, *successor.Hash, *unrelated.Hash)

	return queuedDuplicate{
		engine:      engine,
		sequence:    real.Sequence,
		orphan:      *orphan.Hash,
		successor:   *successor.Hash,
		unrelated:   *unrelated.Hash,
		conflicting: conflicting.Raw,
		conflict:    *conflicting.Hash,
	}
}

type queuedGap struct {
	engine  *Engine
	branch  opset.ChangeHash
	later   opset.ChangeHash
	gap     []byte
	gapHash opset.ChangeHash
}

func queuedGapFixture(t *testing.T) queuedGap {
	t.Helper()

	engine, real := committedChange(t)
	var missing opset.ChangeHash
	missing[0] = 2

	branch := mustEncodeChange(
		t,
		opset.Change{
			Actor:        real.Actor,
			Sequence:     2,
			StartOp:      real.MaxOp + 1,
			MaxOp:        real.MaxOp,
			Message:      "branch",
			Dependencies: []opset.ChangeHash{missing},
		},
	)
	later := mustEncodeChange(
		t,
		opset.Change{
			Actor:        real.Actor,
			Sequence:     5,
			StartOp:      real.MaxOp + 1,
			MaxOp:        real.MaxOp,
			Message:      "later",
			Dependencies: []opset.ChangeHash{*branch.Hash},
		},
	)
	gap := mustEncodeChange(
		t,
		opset.Change{
			Actor:        real.Actor,
			Sequence:     4,
			StartOp:      real.MaxOp + 1,
			MaxOp:        real.MaxOp,
			Message:      "gap",
			Dependencies: []opset.ChangeHash{*real.Hash},
		},
	)

	require.NoError(t, engine.ApplyChanges([][]byte{branch.Raw, later.Raw}))
	requireQueued(t, engine, *branch.Hash, *later.Hash)

	return queuedGap{
		engine:  engine,
		branch:  *branch.Hash,
		later:   *later.Hash,
		gap:     gap.Raw,
		gapHash: *gap.Hash,
	}
}

func committedChange(t *testing.T) (*Engine, opset.Change) {
	t.Helper()

	engine, _ := committedScalar(t, "remote", "real")
	require.NotEmpty(t, engine.appended)

	raw := append([]byte(nil), engine.appended[len(engine.appended)-1]...)
	document, err := storage.DecodePartial(raw)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	change := document.Changes[0]
	require.NotNil(t, change.Hash)
	require.Equal(t, uint64(1), change.Sequence)
	require.NotZero(t, change.MaxOp)

	return engine, change
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

func mustEncodeChange(t *testing.T, change opset.Change) opset.Change {
	t.Helper()

	change.Hash = nil
	change.Raw = nil
	change.DependencyIndexes = nil
	_, err := storage.EncodeChange(&change)
	require.NoError(t, err)
	require.NotNil(t, change.Hash)
	require.NotEmpty(t, change.Raw)

	return change
}

func requireQueued(t *testing.T, engine *Engine, hashes ...opset.ChangeHash) {
	t.Helper()

	require.Len(t, engine.queuedChanges, len(hashes))
	for _, hash := range hashes {
		require.Contains(t, engine.queuedChanges, hash)
	}
}
