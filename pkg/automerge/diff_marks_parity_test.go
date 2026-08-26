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

// The tests in this file reproduce the text-and-mark scenarios from upstream
// Rust automerge 0.11 (rust/automerge/tests/diff_marks.rs). Each scenario runs
// update_spans identically on the native Go engine and the Rust/WASM reference
// engine and asserts their materialized spans agree. Block-valued scenarios are
// tracked separately and excluded here.

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func markBool() automerge.Scalar {
	return automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}
}

func markStr(value string) automerge.Scalar {
	return automerge.Scalar{Type: automerge.ScalarTypeString, String: value}
}

type diffMarksScenario struct {
	name   string
	setup  func(t *testing.T, text *automerge.Text)
	spans  []automerge.SpanInput
	config automerge.UpdateSpansConfig
	post   func(t *testing.T, text *automerge.Text)
}

func TestRustDiffMarks(t *testing.T) {
	t.Parallel()

	defaultConfig := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}

	spliceSetup := func(content string) func(*testing.T, *automerge.Text) {
		return func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, content))
		}
	}

	markSetup := func(content string, marks ...markSpec) func(*testing.T, *automerge.Text) {
		return func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, content))

			for _, mark := range marks {
				require.NoError(
					t,
					text.Mark(

						mark.start,
						mark.end,
						mark.name,
						mark.value,
						automerge.MarkExpandBoth,
					),
				)
			}
		}
	}

	scenarios := []diffMarksScenario{
		{
			name:  "overlapping_marks_remove_one_keep_other",
			setup: markSetup("hello world", markSpec{"bold", markBool(), 6, 11}, markSpec{"italic", markBool(), 6, 11}),
			spans: []automerge.SpanInput{
				{Text: "hello "},
				{Text: "world", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "overlapping_marks_change_boundaries",
			setup: markSetup("hello beautiful world", markSpec{"bold", markBool(), 0, 15}, markSpec{"italic", markBool(), 6, 21}),
			spans: []automerge.SpanInput{
				{Text: "hello", Marks: marks("bold", markBool())},
				{Text: " beautiful "},
				{Text: "world", Marks: marks("italic", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "overlapping_marks_add_third_mark",
			setup: markSetup("hello world", markSpec{"bold", markBool(), 0, 11}, markSpec{"italic", markBool(), 6, 11}),
			spans: []automerge.SpanInput{
				{Text: "hel", Marks: marks("bold", markBool())},
				{Text: "lo wo", Marks: marks("bold", markBool(), "underline", markBool())},
				{Text: "rld", Marks: marks("bold", markBool(), "italic", markBool(), "underline", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "adjacent_marks_stay_separate",
			setup: spliceSetup("bold text"),
			spans: []automerge.SpanInput{
				{Text: "bold", Marks: marks("bold", markBool())},
				{Text: " "},
				{Text: "text", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_expands",
			setup: markSetup("bold text", markSpec{"bold", markBool(), 0, 4}),
			spans: []automerge.SpanInput{
				{Text: "bold text", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_contracts",
			setup: markSetup("bold text", markSpec{"bold", markBool(), 0, 9}),
			spans: []automerge.SpanInput{
				{Text: "bold", Marks: marks("bold", markBool())},
				{Text: " text"},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_shifts_position",
			setup: markSetup("bold text", markSpec{"bold", markBool(), 0, 4}),
			spans: []automerge.SpanInput{
				{Text: "text "},
				{Text: "bold", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_splits",
			setup: markSetup("bold text here", markSpec{"bold", markBool(), 0, 14}),
			spans: []automerge.SpanInput{
				{Text: "bold", Marks: marks("bold", markBool())},
				{Text: " text "},
				{Text: "here", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "adjacent_marks_merge",
			setup: markSetup("bold text", markSpec{"bold", markBool(), 0, 4}, markSpec{"bold", markBool(), 5, 9}),
			spans: []automerge.SpanInput{
				{Text: "bold text", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "different_adjacent_marks",
			setup: spliceSetup("bolditalic"),
			spans: []automerge.SpanInput{
				{Text: "bold", Marks: marks("bold", markBool())},
				{Text: "italic", Marks: marks("italic", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_on_empty_string",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: "", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_on_whitespace",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: " ", Marks: marks("bold", markBool())},
				{Text: "\n", Marks: marks("italic", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "removing_all_text_from_marked_span",
			setup: markSetup("hello world", markSpec{"bold", markBool(), 0, 5}),
			spans: []automerge.SpanInput{
				{Text: " world"},
			},
			config: defaultConfig,
		},
		{
			name:  "nested_marks",
			setup: spliceSetup("italic bold and italic just italic"),
			spans: []automerge.SpanInput{
				{Text: "italic ", Marks: marks("italic", markBool())},
				{Text: "bold and italic", Marks: marks("italic", markBool(), "bold", markBool())},
				{Text: " just italic", Marks: marks("italic", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "many_marks_on_same_text",
			setup: spliceSetup("formatted"),
			spans: []automerge.SpanInput{
				{Text: "formatted", Marks: marks(
					"bold",
					markBool(),
					"italic",
					markBool(),
					"underline",
					markBool(),
					"link",
					markStr("https://example.com"),
				)},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_value_changes_link_url",
			setup: markSetup("click here", markSpec{"link", markStr("https://old.com"), 0, 10}),
			spans: []automerge.SpanInput{
				{Text: "click here", Marks: marks("link", markStr("https://new.com"))},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_value_changes_color",
			setup: markSetup("colored", markSpec{"color", markStr("red"), 0, 7}),
			spans: []automerge.SpanInput{
				{Text: "colored", Marks: marks("color", markStr("blue"))},
			},
			config: defaultConfig,
		},
		{
			name:  "mark_value_type_changes",
			setup: markSetup("text", markSpec{"custom", markBool(), 0, 4}),
			spans: []automerge.SpanInput{
				{Text: "text", Marks: marks("custom", markStr("value"))},
			},
			config: defaultConfig,
		},
		{
			name:  "marks_on_emoji",
			setup: spliceSetup("Hello 👨‍👩‍👧‍👦 world"),
			spans: []automerge.SpanInput{
				{Text: "Hello "},
				{Text: "👨‍👩‍👧‍👦", Marks: marks("emoji", markBool())},
				{Text: " world"},
			},
			config: defaultConfig,
		},
		{
			name:  "marks_on_combining_characters",
			setup: spliceSetup("café"),
			spans: []automerge.SpanInput{
				{Text: "café", Marks: marks("accented", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "unmark_part_of_range",
			setup: markSetup("bold text here", markSpec{"bold", markBool(), 0, 14}),
			spans: []automerge.SpanInput{
				{Text: "bold", Marks: marks("bold", markBool())},
				{Text: " text "},
				{Text: "here", Marks: marks("bold", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "unmark_creates_gaps",
			setup: markSetup("a b c d e", markSpec{"mark", markBool(), 0, 9}),
			spans: []automerge.SpanInput{
				{Text: "a", Marks: marks("mark", markBool())},
				{Text: " b "},
				{Text: "c", Marks: marks("mark", markBool())},
				{Text: " d "},
				{Text: "e", Marks: marks("mark", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "complex_unicode_text",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: "Hello "},
				{Text: "😊", Marks: marks("emoji", markBool())},
				{Text: " 世界 ", Marks: marks("chinese", markBool())},
				{Text: "🌍", Marks: marks("emoji", markBool())},
				{Text: " مرحبا", Marks: marks("arabic", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "empty_spans_between_marks",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: "a", Marks: marks("mark", markBool())},
				{Text: ""},
				{Text: "b", Marks: marks("mark", markBool())},
			},
			config: defaultConfig,
		},
		{
			name:  "marks_with_different_values_same_name",
			setup: spliceSetup("red blue green"),
			spans: []automerge.SpanInput{
				{Text: "red", Marks: marks("color", markStr("red"))},
				{Text: " "},
				{Text: "blue", Marks: marks("color", markStr("blue"))},
				{Text: " "},
				{Text: "green", Marks: marks("color", markStr("green"))},
			},
			config: defaultConfig,
		},
		{
			name:  "marks_with_expand_none_at_boundaries",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: "text", Marks: marks("mark", markBool())},
			},
			config: automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandNone},
			post: func(t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(0, 0, "before "))
				require.NoError(t, text.Splice(11, 0, " after"))
			},
		},
		{
			name:  "multiple_marks_different_expand_behaviors",
			setup: spliceSetup(""),
			spans: []automerge.SpanInput{
				{Text: "text", Marks: marks("before", markBool(), "after", markBool(), "none", markBool())},
			},
			config: automerge.UpdateSpansConfig{
				DefaultExpand: automerge.MarkExpandAfter,
				PerMarkExpands: map[string]automerge.MarkExpand{
					"before": automerge.MarkExpandBefore,
					"after":  automerge.MarkExpandAfter,
					"none":   automerge.MarkExpandNone,
				},
			},
			post: func(t *testing.T, text *automerge.Text) {
				require.NoError(t, text.Splice(0, 0, "a"))
				require.NoError(t, text.Splice(5, 0, "b"))
			},
		},
		{
			name:  "update_spans_which_inserts_at_the_end_of_expand_mark",
			setup: markSetup("hello world", markSpec{"bold", markBool(), 6, 11}),
			spans: []automerge.SpanInput{
				{Text: "hello "},
				{Text: "wworldd", Marks: marks("bold", markBool())},
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

					scenario.setup(t, text)
					_, err = document.Commit("setup", commitTime)
					require.NoError(t, err)

					require.NoError(t, text.UpdateSpans(scenario.spans, scenario.config))
					_, err = document.Commit("update", commitTime)
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

// TestRustDiffMarks_Idempotent reproduces idempotent_update_spans: repeating the
// same update_spans call produces no additional changes.
func TestRustDiffMarks_Idempotent(t *testing.T) {
	t.Parallel()

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

				spans := []automerge.SpanInput{
					{Text: "hello ", Marks: marks("bold", markBool())},
					{Text: "world", Marks: marks("italic", markBool())},
				}
				config := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}

				require.NoError(t, text.UpdateSpans(spans, config))
				_, err = document.Commit("first", commitTime)
				require.NoError(t, err)

				first, err := document.Heads()
				require.NoError(t, err)

				require.NoError(t, text.UpdateSpans(spans, config))
				second, err := document.Heads()
				require.NoError(t, err)

				require.NoError(t, text.UpdateSpans(spans, config))
				third, err := document.Heads()
				require.NoError(t, err)

				assert.Equal(t, headHexes(first), headHexes(second))
				assert.Equal(t, headHexes(first), headHexes(third))
			},
		)
	}
}

// TestRustDiffMarks_Alternating reproduces alternating_mark_changes: repeatedly
// adding and removing a mark on the same text converges on the final span set.
func TestRustDiffMarks_Alternating(t *testing.T) {
	t.Parallel()

	config := automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}
	result := make(map[string][]automerge.Span)

	rounds := [][]automerge.SpanInput{
		{{Text: "text", Marks: marks("bold", markBool())}},
		{{Text: "text"}},
		{{Text: "text", Marks: marks("italic", markBool())}},
	}

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "text"))
		_, err = document.Commit("seed", commitTime)
		require.NoError(t, err)

		for index, round := range rounds {
			require.NoError(t, text.UpdateSpans(round, config))
			_, err = document.Commit("round", commitTime.Add(time.Duration(index+1)*time.Second))
			require.NoError(t, err)
		}

		spans, err := text.Spans()
		require.NoError(t, err)

		result[engine.name] = spans
	}

	assert.Equal(t, result["reference"], result["native"])
}

type markSpec struct {
	name  string
	value automerge.Scalar
	start uint32
	end   uint32
}

func marks(pairs ...any) map[string]automerge.Scalar {
	result := make(map[string]automerge.Scalar, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		result[pairs[i].(string)] = pairs[i+1].(automerge.Scalar)
	}

	return result
}

func headHexes(heads []automerge.Hash) []string {
	hexes := make([]string, len(heads))
	for i, head := range heads {
		hexes[i] = head.String()
	}

	return hexes
}
