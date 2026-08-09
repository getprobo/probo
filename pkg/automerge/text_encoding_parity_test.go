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
