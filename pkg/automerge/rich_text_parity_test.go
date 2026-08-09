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

// The tests in this file reproduce upstream Rust rich-text tests from automerge
// 0.10 (rust/automerge/tests/block_tests.rs and text.rs) that assert on the
// materialized span stream after mark, splice, and block operations. Each
// scenario runs identically on the native Go engine and the Rust/WASM reference
// engine and asserts their span output agrees.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func markTrue() automerge.Scalar {
	return automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}
}

// richTextSpans builds a rich-text document with the given closure on each
// engine and returns the resulting span stream keyed by engine name.
func richTextSpans(
	t *testing.T,
	build func(t *testing.T, ctx context.Context, text *automerge.Text),
) map[string][]automerge.Span {
	t.Helper()

	ctx := context.Background()
	spans := make(map[string][]automerge.Span)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)

		build(t, ctx, text)

		_, err = document.CommitNow(ctx, "rich text")
		require.NoError(t, err)

		result, err := text.Spans(ctx)
		require.NoError(t, err)

		spans[engine.name] = result
	}

	return spans
}

// TestRustRichText_MarksInSpansCrossBlockMarkers reproduces
// marks_in_spans_cross_block_markers.
func TestRustRichText_MarksInSpansCrossBlockMarkers(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "lix"))
		require.NoError(t, text.Mark(ctx, 0, 3, "bold", markTrue(), automerge.MarkExpandAfter))
		_, err := text.SplitBlock(ctx, 1)
		require.NoError(t, err)
	})

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_MarkBehaviorOnDeleteInsert reproduces
// test_mark_behavior_on_delete_insert.
func TestRustRichText_MarkBehaviorOnDeleteInsert(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "hello"))
		require.NoError(t, text.Mark(ctx, 0, 5, "bold", markTrue(), automerge.MarkExpandBoth))
		require.NoError(t, text.Splice(ctx, 0, 5, ""))
		require.NoError(t, text.Splice(ctx, 0, 0, "hi"))
	})

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hi", spans["native"][0].Text)
	assert.Empty(t, spans["native"][0].Marks)
}

// TestRustRichText_SpansConsolidateEmptyDueToDeletedMarks reproduces
// spans_consolidates_marks_which_are_empty_due_to_deleted_marks.
func TestRustRichText_SpansConsolidateEmptyDueToDeletedMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "hello middle world"))
		require.NoError(t, text.Mark(ctx, 0, 9, "bold", markTrue(), automerge.MarkExpandNone))
		require.NoError(t, text.Mark(ctx, 9, 18, "italic", markTrue(), automerge.MarkExpandNone))
		require.NoError(t, text.Unmark(ctx, 6, 9, "bold", automerge.MarkExpandNone))
		require.NoError(t, text.Unmark(ctx, 9, 12, "italic", automerge.MarkExpandNone))
	})

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_SpansConsolidateDeletedThenEmptyMarks reproduces
// spans_consolidates_marks_with_deleted_marks_followed_by_empty_marks.
func TestRustRichText_SpansConsolidateDeletedThenEmptyMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "hello world"))
		require.NoError(t, text.Mark(ctx, 0, 6, "bold", markTrue(), automerge.MarkExpandNone))
		require.NoError(t, text.Unmark(ctx, 0, 6, "bold", automerge.MarkExpandNone))
	})

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hello world", spans["native"][0].Text)
	assert.Empty(t, spans["native"][0].Marks)
}

// TestRustRichText_SpansConsolidateEmptyThenDeletedMarks reproduces
// spans_consolidates_marks_with_empty_marks_followed_by_deleted_marks.
func TestRustRichText_SpansConsolidateEmptyThenDeletedMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "hello world"))
		require.NoError(t, text.Mark(ctx, 6, 11, "bold", markTrue(), automerge.MarkExpandNone))
		require.NoError(t, text.Unmark(ctx, 6, 11, "bold", automerge.MarkExpandNone))
	})

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hello world", spans["native"][0].Text)
}

// TestRustRichText_EmptyMarksBeforeBlockMarker reproduces
// empty_marks_before_block_marker_dont_repeat_text.
func TestRustRichText_EmptyMarksBeforeBlockMarker(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		_, err := text.SplitBlock(ctx, 0)
		require.NoError(t, err)
		_, err = text.SplitBlock(ctx, 0)
		require.NoError(t, err)
		require.NoError(t, text.Mark(ctx, 1, 1, "strong", markTrue(), automerge.MarkExpandBoth))
		require.NoError(t, text.Splice(ctx, 2, 0, "a"))
	})

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 3)
	assert.Equal(t, automerge.SpanTypeBlock, spans["native"][0].Type)
	assert.Equal(t, automerge.SpanTypeBlock, spans["native"][1].Type)
	assert.Equal(t, automerge.SpanTypeText, spans["native"][2].Type)
	assert.Equal(t, "a", spans["native"][2].Text)
}

// TestRustRichText_ComplexBlockProperties reproduces
// text_complex_block_properties.
func TestRustRichText_ComplexBlockProperties(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		block, err := text.SplitBlock(ctx, 0)
		require.NoError(t, err)
		require.NoError(t, block.PutValue(ctx, "type", automerge.Value{
			Type: automerge.ValueTypeText,
			Text: "ordered-list-item",
		}))
		require.NoError(t, block.PutValue(ctx, "parents", automerge.Value{
			Type: automerge.ValueTypeList,
			List: []automerge.Value{{Type: automerge.ValueTypeText, Text: "div"}},
		}))
	})

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_MarkCreatedAfterInsertion reproduces
// mark_created_after_insertion.
func TestRustRichText_MarkCreatedAfterInsertion(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "12345"))
		require.NoError(t, text.Mark(ctx, 1, 2, "strong", markTrue(), automerge.MarkExpandBoth))
		require.NoError(t, text.Mark(ctx, 3, 4, "strong", markTrue(), automerge.MarkExpandBoth))
	})

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_SpansConsolidatedWithZeroLengthSpans reproduces
// spans_are_consolidated_in_the_presence_of_zero_length_spans.
func TestRustRichText_SpansConsolidatedWithZeroLengthSpans(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(t, func(t *testing.T, ctx context.Context, text *automerge.Text) {
		require.NoError(t, text.Splice(ctx, 0, 0, "1234"))
		require.NoError(t, text.Mark(ctx, 1, 1, "strong", markTrue(), automerge.MarkExpandBoth))
		require.NoError(t, text.Mark(ctx, 2, 2, "strong", markTrue(), automerge.MarkExpandBoth))
	})

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "1234", spans["native"][0].Text)
}
