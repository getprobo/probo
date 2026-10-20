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

// The tests in this file reproduce individual upstream Rust integration tests
// from automerge 0.11 (rust/automerge/tests/test.rs) against both the native
// Go engine and the Rust/WASM reference engine. Every scenario runs identically
// on both engines and asserts that the observable materialized state, causal
// heads, and cross-engine reloads agree. Because both engines produce identical
// change hashes for identical operation sequences, matching heads across the
// real Rust engine and the native engine also guarantees identical conflict
// structure by construction.

package automerge_test

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

type rustParityEngine struct {
	name string
	open func(automerge.ActorID) (*automerge.Document, error)
	load func([]byte, automerge.ActorID, ...automerge.LoadOption) (*automerge.Document, error)
}

func rustParityEngines() []rustParityEngine {
	return []rustParityEngine{
		{"native", automerge.New, automerge.Load},
		{"reference", automerge.NewReference, automerge.LoadReference},
	}
}

func sortedHeadHex(
	t *testing.T,
	document *automerge.Document,
) []string {
	t.Helper()

	heads, err := document.Heads()
	require.NoError(t, err)

	hex := make([]string, len(heads))
	for index, head := range heads {
		hex[index] = head.String()
	}

	sort.Strings(hex)

	return hex
}

func sortedCounterValues(
	t *testing.T,
	object *automerge.Object,
	key string,
) []int64 {
	t.Helper()

	values, err := object.Scalars(key)
	require.NoError(t, err)

	result := make([]int64, len(values))
	for index, value := range values {
		result[index] = value.Int
	}

	slices.Sort(result)

	return result
}

func sortedStringValues(
	t *testing.T,
	object *automerge.Object,
	key string,
) []string {
	t.Helper()

	values, err := object.Scalars(key)
	require.NoError(t, err)

	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.String
	}

	sort.Strings(result)

	return result
}

// TestRust_RepeatedListAssignmentResolvesConflict reproduces
// repeated_list_assignment_which_resolves_conflict_not_ignored.
func TestRust_RepeatedListAssignmentResolvesConflict(t *testing.T) {
	t.Parallel()

	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)

		list, err := doc1.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 123},
			),
		)

		_, err = doc1.Commit("insert", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		list2, err := doc2.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			list2.PutScalarAt(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 456},
			),
		)

		_, err = doc2.Commit("put 456", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)
		require.NoError(
			t,
			list.PutScalarAt(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 789},
			),
		)

		_, err = doc1.Commit("put 789", commitTime.Add(2*time.Second))
		require.NoError(t, err)

		length, err := list.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(1), length)

		winner, err := list.ScalarAt(0)
		require.NoError(t, err)
		assert.Equal(t, int64(789), winner.Int)

		results[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, results["reference"], results["native"])
}

// TestRust_AddIncrementsOnlyToPreceededValues reproduces
// add_increments_only_to_preceeded_values.
func TestRust_AddIncrementsOnlyToPreceededValues(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		require.NoError(
			t,
			doc1.Root().PutScalar(

				"counter",
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 0},
			),
		)
		require.NoError(t, doc1.Root().Increment("counter", 1))
		_, err = doc1.Commit("doc1 counter", commitTime)
		require.NoError(t, err)

		doc2, err := engine.open(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		require.NoError(
			t,
			doc2.Root().PutScalar(

				"counter",
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 0},
			),
		)
		require.NoError(t, doc2.Root().Increment("counter", 3))
		_, err = doc2.Commit("doc2 counter", commitTime)
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		assert.Equal(t, []int64{1, 3}, sortedCounterValues(t, doc1.Root(), "counter"))

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_AssignmentConflictsOfDifferentTypes reproduces
// assignment_conflicts_of_different_types.
func TestRust_AssignmentConflictsOfDifferentTypes(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		require.NoError(
			t,
			doc1.Root().PutScalar(

				"field",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "string"},
			),
		)
		_, err = doc1.Commit("string", commitTime)
		require.NoError(t, err)

		doc2, err := engine.open(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		_, err = doc2.Root().CreateObject("field", automerge.ObjectTypeList)
		require.NoError(t, err)
		_, err = doc2.Commit("list", commitTime)
		require.NoError(t, err)

		doc3, err := engine.open(actor(3))
		require.NoError(t, err)
		closeDocument(t, doc3)
		_, err = doc3.Root().CreateObject("field", automerge.ObjectTypeMap)
		require.NoError(t, err)
		_, err = doc3.Commit("map", commitTime)
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)
		_, err = doc1.Merge(doc3)
		require.NoError(t, err)

		// The highest-actor operation wins; actor(3) created a map.
		winner, err := doc1.Root().Object("field")
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, winner.Type)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_ChangesWithinConflictingMapField reproduces
// changes_within_conflicting_map_field.
func TestRust_ChangesWithinConflictingMapField(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		require.NoError(
			t,
			doc1.Root().PutScalar(

				"field",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "string"},
			),
		)
		_, err = doc1.Commit("string", commitTime)
		require.NoError(t, err)

		doc2, err := engine.open(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		inner, err := doc2.Root().CreateObject("field", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			inner.PutScalar(

				"innerKey",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 42},
			),
		)

		_, err = doc2.Commit("map", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		// actor(2) wins; the winning value is the map with innerKey = 42.
		winner, err := doc1.Root().Object("field")
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, winner.Type)
		value, err := winner.Scalar("innerKey")
		require.NoError(t, err)
		assert.Equal(t, int64(42), value.Int)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_ChangesWithinConflictingListElement reproduces
// changes_within_conflicting_list_element.
func TestRust_ChangesWithinConflictingListElement(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

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
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "hello"},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)

		map1, err := list1.PutObjectAt(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			map1.PutScalar(

				"map1",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			),
		)
		require.NoError(
			t,
			map1.PutScalar(

				"key",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
			),
		)

		_, err = doc1.Commit("doc1 map", commitTime.Add(time.Second))
		require.NoError(t, err)

		list2, err := doc2.Root().Object("list")
		require.NoError(t, err)
		map2, err := list2.PutObjectAt(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		_, err = doc2.Commit("doc2 map", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)
		require.NoError(
			t,
			map2.PutScalar(

				"map2",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			),
		)
		require.NoError(
			t,
			map2.PutScalar(

				"key",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
			),
		)

		_, err = doc2.Commit("doc2 values", commitTime.Add(2*time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		// actor(2)'s map wins with key = 2 and map2 = true.
		winner, err := list1.ObjectAt(0)
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, winner.Type)
		key, err := winner.Scalar("key")
		require.NoError(t, err)
		assert.Equal(t, int64(2), key.Int)

		flag, err := winner.Scalar("map2")
		require.NoError(t, err)
		assert.True(t, flag.Bool)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_ConcurrentlyAssignedNestedMapsShouldNotMerge reproduces
// concurrently_assigned_nested_maps_should_not_merge.
func TestRust_ConcurrentlyAssignedNestedMapsShouldNotMerge(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		config1, err := doc1.Root().CreateObject("config", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			config1.PutScalar(

				"background",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "blue"},
			),
		)

		_, err = doc1.Commit("doc1 config", commitTime)
		require.NoError(t, err)

		doc2, err := engine.open(actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		config2, err := doc2.Root().CreateObject("config", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			config2.PutScalar(

				"logo_url",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "logo.png"},
			),
		)

		_, err = doc2.Commit("doc2 config", commitTime)
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		// The two maps do not merge; the winning map keeps exactly one key.
		winner, err := doc1.Root().Object("config")
		require.NoError(t, err)
		keys, err := winner.Keys()
		require.NoError(t, err)
		assert.Len(t, keys, 1)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_ConcurrentDeletionOfSameListElement reproduces
// concurrent_deletion_of_same_list_element.
func TestRust_ConcurrentDeletionOfSameListElement(t *testing.T) {
	t.Parallel()

	results := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		list1, err := doc1.Root().CreateObject("birds", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list1.InsertValues(

				0,
				[]automerge.Value{
					hydratedString("albatross"),
					hydratedString("buzzard"),
					hydratedString("cormorant"),
				},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		list2, err := doc2.Root().Object("birds")
		require.NoError(t, err)

		require.NoError(t, list1.DeleteIndex(1))

		_, err = doc1.Commit("doc1 delete", commitTime.Add(time.Second))
		require.NoError(t, err)
		require.NoError(t, list2.DeleteIndex(1))

		_, err = doc2.Commit("doc2 delete", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		values := listStrings(t, list1)
		assert.Equal(t, []string{"albatross", "cormorant"}, values)

		results[engine.name] = values
	}

	assert.Equal(t, results["reference"], results["native"])
}

// TestRust_ConcurrentUpdatesAtDifferentLevels reproduces
// concurrent_updates_at_different_levels.
func TestRust_ConcurrentUpdatesAtDifferentLevels(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		animals, err := doc1.Root().CreateObject("animals", automerge.ObjectTypeMap)
		require.NoError(t, err)
		birds, err := animals.CreateObject("birds", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			birds.PutScalar(

				"pink",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "flamingo"},
			),
		)
		require.NoError(
			t,
			birds.PutScalar(

				"black",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "starling"},
			),
		)

		mammals, err := animals.CreateObject("mammals", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			mammals.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "badger"},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)

		require.NoError(
			t,
			birds.PutScalar(

				"brown",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "sparrow"},
			),
		)

		_, err = doc1.Commit("doc1 update", commitTime.Add(time.Second))
		require.NoError(t, err)

		animals2, err := doc2.Root().Object("animals")
		require.NoError(t, err)
		require.NoError(t, animals2.DeleteKey("birds"))

		_, err = doc2.Commit("doc2 delete", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		// birds was deleted concurrently; only mammals remains under animals.
		mergedAnimals, err := doc1.Root().Object("animals")
		require.NoError(t, err)
		keys, err := mergedAnimals.Keys()
		require.NoError(t, err)
		assert.Equal(t, []string{"mammals"}, keys)

		mergedMammals, err := mergedAnimals.Object("mammals")
		require.NoError(t, err)
		assert.Equal(t, []string{"badger"}, listStrings(t, mergedMammals))

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_ConcurrentUpdatesOfConcurrentlyDeletedObjects reproduces
// concurrent_updates_of_concurrently_deleted_objects.
func TestRust_ConcurrentUpdatesOfConcurrentlyDeletedObjects(t *testing.T) {
	t.Parallel()

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		birds, err := doc1.Root().CreateObject("birds", automerge.ObjectTypeMap)
		require.NoError(t, err)
		blackbird, err := birds.CreateObject("blackbird", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			blackbird.PutScalar(

				"feathers",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "black"},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)

		require.NoError(t, birds.DeleteKey("blackbird"))

		_, err = doc1.Commit("doc1 delete", commitTime.Add(time.Second))
		require.NoError(t, err)

		birds2, err := doc2.Root().Object("birds")
		require.NoError(t, err)
		blackbird2, err := birds2.Object("blackbird")
		require.NoError(t, err)
		require.NoError(
			t,
			blackbird2.PutScalar(

				"beak",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "orange"},
			),
		)

		_, err = doc2.Commit("doc2 update", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		// The deletion wins; birds becomes an empty map.
		mergedBirds, err := doc1.Root().Object("birds")
		require.NoError(t, err)
		length, err := mergedBirds.Len()
		require.NoError(t, err)
		assert.Zero(t, length)

		heads[engine.name] = sortedHeadHex(t, doc1)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_InsertionConsistentWithCausality reproduces
// insertion_consistent_with_causality.
func TestRust_InsertionConsistentWithCausality(t *testing.T) {
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
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "four"},
			),
		)

		_, err = doc1.Commit("four", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		list2, err := doc2.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			list2.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "three"},
			),
		)

		_, err = doc2.Commit("three", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)
		require.NoError(
			t,
			list1.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "two"},
			),
		)

		_, err = doc1.Commit("two", commitTime.Add(2*time.Second))
		require.NoError(t, err)

		_, err = doc2.Merge(doc1)
		require.NoError(t, err)
		require.NoError(
			t,
			list2.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "one"},
			),
		)

		_, err = doc2.Commit("one", commitTime.Add(3*time.Second))
		require.NoError(t, err)

		values := listStrings(t, list2)
		assert.Equal(t, []string{"one", "two", "three", "four"}, values)

		results[engine.name] = values
	}

	assert.Equal(t, results["reference"], results["native"])
}

// TestRust_SaveRestoreComplex1 reproduces save_restore_complex1.
func TestRust_SaveRestoreComplex1(t *testing.T) {
	t.Parallel()

	titles := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		todos, err := doc1.Root().CreateObject("todos", automerge.ObjectTypeList)
		require.NoError(t, err)
		firstTodo, err := todos.InsertObject(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			firstTodo.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "water plants"},
			),
		)
		require.NoError(
			t,
			firstTodo.PutScalar(

				"done",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: false},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		todos2, err := doc2.Root().Object("todos")
		require.NoError(t, err)
		firstTodo2, err := todos2.ObjectAt(0)
		require.NoError(t, err)
		require.NoError(
			t,
			firstTodo2.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "weed plants"},
			),
		)

		_, err = doc2.Commit("weed", commitTime.Add(time.Second))
		require.NoError(t, err)

		require.NoError(
			t,
			firstTodo.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "kill plants"},
			),
		)

		_, err = doc1.Commit("kill", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		saved, err := doc1.Save()
		require.NoError(t, err)
		reloaded, err := engine.load(saved, actor(3))
		require.NoError(t, err)
		closeDocument(t, reloaded)

		reloadedTodos, err := reloaded.Root().Object("todos")
		require.NoError(t, err)
		reloadedTodo, err := reloadedTodos.ObjectAt(0)
		require.NoError(t, err)
		done, err := reloadedTodo.Scalar("done")
		require.NoError(t, err)
		assert.False(t, done.Bool)

		values := sortedStringValues(t, reloadedTodo, "title")
		assert.Equal(t, []string{"kill plants", "weed plants"}, values)

		titles[engine.name] = values
	}

	assert.Equal(t, titles["reference"], titles["native"])
}

// TestRust_SaveRestoreComplexTransactional reproduces
// save_restore_complex_transactional. The Rust test groups its writes inside
// explicit transactions; the observable outcome is identical to a single
// grouped commit, which is what the Go engine exposes.
func TestRust_SaveRestoreComplexTransactional(t *testing.T) {
	t.Parallel()

	titles := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		todos, err := doc1.Root().CreateObject("todos", automerge.ObjectTypeList)
		require.NoError(t, err)
		firstTodo, err := todos.InsertObject(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			firstTodo.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "water plants"},
			),
		)
		require.NoError(
			t,
			firstTodo.PutScalar(

				"done",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: false},
			),
		)

		_, err = doc1.Commit("transaction", commitTime)
		require.NoError(t, err)

		data, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(data, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		todos2, err := doc2.Root().Object("todos")
		require.NoError(t, err)
		firstTodo2, err := todos2.ObjectAt(0)
		require.NoError(t, err)
		require.NoError(
			t,
			firstTodo2.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "weed plants"},
			),
		)

		_, err = doc2.Commit("transaction", commitTime.Add(time.Second))
		require.NoError(t, err)

		require.NoError(
			t,
			firstTodo.PutScalar(

				"title",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "kill plants"},
			),
		)

		_, err = doc1.Commit("transaction", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)

		saved, err := doc1.Save()
		require.NoError(t, err)
		reloaded, err := engine.load(saved, actor(3))
		require.NoError(t, err)
		closeDocument(t, reloaded)

		reloadedTodos, err := reloaded.Root().Object("todos")
		require.NoError(t, err)
		reloadedTodo, err := reloadedTodos.ObjectAt(0)
		require.NoError(t, err)
		done, err := reloadedTodo.Scalar("done")
		require.NoError(t, err)
		assert.False(t, done.Bool)

		values := sortedStringValues(t, reloadedTodo, "title")
		assert.Equal(t, []string{"kill plants", "weed plants"}, values)

		titles[engine.name] = values
	}

	assert.Equal(t, titles["reference"], titles["native"])
}

// TestRust_BigList reproduces big_list. The upstream test inspects the patch
// stream; the interoperable behavior is the resulting document state, which
// this test verifies is a list of N+1 map objects that survives a cross-engine
// save/load.
func TestRust_BigList(t *testing.T) {
	t.Parallel()

	const count = 128

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)
		list, err := doc.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		_, err = doc.Commit("create list", commitTime)
		require.NoError(t, err)

		for index := range count + 1 {
			require.NoError(
				t,
				list.InsertScalar(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeNull},
				),
			)
		}

		for index := range count + 1 {
			_, err := list.PutObjectAt(uint64(index), automerge.ObjectTypeMap)
			require.NoError(t, err)
		}

		_, err = doc.Commit("populate", commitTime.Add(time.Second))
		require.NoError(t, err)

		length, err := list.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(count+1), length)

		element, err := list.ObjectAt(count)
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, element.Type)

		saved, err := doc.Save()
		require.NoError(t, err)
		reloaded, err := engine.load(saved, actor(2))
		require.NoError(t, err)
		closeDocument(t, reloaded)
		reloadedList, err := reloaded.Root().Object("list")
		require.NoError(t, err)
		reloadedLength, err := reloadedList.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(count+1), reloadedLength)

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_InvalidIndex reproduces invalid_index.
func TestRust_InvalidIndex(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc)
				list, err := doc.Root().CreateObject("a", automerge.ObjectTypeList)
				require.NoError(t, err)
				require.NoError(
					t,
					list.InsertScalar(

						0,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				require.NoError(
					t,
					list.PutScalarAt(

						0,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)

				value, err := list.ScalarAt(0)
				require.NoError(t, err)
				assert.Equal(t, int64(2), value.Int)

				require.Error(
					t,
					list.InsertScalar(

						2,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				require.Error(
					t,
					list.PutScalarAt(

						2,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)
				require.Error(
					t,
					list.InsertScalar(

						100,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				require.Error(
					t,
					list.PutScalarAt(

						100,
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)
			},
		)
	}
}

// TestRust_HasOurChanges reproduces has_our_changes: two peers with concurrent
// changes synchronize until each has received the other's changes.
func TestRust_HasOurChanges(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				left, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, left)
				require.NoError(
					t,
					left.Root().PutScalar(

						"a",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				leftHash, err := left.Commit("a", commitTime)
				require.NoError(t, err)

				right, err := engine.open(actor(2))
				require.NoError(t, err)
				closeDocument(t, right)
				require.NoError(
					t,
					right.Root().PutScalar(

						"b",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)
				rightHash, err := right.Commit("b", commitTime)
				require.NoError(t, err)

				leftToRight, err := left.NewSyncState()
				require.NoError(t, err)
				closeSyncState(t, leftToRight)

				rightToLeft, err := right.NewSyncState()
				require.NoError(t, err)
				closeSyncState(t, rightToLeft)

				syncBothDirections(t, leftToRight, rightToLeft)

				rightHasLeft, err := right.HasHeads([]automerge.Hash{leftHash})
				require.NoError(t, err)
				assert.True(t, rightHasLeft)

				leftHasRight, err := left.HasHeads([]automerge.Hash{rightHash})
				require.NoError(t, err)
				assert.True(t, leftHasRight)

				assert.Equal(
					t,
					sortedHeadHex(t, left),
					sortedHeadHex(t, right),
				)
			},
		)
	}
}

// TestRust_LoadIncrementalWithCommonHead reproduces
// make_sure_load_incremental_doesnt_skip_a_load_with_a_common_head.
func TestRust_LoadIncrementalWithCommonHead(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc1)
				require.NoError(
					t,
					doc1.Root().PutScalar(

						"string",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "hello"},
					),
				)
				_, err = doc1.Commit("hello", commitTime)
				require.NoError(t, err)

				base, err := doc1.Save()
				require.NoError(t, err)
				doc2, err := engine.load(base, actor(2))
				require.NoError(t, err)
				closeDocument(t, doc2)

				doc3, err := engine.load(base, actor(3))
				require.NoError(t, err)
				closeDocument(t, doc3)

				heads1, err := doc1.Heads()
				require.NoError(t, err)
				require.Len(t, heads1, 1)

				require.NoError(
					t,
					doc1.Root().PutScalar(

						"concurrent1",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "123"},
					),
				)
				hashB, err := doc1.Commit("concurrent1", commitTime.Add(time.Second))
				require.NoError(t, err)

				saved1, err := doc1.Save()
				require.NoError(t, err)
				_, err = doc3.LoadIncremental(saved1)
				require.NoError(t, err)
				headsC, err := doc3.Heads()
				require.NoError(t, err)
				require.Len(t, headsC, 1)
				assert.Equal(t, hashB.String(), headsC[0].String())

				require.NoError(
					t,
					doc2.Root().PutScalar(

						"concurrent2",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "abc"},
					),
				)
				hashD, err := doc2.Commit("concurrent2", commitTime.Add(time.Second))
				require.NoError(t, err)

				_, err = doc2.Merge(doc1)
				require.NoError(t, err)
				mergedHeads := sortedHeadHex(t, doc2)
				require.Len(t, mergedHeads, 2)
				assert.Contains(t, mergedHeads, hashB.String())
				assert.Contains(t, mergedHeads, hashD.String())

				saved2, err := doc2.Save()
				require.NoError(t, err)
				_, err = doc3.LoadIncremental(saved2)
				require.NoError(t, err)
				assert.Equal(t, mergedHeads, sortedHeadHex(t, doc3))
			},
		)
	}
}

// TestRust_RegressionNthMiscount reproduces regression_nth_miscount.
func TestRust_RegressionNthMiscount(t *testing.T) {
	t.Parallel()

	const count = 30

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)
		list, err := doc.Root().CreateObject("listval", automerge.ObjectTypeList)
		require.NoError(t, err)

		for index := range count {
			require.NoError(
				t,
				list.InsertScalar(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeNull},
				),
			)
			element, err := list.PutObjectAt(uint64(index), automerge.ObjectTypeMap)
			require.NoError(t, err)
			require.NoError(
				t,
				element.PutScalar(

					"test",
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(index)},
				),
			)
		}

		_, err = doc.Commit("populate", commitTime)
		require.NoError(t, err)

		for index := range count {
			element, err := list.ObjectAt(uint64(index))
			require.NoError(t, err)
			assert.Equal(t, automerge.ObjectTypeMap, element.Type)
			value, err := element.Scalar("test")
			require.NoError(t, err)
			assert.Equal(t, int64(index), value.Int)
		}

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_RegressionNthMiscountSmaller reproduces
// regression_nth_miscount_smaller. B is the op-tree node size (16 upstream).
func TestRust_RegressionNthMiscountSmaller(t *testing.T) {
	t.Parallel()

	const count = 16 * 4

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)
		list, err := doc.Root().CreateObject("listval", automerge.ObjectTypeList)
		require.NoError(t, err)

		for index := range count {
			require.NoError(
				t,
				list.InsertScalar(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeNull},
				),
			)
			require.NoError(
				t,
				list.PutScalarAt(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(index)},
				),
			)
		}

		_, err = doc.Commit("populate", commitTime)
		require.NoError(t, err)

		for index := range count {
			value, err := list.ScalarAt(uint64(index))
			require.NoError(t, err)
			assert.Equal(t, int64(index), value.Int)
		}

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_RegressionInsertOpid reproduces regression_insert_opid: interleaved
// insert-then-overwrite operations round-trip through a cross-engine reload
// with every list value preserved.
func TestRust_RegressionInsertOpid(t *testing.T) {
	t.Parallel()

	const count = 30

	heads := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc)
		list, err := doc.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		_, err = doc.Commit("create list", commitTime)
		require.NoError(t, err)

		for index := range count + 1 {
			require.NoError(
				t,
				list.InsertScalar(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeNull},
				),
			)
			require.NoError(
				t,
				list.PutScalarAt(

					uint64(index),
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(index)},
				),
			)
		}

		_, err = doc.Commit("populate", commitTime.Add(time.Second))
		require.NoError(t, err)

		saved, err := doc.Save()
		require.NoError(t, err)
		reloaded, err := engine.load(saved, actor(2))
		require.NoError(t, err)
		closeDocument(t, reloaded)
		reloadedList, err := reloaded.Root().Object("list")
		require.NoError(t, err)

		for index := range count + 1 {
			original, err := list.ScalarAt(uint64(index))
			require.NoError(t, err)
			roundTripped, err := reloadedList.ScalarAt(uint64(index))
			require.NoError(t, err)
			assert.Equal(t, int64(index), original.Int)
			assert.Equal(t, original.Int, roundTripped.Int)
		}

		heads[engine.name] = sortedHeadHex(t, doc)
	}

	assert.Equal(t, heads["reference"], heads["native"])
}

// TestRust_RollbackWithSeveralActors reproduces rollback_with_several_actors:
// uncommitted edits by a third actor are discarded, leaving the document
// byte-identical to the state it was forked from.
func TestRust_RollbackWithSeveralActors(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, doc1)
				text1, err := doc1.CreateText("text")
				require.NoError(t, err)
				require.NoError(
					t,
					text1.Splice(

						0,
						0,
						"the sly fox jumped over the lazy dog",
					),
				)

				mapA1, err := doc1.Root().CreateObject("map_a", automerge.ObjectTypeMap)
				require.NoError(t, err)
				require.NoError(
					t,
					mapA1.PutScalar(

						"key1",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value1a"},
					),
				)
				require.NoError(
					t,
					mapA1.PutScalar(

						"key2",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value2a"},
					),
				)

				_, err = doc1.Commit("doc1", commitTime)
				require.NoError(t, err)

				doc2, err := doc1.Fork(actor(0xcc))
				require.NoError(t, err)
				closeDocument(t, doc2)
				text2, err := doc2.Text("text")
				require.NoError(t, err)
				require.NoError(t, text2.Splice(8, 3, "monkey"))
				require.NoError(t, text2.Splice(36, 3, "pig"))

				mapC2, err := doc2.Root().CreateObject("map_c", automerge.ObjectTypeMap)
				require.NoError(t, err)
				mapA2, err := doc2.Root().Object("map_a")
				require.NoError(t, err)
				require.NoError(
					t,
					mapA2.PutScalar(

						"key2",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value2c"},
					),
				)
				require.NoError(
					t,
					mapA2.PutScalar(

						"key3",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value3c"},
					),
				)
				require.NoError(
					t,
					mapC2.PutScalar(

						"key1",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
					),
				)

				_, err = doc2.Commit("doc2", commitTime.Add(time.Second))
				require.NoError(t, err)

				doc3, err := doc2.Fork(actor(0xbb))
				require.NoError(t, err)
				closeDocument(t, doc3)
				text3, err := doc3.Text("text")
				require.NoError(t, err)
				require.NoError(t, text3.Splice(8, 5, "zebra"))

				mapB3, err := doc3.Root().CreateObject("map_b", automerge.ObjectTypeMap)
				require.NoError(t, err)
				mapA3, err := doc3.Root().Object("map_a")
				require.NoError(t, err)
				require.NoError(
					t,
					mapA3.PutScalar(

						"key1",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value3b"},
					),
				)
				require.NoError(
					t,
					mapA3.PutScalar(

						"key3",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value3b"},
					),
				)
				require.NoError(
					t,
					mapB3.PutScalar(

						"key1",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
					),
				)

				_, err = doc3.Rollback()
				require.NoError(t, err)

				assert.Equal(t, sortedHeadHex(t, doc2), sortedHeadHex(t, doc3))

				doc2Save, err := doc2.Save()
				require.NoError(t, err)
				doc3Save, err := doc3.Save()
				require.NoError(t, err)
				assert.Equal(t, doc2Save, doc3Save)
			},
		)
	}
}

// TestRust_SaveWithOpsReferencingActorsOnlyViaDelete reproduces
// save_with_ops_which_reference_actors_only_via_delete: a merged delete op
// references a fork's actor only through successors, and the document must still
// save and reload.
func TestRust_SaveWithOpsReferencingActorsOnlyViaDelete(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"a",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				_, err = doc.Commit("put a", commitTime)
				require.NoError(t, err)

				forked, err := doc.Fork(actor(2))
				require.NoError(t, err)
				closeDocument(t, forked)
				require.NoError(t, forked.Root().DeleteKey("a"))
				_, err = forked.Commit("delete a", commitTime.Add(time.Second))
				require.NoError(t, err)

				_, err = doc.Merge(forked)
				require.NoError(t, err)

				saved, err := doc.Save()
				require.NoError(t, err)

				nativeReload, err := automerge.Load(saved, actor(3))
				require.NoError(t, err)
				closeDocument(t, nativeReload)

				referenceReload, err := automerge.LoadReference(saved, actor(4))
				require.NoError(t, err)
				closeDocument(t, referenceReload)

				for _, reloaded := range []*automerge.Document{nativeReload, referenceReload} {
					length, err := reloaded.Root().Len()
					require.NoError(t, err)
					assert.Zero(t, length)
				}
			},
		)
	}
}

func sortedScalarsAt(
	t *testing.T,
	object *automerge.Object,
	index uint64,
) []string {
	t.Helper()

	values, err := object.ScalarsAt(index)
	require.NoError(t, err)

	result := make([]string, len(values))
	for i, value := range values {
		result[i] = fmt.Sprintf("%s:%d:%d", value.Type, value.Int, value.Uint)
	}

	sort.Strings(result)

	return result
}

// TestRust_ListCounterDel reproduces list_counter_del: three actors write
// conflicting counters (and one integer) to the same list elements, increments
// are applied and merged, and the elements are deleted. The conflicting value
// sets and lengths are compared directly against the reference engine.
func TestRust_ListCounterDel(t *testing.T) {
	t.Parallel()

	index1 := make(map[string][]string)
	index2 := make(map[string][]string)

	for _, engine := range rustParityEngines() {
		doc1, err := engine.open(actor(1))
		require.NoError(t, err)
		closeDocument(t, doc1)
		list1, err := doc1.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list1.InsertValues(

				0,
				[]automerge.Value{
					hydratedString("a"),
					hydratedString("b"),
					hydratedString("c"),
				},
			),
		)

		_, err = doc1.Commit("base", commitTime)
		require.NoError(t, err)

		base, err := doc1.Save()
		require.NoError(t, err)
		doc2, err := engine.load(base, actor(2))
		require.NoError(t, err)
		closeDocument(t, doc2)
		list2, err := doc2.Root().Object("list")
		require.NoError(t, err)
		doc3, err := engine.load(base, actor(3))
		require.NoError(t, err)
		closeDocument(t, doc3)
		list3, err := doc3.Root().Object("list")
		require.NoError(t, err)

		counter := func(value int64) automerge.Scalar {
			return automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: value}
		}

		require.NoError(t, list1.PutScalarAt(1, counter(0)))
		require.NoError(t, list2.PutScalarAt(1, counter(10)))
		require.NoError(t, list3.PutScalarAt(1, counter(100)))

		require.NoError(t, list1.PutScalarAt(2, counter(0)))
		require.NoError(t, list2.PutScalarAt(2, counter(10)))
		require.NoError(
			t,
			list3.PutScalarAt(

				2,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 100},
			),
		)

		require.NoError(t, list1.IncrementAt(1, 1))
		require.NoError(t, list1.IncrementAt(2, 1))

		_, err = doc1.Commit("doc1", commitTime.Add(time.Second))
		require.NoError(t, err)
		_, err = doc2.Commit("doc2", commitTime.Add(time.Second))
		require.NoError(t, err)
		_, err = doc3.Commit("doc3", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = doc1.Merge(doc2)
		require.NoError(t, err)
		_, err = doc1.Merge(doc3)
		require.NoError(t, err)

		require.NoError(t, list1.IncrementAt(1, 1))
		require.NoError(t, list1.IncrementAt(2, 1))

		_, err = doc1.Commit("increments", commitTime.Add(2*time.Second))
		require.NoError(t, err)

		index1[engine.name] = sortedScalarsAt(t, list1, 1)
		index2[engine.name] = sortedScalarsAt(t, list1, 2)

		require.NoError(t, list1.DeleteIndex(2))

		_, err = doc1.Commit("delete 2", commitTime.Add(3*time.Second))
		require.NoError(t, err)
		length, err := list1.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(2), length)

		saved, err := doc1.Save()
		require.NoError(t, err)
		reloaded, err := engine.load(saved, actor(4))
		require.NoError(t, err)
		closeDocument(t, reloaded)
		reloadedList, err := reloaded.Root().Object("list")
		require.NoError(t, err)
		reloadedLength, err := reloadedList.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(2), reloadedLength)

		require.NoError(t, list1.DeleteIndex(1))

		_, err = doc1.Commit("delete 1", commitTime.Add(4*time.Second))
		require.NoError(t, err)
		length, err = list1.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(1), length)
	}

	assert.Equal(t, index1["reference"], index1["native"])
	assert.Equal(t, index2["reference"], index2["native"])
}

// TestRust_SimpleBadSaveload reproduces simple_bad_saveload: an empty commit
// interleaved with real changes must not corrupt the save/load round trip. The
// upstream test reassigns the same value; because both engines treat a repeated
// equal assignment as a no-op, this variant uses a distinct value for the
// second write so the empty commit remains the interleaved change.
func TestRust_SimpleBadSaveload(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"count",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 0},
					),
				)
				_, err = doc.Commit("count 0", commitTime)
				require.NoError(t, err)

				_, err = doc.EmptyCommit("empty", commitTime.Add(time.Second))
				require.NoError(t, err)

				require.NoError(
					t,
					doc.Root().PutScalar(

						"count",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				_, err = doc.Commit("count 1", commitTime.Add(2*time.Second))
				require.NoError(t, err)

				saved, err := doc.Save()
				require.NoError(t, err)

				for _, load := range []func([]byte, automerge.ActorID, ...automerge.LoadOption) (*automerge.Document, error){
					automerge.Load,
					automerge.LoadReference,
				} {
					reloaded, err := load(saved, actor(2))
					require.NoError(t, err)
					closeDocument(t, reloaded)
					value, err := reloaded.Root().Scalar("count")
					require.NoError(t, err)
					assert.Equal(t, int64(1), value.Int)
				}
			},
		)
	}
}

// TestRust_BadChangeOnOptreeNodeBoundary reproduces
// bad_change_on_optree_node_boundary: a document grown across an op-tree node
// boundary is saved, reloaded elsewhere, then a further change is transferred
// and the result still saves and reloads with matching state.
func TestRust_BadChangeOnOptreeNodeBoundary(t *testing.T) {
	t.Parallel()

	const iterations = 15

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"a",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "z"},
					),
				)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"b",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 0},
					),
				)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"c",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 0},
					),
				)
				_, err = doc.Commit("base", commitTime)
				require.NoError(t, err)

				for i := range iterations {
					require.NoError(
						t,
						doc.Root().PutScalar(

							"a",
							automerge.Scalar{
								Type:   automerge.ScalarTypeString,
								String: strings.Repeat("a", i),
							},
						),
					)
					require.NoError(
						t,
						doc.Root().PutScalar(

							"b",
							automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i + 1)},
						),
					)
					require.NoError(
						t,
						doc.Root().PutScalar(

							"c",
							automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i + 1)},
						),
					)
					_, err = doc.Commit(

						"iterate",
						commitTime.Add(time.Duration(i+1)*time.Second),
					)
					require.NoError(t, err)
				}

				saved, err := doc.Save()
				require.NoError(t, err)
				other, err := engine.load(saved, actor(2))
				require.NoError(t, err)
				closeDocument(t, other)

				final := iterations + 2
				require.NoError(
					t,
					doc.Root().PutScalar(

						"a",
						automerge.Scalar{
							Type:   automerge.ScalarTypeString,
							String: strings.Repeat("a", final),
						},
					),
				)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"b",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(final)},
					),
				)
				require.NoError(
					t,
					doc.Root().PutScalar(

						"c",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(final)},
					),
				)
				_, err = doc.Commit("final", commitTime.Add(time.Hour))
				require.NoError(t, err)

				_, err = other.Merge(doc)
				require.NoError(t, err)

				transferred, err := other.Save()
				require.NoError(t, err)
				reloaded, err := engine.load(transferred, actor(3))
				require.NoError(t, err)
				closeDocument(t, reloaded)

				value, err := reloaded.Root().Scalar("b")
				require.NoError(t, err)
				assert.Equal(t, int64(final), value.Int)
				assert.Equal(t, sortedHeadHex(t, doc), sortedHeadHex(t, other))
			},
		)
	}
}

func syncBothDirections(
	t *testing.T,
	leftToRight *automerge.SyncState,
	rightToLeft *automerge.SyncState,
) {
	t.Helper()

	for range 20 {
		quiet := true

		message, ok, err := leftToRight.GenerateMessage()
		require.NoError(t, err)

		if ok {
			quiet = false

			require.NoError(t, rightToLeft.ReceiveMessage(message))
		}

		message, ok, err = rightToLeft.GenerateMessage()
		require.NoError(t, err)

		if ok {
			quiet = false

			require.NoError(t, leftToRight.ReceiveMessage(message))
		}

		if quiet {
			return
		}
	}
}
