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

// The tests in this file reproduce upstream Rust diff tests from automerge 0.10
// (rust/automerge/tests/test.rs). Each builds the same document on the native
// and Rust/WASM reference engines and asserts their diff patch streams agree.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// TestRustDiff_ReverseDeletionOfObjectInList reproduces
// diff_should_reverse_deletion_of_object_in_list_correctly.
func TestRustDiff_ReverseDeletionOfObjectInList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		list, err := document.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(ctx, 0, automerge.Scalar{Type: automerge.ScalarTypeString, String: "a"}))
		text, err := list.InsertObject(ctx, 1, automerge.ObjectTypeText)
		require.NoError(t, err)
		textValue, err := text.Text()
		require.NoError(t, err)
		require.NoError(t, textValue.Splice(ctx, 0, 0, "b"))
		require.NoError(t, list.InsertScalar(ctx, 2, automerge.Scalar{Type: automerge.ScalarTypeString, String: "c"}))
		_, err = document.Commit(ctx, "build", commitTime)
		require.NoError(t, err)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, list.DeleteIndex(ctx, 1))
		_, err = document.Commit(ctx, "delete", commitTime.Add(time.Second))
		require.NoError(t, err)
		after, err := document.Heads(ctx)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, after, before)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 2)
	assert.Equal(t, automerge.PatchInsert, result["native"][0].Action)
	assert.Equal(t, uint64(1), result["native"][0].Index)
	require.Len(t, result["native"][0].Values, 1)
	assert.Equal(t, automerge.ObjectTypeText, result["native"][0].Values[0].Value.Object)
	assert.Equal(t, automerge.PatchSpliceText, result["native"][1].Action)
	assert.Equal(t, "b", result["native"][1].Text)
}

// TestRustDiff_ReverseDeletionOfObjectInMap reproduces
// diff_should_reverse_deletion_of_object_in_map_correctly.
func TestRustDiff_ReverseDeletionOfObjectInMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		mapObject, err := document.Root().CreateObject(ctx, "map", automerge.ObjectTypeMap)
		require.NoError(t, err)
		_, err = mapObject.CreateObject(ctx, "text", automerge.ObjectTypeText)
		require.NoError(t, err)
		require.NoError(t, mapObject.PutScalar(ctx, "a", automerge.Scalar{Type: automerge.ScalarTypeString, String: "a"}))
		textB, err := mapObject.CreateObject(ctx, "b", automerge.ObjectTypeText)
		require.NoError(t, err)
		textBValue, err := textB.Text()
		require.NoError(t, err)
		require.NoError(t, textBValue.Splice(ctx, 0, 0, "b"))
		require.NoError(t, mapObject.PutScalar(ctx, "c", automerge.Scalar{Type: automerge.ScalarTypeString, String: "c"}))
		_, err = document.Commit(ctx, "build", commitTime)
		require.NoError(t, err)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, mapObject.DeleteKey(ctx, "b"))
		_, err = document.Commit(ctx, "delete", commitTime.Add(time.Second))
		require.NoError(t, err)
		after, err := document.Heads(ctx)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, after, before)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 2)
	assert.Equal(t, automerge.PatchPutMap, result["native"][0].Action)
	assert.Equal(t, "b", result["native"][0].Key)
	assert.Equal(t, automerge.ObjectTypeText, result["native"][0].Value.Object)
	assert.Equal(t, automerge.PatchSpliceText, result["native"][1].Action)
	assert.Equal(t, "b", result["native"][1].Text)
}

// TestRustDiff_ReverseDeletionOfBlockInText reproduces
// diff_should_reverse_deletion_of_block_in_text_correctly.
func TestRustDiff_ReverseDeletionOfBlockInText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.CreateText(ctx, "text")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 0, 0, "a"))
		block, err := text.SplitBlock(ctx, 1)
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 2, 0, "b"))
		require.NoError(t, block.PutScalar(ctx, "key", automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"}))
		_, err = document.Commit(ctx, "build", commitTime)
		require.NoError(t, err)

		before, err := document.Heads(ctx)
		require.NoError(t, err)
		require.NoError(t, text.JoinBlock(ctx, 1))
		_, err = document.Commit(ctx, "delete", commitTime.Add(time.Second))
		require.NoError(t, err)
		after, err := document.Heads(ctx)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, after, before)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	assert.Equal(t, result["reference"], result["native"])
	require.Len(t, result["native"], 2)
	assert.Equal(t, automerge.PatchInsert, result["native"][0].Action)
	assert.Equal(t, uint64(1), result["native"][0].Index)
	require.Len(t, result["native"][0].Values, 1)
	assert.Equal(t, automerge.ObjectTypeMap, result["native"][0].Values[0].Value.Object)
	assert.Equal(t, automerge.PatchPutMap, result["native"][1].Action)
	assert.Equal(t, "key", result["native"][1].Key)
	require.NotNil(t, result["native"][1].Value.Scalar)
	assert.Equal(t, "value", result["native"][1].Value.Scalar.String)
}
