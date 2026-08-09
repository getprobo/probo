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

// The tests in this file reproduce upstream Rust current-state tests from
// automerge 0.10 (rust/automerge/src/automerge/current_state.rs). Each builds
// the same document on the native and Rust/WASM reference engines and asserts
// their materialization patch streams agree.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func basicStateDocument(
	t *testing.T,
	ctx context.Context,
	factory func(context.Context, automerge.ActorID) (*automerge.Document, error),
) *automerge.Document {
	t.Helper()

	document, err := factory(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	require.NoError(t, document.Root().PutScalar(
		ctx,
		"key",
		automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
	))

	mapObject, err := document.Root().CreateObject(ctx, "map", automerge.ObjectTypeMap)
	require.NoError(t, err)
	require.NoError(t, mapObject.PutScalar(
		ctx,
		"nested_key",
		automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
	))

	list, err := document.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
	require.NoError(t, err)
	require.NoError(t, list.InsertScalar(
		ctx,
		0,
		automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
	))

	text, err := document.CreateText(ctx, "text")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "a"))

	_, err = document.Commit(ctx, "basic", commitTime)
	require.NoError(t, err)

	return document
}

// TestRustCurrentState_Basic reproduces the current_state basic_test.
func TestRustCurrentState_Basic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	nativePatches, err := basicStateDocument(t, ctx, automerge.New).CurrentState(ctx)
	require.NoError(t, err)
	referencePatches, err := basicStateDocument(t, ctx, automerge.NewReference).CurrentState(ctx)
	require.NoError(t, err)

	assert.Equal(t, referencePatches, nativePatches)

	require.Len(t, nativePatches, 7)

	assert.Equal(t, automerge.PatchPutMap, nativePatches[0].Action)
	assert.Equal(t, "key", nativePatches[0].Key)
	require.NotNil(t, nativePatches[0].Value.Scalar)
	assert.Equal(t, "value", nativePatches[0].Value.Scalar.String)

	assert.Equal(t, "list", nativePatches[1].Key)
	assert.Equal(t, automerge.ObjectTypeList, nativePatches[1].Value.Object)
	assert.Equal(t, "map", nativePatches[2].Key)
	assert.Equal(t, automerge.ObjectTypeMap, nativePatches[2].Value.Object)
	assert.Equal(t, "text", nativePatches[3].Key)
	assert.Equal(t, automerge.ObjectTypeText, nativePatches[3].Value.Object)

	assert.Equal(t, automerge.PatchPutMap, nativePatches[4].Action)
	assert.Equal(t, "nested_key", nativePatches[4].Key)

	assert.Equal(t, automerge.PatchInsert, nativePatches[5].Action)
	require.Len(t, nativePatches[5].Values, 1)
	assert.Equal(t, "value", nativePatches[5].Values[0].Value.Scalar.String)

	assert.Equal(t, automerge.PatchSpliceText, nativePatches[6].Action)
	assert.Equal(t, "a", nativePatches[6].Text)
}

// TestRustCurrentState_DeletedOpsOmitted reproduces
// current_state test_deleted_ops_omitted.
func TestRustCurrentState_DeletedOpsOmitted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	build := func(factory func(context.Context, automerge.ActorID) (*automerge.Document, error)) []automerge.Patch {
		document, err := factory(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.Root().PutScalar(
			ctx,
			"key",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
		))
		require.NoError(t, document.Root().DeleteKey(ctx, "key"))

		mapObject, err := document.Root().CreateObject(ctx, "map", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(t, mapObject.PutScalar(
			ctx,
			"nested_key",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
		))
		require.NoError(t, mapObject.DeleteKey(ctx, "nested_key"))

		list, err := document.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(
			ctx,
			0,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
		))
		require.NoError(t, list.DeleteIndex(ctx, 0))

		_, err = document.Commit(ctx, "deleted", commitTime)
		require.NoError(t, err)

		patches, err := document.CurrentState(ctx)
		require.NoError(t, err)

		return patches
	}

	nativePatches := build(automerge.New)
	referencePatches := build(automerge.NewReference)

	assert.Equal(t, referencePatches, nativePatches)

	// The deleted scalar, nested key, and list element must not appear; only the
	// three surviving empty objects remain.
	for _, patch := range nativePatches {
		assert.NotEqual(t, "key", patch.Key)
		assert.NotEqual(t, "nested_key", patch.Key)
		assert.NotEqual(t, automerge.PatchInsert, patch.Action)
	}
}
