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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
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

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		require.NoError(
			t,
			doc1.Root().PutValue(

				"obj1",
				hydratedMap(map[string]automerge.Value{"from": hydratedString("doc1")}),
			),
		)
		_, err = doc1.Commit("obj1", commitTime)
		require.NoError(t, err)

		doc2, err := doc1.Fork(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		require.NoError(
			t,
			doc2.Root().PutValue(

				"obj2",
				hydratedMap(map[string]automerge.Value{"from": hydratedString("doc2")}),
			),
		)
		_, err = doc2.Commit("obj2", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		obj1, err := doc1.Root().Object("obj1")
		require.NoError(t, err)
		value, err := obj1.Scalar("from")
		require.NoError(t, err)
		assert.Equal(t, "doc1", value.String)

		obj2, err := doc1.Root().Object("obj2")
		require.NoError(t, err)
		value, err = obj2.Scalar("from")
		require.NoError(t, err)
		assert.Equal(t, "doc2", value.String)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_MultipleInserts reproduces multiple_batch_inserts.
func TestRustBatch_MultipleInserts(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		require.NoError(
			t,
			doc.Root().PutValue(

				"first",
				hydratedMap(map[string]automerge.Value{"a": hydratedInt(1)}),
			),
		)
		require.NoError(
			t,
			doc.Root().PutValue(

				"second",
				hydratedMap(map[string]automerge.Value{"b": hydratedInt(2)}),
			),
		)
		require.NoError(
			t,
			doc.Root().PutValue(

				"third",
				hydratedMap(map[string]automerge.Value{"c": hydratedInt(3)}),
			),
		)
		_, err = doc.Commit("batches", commitTime)
		require.NoError(t, err)

		for key, field := range map[string]struct {
			name  string
			value int64
		}{
			"first":  {"a", 1},
			"second": {"b", 2},
			"third":  {"c", 3},
		} {
			object, err := doc.Root().Object(key)
			require.NoError(t, err)
			value, err := object.Scalar(field.name)
			require.NoError(t, err)
			assert.Equal(t, field.value, value.Int)
		}

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_InsertIntoExistingMap reproduces batch_insert_into_existing_map.
func TestRustBatch_InsertIntoExistingMap(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		parent, err := doc.Root().CreateObject("parent", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			parent.PutScalar(

				"existing",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
			),
		)
		require.NoError(
			t,
			parent.PutValue(

				"child",
				hydratedMap(
					map[string]automerge.Value{
						"x": hydratedInt(1),
						"y": hydratedInt(2),
					},
				),
			),
		)
		_, err = doc.Commit("batch", commitTime)
		require.NoError(t, err)

		existing, err := parent.Scalar("existing")
		require.NoError(t, err)
		assert.Equal(t, "value", existing.String)

		child, err := parent.Object("child")
		require.NoError(t, err)
		x, err := child.Scalar("x")
		require.NoError(t, err)
		assert.Equal(t, int64(1), x.Int)

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_PutOverwriteWithNestedStructure reproduces
// batch_put_overwrite_with_nested_structure.
func TestRustBatch_PutOverwriteWithNestedStructure(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)

		list, err := doc.Root().CreateObject("items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertValues(

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
		_, err = doc.Commit("overwrite", commitTime)
		require.NoError(t, err)

		length, err := list.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(2), length)

		object, err := list.ObjectAt(0)
		require.NoError(t, err)
		name, err := object.Scalar("name")
		require.NoError(t, err)
		assert.Equal(t, "complex", name.String)

		children, err := object.Object("children")
		require.NoError(t, err)
		childrenLength, err := children.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(2), childrenLength)

		firstChild, err := children.ObjectAt(0)
		require.NoError(t, err)
		id, err := firstChild.Scalar("id")
		require.NoError(t, err)
		assert.Equal(t, int64(1), id.Int)

		keep, err := list.ScalarAt(1)
		require.NoError(t, err)
		assert.Equal(t, "keep", keep.String)

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRustBatch_SpliceMergesCorrectly reproduces splice_merges_correctly.
func TestRustBatch_SpliceMergesCorrectly(t *testing.T) {
	t.Parallel()

	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		list1, err := doc1.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list1.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "shared"},
			),
		)
		_, err = doc1.Commit("shared", commitTime)
		require.NoError(t, err)

		doc2, err := doc1.Fork(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)

		require.NoError(
			t,
			list1.SpliceValues(

				1,
				0,
				[]automerge.Value{
					hydratedMap(map[string]automerge.Value{"from": hydratedString("doc1")}),
				},
			),
		)
		_, err = doc1.Commit("doc1", commitTime.Add(time.Second))
		require.NoError(t, err)

		list2, err := doc2.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			list2.SpliceValues(

				1,
				0,
				[]automerge.Value{
					hydratedMap(map[string]automerge.Value{"from": hydratedString("doc2")}),
				},
			),
		)
		_, err = doc2.Commit("doc2", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		length, err := list1.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(3), length)

		first, err := list1.ScalarAt(0)
		require.NoError(t, err)
		assert.Equal(t, "shared", first.String)

		results[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, results["reference"], results["native"])
}
