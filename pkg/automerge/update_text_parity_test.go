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

// The tests in this file reproduce upstream Rust update_text tests from
// automerge 0.10 (rust/automerge/tests/text.rs). update_text computes a minimal
// grapheme-aware diff so concurrent edits to disjoint regions merge cleanly.
// Each scenario runs identically on the native Go engine and the Rust/WASM
// reference engine and asserts their materialized text and change history agree.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRustText_SimpleUpdateText reproduces simple_update_text: two forks edit
// disjoint words with update_text and merge into a document combining both.
func TestRustText_SimpleUpdateText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	merged := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 0, 0, "Hello, world!"))
		_, err = document.Commit(ctx, "seed", commitTime)
		require.NoError(t, err)

		other, err := document.Fork(ctx, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, other)

		otherObject, err := other.Root().Object(ctx, "text")
		require.NoError(t, err)
		otherText, err := otherObject.Text()
		require.NoError(t, err)
		require.NoError(t, otherText.Update(ctx, "Goodbye, world!"))
		_, err = other.Commit(ctx, "goodbye", commitTime)
		require.NoError(t, err)

		require.NoError(t, text.Update(ctx, "Hello, friends!"))
		_, err = document.Commit(ctx, "friends", commitTime)
		require.NoError(t, err)

		_, err = document.Merge(ctx, other)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		merged[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "Goodbye, friends!", merged["reference"])
	assert.Equal(t, merged["reference"], merged["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustText_UpdateTextBigOleGraphemes reproduces update_text_big_ole_graphemes:
// update_text treats emoji ZWJ sequences as single grapheme clusters, so two
// forks that swap the family emoji merge into both new families side by side.
func TestRustText_UpdateTextBigOleGraphemes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	merged := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 0, 0, "left👨‍👩‍👦right"))
		_, err = document.Commit(ctx, "seed", commitTime)
		require.NoError(t, err)

		other, err := document.Fork(ctx, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, other)

		otherObject, err := other.Root().Object(ctx, "text")
		require.NoError(t, err)
		otherText, err := otherObject.Text()
		require.NoError(t, err)
		require.NoError(t, otherText.Update(ctx, "left👨‍👩‍👧right"))
		_, err = other.Commit(ctx, "girl", commitTime)
		require.NoError(t, err)

		require.NoError(t, text.Update(ctx, "left👨‍👩‍👦‍👦right"))
		_, err = document.Commit(ctx, "boys", commitTime)
		require.NoError(t, err)

		_, err = document.Merge(ctx, other)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		merged[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "left👨‍👩‍👧👨‍👩‍👦‍👦right", merged["reference"])
	assert.Equal(t, merged["reference"], merged["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}
