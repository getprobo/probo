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

// This file reproduces the orphan save/load tests
// (rust/automerge/tests/test_save_load_orphans.rs): a document that holds an
// orphan change (a change whose dependency has not been applied) must retain
// that orphan across a save by default, so applying the missing dependency after
// a reload resolves it, and must be able to discard it on request.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// orphanScenario builds a document that has applied one change ("value") and
// holds an orphan change ("value3") whose dependency ("value2") is missing. It
// returns that document and the missing dependency change so a caller can apply
// it after a save/load round trip.
func orphanScenario(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
) (*automerge.Document, []byte) {
	t.Helper()

	doc1, err := engine.open(ctx, actor(0x01))
	require.NoError(t, err)

	putRoot(t, ctx, doc1, "key", "value", "value", commitTime)

	doc2, err := doc1.Fork(ctx, actor(0x02))
	require.NoError(t, err)
	closeDocument(t, doc2)

	_, err = doc2.SaveIncremental(ctx)
	require.NoError(t, err)

	putRoot(t, ctx, doc2, "key", "value2", "value2", commitTime)

	missing, err := doc2.SaveIncremental(ctx)
	require.NoError(t, err)

	putRoot(t, ctx, doc2, "key", "value3", "value3", commitTime)

	dependent, err := doc2.SaveIncremental(ctx)
	require.NoError(t, err)

	// Applying the second remote change orphans it because doc1 lacks the first.
	_, err = doc1.LoadIncremental(ctx, dependent)
	require.NoError(t, err)

	return doc1, missing
}

func rootKey(t *testing.T, ctx context.Context, document *automerge.Document) string {
	t.Helper()

	value, err := document.Root().Scalar(ctx, "key")
	require.NoError(t, err)

	return value.String
}

// TestRustOrphans_SaveOrphanedChanges reproduces save_orphaned_changes: the
// orphan survives a default save, so applying the missing dependency after a
// reload yields value3.
func TestRustOrphans_SaveOrphanedChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc, missing := orphanScenario(t, ctx, engine)
			closeDocument(t, doc)

			saved, err := doc.Save(ctx)
			require.NoError(t, err)

			loaded, err := engine.load(ctx, saved, actor(0x03))
			require.NoError(t, err)
			closeDocument(t, loaded)

			_, err = loaded.LoadIncremental(ctx, missing)
			require.NoError(t, err)

			require.Equal(t, "value3", rootKey(t, ctx, loaded))
		})
	}
}

// TestRustOrphans_DiscardOrphans reproduces discard_orphans: saving with
// retain_orphans disabled drops the orphan, so after a reload only the value
// from the applicable change (value2) is seen.
func TestRustOrphans_DiscardOrphans(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc, missing := orphanScenario(t, ctx, engine)
			closeDocument(t, doc)

			saved, err := doc.Save(ctx, automerge.DiscardOrphans())
			require.NoError(t, err)

			loaded, err := engine.load(ctx, saved, actor(0x03))
			require.NoError(t, err)
			closeDocument(t, loaded)

			_, err = loaded.LoadIncremental(ctx, missing)
			require.NoError(t, err)

			require.Equal(t, "value2", rootKey(t, ctx, loaded))
		})
	}
}
