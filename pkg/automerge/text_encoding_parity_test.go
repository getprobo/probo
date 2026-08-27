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

// The tests in this file reproduce upstream Rust text-encoding tests from
// automerge 0.11 (rust/automerge/tests/text_encoding.rs). The reference backend
// is built with the utf16-indexing feature, so every text index is expressed in
// UTF-16 code units. Each scenario runs identically on the native Go engine and
// the Rust/WASM reference engine and asserts their results agree with the
// documented UTF-16 expectation. The 👩‍👩‍👧‍👦 family emoji used throughout is a
// single grapheme cluster spanning 7 code points and 11 UTF-16 code units.

package automerge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

const familyEmoji = "👩‍👩‍👧‍👦"

// seedText creates a text object seeded with the given content and returns the
// committed document together with map- and text-typed handles to it.
func seedText(
	t *testing.T,
	engine rustParityEngine,
	content string,
) (*automerge.Document, *automerge.Object, *automerge.Text) {
	t.Helper()

	document, err := engine.open(actor(0xaa))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText("text")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, content))

	_, err = document.Commit("seed", commitTime)
	require.NoError(t, err)

	object, err := document.Root().Object("text")
	require.NoError(t, err)

	return document, object, text
}

// TestRustTextEncoding_Length reproduces the utf16 case of length.
func TestRustTextEncoding_Length(t *testing.T) {
	t.Parallel()

	lengths := make(map[string]uint64)

	for _, engine := range rustParityEngines() {
		_, object, _ := seedText(t, engine, "hello"+familyEmoji)

		length, err := object.Len()
		require.NoError(t, err)

		lengths[engine.name] = length
	}

	assert.Equal(t, uint64(16), lengths["reference"])
	assert.Equal(t, lengths["reference"], lengths["native"])
}

// TestRustTextEncoding_SpliceText reproduces the utf16 case of splice_text.
func TestRustTextEncoding_SpliceText(t *testing.T) {
	t.Parallel()

	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "hello "+familyEmoji+" world")
		require.NoError(t, text.Splice(18, 0, "beautiful "))

		_, err := document.Commit("splice", commitTime)
		require.NoError(t, err)

		result, err := text.String()
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, document)
	}

	assert.Equal(t, "hello "+familyEmoji+" beautiful world", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Get reproduces the utf16 case of get.
func TestRustTextEncoding_Get(t *testing.T) {
	t.Parallel()

	values := make(map[string]string)

	for _, engine := range rustParityEngines() {
		_, object, _ := seedText(t, engine, "he"+familyEmoji+"lo")

		scalar, err := object.ScalarAt(13)
		require.NoError(t, err)
		require.Equal(t, automerge.ScalarTypeString, scalar.Type)
		values[engine.name] = scalar.String
	}

	assert.Equal(t, "l", values["reference"])
	assert.Equal(t, values["reference"], values["native"])
}

// TestRustTextEncoding_Put reproduces the utf16 case of put.
func TestRustTextEncoding_Put(t *testing.T) {
	t.Parallel()

	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, engine, "he"+familyEmoji+"llo")
		require.NoError(
			t,
			object.PutScalarAt(

				13,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
			),
		)

		_, err := document.Commit("put", commitTime)
		require.NoError(t, err)

		result, err := text.String()
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, document)
	}

	assert.Equal(t, "he"+familyEmoji+"Llo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Insert reproduces the utf16 case of insert.
func TestRustTextEncoding_Insert(t *testing.T) {
	t.Parallel()

	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, engine, "he"+familyEmoji+"llo")
		require.NoError(
			t,
			object.InsertScalar(

				13,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
			),
		)

		_, err := document.Commit("insert", commitTime)
		require.NoError(t, err)

		result, err := text.String()
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, document)
	}

	assert.Equal(t, "he"+familyEmoji+"Lllo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Delete reproduces the utf16 case of delete.
func TestRustTextEncoding_Delete(t *testing.T) {
	t.Parallel()

	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, engine, "he"+familyEmoji+"llo")
		require.NoError(t, object.DeleteIndex(13))

		_, err := document.Commit("delete", commitTime)
		require.NoError(t, err)

		result, err := text.String()
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, document)
	}

	assert.Equal(t, "he"+familyEmoji+"lo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// diffTextPatches seeds a text object, runs the mutation, and returns the diff
// between the states before and after the mutation for each engine.
func diffTextPatches(
	t *testing.T,
	content string,
	mutate func(object *automerge.Object, text *automerge.Text) error,
) map[string][]automerge.Patch {
	t.Helper()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, engine, content)

		before, err := document.Heads()
		require.NoError(t, err)
		require.NoError(t, mutate(object, text))

		after, err := document.Commit("mutate", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(before, []automerge.Hash{after})
		require.NoError(t, err)

		result[engine.name] = patches
	}

	return result
}

// TestRustTextEncoding_PatchInsert reproduces the utf16 case of patch_insert:
// an insert produces a SpliceText patch addressed by UTF-16 code units.
func TestRustTextEncoding_PatchInsert(t *testing.T) {
	t.Parallel()

	patches := diffTextPatches(
		t,
		"he"+familyEmoji+"llo",
		func(object *automerge.Object, _ *automerge.Text) error {
			return object.InsertScalar(

				13,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
			)
		},
	)

	require.Len(t, patches["reference"], 1)
	assert.Equal(t, automerge.PatchSpliceText, patches["reference"][0].Action)
	assert.Equal(t, uint64(13), patches["reference"][0].Index)
	assert.Equal(t, "L", patches["reference"][0].Text)
	assert.Equal(t, patches["reference"], patches["native"])
}

// TestRustTextEncoding_PatchSpliceText reproduces the utf16 case of
// patch_splice_text: a splice produces a SpliceText patch at a UTF-16 index.
func TestRustTextEncoding_PatchSpliceText(t *testing.T) {
	t.Parallel()

	patches := diffTextPatches(
		t,
		"he"+familyEmoji+"llo",
		func(_ *automerge.Object, text *automerge.Text) error {
			return text.Splice(13, 0, "L")
		},
	)

	require.Len(t, patches["reference"], 1)
	assert.Equal(t, automerge.PatchSpliceText, patches["reference"][0].Action)
	assert.Equal(t, uint64(13), patches["reference"][0].Index)
	assert.Equal(t, "L", patches["reference"][0].Text)
	assert.Equal(t, patches["reference"], patches["native"])
}

// TestRustTextEncoding_PatchDelete reproduces the utf16 case of patch_delete:
// a delete produces a DeleteSeq patch at a UTF-16 index with length one.
func TestRustTextEncoding_PatchDelete(t *testing.T) {
	t.Parallel()

	patches := diffTextPatches(
		t,
		"he"+familyEmoji+"llo",
		func(object *automerge.Object, _ *automerge.Text) error {
			return object.DeleteIndex(13)
		},
	)

	require.Len(t, patches["reference"], 1)
	assert.Equal(t, automerge.PatchDeleteSeq, patches["reference"][0].Action)
	assert.Equal(t, uint64(13), patches["reference"][0].Index)
	assert.Equal(t, uint64(1), patches["reference"][0].Length)
	assert.Equal(t, patches["reference"], patches["native"])
}

// TestRustText_IncrementalSplicePatchesIncludeMarks reproduces
// incremental_splice_patches_include_marks: text spliced inside an expanding
// mark is reported as a splice_text patch carrying that mark, with no separate
// mark patch for the range growth.
func TestRustText_IncrementalSplicePatchesIncludeMarks(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "12345")
		require.NoError(
			t,
			text.Mark(

				1,
				2,
				"strong",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			),
		)

		_, err := document.Commit("mark", commitTime)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())

		var patches []automerge.Patch

		require.NoError(t, text.Splice(1, 0, "-"))

		_, err = document.Commit("s1", commitTime)
		require.NoError(t, err)
		first, err := document.DiffIncremental()
		require.NoError(t, err)

		patches = append(patches, first...)

		require.NoError(t, text.Splice(2, 0, "-"))

		_, err = document.Commit("s2", commitTime)
		require.NoError(t, err)
		second, err := document.DiffIncremental()
		require.NoError(t, err)

		patches = append(patches, second...)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 2)

	for _, patch := range result["reference"] {
		assert.Equal(t, automerge.PatchSpliceText, patch.Action)
		require.Len(t, patch.Marks, 1)
		assert.Equal(t, "strong", patch.Marks[0].Name)
	}

	assert.Equal(t, uint64(1), result["reference"][0].Index)
	assert.Equal(t, uint64(2), result["reference"][1].Index)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustText_NoexpandMarksAtEndOfText reproduces
// noexpand_marks_at_the_end_of_text_should_not_emit_marked_patches_on_following_insertions:
// text appended after a non-expanding mark does not inherit it, so the splice
// patch carries no marks.
func TestRustText_NoexpandMarksAtEndOfText(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "Hello world")
		require.NoError(
			t,
			text.Mark(

				10,
				11,
				"strong",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandNone,
			),
		)

		_, err := document.Commit("mark", commitTime)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())

		require.NoError(t, text.Splice(11, 0, "a"))

		_, err = document.Commit("append", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 1)
	assert.Equal(t, automerge.PatchSpliceText, result["reference"][0].Action)
	assert.Empty(t, result["reference"][0].Marks)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustText_LocalPatchesCreatedForMarks reproduces local_patches_created_for_marks:
// materializing marked text through the diff cursor splits it into one
// splice_text patch per mark run, each carrying the marks active on that run.
func TestRustText_LocalPatchesCreatedForMarks(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText("text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "the quick fox jumps over the lazy dog"))
		require.NoError(
			t,
			text.Mark(

				0,
				37,
				"bold",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			),
		)
		require.NoError(
			t,
			text.Mark(

				4,
				19,
				"italic",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			),
		)
		require.NoError(
			t,
			text.Mark(

				10,
				13,
				"comment:somerandomcommentid",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "foxes are my favorite animal!"},
				automerge.MarkExpandBoth,
			),
		)

		_, err = document.Commit("seed", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	require.NotEmpty(t, reference)
	assert.Equal(t, automerge.PatchPutMap, reference[0].Action)

	runs := reference[1:]
	require.Len(t, runs, 5)

	expected := []struct {
		text  string
		names []string
	}{
		{"the ", []string{"bold"}},
		{"quick ", []string{"bold", "italic"}},
		{"fox", []string{"bold", "comment:somerandomcommentid", "italic"}},
		{" jumps", []string{"bold", "italic"}},
		{" over the lazy dog", []string{"bold"}},
	}

	for index, want := range expected {
		assert.Equal(t, automerge.PatchSpliceText, runs[index].Action)
		assert.Equal(t, want.text, runs[index].Text)

		names := make([]string, 0, len(runs[index].Marks))
		for _, mark := range runs[index].Marks {
			names = append(names, mark.Name)
		}

		assert.Equal(t, want.names, names)
	}

	assert.Equal(t, result["reference"], result["native"])
}

// TestRustTextEncoding_PatchPutText reproduces the utf16 case of patch_put_text:
// an in-place string put in text is reported as a UTF-16-addressed deletion
// followed by a text splice.
func TestRustTextEncoding_PatchPutText(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, object, _ := seedText(t, engine, "he"+familyEmoji+"llo")

		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(
			t,
			object.PutScalarAt(

				13,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
			),
		)

		_, err := document.Commit("put", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 2)
	assert.Equal(t, automerge.PatchDeleteSeq, result["reference"][0].Action)
	assert.Equal(t, uint64(13), result["reference"][0].Index)
	assert.Equal(t, uint64(1), result["reference"][0].Length)
	assert.Equal(t, automerge.PatchSpliceText, result["reference"][1].Action)
	assert.Equal(t, uint64(13), result["reference"][1].Index)
	assert.Equal(t, "L", result["reference"][1].Text)
	assert.Equal(t, result["reference"], result["native"])
}

// TestDocument_IncrementalDiffMatchesReference exercises the incremental diff
// cursor across map, list, and text mutations and asserts the native and
// reference patch streams agree for each committed change.
func TestDocument_IncrementalDiffMatchesReference(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name   string
		mutate func(object *automerge.Object, text *automerge.Text) error
	}{
		{"text_put", func(object *automerge.Object, _ *automerge.Text) error {
			return object.PutScalarAt(13, automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"})
		}},
		{"text_insert", func(object *automerge.Object, _ *automerge.Text) error {
			return object.InsertScalar(13, automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"})
		}},
		{"text_splice", func(_ *automerge.Object, text *automerge.Text) error {
			return text.Splice(13, 0, "AB")
		}},
		{"text_delete", func(object *automerge.Object, _ *automerge.Text) error {
			return object.DeleteIndex(13)
		}},
		{"text_mark", func(_ *automerge.Object, text *automerge.Text) error {
			return text.Mark(

				1,
				13,
				"bold",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			)
		}},
	}

	for _, scenario := range scenarios {
		t.Run(
			scenario.name,
			func(t *testing.T) {
				t.Parallel()

				result := make(map[string][]automerge.Patch)

				for _, engine := range rustParityEngines() {
					document, object, text := seedText(t, engine, "he"+familyEmoji+"llo")

					require.NoError(t, document.UpdateDiffCursor())
					require.NoError(t, scenario.mutate(object, text))
					_, err := document.Commit(scenario.name, commitTime)
					require.NoError(t, err)

					patches, err := document.DiffIncremental()
					require.NoError(t, err)

					result[engine.name] = patches
				}

				assert.NotEmpty(t, result["reference"])
				assert.Equal(t, result["reference"], result["native"])
			},
		)
	}
}

// TestRustTextEncoding_PatchMark reproduces the utf16 case of patch_mark: a
// mark produces a Mark patch whose start and end are UTF-16 code units.
func TestRustTextEncoding_PatchMark(t *testing.T) {
	t.Parallel()

	patches := diffTextPatches(
		t,
		"he"+familyEmoji+"llo",
		func(_ *automerge.Object, text *automerge.Text) error {
			return text.Mark(

				1,
				13,
				"bold",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			)
		},
	)

	require.Len(t, patches["reference"], 1)
	assert.Equal(t, automerge.PatchMark, patches["reference"][0].Action)
	require.Len(t, patches["reference"][0].Marks, 1)
	assert.Equal(t, uint32(1), patches["reference"][0].Marks[0].Start)
	assert.Equal(t, uint32(13), patches["reference"][0].Marks[0].End)
	assert.Equal(t, "bold", patches["reference"][0].Marks[0].Name)
	assert.Equal(t, patches["reference"], patches["native"])
}

// TestTextDiff_MarkRemovalMatchesReference verifies that removing a mark emits a
// mark patch carrying a null value on both engines.
func TestTextDiff_MarkRemovalMatchesReference(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "hello world")
		require.NoError(
			t,
			text.Mark(

				0,
				5,
				"bold",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			),
		)

		_, err := document.Commit("mark", commitTime)
		require.NoError(t, err)

		before, err := document.Heads()
		require.NoError(t, err)
		require.NoError(t, text.Unmark(0, 5, "bold", automerge.MarkExpandBoth))

		after, err := document.Commit("unmark", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(before, []automerge.Hash{after})
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 1)
	assert.Equal(t, automerge.PatchMark, result["reference"][0].Action)
	require.Len(t, result["reference"][0].Marks, 1)
	assert.Equal(t, automerge.ScalarTypeNull, result["reference"][0].Marks[0].Value.Type)
	assert.Equal(t, result["reference"], result["native"])
}

// TestTextDiff_MarkValueChangeMatchesReference verifies that changing a mark
// value emits a mark patch with the new value on both engines.
func TestTextDiff_MarkValueChangeMatchesReference(t *testing.T) {
	t.Parallel()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "hello world")
		require.NoError(
			t,
			text.Mark(

				0,
				5,
				"color",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "red"},
				automerge.MarkExpandBoth,
			),
		)

		_, err := document.Commit("red", commitTime)
		require.NoError(t, err)

		before, err := document.Heads()
		require.NoError(t, err)
		require.NoError(
			t,
			text.Mark(

				0,
				5,
				"color",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "blue"},
				automerge.MarkExpandBoth,
			),
		)

		after, err := document.Commit("blue", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(before, []automerge.Hash{after})
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 1)
	require.Len(t, result["reference"][0].Marks, 1)
	assert.Equal(t, "blue", result["reference"][0].Marks[0].Value.String)
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustTextEncoding_SplitBlock reproduces the utf16 case of split_block.
func TestRustTextEncoding_SplitBlock(t *testing.T) {
	t.Parallel()

	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "he"+familyEmoji+"llo")
		_, err := text.SplitBlock(13)
		require.NoError(t, err)
		_, err = document.Commit("split", commitTime)
		require.NoError(t, err)

		spans, err := text.Spans()
		require.NoError(t, err)

		texts := make([]string, 0, len(spans))

		for _, span := range spans {
			if span.Type == automerge.SpanTypeText {
				texts = append(texts, span.Text)
			}
		}

		results[engine.name] = texts
	}

	assert.Equal(t, []string{"he" + familyEmoji, "llo"}, results["reference"])
	assert.Equal(t, results["reference"], results["native"])
}
