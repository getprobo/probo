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

// The tests in this file reproduce upstream Rust batch-insertion tests from
// automerge 0.10 (rust/automerge/tests/batch_insert.rs) against both the native
// Go engine and the Rust/WASM reference engine, driving the public hydration
// API (PutValue, PutValueAt, SpliceValues) that mirrors batch_create_object.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func hydratedMap(pairs map[string]automerge.Value) automerge.Value {
	return automerge.Value{Type: automerge.ValueTypeMap, Map: pairs}
}

func hydratedList(values ...automerge.Value) automerge.Value {
	return automerge.Value{Type: automerge.ValueTypeList, List: values}
}

// TestRustBatch_MergesCorrectly reproduces batch_insert_merges_correctly.
func TestRustBatch_MergesCorrectly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		require.NoError(
			t,
			doc1.Root().PutValue(
				ctx,
				"obj1",
				hydratedMap(map[string]automerge.Value{"from": hydratedString("doc1")}),
			),
		)
		_, err = doc1.Commit(ctx, "obj1", commitTime)
		require.NoError(t, err)

		doc2, err := doc1.Fork(ctx, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		require.NoError(
			t,
			doc2.Root().PutValue(
				ctx,
				"obj2",
				hydratedMap(map[string]automerge.Value{"from": hydratedString("doc2")}),
			),
		)
		_, err = doc2.Commit(ctx, "obj2", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(ctx, doc2)
		require.NoError(t, err)

		obj1, err := doc1.Root().Object(ctx, "obj1")
		require.NoError(t, err)
		value, err := obj1.Scalar(ctx, "from")
		require.NoError(t, err)
		assert.Equal(t, "doc1", value.String)

		obj2, err := doc1.Root().Object(ctx, "obj2")
		require.NoError(t, err)
		value, err = obj2.Scalar(ctx, "from")
		require.NoError(t, err)
		assert.Equal(t, "doc2", value.String)

		heads[engine.name] = sortedHeadHex(t, ctx, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_MultipleInserts reproduces multiple_batch_inserts.
func TestRustBatch_MultipleInserts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		require.NoError(
			t,
			doc.Root().PutValue(
				ctx,
				"first",
				hydratedMap(map[string]automerge.Value{"a": hydratedInt(1)}),
			),
		)
		require.NoError(
			t,
			doc.Root().PutValue(
				ctx,
				"second",
				hydratedMap(map[string]automerge.Value{"b": hydratedInt(2)}),
			),
		)
		require.NoError(
			t,
			doc.Root().PutValue(
				ctx,
				"third",
				hydratedMap(map[string]automerge.Value{"c": hydratedInt(3)}),
			),
		)
		_, err = doc.Commit(ctx, "batches", commitTime)
		require.NoError(t, err)

		for key, field := range map[string]struct {
			name  string
			value int64
		}{
			"first":  {"a", 1},
			"second": {"b", 2},
			"third":  {"c", 3},
		} {
			object, err := doc.Root().Object(ctx, key)
			require.NoError(t, err)
			value, err := object.Scalar(ctx, field.name)
			require.NoError(t, err)
			assert.Equal(t, field.value, value.Int)
		}

		heads[engine.name] = sortedHeadHex(t, ctx, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_InsertIntoExistingMap reproduces batch_insert_into_existing_map.
func TestRustBatch_InsertIntoExistingMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		parent, err := doc.Root().CreateObject(ctx, "parent", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			parent.PutScalar(
				ctx,
				"existing",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
			),
		)
		require.NoError(
			t,
			parent.PutValue(
				ctx,
				"child",
				hydratedMap(
					map[string]automerge.Value{
						"x": hydratedInt(1),
						"y": hydratedInt(2),
					},
				),
			),
		)
		_, err = doc.Commit(ctx, "batch", commitTime)
		require.NoError(t, err)

		existing, err := parent.Scalar(ctx, "existing")
		require.NoError(t, err)
		assert.Equal(t, "value", existing.String)

		child, err := parent.Object(ctx, "child")
		require.NoError(t, err)
		x, err := child.Scalar(ctx, "x")
		require.NoError(t, err)
		assert.Equal(t, int64(1), x.Int)

		heads[engine.name] = sortedHeadHex(t, ctx, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_PutOverwriteWithNestedStructure reproduces
// batch_put_overwrite_with_nested_structure.
func TestRustBatch_PutOverwriteWithNestedStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		list, err := doc.Root().CreateObject(ctx, "items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertValues(
				ctx,
				0,
				[]automerge.Value{
					hydratedString("placeholder"),
					hydratedString("keep"),
				},
			),
		)

		require.NoError(
			t,
			list.PutValueAt(
				ctx,
				0,
				hydratedMap(
					map[string]automerge.Value{
						"name": hydratedString("complex"),
						"children": hydratedList(
							hydratedMap(map[string]automerge.Value{"id": hydratedInt(1)}),
							hydratedMap(map[string]automerge.Value{"id": hydratedInt(2)}),
						),
					},
				),
			),
		)
		_, err = doc.Commit(ctx, "overwrite", commitTime)
		require.NoError(t, err)

		length, err := list.Len(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), length)

		object, err := list.ObjectAt(ctx, 0)
		require.NoError(t, err)
		name, err := object.Scalar(ctx, "name")
		require.NoError(t, err)
		assert.Equal(t, "complex", name.String)

		children, err := object.Object(ctx, "children")
		require.NoError(t, err)
		childrenLength, err := children.Len(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), childrenLength)

		firstChild, err := children.ObjectAt(ctx, 0)
		require.NoError(t, err)
		id, err := firstChild.Scalar(ctx, "id")
		require.NoError(t, err)
		assert.Equal(t, int64(1), id.Int)

		keep, err := list.ScalarAt(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "keep", keep.String)

		heads[engine.name] = sortedHeadHex(t, ctx, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_SpliceMergesCorrectly reproduces splice_merges_correctly.
func TestRustBatch_SpliceMergesCorrectly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(ctx, actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		list1, err := doc1.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list1.InsertScalar(
				ctx,
				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "shared"},
			),
		)
		_, err = doc1.Commit(ctx, "shared", commitTime)
		require.NoError(t, err)

		doc2, err := doc1.Fork(ctx, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)

		require.NoError(
			t,
			list1.SpliceValues(
				ctx,
				1,
				0,
				[]automerge.Value{
					hydratedMap(map[string]automerge.Value{"from": hydratedString("doc1")}),
				},
			),
		)
		_, err = doc1.Commit(ctx, "doc1", commitTime.Add(time.Second))
		require.NoError(t, err)

		list2, err := doc2.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(
			t,
			list2.SpliceValues(
				ctx,
				1,
				0,
				[]automerge.Value{
					hydratedMap(map[string]automerge.Value{"from": hydratedString("doc2")}),
				},
			),
		)
		_, err = doc2.Commit(ctx, "doc2", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(ctx, doc2)
		require.NoError(t, err)

		length, err := list1.Len(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), length)

		first, err := list1.ScalarAt(ctx, 0)
		require.NoError(t, err)
		assert.Equal(t, "shared", first.String)

		results[engine.name] = sortedHeadHex(t, ctx, doc1)
	}

	assert.Equal(t, results["reference"], results["native"])
}
