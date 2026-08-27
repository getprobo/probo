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

// The tests in this file reproduce the block-valued update_spans scenarios from
// upstream Rust automerge 0.11 (rust/automerge/tests/block_tests.rs and
// diff_marks.rs). Each runs update_spans identically on the native Go engine and
// the Rust/WASM reference engine and asserts their materialized spans agree.

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func blockSpan(attributes map[string]any) automerge.SpanInput {
	if attributes == nil {
		attributes = map[string]any{}
	}

	return automerge.SpanInput{Block: attributes}
}

func textSpan(text string, markPairs ...any) automerge.SpanInput {
	return automerge.SpanInput{Text: text, Marks: marks(markPairs...)}
}

func emptyAttrs() map[string]any {
	return map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}
}

type blockSpanScenario struct {
	name    string
	initial []automerge.SpanInput
	target  []automerge.SpanInput
	config  automerge.UpdateSpansConfig
	post    func(t *testing.T, text *automerge.Text)
}

func TestRustBlockSpans(t *testing.T) {
	t.Parallel()

	defaultConfig := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}

	scenarios := []blockSpanScenario{
		{
			name: "update_blocks_change_block_properties",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("item 1"),
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("item 2"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("item 1"),
				blockSpan(map[string]any{"type": "unordered-list-item", "parents": []any{"ordered-list-item"}, "attrs": map[string]any{"key": 1}}),
				textSpan("item 2"),
			},
			config: defaultConfig,
		},
		{
			name: "update_blocks_updates_text",
			initial: []automerge.SpanInput{
				blockSpan(emptyAttrs()),
				textSpan("first thing"),
				blockSpan(emptyAttrs()),
				textSpan("second thing"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("the first thing"),
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("the things are done"),
			},
			config: defaultConfig,
		},
		{
			name: "update_blocks_updates_marks",
			initial: []automerge.SpanInput{
				textSpan("onetwo"),
				blockSpan(emptyAttrs()),
				textSpan("threefour"),
			},
			target: []automerge.SpanInput{
				textSpan("one"),
				textSpan("two", "bold", markBool()),
				blockSpan(emptyAttrs()),
				textSpan("three"),
				textSpan("four", "italic", markBool()),
			},
			config: defaultConfig,
		},
		{
			name: "update_blocks_updates_text_and_blocks_at_once",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("hello world"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "unordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("goodbye world"),
			},
			config: defaultConfig,
		},
		{
			name: "update_spans_delete_attribute",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{"div"}}),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}}),
			},
			config: defaultConfig,
		},
		{
			name: "update_spans_diffs_marks",
			initial: []automerge.SpanInput{
				textSpan("hello", "bold", markBool()),
				textSpan(" world"),
			},
			target: []automerge.SpanInput{
				textSpan("hello", "italic", markBool()),
				textSpan(" "),
				textSpan("world", "bold", markBool(), "italic", markBool()),
			},
			config: defaultConfig,
		},
		{
			name:    "update_spans_uses_expand_config",
			initial: nil,
			target: []automerge.SpanInput{
				textSpan("hello", "bold", markBool()),
				textSpan(" world"),
			},
			config: automerge.UpdateSpansConfig{
				DefaultExpand:  automerge.MarkExpandNone,
				PerMarkExpands: map[string]automerge.MarkExpand{"bold": automerge.MarkExpandAfter},
			},
			post: func(t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(5, 0, "!"))
				require.NoError(t, text.Splice(0, 0, "Oh "))
			},
		},
		{
			name:    "mark_spans_across_block",
			initial: nil,
			target: []automerge.SpanInput{
				textSpan("bold", "bold", markBool()),
				blockSpan(nil),
				textSpan("text", "bold", markBool()),
			},
			config: defaultConfig,
		},
		{
			name:    "mark_ends_at_block_boundary",
			initial: nil,
			target: []automerge.SpanInput{
				textSpan("bold", "bold", markBool()),
				blockSpan(nil),
				textSpan("text"),
			},
			config: defaultConfig,
		},
		{
			name: "block_properties_change_with_marks",
			initial: []automerge.SpanInput{
				blockSpan(emptyAttrs()),
				textSpan("marked text"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "level": 1}),
				textSpan("marked", "bold", markBool()),
				textSpan(" text"),
			},
			config: defaultConfig,
		},
		{
			name:    "block_with_marked_content",
			initial: nil,
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "heading", "level": 1}),
				textSpan("Chapter "),
				textSpan("One", "emphasis", markBool()),
				blockSpan(map[string]any{"type": "paragraph"}),
				textSpan("This is the "),
				textSpan("first", "bold", markBool()),
				textSpan(" chapter."),
			},
			config: defaultConfig,
		},
		{
			name: "update_spans_with_only_blocks",
			initial: []automerge.SpanInput{
				blockSpan(emptyAttrs()),
				textSpan("text"),
				blockSpan(emptyAttrs()),
				textSpan("more"),
			},
			target: []automerge.SpanInput{
				blockSpan(nil),
				blockSpan(nil),
			},
			config: defaultConfig,
		},
		{
			name: "marks_survive_block_updates",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "p"}),
				textSpan("marked", "bold", markBool()),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "h1", "level": 1}),
				textSpan("marked", "bold", markBool()),
			},
			config: defaultConfig,
		},
	}

	for _, scenario := range scenarios {
		t.Run(
			scenario.name,
			func(t *testing.T) {
				t.Parallel()

				result := make(map[string][]automerge.Span)

				for _, engine := range rustParityEngines() {
					document, err := engine.open(actor(0xaa))
					require.NoError(t, err)
					closeDocument(t, document)

					text, err := document.CreateText("text")
					require.NoError(t, err)

					if scenario.initial != nil {
						require.NoError(t, text.UpdateSpans(scenario.initial, scenario.config))

						_, err = document.Commit("initial", commitTime)
						require.NoError(t, err)
					}

					require.NoError(t, text.UpdateSpans(scenario.target, scenario.config))

					_, err = document.Commit("target", commitTime)
					require.NoError(t, err)

					if scenario.post != nil {
						scenario.post(t, text)

						_, err = document.Commit("post", commitTime)
						require.NoError(t, err)
					}

					spans, err := text.Spans()
					require.NoError(t, err)

					result[engine.name] = spans
				}

				assert.Equal(t, result["reference"], result["native"])
			},
		)
	}
}

// TestRustText_InsertionsAfterNoexpandSpans reproduces
// insertions_after_noexpand_spans_are_not_marked: text appended after a block,
// with no expanding mark in scope, is reported by a diff as an unmarked splice.
func TestRustText_InsertionsAfterNoexpandSpans(t *testing.T) {
	t.Parallel()

	config := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandNone}
	heading := map[string]any{"type": "heading", "parents": []any{}, "attrs": map[string]any{}}
	paragraph := map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "")

		spans := []automerge.SpanInput{
			blockSpan(heading),
			textSpan("Heading"),
			blockSpan(paragraph),
			textSpan("a"),
			blockSpan(paragraph),
		}
		require.NoError(t, text.UpdateSpans(spans, config))

		_, err := document.Commit("spans", commitTime)
		require.NoError(t, err)

		before, err := document.Heads()
		require.NoError(t, err)
		require.NoError(t, text.Splice(11, 0, "a"))

		after, err := document.Commit("append", commitTime.Add(time.Second))
		require.NoError(t, err)

		patches, err := document.Diff(before, []automerge.Hash{after})
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 1)
	assert.Equal(t, automerge.PatchSpliceText, result["reference"][0].Action)
	assert.Empty(t, result["reference"][0].Marks)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustBlock_MarksOnSpansRespectHeads reproduces marks_on_spans_respect_heads:
// spans_at reports the marks active at a historical frontier, excluding marks
// added afterward.
func TestRustBlock_MarksOnSpansRespectHeads(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Span)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "hello world"))
		require.NoError(t, text.Mark(0, 5, "bold", markBool(), automerge.MarkExpandAfter))

		heads, err := document.Commit("bold", commitTime)
		require.NoError(t, err)

		require.NoError(t, text.Mark(5, 11, "italic", markBool(), automerge.MarkExpandAfter))

		_, err = document.Commit("italic", commitTime.Add(time.Second))
		require.NoError(t, err)

		spans, err := text.SpansAt([]automerge.Hash{heads})
		require.NoError(t, err)

		result[engine.name] = spans
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 2)
	assert.Equal(t, "hello", result["native"][0].Text)
	assert.Equal(t, map[string]any{"bold": true}, result["native"][0].Marks)
	assert.Equal(t, " world", result["native"][1].Text)
}

// TestRustBlock_DiffEmitsBlockUpdates reproduces diff_emits_block_updates: a diff
// from the empty frontier inserts the block and materializes its nested parents
// list.
func TestRustBlock_DiffEmitsBlockUpdates(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)

		block, err := text.SplitBlock(0)
		require.NoError(t, err)
		_, err = block.CreateObject("parents", automerge.ObjectTypeList)
		require.NoError(t, err)
		_, err = document.Commit("block", commitTime)
		require.NoError(t, err)

		heads, err := document.Heads()
		require.NoError(t, err)

		patches, err := document.Diff(nil, heads)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	require.NotEmpty(t, reference)

	hasBlockInsert := false
	hasParents := false

	for _, patch := range reference {
		if patch.Action == automerge.PatchInsert && len(patch.Values) == 1 &&
			patch.Values[0].Value.Object == automerge.ObjectTypeMap {
			hasBlockInsert = true
		}

		if patch.Action == automerge.PatchPutMap && patch.Key == "parents" {
			hasParents = true
		}
	}

	assert.True(t, hasBlockInsert, "expected a block insert patch")
	assert.True(t, hasParents, "expected a parents put_map patch")
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustBlock_MergeProducesBlockInsertionDiffs reproduces
// merge_produces_block_insertion_diffs: merging a peer that inserted a block
// yields a block insertion patch.
func TestRustBlock_MergeProducesBlockInsertionDiffs(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		_, err = document.Commit("seed", commitTime)
		require.NoError(t, err)

		other, err := document.Fork(actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, other)

		_, err = text.SplitBlock(0)
		require.NoError(t, err)
		_, err = document.Commit("block", commitTime.Add(time.Second))
		require.NoError(t, err)

		require.NoError(t, other.UpdateDiffCursor())
		before, err := other.Heads()
		require.NoError(t, err)
		_, err = other.Merge(document)
		require.NoError(t, err)
		after, err := other.Heads()
		require.NoError(t, err)

		patches, err := other.Diff(before, after)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	require.NotEmpty(t, reference)
	assert.Equal(t, automerge.PatchInsert, reference[0].Action)
	require.Len(t, reference[0].Values, 1)
	assert.Equal(t, automerge.ObjectTypeMap, reference[0].Values[0].Value.Object)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustBlockSpans_Noop reproduces update_blocks_noop: re-applying the current
// spans through the diff cursor produces no patches.
func TestRustBlockSpans_Noop(t *testing.T) {
	t.Parallel()

	config := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}

	spans := []automerge.SpanInput{
		blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
		textSpan("item 1"),
	}

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				document, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, document)

				text, err := document.CreateText("text")
				require.NoError(t, err)

				require.NoError(t, text.UpdateSpans(spans, config))

				_, err = document.Commit("seed", commitTime)
				require.NoError(t, err)

				require.NoError(t, document.UpdateDiffCursor())
				require.NoError(t, text.UpdateSpans(spans, config))

				patches, err := document.DiffIncremental()
				require.NoError(t, err)
				assert.Empty(t, patches)
			},
		)
	}
}
