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

package native

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// committedTextBackend returns a backend with a text object edited `edits` times,
// each edit its own committed change.
func committedTextBackend(t *testing.T, id byte, edits int) *Engine {
	t.Helper()

	ctx := context.Background()

	backend, err := NewEngine(ctx)
	require.NoError(t, err)
	require.NoError(
		t,
		backend.SetActor(
			ctx,
			[]byte{
				id, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
			},
		),
	)

	handle, err := backend.PutText(ctx, 0, "body")
	require.NoError(t, err)

	for i := range edits {
		require.NoError(t, backend.SpliceText(ctx, handle, uint32(i), 0, "x"))
		_, err = backend.Commit(ctx, "edit", time.Unix(int64(i), 0))
		require.NoError(t, err)
	}

	return backend
}

func snapshotFromBackend(backend *Engine) *Document {
	document := &Document{}

	for hash, change := range backend.state.changes {
		clone := *change
		clone.Hash = new(hash)
		document.Changes = append(document.Changes, clone)
	}

	for head := range backend.state.heads {
		document.Heads = append(document.Heads, head)
	}

	return document
}

// TestNewStateFromDocument_RebuildsInconsistentFrontier is the regression for the
// production "cannot compute changes from unknown heads" outage. A frontier that
// references a change the document does not carry must be rebuilt from the graph
// so Heads() stays consistent with changes and incremental reads keep working.
func TestNewStateFromDocument_RebuildsInconsistentFrontier(t *testing.T) {
	t.Parallel()

	document := snapshotFromBackend(committedTextBackend(t, 1, 3))

	var phantom ChangeHash

	phantom[0] = 0xAB
	document.Heads = append(document.Heads, phantom)

	state, err := NewStateFromDocument(document)
	require.NoError(t, err)

	for _, head := range state.Heads() {
		_, ok := state.changes[head]
		assert.Truef(t, ok, "head %s must exist in the change graph", head)
	}

	_, ok := state.changesSince(nil)
	assert.True(t, ok, "changesSince(nil) must succeed after rebuild")

	_, ok = state.changesSince(state.Heads())
	assert.True(t, ok, "changesSince(own heads) must succeed after rebuild")
}

// TestChangesSince_DegradesToReachablePrefix is the regression for the wedge
// where one unreachable ancestor failed the whole read. A branched history has
// an intact branch and a branch whose middle change is removed; the intact
// branch must still come back as a consistent, replayable prefix while the
// broken branch is dropped, and completeness must report false.
func TestChangesSince_DegradesToReachablePrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	actorBytes := func(id byte) []byte {
		return []byte{id, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	}

	// Shared base commit.
	base, err := NewEngine(ctx)
	require.NoError(t, err)
	require.NoError(t, base.SetActor(ctx, actorBytes(0x10)))

	handle, err := base.PutText(ctx, 0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(ctx, handle, 0, 0, "a"))
	_, err = base.Commit(ctx, "base", time.Unix(0, 0))
	require.NoError(t, err)

	shared, err := base.Save(ctx, true, true)
	require.NoError(t, err)

	// A branch two commits deep, authored by a second actor.
	deep, err := LoadEngine(ctx, shared)
	require.NoError(t, err)
	require.NoError(t, deep.SetActor(ctx, actorBytes(0x20)))

	deepHandle, _, err := deep.GetObject(ctx, 0, "body")
	require.NoError(t, err)
	require.NoError(t, deep.SpliceText(ctx, deepHandle, 1, 0, "b"))
	_, err = deep.Commit(ctx, "deep-1", time.Unix(1, 0))
	require.NoError(t, err)
	require.NoError(t, deep.SpliceText(ctx, deepHandle, 2, 0, "c"))
	_, err = deep.Commit(ctx, "deep-2", time.Unix(2, 0))
	require.NoError(t, err)

	deepSave, err := deep.Save(ctx, true, true)
	require.NoError(t, err)

	// The base adds its own branch commit, then merges the deep branch, so the
	// frontier holds two heads: the base branch and the deep branch.
	require.NoError(t, base.SpliceText(ctx, handle, 1, 0, "z"))
	_, err = base.Commit(ctx, "base-2", time.Unix(3, 0))
	require.NoError(t, err)
	_, err = base.Merge(ctx, deepSave)
	require.NoError(t, err)

	all, complete := base.state.changesSince(nil)
	require.True(t, complete)
	require.Len(t, all, 4)

	// Remove the deep branch's first change, the ancestor of its head.
	deepActor, err := NewActorID(actorBytes(0x20))
	require.NoError(t, err)

	var removed ChangeHash

	for hash, change := range base.state.changes {
		if change.Actor == deepActor && change.Sequence == 1 {
			removed = hash
		}
	}

	delete(base.state.changes, removed)

	changes, complete := base.state.changesSince(nil)
	assert.False(t, complete, "the broken branch leaves the walk incomplete")
	assert.Len(t, changes, 2, "the intact branch is still emitted as a prefix")

	for i, change := range changes {
		require.NotNil(t, change.Hash)
		assert.NotEqual(t, removed, *change.Hash)

		// Every emitted change's dependencies precede it, so the prefix replays.
		for _, dependency := range change.Dependencies {
			found := false
			for _, earlier := range changes[:i] {
				if earlier.Hash != nil && *earlier.Hash == dependency {
					found = true
				}
			}

			assert.True(t, found, "dependency of an emitted change must precede it")
		}
	}

	// The engine method must not wedge: it returns the reachable prefix.
	raw, hashes, err := base.ChangesSince(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, raw, 2)
	assert.Len(t, hashes, 2)
}

// TestChangesSince_ToleratesUnknownBaseline mirrors Rust's get_changes, which
// takes have_deps by value and never errors: an unknown baseline excludes
// nothing and the full history is returned.
func TestChangesSince_ToleratesUnknownBaseline(t *testing.T) {
	t.Parallel()

	backend := committedTextBackend(t, 2, 3)

	var unknown ChangeHash

	unknown[0] = 0xCD

	changes, ok := backend.state.changesSince([]ChangeHash{unknown})
	require.True(t, ok)
	assert.Len(t, changes, 3)
}

// TestChangesSince_ToleratesFrontierWithMissingChange guards the frontier walk:
// a head recorded without a retrievable change must not abort the computation.
// The reachable changes are still returned so an incremental read keeps working,
// and completeness reports false so sync knows to fall back to a full document.
func TestChangesSince_ToleratesFrontierWithMissingChange(t *testing.T) {
	t.Parallel()

	backend := committedTextBackend(t, 3, 2)

	var phantom ChangeHash

	phantom[0] = 0xEF
	backend.state.heads[phantom] = struct{}{}

	changes, complete := backend.state.changesSince(nil)
	assert.False(t, complete, "an unretrievable head leaves the walk incomplete")
	assert.Len(t, changes, 2, "the reachable changes are still returned")
}
