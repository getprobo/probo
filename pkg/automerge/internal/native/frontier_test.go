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
func committedTextBackend(t *testing.T, id byte, edits int) *Backend {
	t.Helper()

	ctx := context.Background()

	backend, err := NewBackend(ctx)
	require.NoError(t, err)
	require.NoError(t, backend.SetActor(ctx, []byte{
		id, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}))

	handle, err := backend.PutText(ctx, 0, "body")
	require.NoError(t, err)

	for i := range edits {
		require.NoError(t, backend.SpliceText(ctx, handle, uint32(i), 0, "x"))
		_, err = backend.Commit(ctx, "edit", time.Unix(int64(i), 0))
		require.NoError(t, err)
	}

	return backend
}

func snapshotFromBackend(backend *Backend) *Document {
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
// a head recorded without a retrievable change contributes nothing and must not
// abort the incremental computation.
func TestChangesSince_ToleratesFrontierWithMissingChange(t *testing.T) {
	t.Parallel()

	backend := committedTextBackend(t, 3, 2)

	var phantom ChangeHash

	phantom[0] = 0xEF
	backend.state.heads[phantom] = struct{}{}

	changes, ok := backend.state.changesSince(nil)
	require.True(t, ok)
	assert.Len(t, changes, 2)
}
