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

// The tests in this file reproduce the cross-engine block behaviors from the
// upstream JavaScript suite (javascript/test/block_test.ts). Each runs on the
// native Go engine and the Rust/WASM reference engine and asserts they agree.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// TestJSBlock_UpdateSpansScenarios reproduces the update_spans block behaviors
// from the JavaScript block suite by comparing the materialized spans.
func TestJSBlock_UpdateSpansScenarios(t *testing.T) {
	t.Parallel()

	none := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandNone}
	overriding := automerge.UpdateSpansConfig{
		DefaultExpand:  automerge.MarkExpandNone,
		PerMarkExpands: map[string]automerge.MarkExpand{"bold": automerge.MarkExpandBoth},
	}
	defaultConfig := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}

	scenarios := []blockSpanScenario{
		{
			name: "allows_updating_all_blocks_at_once",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("first thing"),
				blockSpan(map[string]any{"type": "ordered-list-item", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("second thing"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("the first thing"),
				blockSpan(map[string]any{"type": "unordered-list-item", "parents": []any{"ordered-list-item"}, "attrs": map[string]any{}}),
				textSpan("the second thing"),
			},
			config: defaultConfig,
		},
		{
			name:    "should_update_marks",
			initial: []automerge.SpanInput{textSpan("hello world")},
			target: []automerge.SpanInput{
				textSpan("hello", "bold", markBool()),
				textSpan(" "),
				textSpan(" world", "italic", markBool()),
			},
			config: defaultConfig,
		},
		{
			name:    "configuring_default_expand",
			initial: nil,
			target: []automerge.SpanInput{
				textSpan("hello", "bold", markBool()),
				textSpan(" world"),
			},
			config: none,
			post: func(ctx context.Context, t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(ctx, 5, 0, "!"))
			},
		},
		{
			name:    "override_default_expand_per_mark",
			initial: nil,
			target: []automerge.SpanInput{
				textSpan("hello", "bold", markBool()),
				textSpan(" world"),
			},
			config: overriding,
			post: func(ctx context.Context, t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(ctx, 5, 0, "!"))
			},
		},
		{
			name: "updates_document_on_block_attribute_change",
			initial: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{}, "attrs": map[string]any{}}),
				textSpan("item"),
			},
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"type": "paragraph", "parents": []any{"ordered-list-item"}, "attrs": map[string]any{}}),
				textSpan("item"),
			},
			config: defaultConfig,
			post: func(ctx context.Context, t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(ctx, 0, 1, "A"))
			},
		},
		{
			name:    "small_values_in_block_attributes",
			initial: nil,
			target: []automerge.SpanInput{
				blockSpan(map[string]any{"smallnum": 1.401298464324817e-45}),
			},
			config: defaultConfig,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := make(map[string][]automerge.Span)

			for _, engine := range rustParityEngines() {
				document, err := engine.open(ctx, actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, document)

				text, err := document.CreateText(ctx, "text")
				require.NoError(t, err)

				if scenario.initial != nil {
					require.NoError(t, text.UpdateSpans(ctx, scenario.initial, scenario.config))
					_, err = document.Commit(ctx, "initial", commitTime)
					require.NoError(t, err)
				}

				require.NoError(t, text.UpdateSpans(ctx, scenario.target, scenario.config))
				_, err = document.Commit(ctx, "target", commitTime)
				require.NoError(t, err)

				if scenario.post != nil {
					scenario.post(ctx, t, text)
					_, err = document.Commit(ctx, "post", commitTime)
					require.NoError(t, err)
				}

				spans, err := text.Spans(ctx)
				require.NoError(t, err)

				result[engine.name] = spans
			}

			assert.Equal(t, result["reference"], result["native"])
		})
	}
}

// TestJSBlock_OmittingConfigParts reproduces "should allow omitting any part of
// the update spans config": partial and absent configs are all accepted.
func TestJSBlock_OmittingConfigParts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Span)

	spans := []automerge.SpanInput{
		textSpan("hello", "bold", markBool()),
		textSpan(" world"),
	}
	configs := []automerge.UpdateSpansConfig{
		{DefaultExpand: automerge.MarkExpandNone},
		{PerMarkExpands: map[string]automerge.MarkExpand{"bold": automerge.MarkExpandNone}},
		{},
	}

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)

		for _, config := range configs {
			require.NoError(t, text.UpdateSpans(ctx, spans, config))
		}

		_, err = document.Commit(ctx, "rounds", commitTime)
		require.NoError(t, err)

		got, err := text.Spans(ctx)
		require.NoError(t, err)

		result[engine.name] = got
	}

	assert.Equal(t, result["reference"], result["native"])
}

// TestJSBlock_ShowHistoricalMarks reproduces "should show historical marks":
// viewing spans at a past frontier omits marks added afterward.
func TestJSBlock_ShowHistoricalMarks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Span)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "hello world")
		require.NoError(t, text.Mark(ctx, 0, 5, "bold", markBool(), automerge.MarkExpandAfter))
		heads, err := document.Commit(ctx, "bold", commitTime)
		require.NoError(t, err)

		require.NoError(t, text.Mark(ctx, 5, 11, "italic", markBool(), automerge.MarkExpandAfter))
		_, err = document.Commit(ctx, "italic", commitTime.Add(time.Second))
		require.NoError(t, err)

		spans, err := text.SpansAt(ctx, []automerge.Hash{heads})
		require.NoError(t, err)

		result[engine.name] = spans
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 2)
	assert.Equal(t, map[string]any{"bold": true}, result["native"][0].Marks)
}
