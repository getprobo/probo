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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// TestRustRichText_LoadRejectsMarkEndBeforeBegin reproduces
// load_rejects_historical_zero_width_mark_end_before_begin.
func TestRustRichText_LoadRejectsMarkEndBeforeBegin(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/fixtures/broken_zero_width_mark.automerge")
	require.NoError(t, err)

	for _, engine := range rustParityEngines() {
		document, err := engine.load(data, actor(1))
		assert.Error(t, err, engine.name)

		if document != nil {
			require.NoError(t, document.Close())
		}
	}
}

func markTrue() automerge.Scalar {
	return automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}
}

func markString(value string) automerge.Scalar {
	return automerge.Scalar{Type: automerge.ScalarTypeString, String: value}
}

// richTextSpans builds a rich-text document with the given closure on each
// engine and returns the resulting span stream keyed by engine name.
func richTextSpans(
	t *testing.T,
	build func(t *testing.T, text *automerge.Text),
) map[string][]automerge.Span {
	t.Helper()

	spans := make(map[string][]automerge.Span)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)

		build(t, text)

		_, err = document.CommitNow("rich text")
		require.NoError(t, err)

		result, err := text.Spans()
		require.NoError(t, err)

		spans[engine.name] = result
	}

	return spans
}

// TestRustRichText_MarksInSpansCrossBlockMarkers reproduces
// marks_in_spans_cross_block_markers.
func TestRustRichText_MarksInSpansCrossBlockMarkers(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "lix"))
			require.NoError(t, text.Mark(0, 3, "bold", markTrue(), automerge.MarkExpandAfter))
			_, err := text.SplitBlock(1)
			require.NoError(t, err)
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_MarkBehaviorOnDeleteInsert reproduces
// test_mark_behavior_on_delete_insert.
func TestRustRichText_MarkBehaviorOnDeleteInsert(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "hello"))
			require.NoError(t, text.Mark(0, 5, "bold", markTrue(), automerge.MarkExpandBoth))
			require.NoError(t, text.Splice(0, 5, ""))
			require.NoError(t, text.Splice(0, 0, "hi"))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hi", spans["native"][0].Text)
	assert.Empty(t, spans["native"][0].Marks)
}

// TestRustRichText_SpansConsolidateEmptyDueToDeletedMarks reproduces
// spans_consolidates_marks_which_are_empty_due_to_deleted_marks.
func TestRustRichText_SpansConsolidateEmptyDueToDeletedMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "hello middle world"))
			require.NoError(t, text.Mark(0, 9, "bold", markTrue(), automerge.MarkExpandNone))
			require.NoError(t, text.Mark(9, 18, "italic", markTrue(), automerge.MarkExpandNone))
			require.NoError(t, text.Unmark(6, 9, "bold", automerge.MarkExpandNone))
			require.NoError(t, text.Unmark(9, 12, "italic", automerge.MarkExpandNone))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_SpansConsolidateDeletedThenEmptyMarks reproduces
// spans_consolidates_marks_with_deleted_marks_followed_by_empty_marks.
func TestRustRichText_SpansConsolidateDeletedThenEmptyMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "hello world"))
			require.NoError(t, text.Mark(0, 6, "bold", markTrue(), automerge.MarkExpandNone))
			require.NoError(t, text.Unmark(0, 6, "bold", automerge.MarkExpandNone))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hello world", spans["native"][0].Text)
	assert.Empty(t, spans["native"][0].Marks)
}

// TestRustRichText_SpansConsolidateEmptyThenDeletedMarks reproduces
// spans_consolidates_marks_with_empty_marks_followed_by_deleted_marks.
func TestRustRichText_SpansConsolidateEmptyThenDeletedMarks(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "hello world"))
			require.NoError(t, text.Mark(6, 11, "bold", markTrue(), automerge.MarkExpandNone))
			require.NoError(t, text.Unmark(6, 11, "bold", automerge.MarkExpandNone))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "hello world", spans["native"][0].Text)
}

// TestRustRichText_SpliceWithMark reproduces test_splice_with_mark.
func TestRustRichText_SpliceWithMark(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "abc"))
			require.NoError(
				t,
				text.Mark(

					1,
					2,
					"some_nonexpanding_mark_type",
					markString("marked"),
					automerge.MarkExpandNone,
				),
			)
			require.NoError(
				t,
				text.Mark(

					1,
					2,
					"some_expanding_mark_type",
					markString("marked"),
					automerge.MarkExpandBoth,
				),
			)
			require.NoError(t, text.Splice(1, 1, "d"))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
}

func textMarks(
	t *testing.T,
	build func(t *testing.T, text *automerge.Text),
) map[string][]automerge.Mark {
	t.Helper()

	marks := make(map[string][]automerge.Mark)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)

		build(t, text)

		_, err = document.CommitNow("marks")
		require.NoError(t, err)

		result, err := text.Marks()
		require.NoError(t, err)

		marks[engine.name] = result
	}

	return marks
}

// TestRustRichText_RemovedMarksNotInGetMarks reproduces
// removed_marks_should_not_appear_in_get_marks.
func TestRustRichText_RemovedMarksNotInGetMarks(t *testing.T) {
	t.Parallel()

	marks := textMarks(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "abcdefg"))
			require.NoError(
				t,
				text.Mark(

					0,
					1,
					"name1",
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					automerge.MarkExpandNone,
				),
			)
			require.NoError(t, text.Unmark(0, 1, "name1", automerge.MarkExpandNone))
		},
	)

	assert.Equal(t, marks["reference"], marks["native"])
	assert.Empty(t, marks["native"])
}

// TestRustRichText_InsertingTextNearDeletedMarks reproduces
// inserting_text_near_deleted_marks.
func TestRustRichText_InsertingTextNearDeletedMarks(t *testing.T) {
	t.Parallel()

	marks := textMarks(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "hello world"))
			require.NoError(t, text.Mark(2, 8, "bold", markTrue(), automerge.MarkExpandAfter))
			require.NoError(t, text.Mark(3, 6, "link", markTrue(), automerge.MarkExpandNone))
			require.NoError(t, text.Splice(1, 10, ""))
			require.NoError(t, text.Splice(0, 0, "a"))
			require.NoError(t, text.Splice(2, 0, "a"))
		},
	)

	assert.Equal(t, marks["reference"], marks["native"])
}

// TestRustRichText_GetMarksAtHeads reproduces get_marks_at_heads: marks active
// at a specific index resolved at a historical frontier.
func TestRustRichText_GetMarksAtHeads(t *testing.T) {
	t.Parallel()

	active := make(map[string]map[string]int64)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "hello world"))
		require.NoError(t, text.Mark(0, 10, "bold", markTrue(), automerge.MarkExpandAfter))

		_, err = document.Commit("bold", commitTime)
		require.NoError(t, err)

		heads, err := document.Heads()
		require.NoError(t, err)

		require.NoError(t, text.Unmark(0, 10, "bold", automerge.MarkExpandNone))

		_, err = document.Commit("unbold", commitTime.Add(time.Second))
		require.NoError(t, err)

		marks, err := text.MarksAt(heads)
		require.NoError(t, err)

		atIndex := make(map[string]int64)

		for _, mark := range marks {
			if uint32(1) >= mark.Start && uint32(1) < mark.End {
				value := int64(0)
				if mark.Value.Bool {
					value = 1
				}

				atIndex[mark.Name] = value
			}
		}

		active[engine.name] = atIndex
	}

	assert.Equal(t, active["reference"], active["native"])
	assert.Equal(t, map[string]int64{"bold": 1}, active["native"])
}

// TestRustText_ExpandMarksAreReportedInPatches reproduces
// expand_marks_are_reported_in_patches: a both-expanding mark includes text
// inserted at either boundary and both incremental splice patches carry it.
func TestRustText_ExpandMarksAreReportedInPatches(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)
	marks := make(map[string][]automerge.Mark)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "aaabbbccc"))
		require.NoError(
			t,
			text.Mark(

				3,
				6,
				"strong",
				markTrue(),
				automerge.MarkExpandBoth,
			),
		)

		_, err = document.Commit("seed", commitTime)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())

		var patches []automerge.Patch

		require.NoError(t, text.Splice(6, 0, "<"))

		_, err = document.Commit("end", commitTime.Add(time.Second))
		require.NoError(t, err)
		endPatches, err := document.DiffIncremental()
		require.NoError(t, err)

		patches = append(patches, endPatches...)

		require.NoError(t, text.Splice(3, 0, ">"))

		_, err = document.Commit("start", commitTime.Add(2*time.Second))
		require.NoError(t, err)
		startPatches, err := document.DiffIncremental()
		require.NoError(t, err)

		patches = append(patches, startPatches...)

		result[engine.name] = patches
		marks[engine.name], err = text.Marks()
		require.NoError(t, err)
	}

	reference := result["reference"]
	require.Len(t, reference, 2)

	for _, patch := range reference {
		assert.Equal(t, automerge.PatchSpliceText, patch.Action)
		require.Len(t, patch.Marks, 1)
		assert.Equal(t, "strong", patch.Marks[0].Name)
	}

	assert.Equal(t, uint64(6), reference[0].Index)
	assert.Equal(t, "<", reference[0].Text)
	assert.Equal(t, uint64(3), reference[1].Index)
	assert.Equal(t, ">", reference[1].Text)
	assert.Equal(t, result["reference"], result["native"])
	assert.Equal(t, marks["reference"], marks["native"])
	assert.Equal(t, uint32(3), marks["native"][0].Start)
	assert.Equal(t, uint32(8), marks["native"][0].End)
}

// TestRustText_RemotePatchesForExpandAfter reproduces
// test_remote_patches_for_marks_with_expand_after: applying a remote insertion
// at an after-expanding boundary produces the same marked splice patch as the
// local document.
func TestRustText_RemotePatchesForExpandAfter(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		documentA, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, documentA)

		textA, err := documentA.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, textA.Splice(0, 0, "fox"))
		require.NoError(
			t,
			textA.Mark(

				0,
				3,
				"strong",
				markTrue(),
				automerge.MarkExpandAfter,
			),
		)

		_, err = documentA.Commit("seed", commitTime)
		require.NoError(t, err)

		documentB, err := documentA.Fork(actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, documentB)

		beforeA, err := documentA.Heads()
		require.NoError(t, err)
		require.NoError(t, textA.Splice(3, 0, "a"))

		afterA, err := documentA.Commit("append", commitTime.Add(time.Second))
		require.NoError(t, err)

		require.NoError(t, documentB.UpdateDiffCursor())
		beforeB, err := documentB.Heads()
		require.NoError(t, err)
		_, err = documentB.Merge(documentA)
		require.NoError(t, err)
		afterB, err := documentB.Heads()
		require.NoError(t, err)

		local, err := documentA.Diff(beforeA, []automerge.Hash{afterA})
		require.NoError(t, err)
		remote, err := documentB.Diff(beforeB, afterB)
		require.NoError(t, err)

		assert.Equal(t, local, remote)
		result[engine.name] = local
	}

	reference := result["reference"]
	require.Len(t, reference, 1)
	assert.Equal(t, automerge.PatchSpliceText, reference[0].Action)
	assert.Equal(t, uint64(3), reference[0].Index)
	assert.Equal(t, "a", reference[0].Text)
	require.Len(t, reference[0].Marks, 1)
	assert.Equal(t, "strong", reference[0].Marks[0].Name)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustMarks_ExpansionAndUnmark reproduces tests/test.rs marks: a
// both-expanding mark grows at its end, an unmark removes only the original
// prefix, and text inserted before the unmarked range remains unmarked.
func TestRustMarks_ExpansionAndUnmark(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Mark)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "hello world"))
		require.NoError(
			t,
			text.Mark(

				0,
				5,
				"bold",
				markTrue(),
				automerge.MarkExpandBoth,
			),
		)
		require.NoError(t, text.Splice(5, 0, " cool"))
		require.NoError(
			t,
			text.Unmark(

				0,
				5,
				"bold",
				automerge.MarkExpandBefore,
			),
		)
		require.NoError(t, text.Splice(0, 0, "why "))

		_, err = document.Commit("marks", commitTime)
		require.NoError(t, err)

		result[engine.name], err = text.Marks()
		require.NoError(t, err)
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 1)
	assert.Equal(t, uint32(9), result["native"][0].Start)
	assert.Equal(t, uint32(14), result["native"][0].End)
	assert.Equal(t, "bold", result["native"][0].Name)
	assert.Equal(t, markTrue(), result["native"][0].Value)
}

// TestRustText_CrossPageMarksNotDoubleCounted reproduces
// marks_which_cross_optree_boundaries_are_not_double_counted_in_splice_patches.
// A mark crossing the reference engine's operation-tree page boundary must not
// leak onto text appended much later after unrelated block insertions.
func TestRustText_CrossPageMarksNotDoubleCounted(t *testing.T) {
	t.Parallel()

	const pageSize = 16

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				document, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, document)

				text, err := document.CreateText("text")
				require.NoError(t, err)
				textObject, err := document.Root().Object("text")
				require.NoError(t, err)

				require.NoError(t, text.Splice(0, 0, strings.Repeat("a", pageSize*2)))
				require.NoError(
					t,
					text.Mark(

						pageSize-1,
						pageSize+1,
						"strong",
						markTrue(),
						automerge.MarkExpandNone,
					),
				)

				_, err = document.Commit("seed", commitTime)
				require.NoError(t, err)

				for iteration := range 100 {
					length, err := textObject.Len()
					require.NoError(t, err)
					_, err = text.SplitBlock(uint32(length))
					require.NoError(t, err)
					_, err = document.Commit("block", commitTime.Add(time.Duration(iteration+1)*time.Second))
					require.NoError(t, err)
					require.NoError(t, document.UpdateDiffCursor())

					length, err = textObject.Len()
					require.NoError(t, err)
					require.NoError(t, text.Splice(uint32(length), 0, "a"))

					_, err = document.Commit("append", commitTime.Add(time.Duration(iteration+101)*time.Second))
					require.NoError(t, err)

					patches, err := document.DiffIncremental()
					require.NoError(t, err)
					require.Len(t, patches, 1)
					assert.Equal(t, automerge.PatchSpliceText, patches[0].Action)
					assert.Empty(t, patches[0].Marks)
				}
			},
		)
	}
}

// TestRustRichText_EmptyMarksBeforeBlockMarker reproduces
// empty_marks_before_block_marker_dont_repeat_text.
func TestRustRichText_EmptyMarksBeforeBlockMarker(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			_, err := text.SplitBlock(0)
			require.NoError(t, err)
			_, err = text.SplitBlock(0)
			require.NoError(t, err)
			require.NoError(t, text.Mark(1, 1, "strong", markTrue(), automerge.MarkExpandBoth))
			require.NoError(t, text.Splice(2, 0, "a"))
		},
	)

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

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			block, err := text.SplitBlock(0)
			require.NoError(t, err)
			require.NoError(
				t,
				block.PutValue(

					"type",
					automerge.Value{
						Type: automerge.ValueTypeText,
						Text: "ordered-list-item",
					},
				),
			)
			require.NoError(
				t,
				block.PutValue(

					"parents",
					automerge.Value{
						Type: automerge.ValueTypeList,
						List: []automerge.Value{{Type: automerge.ValueTypeText, Text: "div"}},
					},
				),
			)
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_MarkCreatedAfterInsertion reproduces
// mark_created_after_insertion.
func TestRustRichText_MarkCreatedAfterInsertion(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "12345"))
			require.NoError(t, text.Mark(1, 2, "strong", markTrue(), automerge.MarkExpandBoth))
			require.NoError(t, text.Mark(3, 4, "strong", markTrue(), automerge.MarkExpandBoth))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
}

// TestRustRichText_SpansConsolidatedWithZeroLengthSpans reproduces
// spans_are_consolidated_in_the_presence_of_zero_length_spans.
func TestRustRichText_SpansConsolidatedWithZeroLengthSpans(t *testing.T) {
	t.Parallel()

	spans := richTextSpans(
		t,
		func(t *testing.T, text *automerge.Text) {
			require.NoError(t, text.Splice(0, 0, "1234"))
			require.NoError(t, text.Mark(1, 1, "strong", markTrue(), automerge.MarkExpandBoth))
			require.NoError(t, text.Mark(2, 2, "strong", markTrue(), automerge.MarkExpandBoth))
		},
	)

	assert.Equal(t, spans["reference"], spans["native"])
	require.Len(t, spans["native"], 1)
	assert.Equal(t, "1234", spans["native"][0].Text)
}

// TestRustRichText_DeletingInMiddleOfMultibyteChar reproduces
// deleting_in_middle_of_multibyte_char_moves_the_cursor_to_after_the_character.
func TestRustRichText_DeletingInMiddleOfMultibyteChar(t *testing.T) {
	t.Parallel()

	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)

		var observed []string

		require.NoError(t, text.Splice(0, 0, "🐻🐻🐻🐻🐻🐻"))
		value, err := text.String()
		require.NoError(t, err)

		observed = append(observed, value)

		require.NoError(t, text.Splice(2, 2, "A🐻A"))
		value, err = text.String()
		require.NoError(t, err)

		observed = append(observed, value)

		require.NoError(t, text.Splice(4, 1, "X"))
		value, err = text.String()
		require.NoError(t, err)

		observed = append(observed, value)

		require.NoError(t, text.Splice(4, 2, "Y"))
		value, err = text.String()
		require.NoError(t, err)

		observed = append(observed, value)

		_, err = document.CommitNow("multibyte")
		require.NoError(t, err)

		results[engine.name] = observed
	}

	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(
		t,
		[]string{
			"🐻🐻🐻🐻🐻🐻",
			"🐻A🐻A🐻🐻🐻🐻",
			"🐻A🐻X🐻🐻🐻🐻",
			"🐻A🐻Y🐻🐻🐻",
		},
		results["native"],
	)
}
