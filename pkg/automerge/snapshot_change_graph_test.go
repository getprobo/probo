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

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// buildSnapshotHistory authors a multi-commit history exercising the operations
// a snapshot stores differently from a change: marks, whose expand column is
// shared across every change, and deletes, which a snapshot keeps only as
// successor entries on the operations they removed.
func buildSnapshotHistory(
	t *testing.T,
	document *automerge.Document,
) {
	t.Helper()

	base := time.Unix(1786147200, 0).UTC()

	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "hello brave world"))
	_, err = document.Commit("write", base)
	require.NoError(t, err)

	require.NoError(
		t,
		text.Mark(

			0,
			5,
			"strong",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandBoth,
		),
	)
	_, err = document.Commit("mark", base.Add(time.Second))
	require.NoError(t, err)

	require.NoError(t, text.Unmark(1, 3, "strong", automerge.MarkExpandNone))
	_, err = document.Commit("unmark", base.Add(2*time.Second))
	require.NoError(t, err)

	require.NoError(t, text.Splice(5, 6, ""))
	_, err = document.Commit("delete", base.Add(3*time.Second))
	require.NoError(t, err)

	require.NoError(
		t,
		document.PutScalar(

			"counter",
			automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
		),
	)
	_, err = document.Commit("counter", base.Add(4*time.Second))
	require.NoError(t, err)

	require.NoError(t, document.Root().Increment("counter", 3))
	_, err = document.Commit("increment", base.Add(5*time.Second))
	require.NoError(t, err)
}

// TestLoadedSnapshotExposesEveryChange is the regression for the production
// outage where collaboration failed with "cannot compute changes from unknown
// heads" on every request for an affected document.
//
// A snapshot records hashes for the frontier only and names ancestry by column
// index, so before the decoder rebuilt them, every non-head change loaded
// without a hash and never entered the change graph. Reading the document still
// worked, which is why the corruption stayed invisible, but walking a head's
// ancestry immediately hit a change that was not there and the walk aborted.
func TestLoadedSnapshotExposesEveryChange(t *testing.T) {
	t.Parallel()

	reference, err := automerge.NewReference(actor(11))
	require.NoError(t, err)
	closeDocument(t, reference)

	buildSnapshotHistory(t, reference)

	// The browser client persists exactly this: a document chunk, not a stream of
	// change chunks.
	snapshot, err := reference.Save()
	require.NoError(t, err)

	referenceHeads, err := reference.Heads()
	require.NoError(t, err)

	loaded, err := automerge.Load(snapshot, actor(12))
	require.NoError(t, err)
	closeDocument(t, loaded)

	loadedHeads, err := loaded.Heads()
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, loadedHeads, "snapshot must load onto the same frontier")

	changes, err := loaded.ChangesSince(nil)
	require.NoError(t, err, "every change in a loaded snapshot must be reachable")
	assert.Len(t, changes, 6, "each commit must survive as an addressable change")

	// Replaying the rebuilt changes has to land on the same frontier, which only
	// holds when each one carries the bytes the original writer hashed.
	replayed, err := automerge.New(actor(13))
	require.NoError(t, err)
	closeDocument(t, replayed)

	require.NoError(t, replayed.ApplyChanges(changes))

	replayedHeads, err := replayed.Heads()
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, replayedHeads, "rebuilt changes must reproduce the frontier")

	// Concatenated change chunks are a document the reference can load, so this
	// proves Rust accepts the rebuilt bytes as the changes it originally wrote.
	var concatenated []byte
	for _, change := range changes {
		concatenated = append(concatenated, change.Bytes...)
	}

	roundTripped, err := automerge.LoadReference(concatenated, actor(14))
	require.NoError(t, err)
	closeDocument(t, roundTripped)

	roundTrippedHeads, err := roundTripped.Heads()
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, roundTrippedHeads)
}

// TestSnapshotMergeReportsIncrementalChanges reproduces the collaboration
// service's persist path: a canonical document restored from a stored snapshot,
// merged with a peer's document, then asked for the changes the merge added.
func TestSnapshotMergeReportsIncrementalChanges(t *testing.T) {
	t.Parallel()

	origin, err := automerge.NewReference(actor(21))
	require.NoError(t, err)
	closeDocument(t, origin)

	buildSnapshotHistory(t, origin)

	snapshot, err := origin.Save()
	require.NoError(t, err)

	canonical, err := automerge.Load(snapshot, actor(22))
	require.NoError(t, err)
	closeDocument(t, canonical)

	peer, err := automerge.Load(snapshot, actor(23))
	require.NoError(t, err)
	closeDocument(t, peer)

	peerText, err := peer.Text("body")
	require.NoError(t, err)
	require.NoError(t, peerText.Splice(0, 0, "new "))
	_, err = peer.Commit("peer edit", time.Unix(1786147300, 0).UTC())
	require.NoError(t, err)

	before, err := canonical.Heads()
	require.NoError(t, err)

	_, err = canonical.Merge(peer)
	require.NoError(t, err)

	incremental, err := canonical.ChangesSince(before)
	require.NoError(t, err, "merging a peer must not break incremental reads")
	assert.Len(t, incremental, 1, "only the peer's commit is new")

	after, err := canonical.Heads()
	require.NoError(t, err)

	peerHeads, err := peer.Heads()
	require.NoError(t, err)
	assert.Equal(t, peerHeads, after, "the merge must adopt the peer's frontier")

	canonicalText, err := canonical.Text("body")
	require.NoError(t, err)

	value, err := canonicalText.String()
	require.NoError(t, err)
	assert.Equal(t, "new hello world", value)
}
