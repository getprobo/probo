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

package collaboration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	persistTestBase = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Acme"}]}]}`
	persistTestEdit = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Acme and Beta"}]}]}`
	persistTestLast = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Acme, Beta and Gamma"}]}]}`
)

// newPersistTestDocuments returns a canonical document and a live document
// loaded from the canonical snapshot, the state a room is in after Open.
func newPersistTestDocuments(t *testing.T) (*automerge.Document, *automerge.Document) {
	t.Helper()

	seedActor, err := automerge.NewActorID()
	require.NoError(t, err)

	seed, err := automerge.New(seedActor)
	require.NoError(t, err)

	require.NoError(t, writeSpans(seed, persistTestBase, true))
	_, err = seed.Commit("Seed collaboration document", time.Now())
	require.NoError(t, err)

	snapshot, err := seed.Save()
	require.NoError(t, err)
	require.NoError(t, seed.Close())

	canonical, err := loadAutomergeDocument(context.Background(), snapshot)
	require.NoError(t, err)
	t.Cleanup(func() { _ = canonical.Close() })

	live, err := loadAutomergeDocument(context.Background(), snapshot)
	require.NoError(t, err)
	t.Cleanup(func() { _ = live.Close() })

	return canonical, live
}

func commitPersistTestContent(t *testing.T, document *automerge.Document, content string) {
	t.Helper()

	require.NoError(t, writeSpans(document, content, false))
	_, err := document.Commit("Set document content", time.Now())
	require.NoError(t, err)
}

func TestReadStableChanges_MatchesMerge(t *testing.T) {
	t.Parallel()

	canonical, live := newPersistTestDocuments(t)

	canonicalHeads, err := canonical.Heads()
	require.NoError(t, err)

	commitPersistTestContent(t, live, persistTestEdit)
	commitPersistTestContent(t, live, persistTestLast)

	covered, err := live.HasHeads(canonicalHeads)
	require.NoError(t, err)
	require.True(t, covered)

	heads, changes, stable, err := readStableChanges(live, canonicalHeads)
	require.NoError(t, err)
	require.True(t, stable)
	require.Len(t, changes, 2)

	fastContent, err := materializeAtHeads(live, heads)
	require.NoError(t, err)

	// Replaying the fast-path changes on the canonical document must land on
	// the same heads and content as the merge the slow path performs.
	require.NoError(t, canonical.ApplyChanges(changes))

	replayedHeads, err := canonical.Heads()
	require.NoError(t, err)
	assert.Equal(t, heads, replayedHeads)

	replayedContent, err := materialize(context.Background(), canonical)
	require.NoError(t, err)
	assert.Equal(t, replayedContent, fastContent)
	assert.Contains(t, fastContent, "Gamma")
}

func TestReadStableChanges_NothingNew(t *testing.T) {
	t.Parallel()

	canonical, live := newPersistTestDocuments(t)

	canonicalHeads, err := canonical.Heads()
	require.NoError(t, err)

	heads, changes, stable, err := readStableChanges(live, canonicalHeads)
	require.NoError(t, err)
	require.True(t, stable)
	assert.Empty(t, changes)
	assert.Equal(t, canonicalHeads, heads)
}

func TestMergeCanonical_CatchesUpLiveDocument(t *testing.T) {
	t.Parallel()

	canonical, live := newPersistTestDocuments(t)

	// Another instance advanced the canonical document while this one made
	// its own edit. The live document does not hold the canonical heads,
	// which is what makes persist catch up before it appends.
	commitPersistTestContent(t, canonical, persistTestEdit)
	require.NoError(t, live.PutString("note", "local edit"))
	_, err := live.Commit("Local edit", time.Now())
	require.NoError(t, err)

	canonicalHeads, err := canonical.Heads()
	require.NoError(t, err)

	covered, err := live.HasHeads(canonicalHeads)
	require.NoError(t, err)
	require.False(t, covered)

	snapshot, err := canonical.Save()
	require.NoError(t, err)

	require.NoError(
		t,
		mergeCanonical(
			context.Background(),
			&coredata.CollaborationDocument{Snapshot: snapshot},
			nil,
			live,
		),
	)

	covered, err = live.HasHeads(canonicalHeads)
	require.NoError(t, err)
	require.True(t, covered)

	// Only the local edit is left to append, and replaying it on the
	// canonical document converges both sides.
	heads, changes, stable, err := readStableChanges(live, canonicalHeads)
	require.NoError(t, err)
	require.True(t, stable)
	require.Len(t, changes, 1)

	require.NoError(t, canonical.ApplyChanges(changes))

	replayedHeads, err := canonical.Heads()
	require.NoError(t, err)
	assert.Equal(t, heads, replayedHeads)

	content, err := materializeAtHeads(live, heads)
	require.NoError(t, err)
	assert.Contains(t, content, "Beta")
}

func TestMaterializeAtHeads_PinsOlderHeads(t *testing.T) {
	t.Parallel()

	_, live := newPersistTestDocuments(t)

	commitPersistTestContent(t, live, persistTestEdit)

	pinned, err := live.Heads()
	require.NoError(t, err)

	// An edit landing after the heads were read must not leak into the
	// content written for those heads.
	commitPersistTestContent(t, live, persistTestLast)

	content, err := materializeAtHeads(live, pinned)
	require.NoError(t, err)
	assert.Contains(t, content, "Beta")
	assert.NotContains(t, content, "Gamma")
}
