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

// The tests in this file reproduce the patch-generating batch-insert scenarios
// from upstream Rust automerge 0.10 (rust/automerge/tests/batch_insert.rs),
// asserting the native Go and Rust/WASM reference engines produce identical
// patch streams for a hydrated batch insertion.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func intValue(value int64) automerge.Value {
	return automerge.Value{
		Type:   automerge.ValueTypeScalar,
		Scalar: automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
	}
}

func stringValue(value string) automerge.Value {
	return automerge.Value{
		Type:   automerge.ValueTypeScalar,
		Scalar: automerge.Scalar{Type: automerge.ScalarTypeString, String: value},
	}
}

// TestRustBatchInsert_GeneratesPatches reproduces batch_insert_generates_patches.
func TestRustBatchInsert_GeneratesPatches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	value := automerge.Value{
		Type: automerge.ValueTypeMap,
		Map: map[string]automerge.Value{
			"name":  stringValue("test"),
			"items": {Type: automerge.ValueTypeList, List: []automerge.Value{intValue(1), intValue(2)}},
		},
	}

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.UpdateDiffCursor(ctx))
		require.NoError(t, document.Root().PutValue(ctx, "data", value))
		_, err = document.Commit(ctx, "batch", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	assert.NotEmpty(t, reference)

	hasData := false

	for _, patch := range reference {
		if patch.Action == automerge.PatchPutMap && patch.Key == "data" {
			hasData = true
		}
	}

	assert.True(t, hasData, "expected a put_map patch for data")
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustBatchInsert_TextGeneratesSplicePatch reproduces
// batch_insert_text_generates_splice_patch.
func TestRustBatchInsert_TextGeneratesSplicePatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	value := automerge.Value{
		Type: automerge.ValueTypeMap,
		Map: map[string]automerge.Value{
			"greeting": {Type: automerge.ValueTypeText, Text: "hi"},
		},
	}

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.UpdateDiffCursor(ctx))
		require.NoError(t, document.Root().PutValue(ctx, "data", value))
		_, err = document.Commit(ctx, "batch", commitTime)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	hasSplice := false

	for _, patch := range result["reference"] {
		if patch.Action == automerge.PatchSpliceText {
			hasSplice = true
		}
	}

	assert.True(t, hasSplice, "expected a splice_text patch")
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustBatchInit_MapGeneratesPatches reproduces batch_init_map_generates_patches.
func TestRustBatchInit_MapGeneratesPatches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	root := map[string]automerge.Value{
		"name":  stringValue("test"),
		"items": {Type: automerge.ValueTypeList, List: []automerge.Value{intValue(1), intValue(2)}},
	}

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.Root().PutMap(ctx, root))
		_, err = document.Commit(ctx, "init", commitTime)
		require.NoError(t, err)

		heads, err := document.Heads(ctx)
		require.NoError(t, err)

		patches, err := document.Diff(ctx, nil, heads)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	hasName := false

	for _, patch := range result["reference"] {
		if patch.Action == automerge.PatchPutMap && patch.Key == "name" {
			hasName = true
		}
	}

	assert.True(t, hasName, "expected a put_map patch for name")
	assert.Equal(t, result["reference"], result["native"])
}
