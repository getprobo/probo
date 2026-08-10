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

// The tests in this file reproduce the transaction rollback behaviors from the
// upstream Rust owned-transaction suite (rust/automerge/src/transaction/
// owned_transaction.rs), asserting the native Go and Rust/WASM reference engines
// agree on the number of discarded operations and the resulting state.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// TestRustMarkPatches_AtEndOfText reproduces mark_patches_at_end_of_text: a mark
// applied at the end of text and loaded incrementally into another document
// produces a single Mark patch through the diff cursor.
func TestRustMarkPatches_AtEndOfText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		author, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, author)

		text, err := author.CreateText(ctx, "text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 0, 0, "sample"))
		_, err = author.Commit(ctx, "seed", commitTime)
		require.NoError(t, err)

		saved, err := author.Save(ctx)
		require.NoError(t, err)

		follower, err := engine.load(ctx, saved, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, follower)

		require.NoError(t, text.Mark(ctx, 5, 6, "bold", markBool(), automerge.MarkExpandAfter))
		_, err = author.Commit(ctx, "mark", commitTime.Add(time.Second))
		require.NoError(t, err)

		incremental, err := author.SaveIncremental(ctx)
		require.NoError(t, err)

		require.NoError(t, follower.UpdateDiffCursor(ctx))
		_, err = follower.LoadIncremental(ctx, incremental)
		require.NoError(t, err)

		patches, err := follower.DiffIncremental(ctx)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	require.Len(t, reference, 1)
	assert.Equal(t, automerge.PatchMark, reference[0].Action)
	require.Len(t, reference[0].Marks, 1)
	assert.Equal(t, "bold", reference[0].Marks[0].Name)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustTransaction_RollbackDiscardsOps reproduces rollback_discards_ops: a
// rollback with no pending writes discards nothing and preserves prior state.
func TestRustTransaction_RollbackDiscardsOps(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cancelled := make(map[string]uint64)
	values := make(map[string]string)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.Root().PutScalar(
			ctx,
			"keep",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "yes"},
		))
		_, err = document.Commit(ctx, "keep", commitTime)
		require.NoError(t, err)

		count, err := document.Rollback(ctx)
		require.NoError(t, err)

		cancelled[engine.name] = count

		value, err := document.Root().Scalar(ctx, "keep")
		require.NoError(t, err)

		values[engine.name] = value.String
	}

	assert.Equal(t, uint64(0), cancelled["reference"])
	assert.Equal(t, "yes", values["reference"])
	assert.Equal(t, cancelled["reference"], cancelled["native"])
	assert.Equal(t, values["reference"], values["native"])
}

// TestRustTransaction_RollbackUndoesWrites reproduces rollback_undoes_writes: a
// rollback discards the uncommitted write and reports the discarded op count.
func TestRustTransaction_RollbackUndoesWrites(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cancelled := make(map[string]uint64)
	present := make(map[string]bool)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.Root().PutScalar(
			ctx,
			"gone",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "soon"},
		))

		count, err := document.Rollback(ctx)
		require.NoError(t, err)

		cancelled[engine.name] = count

		_, err = document.Root().Scalar(ctx, "gone")
		present[engine.name] = err == nil
	}

	assert.Equal(t, uint64(1), cancelled["reference"])
	assert.False(t, present["reference"])
	assert.Equal(t, cancelled["reference"], cancelled["native"])
	assert.Equal(t, present["reference"], present["native"])
}
