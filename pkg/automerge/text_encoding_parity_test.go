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
// automerge 0.10 (rust/automerge/tests/text_encoding.rs). The reference backend
// is built with the utf16-indexing feature, so every text index is expressed in
// UTF-16 code units. Each scenario runs identically on the native Go engine and
// the Rust/WASM reference engine and asserts their results agree with the
// documented UTF-16 expectation. The 👩‍👩‍👧‍👦 family emoji used throughout is a
// single grapheme cluster spanning 7 code points and 11 UTF-16 code units.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

const familyEmoji = "👩‍👩‍👧‍👦"

// seedText creates a text object seeded with the given content and returns the
// committed document together with map- and text-typed handles to it.
func seedText(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
	content string,
) (*automerge.Document, *automerge.Object, *automerge.Text) {
	t.Helper()

	document, err := engine.open(ctx, actor(0xaa))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText(ctx, "text")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, content))
	_, err = document.Commit(ctx, "seed", commitTime)
	require.NoError(t, err)

	object, err := document.Root().Object(ctx, "text")
	require.NoError(t, err)

	return document, object, text
}

// TestRustTextEncoding_Length reproduces the utf16 case of length.
func TestRustTextEncoding_Length(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	lengths := make(map[string]uint64)

	for _, engine := range rustParityEngines() {
		_, object, _ := seedText(t, ctx, engine, "hello"+familyEmoji)

		length, err := object.Len(ctx)
		require.NoError(t, err)

		lengths[engine.name] = length
	}

	assert.Equal(t, uint64(16), lengths["reference"])
	assert.Equal(t, lengths["reference"], lengths["native"])
}

// TestRustTextEncoding_SpliceText reproduces the utf16 case of splice_text.
func TestRustTextEncoding_SpliceText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "hello "+familyEmoji+" world")
		require.NoError(t, text.Splice(ctx, 18, 0, "beautiful "))
		_, err := document.Commit(ctx, "splice", commitTime)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "hello "+familyEmoji+" beautiful world", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Get reproduces the utf16 case of get.
func TestRustTextEncoding_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	values := make(map[string]string)

	for _, engine := range rustParityEngines() {
		_, object, _ := seedText(t, ctx, engine, "he"+familyEmoji+"lo")

		scalar, err := object.ScalarAt(ctx, 13)
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

	ctx := context.Background()
	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, ctx, engine, "he"+familyEmoji+"llo")
		require.NoError(t, object.PutScalarAt(
			ctx,
			13,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
		))
		_, err := document.Commit(ctx, "put", commitTime)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "he"+familyEmoji+"Llo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Insert reproduces the utf16 case of insert.
func TestRustTextEncoding_Insert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, ctx, engine, "he"+familyEmoji+"llo")
		require.NoError(t, object.InsertScalar(
			ctx,
			13,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
		))
		_, err := document.Commit(ctx, "insert", commitTime)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "he"+familyEmoji+"Lllo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustTextEncoding_Delete reproduces the utf16 case of delete.
func TestRustTextEncoding_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	results := make(map[string]string)
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, ctx, engine, "he"+familyEmoji+"llo")
		require.NoError(t, object.DeleteIndex(ctx, 13))
		_, err := document.Commit(ctx, "delete", commitTime)
		require.NoError(t, err)

		result, err := text.String(ctx)
		require.NoError(t, err)

		results[engine.name] = result
		heads[engine.name] = sortedHeadHex(t, ctx, document)
	}

	assert.Equal(t, "he"+familyEmoji+"lo", results["reference"])
	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, heads["reference"], heads["native"])
}

// diffTextPatches seeds a text object, runs the mutation, and returns the diff
// between the states before and after the mutation for each engine.
func diffTextPatches(
	t *testing.T,
	ctx context.Context,
	content string,
	mutate func(ctx context.Context, object *automerge.Object, text *automerge.Text) error,
) map[string][]automerge.Patch {
	t.Helper()

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, object, text := seedText(t, ctx, engine, content)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, mutate(ctx, object, text))
		after, err := document.Commit(ctx, "mutate", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, before, []automerge.Hash{after})
		require.NoError(t, err)

		result[engine.name] = patches
	}

	return result
}

// TestRustTextEncoding_PatchInsert reproduces the utf16 case of patch_insert:
// an insert produces a SpliceText patch addressed by UTF-16 code units.
func TestRustTextEncoding_PatchInsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	patches := diffTextPatches(
		t,
		ctx,
		"he"+familyEmoji+"llo",
		func(ctx context.Context, object *automerge.Object, _ *automerge.Text) error {
			return object.InsertScalar(
				ctx,
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

	ctx := context.Background()
	patches := diffTextPatches(
		t,
		ctx,
		"he"+familyEmoji+"llo",
		func(ctx context.Context, _ *automerge.Object, text *automerge.Text) error {
			return text.Splice(ctx, 13, 0, "L")
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

	ctx := context.Background()
	patches := diffTextPatches(
		t,
		ctx,
		"he"+familyEmoji+"llo",
		func(ctx context.Context, object *automerge.Object, _ *automerge.Text) error {
			return object.DeleteIndex(ctx, 13)
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

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "12345")
		require.NoError(t, text.Mark(
			ctx, 1, 2, "strong",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandBoth,
		))
		_, err := document.Commit(ctx, "mark", commitTime)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor(ctx))

		var patches []automerge.Patch

		require.NoError(t, text.Splice(ctx, 1, 0, "-"))
		_, err = document.Commit(ctx, "s1", commitTime)
		require.NoError(t, err)
		first, err := document.DiffIncremental(ctx)
		require.NoError(t, err)

		patches = append(patches, first...)

		require.NoError(t, text.Splice(ctx, 2, 0, "-"))
		_, err = document.Commit(ctx, "s2", commitTime)
		require.NoError(t, err)
		second, err := document.DiffIncremental(ctx)
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

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "Hello world")
		require.NoError(t, text.Mark(
			ctx, 10, 11, "strong",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandNone,
		))
		_, err := document.Commit(ctx, "mark", commitTime)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor(ctx))

		require.NoError(t, text.Splice(ctx, 11, 0, "a"))
		_, err = document.Commit(ctx, "append", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
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

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 0, 0, "the quick fox jumps over the lazy dog"))
		require.NoError(t, text.Mark(
			ctx, 0, 37, "bold",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandBoth,
		))
		require.NoError(t, text.Mark(
			ctx, 4, 19, "italic",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandBoth,
		))
		require.NoError(t, text.Mark(
			ctx, 10, 13, "comment:somerandomcommentid",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "foxes are my favorite animal!"},
			automerge.MarkExpandBoth,
		))
		_, err = document.Commit(ctx, "seed", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
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

// TestRustTextEncoding_PatchPutSeq reproduces the utf16 case of patch_put_seq:
// an in-place text put reported through the incremental diff cursor produces a
// PutSeq patch addressed by UTF-16 code units.
func TestRustTextEncoding_PatchPutSeq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, object, _ := seedText(t, ctx, engine, "he"+familyEmoji+"llo")

		require.NoError(t, document.UpdateDiffCursor(ctx))
		require.NoError(t, object.PutScalarAt(
			ctx,
			13,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
		))
		_, err := document.Commit(ctx, "put", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	require.Len(t, result["reference"], 1)
	assert.Equal(t, automerge.PatchPutSeq, result["reference"][0].Action)
	assert.Equal(t, uint64(13), result["reference"][0].Index)
	require.NotNil(t, result["reference"][0].Value.Scalar)
	assert.Equal(t, "L", result["reference"][0].Value.Scalar.String)
	assert.Equal(t, result["reference"], result["native"])
}

// TestDocument_IncrementalDiffMatchesReference exercises the incremental diff
// cursor across map, list, and text mutations and asserts the native and
// reference patch streams agree for each committed change.
func TestDocument_IncrementalDiffMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	scenarios := []struct {
		name   string
		mutate func(ctx context.Context, object *automerge.Object, text *automerge.Text) error
	}{
		{"text_put", func(ctx context.Context, object *automerge.Object, _ *automerge.Text) error {
			return object.PutScalarAt(ctx, 13, automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"})
		}},
		{"text_insert", func(ctx context.Context, object *automerge.Object, _ *automerge.Text) error {
			return object.InsertScalar(ctx, 13, automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"})
		}},
		{"text_splice", func(ctx context.Context, _ *automerge.Object, text *automerge.Text) error {
			return text.Splice(ctx, 13, 0, "AB")
		}},
		{"text_delete", func(ctx context.Context, object *automerge.Object, _ *automerge.Text) error {
			return object.DeleteIndex(ctx, 13)
		}},
		{"text_mark", func(ctx context.Context, _ *automerge.Object, text *automerge.Text) error {
			return text.Mark(
				ctx,
				1,
				13,
				"bold",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
				automerge.MarkExpandBoth,
			)
		}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			result := make(map[string][]automerge.Patch)

			for _, engine := range rustParityEngines() {
				document, object, text := seedText(t, ctx, engine, "he"+familyEmoji+"llo")

				require.NoError(t, document.UpdateDiffCursor(ctx))
				require.NoError(t, scenario.mutate(ctx, object, text))
				_, err := document.Commit(ctx, scenario.name, commitTime)
				require.NoError(t, err)

				patches, err := document.DiffIncremental(ctx)
				require.NoError(t, err)

				result[engine.name] = patches
			}

			assert.NotEmpty(t, result["reference"])
			assert.Equal(t, result["reference"], result["native"])
		})
	}
}

// TestRustTextEncoding_PatchMark reproduces the utf16 case of patch_mark: a
// mark produces a Mark patch whose start and end are UTF-16 code units.
func TestRustTextEncoding_PatchMark(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	patches := diffTextPatches(
		t,
		ctx,
		"he"+familyEmoji+"llo",
		func(ctx context.Context, _ *automerge.Object, text *automerge.Text) error {
			return text.Mark(
				ctx,
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

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "hello world")
		require.NoError(t, text.Mark(
			ctx,
			0,
			5,
			"bold",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			automerge.MarkExpandBoth,
		))
		_, err := document.Commit(ctx, "mark", commitTime)
		require.NoError(t, err)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, text.Unmark(ctx, 0, 5, "bold", automerge.MarkExpandBoth))
		after, err := document.Commit(ctx, "unmark", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, before, []automerge.Hash{after})
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

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "hello world")
		require.NoError(t, text.Mark(
			ctx,
			0,
			5,
			"color",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "red"},
			automerge.MarkExpandBoth,
		))
		_, err := document.Commit(ctx, "red", commitTime)
		require.NoError(t, err)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, text.Mark(
			ctx,
			0,
			5,
			"color",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "blue"},
			automerge.MarkExpandBoth,
		))
		after, err := document.Commit(ctx, "blue", commitTime)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, before, []automerge.Hash{after})
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

	ctx := context.Background()
	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, ctx, engine, "he"+familyEmoji+"llo")
		_, err := text.SplitBlock(ctx, 13)
		require.NoError(t, err)
		_, err = document.Commit(ctx, "split", commitTime)
		require.NoError(t, err)

		spans, err := text.Spans(ctx)
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
