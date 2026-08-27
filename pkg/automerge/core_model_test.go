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

package automerge_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func TestDocument_StringParity(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(125))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(125))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		require.NoError(t, document.PutString("title", "first"))
		require.NoError(t, document.PutString("title", "second"))
		_, err = document.Commit("set title", commitTime)
		require.NoError(t, err)
		value, err := document.String("title")
		require.NoError(t, err)
		assert.Equal(t, "second", value)

		values, err := document.Scalars("title")
		require.NoError(t, err)
		require.Len(t, values, 1)
	}

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceData, err := referenceDocument.Save()
	require.NoError(t, err)
	nativeFromReference, err := automerge.Load(

		referenceData,
		actor(126),
	)
	require.NoError(t, err)
	closeDocument(t, nativeFromReference)

	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(127),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)

	for _, document := range []*automerge.Document{
		nativeFromReference,
		referenceFromNative,
	} {
		value, err := document.String("title")
		require.NoError(t, err)
		assert.Equal(t, "second", value)
	}
}

func TestDocument_ConcurrentStringWinnerMatchesReference(t *testing.T) {
	t.Parallel()

	base, err := automerge.New(actor(128))
	require.NoError(t, err)
	closeDocument(t, base)
	require.NoError(t, base.PutString("title", "base"))
	_, err = base.Commit("base", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save()
	require.NoError(t, err)

	left, err := automerge.Load(baseData, actor(129))
	require.NoError(t, err)
	closeDocument(t, left)
	require.NoError(t, left.PutString("title", "left"))
	_, err = left.Commit("left", commitTime.Add(time.Second))
	require.NoError(t, err)

	right, err := automerge.Load(baseData, actor(130))
	require.NoError(t, err)
	closeDocument(t, right)
	require.NoError(t, right.PutString("title", "right"))
	_, err = right.Commit("right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	_, err = left.Merge(right)
	require.NoError(t, err)
	nativeValue, err := left.String("title")
	require.NoError(t, err)
	nativeConflicts, err := left.Scalars("title")
	require.NoError(t, err)
	require.Len(t, nativeConflicts, 2)

	merged, err := left.Save()
	require.NoError(t, err)

	reference, err := automerge.LoadReference(merged, actor(131))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceValue, err := reference.String("title")
	require.NoError(t, err)
	assert.Equal(t, referenceValue, nativeValue)

	referenceConflicts, err := reference.Scalars("title")
	require.NoError(t, err)
	assert.ElementsMatch(t, referenceConflicts, nativeConflicts)

	require.NoError(t, left.PutString("title", "resolved"))
	_, err = left.Commit("resolve conflict", commitTime.Add(3*time.Second))
	require.NoError(t, err)
	resolved, err := left.Scalars("title")
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, "resolved", resolved[0].String)
}

func TestDocument_AllScalarTypesMatchReference(t *testing.T) {
	t.Parallel()

	values := []automerge.Scalar{
		{Type: automerge.ScalarTypeNull},
		{Type: automerge.ScalarTypeBoolean, Bool: false},
		{Type: automerge.ScalarTypeBoolean, Bool: true},
		{Type: automerge.ScalarTypeUint, Uint: math.MaxUint64},
		{Type: automerge.ScalarTypeInt, Int: math.MinInt64},
		{Type: automerge.ScalarTypeFloat64, Float: math.Inf(1)},
		{Type: automerge.ScalarTypeFloat64, Float: math.NaN()},
		{Type: automerge.ScalarTypeString, String: "Hello 😀"},
		{Type: automerge.ScalarTypeBytes, Bytes: []byte{0, 1, 254, 255}},
		{Type: automerge.ScalarTypeCounter, Int: -42},
		{Type: automerge.ScalarTypeTimestamp, Int: 1_786_147_200_000},
	}

	nativeDocument, err := automerge.New(actor(132))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(132))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for index, value := range values {
		key := fmt.Sprintf("value-%d", index)
		require.NoError(t, nativeDocument.PutScalar(key, value))
		require.NoError(t, referenceDocument.PutScalar(key, value))
	}

	_, err = nativeDocument.Commit("put scalars", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit("put scalars", commitTime)
	require.NoError(t, err)

	for index, expected := range values {
		key := fmt.Sprintf("value-%d", index)
		nativeValue, err := nativeDocument.Scalar(key)
		require.NoError(t, err)
		referenceValue, err := referenceDocument.Scalar(key)
		require.NoError(t, err)
		assertScalarEqual(t, expected, nativeValue)
		assertScalarEqual(t, expected, referenceValue)
		assertScalarEqual(t, referenceValue, nativeValue)
	}

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(133),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)

	for index, expected := range values {
		value, err := referenceFromNative.Scalar(

			fmt.Sprintf("value-%d", index),
		)
		require.NoError(t, err)
		assertScalarEqual(t, expected, value)
	}
}

func TestDocument_NestedMapsAndListsMatchReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(134))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(134))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		root := document.Root()
		config, err := root.CreateObject("config", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			config.PutScalar(

				"enabled",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			),
		)

		items, err := root.CreateObject("items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			items.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "first"},
			),
		)
		nested, err := items.InsertObject(1, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			nested.PutScalar(

				"count",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
			),
		)
		require.NoError(
			t,
			items.InsertScalar(

				2,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "last"},
			),
		)
		require.NoError(t, items.DeleteIndex(0))
		require.NoError(
			t,
			items.PutScalarAt(

				1,
				automerge.Scalar{
					Type:   automerge.ScalarTypeString,
					String: "replaced",
				},
			),
		)
		require.NoError(t, config.DeleteKey("enabled"))

		_, err = document.Commit("nested values", commitTime)
		require.NoError(t, err)

		length, err := items.Len()
		require.NoError(t, err)
		assert.Equal(t, uint64(2), length)

		readNested, err := items.ObjectAt(0)
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, readNested.Type)
		count, err := readNested.Scalar("count")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count.Int)

		last, err := items.ScalarAt(1)
		require.NoError(t, err)
		assert.Equal(t, "replaced", last.String)

		_, err = config.Scalar("enabled")
		require.Error(t, err)
	}

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(135),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	items, err := referenceFromNative.Root().Object("items")
	require.NoError(t, err)
	length, err := items.Len()
	require.NoError(t, err)
	assert.Equal(t, uint64(2), length)

	nested, err := items.ObjectAt(0)
	require.NoError(t, err)
	count, err := nested.Scalar("count")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count.Int)
}

func TestDocument_LoadedObjectRemainsEditable(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(136))
	require.NoError(t, err)
	closeDocument(t, document)
	_, err = document.Root().CreateObject("items", automerge.ObjectTypeList)
	require.NoError(t, err)
	_, err = document.Commit("create list", commitTime)
	require.NoError(t, err)
	data, err := document.Save()
	require.NoError(t, err)

	loaded, err := automerge.Load(data, actor(137))
	require.NoError(t, err)
	closeDocument(t, loaded)
	items, err := loaded.Root().Object("items")
	require.NoError(t, err)
	require.NoError(
		t,
		items.InsertScalar(

			0,
			automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: 1},
		),
	)

	_, err = loaded.Commit("insert item", commitTime.Add(time.Second))
	require.NoError(t, err)
	data, err = loaded.Save()
	require.NoError(t, err)

	reference, err := automerge.LoadReference(data, actor(138))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceItems, err := reference.Root().Object("items")
	require.NoError(t, err)
	value, err := referenceItems.ScalarAt(0)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), value.Uint)
}

func TestDocument_CountersMatchReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(139))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(139))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		root := document.Root()
		require.NoError(
			t,
			root.PutScalar(

				"counter",
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
			),
		)
		require.NoError(
			t,
			root.PutScalar(

				"integer",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 5},
			),
		)
		list, err := root.CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 10},
			),
		)
		require.NoError(
			t,
			list.InsertScalar(

				1,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 10},
			),
		)

		_, err = document.Commit("create counters", commitTime)
		require.NoError(t, err)

		require.Error(t, root.Increment("integer", 1))
		require.Error(t, list.IncrementAt(1, 1))
		require.NoError(t, root.Increment("counter", 3))
		require.NoError(t, root.Increment("counter", -2))
		require.NoError(t, list.IncrementAt(0, -4))

		_, err = document.Commit(

			"increment counters",
			commitTime.Add(time.Second),
		)
		require.NoError(t, err)

		counter, err := root.Scalar("counter")
		require.NoError(t, err)
		assert.Equal(t, int64(6), counter.Int)

		listCounter, err := list.ScalarAt(0)
		require.NoError(t, err)
		assert.Equal(t, int64(6), listCounter.Int)
	}

	base, err := automerge.New(actor(140))
	require.NoError(t, err)
	closeDocument(t, base)
	require.NoError(
		t,
		base.Root().PutScalar(

			"counter",
			automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
		),
	)
	_, err = base.Commit("base counter", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save()
	require.NoError(t, err)

	left, err := automerge.Load(baseData, actor(141))
	require.NoError(t, err)
	closeDocument(t, left)
	require.NoError(t, left.Root().Increment("counter", 2))
	_, err = left.Commit("left increment", commitTime.Add(time.Second))
	require.NoError(t, err)
	right, err := automerge.Load(baseData, actor(142))
	require.NoError(t, err)
	closeDocument(t, right)
	require.NoError(t, right.Root().Increment("counter", 3))
	_, err = right.Commit("right increment", commitTime.Add(time.Second))
	require.NoError(t, err)

	_, err = left.Merge(right)
	require.NoError(t, err)
	value, err := left.Root().Scalar("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), value.Int)

	data, err := left.Save()
	require.NoError(t, err)
	reference, err := automerge.LoadReference(data, actor(143))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceValue, err := reference.Root().Scalar("counter")
	require.NoError(t, err)
	assert.Equal(t, value.Int, referenceValue.Int)
}

func TestDocument_CounterDeletionMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(160))
				require.NoError(t, err)
				closeDocument(t, document)
				root := document.Root()
				require.NoError(
					t,
					root.PutScalar(

						"counter",
						automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 1},
					),
				)
				list, err := root.CreateObject("list", automerge.ObjectTypeList)
				require.NoError(t, err)
				require.NoError(
					t,
					list.InsertScalar(

						0,
						automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 1},
					),
				)

				_, err = document.Commit("counters", commitTime)
				require.NoError(t, err)

				require.NoError(t, root.DeleteKey("counter"))
				require.NoError(t, list.DeleteIndex(0))
			},
		)
	}
}

func TestDocument_RandomListParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 10
		steps     = 100
	)

	for history := range histories {
		random := rand.New(rand.NewSource(int64(20_000 + history)))
		actorID := actor(byte(150 + history))
		nativeDocument, err := automerge.New(actorID)
		require.NoError(t, err)
		closeDocument(t, nativeDocument)

		referenceDocument, err := automerge.NewReference(actorID)
		require.NoError(t, err)
		closeDocument(t, referenceDocument)

		nativeList, err := nativeDocument.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		referenceList, err := referenceDocument.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		nativeHash, err := nativeDocument.Commit("create list", commitTime)
		require.NoError(t, err)
		referenceHash, err := referenceDocument.Commit(

			"create list",
			commitTime,
		)
		require.NoError(t, err)
		assert.Equal(t, referenceHash, nativeHash)

		var model []int64

		for step := range steps {
			switch {
			case len(model) == 0 || random.Intn(3) == 0:
				index := random.Intn(len(model) + 1)
				value := random.Int63()

				model = append(model, 0)
				copy(model[index+1:], model[index:])
				model[index] = value
				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  value,
				}
				require.NoError(
					t,
					nativeList.InsertScalar(uint64(index), scalar),
				)
				require.NoError(
					t,
					referenceList.InsertScalar(uint64(index), scalar),
				)
			case random.Intn(2) == 0:
				index := random.Intn(len(model))
				model = append(model[:index], model[index+1:]...)
				require.NoError(
					t,
					nativeList.DeleteIndex(uint64(index)),
				)
				require.NoError(
					t,
					referenceList.DeleteIndex(uint64(index)),
				)
			default:
				index := random.Intn(len(model))
				value := random.Int63()
				model[index] = value
				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  value,
				}
				require.NoError(
					t,
					nativeList.PutScalarAt(uint64(index), scalar),
				)
				require.NoError(
					t,
					referenceList.PutScalarAt(uint64(index), scalar),
				)
			}

			message := fmt.Sprintf("history %d step %d", history, step)
			timestamp := commitTime.Add(time.Duration(step+1) * time.Second)
			nativeHash, err = nativeDocument.Commit(message, timestamp)
			require.NoError(t, err)
			referenceHash, err = referenceDocument.Commit(

				message,
				timestamp,
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				referenceHash,
				nativeHash,
				"history %d step %d",
				history,
				step,
			)

			nativeLength, err := nativeList.Len()
			require.NoError(t, err)
			referenceLength, err := referenceList.Len()
			require.NoError(t, err)
			assert.Equal(t, uint64(len(model)), nativeLength)
			assert.Equal(t, referenceLength, nativeLength)

			for index, expected := range model {
				nativeValue, err := nativeList.ScalarAt(uint64(index))
				require.NoError(t, err)
				referenceValue, err := referenceList.ScalarAt(

					uint64(index),
				)
				require.NoError(t, err)
				assert.Equal(t, expected, nativeValue.Int)
				assertScalarEqual(t, referenceValue, nativeValue)
			}
		}
	}
}

func TestDocument_RandomMapParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 10
		steps     = 100
	)

	for history := range histories {
		random := rand.New(rand.NewSource(int64(30_000 + history)))
		actorID := actor(byte(170 + history))
		nativeDocument, err := automerge.New(actorID)
		require.NoError(t, err)
		closeDocument(t, nativeDocument)

		referenceDocument, err := automerge.NewReference(actorID)
		require.NoError(t, err)
		closeDocument(t, referenceDocument)

		nativeMap, err := nativeDocument.Root().CreateObject(

			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)
		referenceMap, err := referenceDocument.Root().CreateObject(

			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)
		nativeHash, err := nativeDocument.Commit("create map", commitTime)
		require.NoError(t, err)
		referenceHash, err := referenceDocument.Commit(

			"create map",
			commitTime,
		)
		require.NoError(t, err)
		assert.Equal(t, referenceHash, nativeHash)

		model := make(map[string]int64)
		keys := []string{"", "a", "b", "c", "d", "e"}

		for step := range steps {
			key := keys[random.Intn(len(keys))]

			operation := "put"
			if _, exists := model[key]; exists && random.Intn(3) == 0 {
				operation = "delete"

				delete(model, key)
				require.NoError(t, nativeMap.DeleteKey(key))
				require.NoError(t, referenceMap.DeleteKey(key))
			} else {
				value := random.Int63()
				model[key] = value
				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  value,
				}
				require.NoError(t, nativeMap.PutScalar(key, scalar))
				require.NoError(t, referenceMap.PutScalar(key, scalar))
			}

			message := fmt.Sprintf("history %d step %d", history, step)
			timestamp := commitTime.Add(time.Duration(step+1) * time.Second)
			nativeHash, err = nativeDocument.Commit(message, timestamp)
			require.NoError(t, err)
			referenceHash, err = referenceDocument.Commit(

				message,
				timestamp,
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				referenceHash,
				nativeHash,
				"history %d step %d %s key %q",
				history,
				step,
				operation,
				key,
			)

			for _, candidate := range keys {
				expected, exists := model[candidate]
				nativeValue, nativeErr := nativeMap.Scalar(candidate)
				referenceValue, referenceErr := referenceMap.Scalar(

					candidate,
				)

				if !exists {
					require.Error(t, nativeErr)
					require.Error(t, referenceErr)

					continue
				}

				require.NoError(t, nativeErr)
				require.NoError(t, referenceErr)
				assert.Equal(t, expected, nativeValue.Int)
				assertScalarEqual(t, referenceValue, nativeValue)
			}
		}
	}
}

func TestDocument_RollbackMatchesReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(180))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(180))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		require.NoError(t, document.PutString("value", "committed"))
		_, err = document.Commit("initial", commitTime)
		require.NoError(t, err)
		headsBefore, err := document.Heads()
		require.NoError(t, err)

		require.NoError(t, document.PutString("value", "rolled back"))
		list, err := document.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
			),
		)

		cancelled, err := document.Rollback()
		require.NoError(t, err)
		assert.Equal(t, uint64(3), cancelled)

		value, err := document.String("value")
		require.NoError(t, err)
		assert.Equal(t, "committed", value)

		_, err = document.Root().Object("list")
		require.Error(t, err)
		headsAfter, err := document.Heads()
		require.NoError(t, err)
		assert.Equal(t, headsBefore, headsAfter)

		cancelled, err = document.Rollback()
		require.NoError(t, err)
		assert.Zero(t, cancelled)
	}
}

func TestDocument_ForkMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(161))
				require.NoError(t, err)
				closeDocument(t, document)
				require.NoError(t, document.PutString("base", "value"))
				_, err = document.Commit("base", commitTime)
				require.NoError(t, err)

				fork, err := document.Fork(actor(162))
				require.NoError(t, err)
				closeDocument(t, fork)
				require.NoError(t, fork.PutString("fork", "value"))
				_, err = fork.Commit(

					"fork",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)

				_, err = document.String("fork")
				require.Error(t, err)
				value, err := fork.String("base")
				require.NoError(t, err)
				assert.Equal(t, "value", value)

				_, err = document.Merge(fork)
				require.NoError(t, err)
				value, err = document.String("fork")
				require.NoError(t, err)
				assert.Equal(t, "value", value)
			},
		)
	}
}

func TestDocument_WrongObjectOperationsMatchReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(171))
				require.NoError(t, err)
				closeDocument(t, document)
				root := document.Root()
				mapObject, err := root.CreateObject(

					"map",
					automerge.ObjectTypeMap,
				)
				require.NoError(t, err)
				listObject, err := root.CreateObject(

					"list",
					automerge.ObjectTypeList,
				)
				require.NoError(t, err)
				textObject, err := root.CreateObject(

					"text",
					automerge.ObjectTypeText,
				)
				require.NoError(t, err)

				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  1,
				}
				require.Error(t, listObject.PutScalar("key", scalar))
				require.Error(t, textObject.PutScalar("key", scalar))
				require.Error(t, mapObject.InsertScalar(0, scalar))
				require.Error(t, mapObject.DeleteIndex(0))
				length, err := mapObject.Len()
				require.NoError(t, err)
				assert.Zero(t, length)

				_, err = listObject.Text()
				require.Error(t, err)
			},
		)
	}
}

func TestDocument_DeletedObjectsSaveLoad(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(172))
	require.NoError(t, err)
	closeDocument(t, document)

	root := document.Root()
	for key, objectType := range map[string]automerge.ObjectType{
		"list":  automerge.ObjectTypeList,
		"text":  automerge.ObjectTypeText,
		"map":   automerge.ObjectTypeMap,
		"table": automerge.ObjectTypeTable,
	} {
		_, err := root.CreateObject(key, objectType)
		require.NoError(t, err)
		require.NoError(t, root.DeleteKey(key))
	}

	_, err = document.Commit("deleted objects", commitTime)
	require.NoError(t, err)
	data, err := document.Save()
	require.NoError(t, err)

	nativeLoaded, err := automerge.Load(data, actor(173))
	require.NoError(t, err)
	closeDocument(t, nativeLoaded)

	referenceLoaded, err := automerge.LoadReference(data, actor(174))
	require.NoError(t, err)
	closeDocument(t, referenceLoaded)

	for _, loaded := range []*automerge.Document{
		nativeLoaded,
		referenceLoaded,
	} {
		length, err := loaded.Root().Len()
		require.NoError(t, err)
		assert.Zero(t, length)
	}
}

func TestDocument_ManyMapDeletes(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(175))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(175))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		object, err := document.Root().CreateObject(

			"object",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)

		for index := range 100 {
			key := fmt.Sprintf("%d", index)
			require.NoError(
				t,
				object.PutScalar(

					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
			require.NoError(t, object.DeleteKey(key))
		}

		_, err = document.Commit("many deletes", commitTime)
		require.NoError(t, err)
		length, err := object.Len()
		require.NoError(t, err)
		assert.Zero(t, length)
	}
}

func TestDocument_MapKeysMatchReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(179))
				require.NoError(t, err)
				closeDocument(t, document)
				keys, err := document.Root().Keys()
				require.NoError(t, err)
				assert.Empty(t, keys)

				object, err := document.Root().CreateObject(

					"map",
					automerge.ObjectTypeMap,
				)
				require.NoError(t, err)

				for _, key := range []string{"z", "", "a@b", "a"} {
					require.NoError(
						t,
						object.PutScalar(

							key,
							automerge.Scalar{
								Type: automerge.ScalarTypeInt,
								Int:  1,
							},
						),
					)
				}

				require.NoError(t, object.DeleteKey("z"))
				keys, err = object.Keys()
				require.NoError(t, err)
				assert.Equal(t, []string{"", "a", "a@b"}, keys)

				_, err = document.Commit("keys", commitTime)
				require.NoError(t, err)
				loaded, err := document.Fork(actor(180))
				require.NoError(t, err)
				closeDocument(t, loaded)
				loadedObject, err := loaded.Root().Object("map")
				require.NoError(t, err)
				keys, err = loadedObject.Keys()
				require.NoError(t, err)
				assert.Equal(t, []string{"", "a", "a@b"}, keys)
			},
		)
	}
}

func TestDocument_NoOpMergeAndEqualPutMatchReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(176))
				require.NoError(t, err)
				closeDocument(t, document)
				require.NoError(t, document.PutString("value", "same"))
				_, err = document.Commit("initial", commitTime)
				require.NoError(t, err)

				fork, err := document.Fork(actor(177))
				require.NoError(t, err)
				closeDocument(t, fork)
				_, err = fork.EmptyCommit(

					"noop",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)
				require.NoError(t, fork.PutString("value", "changed"))
				_, err = fork.Commit(

					"real",
					commitTime.Add(2*time.Second),
				)
				require.NoError(t, err)

				require.NoError(t, document.PutString("value", "same"))
				_, err = document.Commit(

					"equal",
					commitTime.Add(time.Second),
				)
				require.Error(t, err)
				_, err = document.Merge(fork)
				require.NoError(t, err)
				value, err := document.String("value")
				require.NoError(t, err)
				assert.Equal(t, "changed", value)

				data, err := document.Save()
				require.NoError(t, err)
				loaded, err := automerge.Load(data, actor(178))
				require.NoError(t, err)
				closeDocument(t, loaded)
				value, err = loaded.String("value")
				require.NoError(t, err)
				assert.Equal(t, "changed", value)
			},
		)
	}
}

func TestDocument_RandomConcurrentMapParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 20
		steps     = 30
	)

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	for history := range histories {
		base, err := automerge.New(actor(byte(190 + history)))
		require.NoError(t, err)
		closeDocument(t, base)
		baseMap, err := base.Root().CreateObject(

			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)

		for index, key := range keys {
			require.NoError(
				t,
				baseMap.PutScalar(

					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
		}

		_, err = base.Commit("base map", commitTime)
		require.NoError(t, err)
		baseData, err := base.Save()
		require.NoError(t, err)

		left, err := automerge.Load(baseData, actor(byte(210+history)))
		require.NoError(t, err)
		closeDocument(t, left)
		leftMap, err := left.Root().Object("map")
		require.NoError(t, err)
		right, err := automerge.Load(baseData, actor(byte(230+history)))
		require.NoError(t, err)
		closeDocument(t, right)
		rightMap, err := right.Root().Object("map")
		require.NoError(t, err)
		referenceLeft, err := automerge.LoadReference(

			baseData,
			actor(byte(210+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceLeft)
		referenceLeftMap, err := referenceLeft.Root().Object("map")
		require.NoError(t, err)
		referenceRight, err := automerge.LoadReference(

			baseData,
			actor(byte(230+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceRight)
		referenceRightMap, err := referenceRight.Root().Object("map")
		require.NoError(t, err)

		leftSeed := int64(40_000 + history)
		mutateRandomMap(
			t,
			rand.New(rand.NewSource(leftSeed)),
			left,
			leftMap,
			keys,
			steps,
			"left",
		)
		mutateRandomMap(
			t,
			rand.New(rand.NewSource(leftSeed)),
			referenceLeft,
			referenceLeftMap,
			keys,
			steps,
			"left",
		)

		rightSeed := int64(50_000 + history)
		mutateRandomMap(
			t,
			rand.New(rand.NewSource(rightSeed)),
			right,
			rightMap,
			keys,
			steps,
			"right",
		)
		mutateRandomMap(
			t,
			rand.New(rand.NewSource(rightSeed)),
			referenceRight,
			referenceRightMap,
			keys,
			steps,
			"right",
		)

		_, err = left.Merge(right)
		require.NoError(t, err)
		_, err = right.Merge(left)
		require.NoError(t, err)
		_, err = referenceLeft.Merge(referenceRight)
		require.NoError(t, err)
		_, err = referenceRight.Merge(referenceLeft)
		require.NoError(t, err)
		leftMap, err = left.Root().Object("map")
		require.NoError(t, err)
		rightMap, err = right.Root().Object("map")
		require.NoError(t, err)
		assertMapParity(t, keys, leftMap, rightMap)

		referenceLeftMap, err = referenceLeft.Root().Object("map")
		require.NoError(t, err)
		referenceRightMap, err = referenceRight.Root().Object("map")
		require.NoError(t, err)
		assertMapParity(t, keys, referenceLeftMap, referenceRightMap)
		assertMapParity(t, keys, leftMap, referenceLeftMap)
	}
}

func TestDocument_RandomConcurrentListParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 20
		steps     = 30
	)

	for history := range histories {
		base, err := automerge.New(actor(byte(10 + history)))
		require.NoError(t, err)
		closeDocument(t, base)
		baseList, err := base.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)

		for index := range 8 {
			require.NoError(
				t,
				baseList.InsertScalar(

					uint64(index),
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
		}

		_, err = base.Commit("base list", commitTime)
		require.NoError(t, err)
		baseData, err := base.Save()
		require.NoError(t, err)

		left, err := automerge.Load(baseData, actor(byte(30+history)))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object("list")
		require.NoError(t, err)
		right, err := automerge.Load(baseData, actor(byte(50+history)))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object("list")
		require.NoError(t, err)
		referenceLeft, err := automerge.LoadReference(

			baseData,
			actor(byte(30+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceLeft)
		referenceLeftList, err := referenceLeft.Root().Object("list")
		require.NoError(t, err)
		referenceRight, err := automerge.LoadReference(

			baseData,
			actor(byte(50+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceRight)
		referenceRightList, err := referenceRight.Root().Object("list")
		require.NoError(t, err)

		leftSeed := int64(60_000 + history)
		mutateRandomList(
			t,
			rand.New(rand.NewSource(leftSeed)),
			left,
			leftList,
			steps,
			"left",
		)
		mutateRandomList(
			t,
			rand.New(rand.NewSource(leftSeed)),
			referenceLeft,
			referenceLeftList,
			steps,
			"left",
		)

		rightSeed := int64(70_000 + history)
		mutateRandomList(
			t,
			rand.New(rand.NewSource(rightSeed)),
			right,
			rightList,
			steps,
			"right",
		)
		mutateRandomList(
			t,
			rand.New(rand.NewSource(rightSeed)),
			referenceRight,
			referenceRightList,
			steps,
			"right",
		)

		_, err = left.Merge(right)
		require.NoError(t, err)
		_, err = right.Merge(left)
		require.NoError(t, err)
		_, err = referenceLeft.Merge(referenceRight)
		require.NoError(t, err)
		_, err = referenceRight.Merge(referenceLeft)
		require.NoError(t, err)
		leftList, err = left.Root().Object("list")
		require.NoError(t, err)
		rightList, err = right.Root().Object("list")
		require.NoError(t, err)
		assertListParity(t, leftList, rightList)

		referenceLeftList, err = referenceLeft.Root().Object("list")
		require.NoError(t, err)
		referenceRightList, err = referenceRight.Root().Object("list")
		require.NoError(t, err)
		assertListParity(t, referenceLeftList, referenceRightList)
		assertListParity(t, leftList, referenceLeftList)
	}
}

func TestDocument_ConcurrentListOrderingMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	results := make(map[string][]string)

	for name, factory := range factories {
		base, err := factory(actor(80))
		require.NoError(t, err)
		closeDocument(t, base)
		list, err := base.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(

				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "A"},
			),
		)

		_, err = base.Commit("base", commitTime)
		require.NoError(t, err)

		left, err := base.Fork(actor(81))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			leftList.InsertValues(

				1,
				[]automerge.Value{
					hydratedString("L1"),
					hydratedString("L2"),
				},
			),
		)

		_, err = left.Commit("left", commitTime.Add(time.Second))
		require.NoError(t, err)

		right, err := base.Fork(actor(82))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			rightList.InsertValues(

				1,
				[]automerge.Value{
					hydratedString("R1"),
					hydratedString("R2"),
				},
			),
		)

		_, err = right.Commit("right", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = left.Merge(right)
		require.NoError(t, err)
		leftList, err = left.Root().Object("list")
		require.NoError(t, err)
		values := listStrings(t, leftList)
		assert.Equal(t, []string{"A", "R1", "R2", "L1", "L2"}, values)
		results[name] = values
	}

	assert.Equal(t, results["reference"], results["native"])
}

func TestDocument_InsertAfterConcurrentDeleteMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	results := make(map[string][]string)

	for name, factory := range factories {
		base, err := factory(actor(83))
		require.NoError(t, err)
		closeDocument(t, base)
		list, err := base.Root().CreateObject(

			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertValues(

				0,
				[]automerge.Value{
					hydratedString("A"),
					hydratedString("B"),
				},
			),
		)

		_, err = base.Commit("base", commitTime)
		require.NoError(t, err)

		left, err := base.Fork(actor(84))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object("list")
		require.NoError(t, err)
		require.NoError(t, leftList.DeleteIndex(0))

		_, err = left.Commit("delete", commitTime.Add(time.Second))
		require.NoError(t, err)

		right, err := base.Fork(actor(85))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object("list")
		require.NoError(t, err)
		require.NoError(
			t,
			rightList.InsertScalar(

				1,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "X"},
			),
		)

		_, err = right.Commit("insert", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = left.Merge(right)
		require.NoError(t, err)
		leftList, err = left.Root().Object("list")
		require.NoError(t, err)
		results[name] = listStrings(t, leftList)
	}

	assert.Equal(t, results["reference"], results["native"])
	assert.Equal(t, []string{"X", "B"}, results["native"])
}

func hydratedString(value string) automerge.Value {
	return automerge.Value{
		Type: automerge.ValueTypeScalar,
		Scalar: automerge.Scalar{
			Type:   automerge.ScalarTypeString,
			String: value,
		},
	}
}

func listStrings(
	t *testing.T,
	list *automerge.Object,
) []string {
	t.Helper()

	length, err := list.Len()
	require.NoError(t, err)

	values := make([]string, length)
	for index := range length {
		value, err := list.ScalarAt(index)
		require.NoError(t, err)

		values[index] = value.String
	}

	return values
}

func mutateRandomMap(
	t *testing.T,
	random *rand.Rand,
	document *automerge.Document,
	object *automerge.Object,
	keys []string,
	steps int,
	side string,
) {
	t.Helper()

	present := make(map[string]bool, len(keys))
	for _, key := range keys {
		present[key] = true
	}

	for step := range steps {
		key := keys[random.Intn(len(keys))]
		if present[key] && random.Intn(4) == 0 {
			require.NoError(t, object.DeleteKey(key))
			present[key] = false
		} else {
			require.NoError(
				t,
				object.PutScalar(

					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  random.Int63(),
					},
				),
			)
			present[key] = true
		}

		if step%5 == 4 {
			_, err := document.Commit(

				fmt.Sprintf("%s map step %d", side, step),
				commitTime.Add(time.Duration(step+1)*time.Second),
			)
			require.NoError(t, err)
		}
	}
}

func mutateRandomList(
	t *testing.T,
	random *rand.Rand,
	document *automerge.Document,
	object *automerge.Object,
	steps int,
	side string,
) {
	t.Helper()

	length := 8

	for step := range steps {
		switch {
		case length == 0 || random.Intn(3) == 0:
			index := random.Intn(length + 1)
			require.NoError(
				t,
				object.InsertScalar(

					uint64(index),
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  random.Int63(),
					},
				),
			)

			length++
		case random.Intn(2) == 0:
			index := random.Intn(length)
			require.NoError(t, object.DeleteIndex(uint64(index)))

			length--
		default:
			index := random.Intn(length)
			require.NoError(
				t,
				object.PutScalarAt(

					uint64(index),
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  random.Int63(),
					},
				),
			)
		}

		if step%5 == 4 {
			_, err := document.Commit(

				fmt.Sprintf("%s list step %d", side, step),
				commitTime.Add(time.Duration(step+1)*time.Second),
			)
			require.NoError(t, err)
		}
	}
}

func assertMapParity(
	t *testing.T,
	keys []string,
	left *automerge.Object,
	right *automerge.Object,
) {
	t.Helper()

	for _, key := range keys {
		leftValues, leftErr := left.Scalars(key)
		rightValues, rightErr := right.Scalars(key)
		assert.Equal(t, leftErr != nil, rightErr != nil, "key %q", key)

		if leftErr == nil && rightErr == nil {
			assert.ElementsMatch(t, leftValues, rightValues, "key %q", key)
		}
	}
}

func assertListParity(
	t *testing.T,
	left *automerge.Object,
	right *automerge.Object,
) {
	t.Helper()

	leftLength, err := left.Len()
	require.NoError(t, err)
	rightLength, err := right.Len()
	require.NoError(t, err)
	require.Equal(t, leftLength, rightLength)

	for index := range leftLength {
		leftValue, err := left.ScalarAt(index)
		require.NoError(t, err)
		rightValue, err := right.ScalarAt(index)
		require.NoError(t, err)
		assertScalarEqual(t, leftValue, rightValue)
	}
}

func assertScalarEqual(
	t *testing.T,
	expected automerge.Scalar,
	actual automerge.Scalar,
) {
	t.Helper()

	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Bool, actual.Bool)
	assert.Equal(t, expected.Uint, actual.Uint)
	assert.Equal(t, expected.Int, actual.Int)
	assert.Equal(t, math.Float64bits(expected.Float), math.Float64bits(actual.Float))
	assert.Equal(t, expected.String, actual.String)
	assert.Equal(t, expected.Bytes, actual.Bytes)
}
