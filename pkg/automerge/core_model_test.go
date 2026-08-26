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
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestDocument_StringParity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(125))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(125))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		require.NoError(t, document.PutString(ctx, "title", "first"))
		require.NoError(t, document.PutString(ctx, "title", "second"))
		_, err = document.Commit(ctx, "set title", commitTime)
		require.NoError(t, err)
		value, err := document.String(ctx, "title")
		require.NoError(t, err)
		assert.Equal(t, "second", value)

		values, err := document.Scalars(ctx, "title")
		require.NoError(t, err)
		require.Len(t, values, 1)
	}

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceData, err := referenceDocument.Save(ctx)
	require.NoError(t, err)
	nativeFromReference, err := automerge.Load(
		ctx,
		referenceData,
		actor(126),
	)
	require.NoError(t, err)
	closeDocument(t, nativeFromReference)

	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(127),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)

	for _, document := range []*automerge.Document{
		nativeFromReference,
		referenceFromNative,
	} {
		value, err := document.String(ctx, "title")
		require.NoError(t, err)
		assert.Equal(t, "second", value)
	}
}

func TestDocument_ConcurrentStringWinnerMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := automerge.New(ctx, actor(128))
	require.NoError(t, err)
	closeDocument(t, base)
	require.NoError(t, base.PutString(ctx, "title", "base"))
	_, err = base.Commit(ctx, "base", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	left, err := automerge.Load(ctx, baseData, actor(129))
	require.NoError(t, err)
	closeDocument(t, left)
	require.NoError(t, left.PutString(ctx, "title", "left"))
	_, err = left.Commit(ctx, "left", commitTime.Add(time.Second))
	require.NoError(t, err)

	right, err := automerge.Load(ctx, baseData, actor(130))
	require.NoError(t, err)
	closeDocument(t, right)
	require.NoError(t, right.PutString(ctx, "title", "right"))
	_, err = right.Commit(ctx, "right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	_, err = left.Merge(ctx, right)
	require.NoError(t, err)
	nativeValue, err := left.String(ctx, "title")
	require.NoError(t, err)
	nativeConflicts, err := left.Scalars(ctx, "title")
	require.NoError(t, err)
	require.Len(t, nativeConflicts, 2)

	merged, err := left.Save(ctx)
	require.NoError(t, err)

	reference, err := automerge.LoadReference(ctx, merged, actor(131))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceValue, err := reference.String(ctx, "title")
	require.NoError(t, err)
	assert.Equal(t, referenceValue, nativeValue)

	referenceConflicts, err := reference.Scalars(ctx, "title")
	require.NoError(t, err)
	assert.ElementsMatch(t, referenceConflicts, nativeConflicts)

	require.NoError(t, left.PutString(ctx, "title", "resolved"))
	_, err = left.Commit(ctx, "resolve conflict", commitTime.Add(3*time.Second))
	require.NoError(t, err)
	resolved, err := left.Scalars(ctx, "title")
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, "resolved", resolved[0].String)
}

func TestDocument_AllScalarTypesMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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

	nativeDocument, err := automerge.New(ctx, actor(132))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(132))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for index, value := range values {
		key := fmt.Sprintf("value-%d", index)
		require.NoError(t, nativeDocument.PutScalar(ctx, key, value))
		require.NoError(t, referenceDocument.PutScalar(ctx, key, value))
	}

	_, err = nativeDocument.Commit(ctx, "put scalars", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit(ctx, "put scalars", commitTime)
	require.NoError(t, err)

	for index, expected := range values {
		key := fmt.Sprintf("value-%d", index)
		nativeValue, err := nativeDocument.Scalar(ctx, key)
		require.NoError(t, err)
		referenceValue, err := referenceDocument.Scalar(ctx, key)
		require.NoError(t, err)
		assertScalarEqual(t, expected, nativeValue)
		assertScalarEqual(t, expected, referenceValue)
		assertScalarEqual(t, referenceValue, nativeValue)
	}

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(133),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)

	for index, expected := range values {
		value, err := referenceFromNative.Scalar(
			ctx,
			fmt.Sprintf("value-%d", index),
		)
		require.NoError(t, err)
		assertScalarEqual(t, expected, value)
	}
}

func TestDocument_NestedMapsAndListsMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(134))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(134))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		root := document.Root()
		config, err := root.CreateObject(ctx, "config", automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			config.PutScalar(
				ctx,
				"enabled",
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
			),
		)

		items, err := root.CreateObject(ctx, "items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			items.InsertScalar(
				ctx,
				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "first"},
			),
		)
		nested, err := items.InsertObject(ctx, 1, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(
			t,
			nested.PutScalar(
				ctx,
				"count",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
			),
		)
		require.NoError(
			t,
			items.InsertScalar(
				ctx,
				2,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "last"},
			),
		)
		require.NoError(t, items.DeleteIndex(ctx, 0))
		require.NoError(
			t,
			items.PutScalarAt(
				ctx,
				1,
				automerge.Scalar{
					Type:   automerge.ScalarTypeString,
					String: "replaced",
				},
			),
		)
		require.NoError(t, config.DeleteKey(ctx, "enabled"))

		_, err = document.Commit(ctx, "nested values", commitTime)
		require.NoError(t, err)

		length, err := items.Len(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), length)

		readNested, err := items.ObjectAt(ctx, 0)
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeMap, readNested.Type)
		count, err := readNested.Scalar(ctx, "count")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count.Int)

		last, err := items.ScalarAt(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "replaced", last.String)

		_, err = config.Scalar(ctx, "enabled")
		require.Error(t, err)
	}

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(135),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	items, err := referenceFromNative.Root().Object(ctx, "items")
	require.NoError(t, err)
	length, err := items.Len(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), length)

	nested, err := items.ObjectAt(ctx, 0)
	require.NoError(t, err)
	count, err := nested.Scalar(ctx, "count")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count.Int)
}

func TestDocument_LoadedObjectRemainsEditable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(136))
	require.NoError(t, err)
	closeDocument(t, document)
	_, err = document.Root().CreateObject(ctx, "items", automerge.ObjectTypeList)
	require.NoError(t, err)
	_, err = document.Commit(ctx, "create list", commitTime)
	require.NoError(t, err)
	data, err := document.Save(ctx)
	require.NoError(t, err)

	loaded, err := automerge.Load(ctx, data, actor(137))
	require.NoError(t, err)
	closeDocument(t, loaded)
	items, err := loaded.Root().Object(ctx, "items")
	require.NoError(t, err)
	require.NoError(
		t,
		items.InsertScalar(
			ctx,
			0,
			automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: 1},
		),
	)
	_, err = loaded.Commit(ctx, "insert item", commitTime.Add(time.Second))
	require.NoError(t, err)
	data, err = loaded.Save(ctx)
	require.NoError(t, err)

	reference, err := automerge.LoadReference(ctx, data, actor(138))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceItems, err := reference.Root().Object(ctx, "items")
	require.NoError(t, err)
	value, err := referenceItems.ScalarAt(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), value.Uint)
}

func TestDocument_CountersMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(139))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(139))
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
				ctx,
				"counter",
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
			),
		)
		require.NoError(
			t,
			root.PutScalar(
				ctx,
				"integer",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 5},
			),
		)
		list, err := root.CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(
				ctx,
				0,
				automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 10},
			),
		)
		require.NoError(
			t,
			list.InsertScalar(
				ctx,
				1,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 10},
			),
		)
		_, err = document.Commit(ctx, "create counters", commitTime)
		require.NoError(t, err)

		require.Error(t, root.Increment(ctx, "integer", 1))
		require.Error(t, list.IncrementAt(ctx, 1, 1))
		require.NoError(t, root.Increment(ctx, "counter", 3))
		require.NoError(t, root.Increment(ctx, "counter", -2))
		require.NoError(t, list.IncrementAt(ctx, 0, -4))
		_, err = document.Commit(
			ctx,
			"increment counters",
			commitTime.Add(time.Second),
		)
		require.NoError(t, err)

		counter, err := root.Scalar(ctx, "counter")
		require.NoError(t, err)
		assert.Equal(t, int64(6), counter.Int)

		listCounter, err := list.ScalarAt(ctx, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(6), listCounter.Int)
	}

	base, err := automerge.New(ctx, actor(140))
	require.NoError(t, err)
	closeDocument(t, base)
	require.NoError(
		t,
		base.Root().PutScalar(
			ctx,
			"counter",
			automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
		),
	)
	_, err = base.Commit(ctx, "base counter", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	left, err := automerge.Load(ctx, baseData, actor(141))
	require.NoError(t, err)
	closeDocument(t, left)
	require.NoError(t, left.Root().Increment(ctx, "counter", 2))
	_, err = left.Commit(ctx, "left increment", commitTime.Add(time.Second))
	require.NoError(t, err)
	right, err := automerge.Load(ctx, baseData, actor(142))
	require.NoError(t, err)
	closeDocument(t, right)
	require.NoError(t, right.Root().Increment(ctx, "counter", 3))
	_, err = right.Commit(ctx, "right increment", commitTime.Add(time.Second))
	require.NoError(t, err)

	_, err = left.Merge(ctx, right)
	require.NoError(t, err)
	value, err := left.Root().Scalar(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), value.Int)

	data, err := left.Save(ctx)
	require.NoError(t, err)
	reference, err := automerge.LoadReference(ctx, data, actor(143))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceValue, err := reference.Root().Scalar(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, value.Int, referenceValue.Int)
}

func TestDocument_CounterDeletionMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
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

				document, err := factory(ctx, actor(160))
				require.NoError(t, err)
				closeDocument(t, document)
				root := document.Root()
				require.NoError(
					t,
					root.PutScalar(
						ctx,
						"counter",
						automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 1},
					),
				)
				list, err := root.CreateObject(ctx, "list", automerge.ObjectTypeList)
				require.NoError(t, err)
				require.NoError(
					t,
					list.InsertScalar(
						ctx,
						0,
						automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 1},
					),
				)
				_, err = document.Commit(ctx, "counters", commitTime)
				require.NoError(t, err)

				require.NoError(t, root.DeleteKey(ctx, "counter"))
				require.NoError(t, list.DeleteIndex(ctx, 0))
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

	ctx := context.Background()

	for history := range histories {
		random := rand.New(rand.NewSource(int64(20_000 + history)))
		actorID := actor(byte(150 + history))
		nativeDocument, err := automerge.New(ctx, actorID)
		require.NoError(t, err)
		closeDocument(t, nativeDocument)

		referenceDocument, err := automerge.NewReference(ctx, actorID)
		require.NoError(t, err)
		closeDocument(t, referenceDocument)

		nativeList, err := nativeDocument.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		referenceList, err := referenceDocument.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		nativeHash, err := nativeDocument.Commit(ctx, "create list", commitTime)
		require.NoError(t, err)
		referenceHash, err := referenceDocument.Commit(
			ctx,
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
					nativeList.InsertScalar(ctx, uint64(index), scalar),
				)
				require.NoError(
					t,
					referenceList.InsertScalar(ctx, uint64(index), scalar),
				)
			case random.Intn(2) == 0:
				index := random.Intn(len(model))
				model = append(model[:index], model[index+1:]...)
				require.NoError(
					t,
					nativeList.DeleteIndex(ctx, uint64(index)),
				)
				require.NoError(
					t,
					referenceList.DeleteIndex(ctx, uint64(index)),
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
					nativeList.PutScalarAt(ctx, uint64(index), scalar),
				)
				require.NoError(
					t,
					referenceList.PutScalarAt(ctx, uint64(index), scalar),
				)
			}

			message := fmt.Sprintf("history %d step %d", history, step)
			timestamp := commitTime.Add(time.Duration(step+1) * time.Second)
			nativeHash, err = nativeDocument.Commit(ctx, message, timestamp)
			require.NoError(t, err)
			referenceHash, err = referenceDocument.Commit(
				ctx,
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

			nativeLength, err := nativeList.Len(ctx)
			require.NoError(t, err)
			referenceLength, err := referenceList.Len(ctx)
			require.NoError(t, err)
			assert.Equal(t, uint64(len(model)), nativeLength)
			assert.Equal(t, referenceLength, nativeLength)

			for index, expected := range model {
				nativeValue, err := nativeList.ScalarAt(ctx, uint64(index))
				require.NoError(t, err)
				referenceValue, err := referenceList.ScalarAt(
					ctx,
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

	ctx := context.Background()

	for history := range histories {
		random := rand.New(rand.NewSource(int64(30_000 + history)))
		actorID := actor(byte(170 + history))
		nativeDocument, err := automerge.New(ctx, actorID)
		require.NoError(t, err)
		closeDocument(t, nativeDocument)

		referenceDocument, err := automerge.NewReference(ctx, actorID)
		require.NoError(t, err)
		closeDocument(t, referenceDocument)

		nativeMap, err := nativeDocument.Root().CreateObject(
			ctx,
			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)
		referenceMap, err := referenceDocument.Root().CreateObject(
			ctx,
			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)
		nativeHash, err := nativeDocument.Commit(ctx, "create map", commitTime)
		require.NoError(t, err)
		referenceHash, err := referenceDocument.Commit(
			ctx,
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
				require.NoError(t, nativeMap.DeleteKey(ctx, key))
				require.NoError(t, referenceMap.DeleteKey(ctx, key))
			} else {
				value := random.Int63()
				model[key] = value
				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  value,
				}
				require.NoError(t, nativeMap.PutScalar(ctx, key, scalar))
				require.NoError(t, referenceMap.PutScalar(ctx, key, scalar))
			}

			message := fmt.Sprintf("history %d step %d", history, step)
			timestamp := commitTime.Add(time.Duration(step+1) * time.Second)
			nativeHash, err = nativeDocument.Commit(ctx, message, timestamp)
			require.NoError(t, err)
			referenceHash, err = referenceDocument.Commit(
				ctx,
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
				nativeValue, nativeErr := nativeMap.Scalar(ctx, candidate)
				referenceValue, referenceErr := referenceMap.Scalar(
					ctx,
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

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(180))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(180))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		require.NoError(t, document.PutString(ctx, "value", "committed"))
		_, err = document.Commit(ctx, "initial", commitTime)
		require.NoError(t, err)
		headsBefore, err := document.Heads(ctx)
		require.NoError(t, err)

		require.NoError(t, document.PutString(ctx, "value", "rolled back"))
		list, err := document.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(
				ctx,
				0,
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
			),
		)
		cancelled, err := document.Rollback(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), cancelled)

		value, err := document.String(ctx, "value")
		require.NoError(t, err)
		assert.Equal(t, "committed", value)

		_, err = document.Root().Object(ctx, "list")
		require.Error(t, err)
		headsAfter, err := document.Heads(ctx)
		require.NoError(t, err)
		assert.Equal(t, headsBefore, headsAfter)

		cancelled, err = document.Rollback(ctx)
		require.NoError(t, err)
		assert.Zero(t, cancelled)
	}
}

func TestDocument_ForkMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
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

				ctx := context.Background()
				document, err := factory(ctx, actor(161))
				require.NoError(t, err)
				closeDocument(t, document)
				require.NoError(t, document.PutString(ctx, "base", "value"))
				_, err = document.Commit(ctx, "base", commitTime)
				require.NoError(t, err)

				fork, err := document.Fork(ctx, actor(162))
				require.NoError(t, err)
				closeDocument(t, fork)
				require.NoError(t, fork.PutString(ctx, "fork", "value"))
				_, err = fork.Commit(
					ctx,
					"fork",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)

				_, err = document.String(ctx, "fork")
				require.Error(t, err)
				value, err := fork.String(ctx, "base")
				require.NoError(t, err)
				assert.Equal(t, "value", value)

				_, err = document.Merge(ctx, fork)
				require.NoError(t, err)
				value, err = document.String(ctx, "fork")
				require.NoError(t, err)
				assert.Equal(t, "value", value)
			},
		)
	}
}

func TestDocument_WrongObjectOperationsMatchReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
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

				ctx := context.Background()
				document, err := factory(ctx, actor(171))
				require.NoError(t, err)
				closeDocument(t, document)
				root := document.Root()
				mapObject, err := root.CreateObject(
					ctx,
					"map",
					automerge.ObjectTypeMap,
				)
				require.NoError(t, err)
				listObject, err := root.CreateObject(
					ctx,
					"list",
					automerge.ObjectTypeList,
				)
				require.NoError(t, err)
				textObject, err := root.CreateObject(
					ctx,
					"text",
					automerge.ObjectTypeText,
				)
				require.NoError(t, err)

				scalar := automerge.Scalar{
					Type: automerge.ScalarTypeInt,
					Int:  1,
				}
				require.Error(t, listObject.PutScalar(ctx, "key", scalar))
				require.Error(t, textObject.PutScalar(ctx, "key", scalar))
				require.Error(t, mapObject.InsertScalar(ctx, 0, scalar))
				require.Error(t, mapObject.DeleteIndex(ctx, 0))
				length, err := mapObject.Len(ctx)
				require.NoError(t, err)
				assert.Zero(t, length)

				_, err = listObject.Text(ctx)
				require.Error(t, err)
			},
		)
	}
}

func TestDocument_DeletedObjectsSaveLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(172))
	require.NoError(t, err)
	closeDocument(t, document)

	root := document.Root()
	for key, objectType := range map[string]automerge.ObjectType{
		"list":  automerge.ObjectTypeList,
		"text":  automerge.ObjectTypeText,
		"map":   automerge.ObjectTypeMap,
		"table": automerge.ObjectTypeTable,
	} {
		_, err := root.CreateObject(ctx, key, objectType)
		require.NoError(t, err)
		require.NoError(t, root.DeleteKey(ctx, key))
	}

	_, err = document.Commit(ctx, "deleted objects", commitTime)
	require.NoError(t, err)
	data, err := document.Save(ctx)
	require.NoError(t, err)

	nativeLoaded, err := automerge.Load(ctx, data, actor(173))
	require.NoError(t, err)
	closeDocument(t, nativeLoaded)

	referenceLoaded, err := automerge.LoadReference(ctx, data, actor(174))
	require.NoError(t, err)
	closeDocument(t, referenceLoaded)

	for _, loaded := range []*automerge.Document{
		nativeLoaded,
		referenceLoaded,
	} {
		length, err := loaded.Root().Len(ctx)
		require.NoError(t, err)
		assert.Zero(t, length)
	}
}

func TestDocument_ManyMapDeletes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(175))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(175))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		object, err := document.Root().CreateObject(
			ctx,
			"object",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)

		for index := range 100 {
			key := fmt.Sprintf("%d", index)
			require.NoError(
				t,
				object.PutScalar(
					ctx,
					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
			require.NoError(t, object.DeleteKey(ctx, key))
		}

		_, err = document.Commit(ctx, "many deletes", commitTime)
		require.NoError(t, err)
		length, err := object.Len(ctx)
		require.NoError(t, err)
		assert.Zero(t, length)
	}
}

func TestDocument_MapKeysMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
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

				document, err := factory(ctx, actor(179))
				require.NoError(t, err)
				closeDocument(t, document)
				keys, err := document.Root().Keys(ctx)
				require.NoError(t, err)
				assert.Empty(t, keys)

				object, err := document.Root().CreateObject(
					ctx,
					"map",
					automerge.ObjectTypeMap,
				)
				require.NoError(t, err)

				for _, key := range []string{"z", "", "a@b", "a"} {
					require.NoError(
						t,
						object.PutScalar(
							ctx,
							key,
							automerge.Scalar{
								Type: automerge.ScalarTypeInt,
								Int:  1,
							},
						),
					)
				}

				require.NoError(t, object.DeleteKey(ctx, "z"))
				keys, err = object.Keys(ctx)
				require.NoError(t, err)
				assert.Equal(t, []string{"", "a", "a@b"}, keys)

				_, err = document.Commit(ctx, "keys", commitTime)
				require.NoError(t, err)
				loaded, err := document.Fork(ctx, actor(180))
				require.NoError(t, err)
				closeDocument(t, loaded)
				loadedObject, err := loaded.Root().Object(ctx, "map")
				require.NoError(t, err)
				keys, err = loadedObject.Keys(ctx)
				require.NoError(t, err)
				assert.Equal(t, []string{"", "a", "a@b"}, keys)
			},
		)
	}
}

func TestDocument_NoOpMergeAndEqualPutMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
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

				document, err := factory(ctx, actor(176))
				require.NoError(t, err)
				closeDocument(t, document)
				require.NoError(t, document.PutString(ctx, "value", "same"))
				_, err = document.Commit(ctx, "initial", commitTime)
				require.NoError(t, err)

				fork, err := document.Fork(ctx, actor(177))
				require.NoError(t, err)
				closeDocument(t, fork)
				_, err = fork.EmptyCommit(
					ctx,
					"noop",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)
				require.NoError(t, fork.PutString(ctx, "value", "changed"))
				_, err = fork.Commit(
					ctx,
					"real",
					commitTime.Add(2*time.Second),
				)
				require.NoError(t, err)

				require.NoError(t, document.PutString(ctx, "value", "same"))
				_, err = document.Commit(
					ctx,
					"equal",
					commitTime.Add(time.Second),
				)
				require.Error(t, err)
				_, err = document.Merge(ctx, fork)
				require.NoError(t, err)
				value, err := document.String(ctx, "value")
				require.NoError(t, err)
				assert.Equal(t, "changed", value)

				data, err := document.Save(ctx)
				require.NoError(t, err)
				loaded, err := automerge.Load(ctx, data, actor(178))
				require.NoError(t, err)
				closeDocument(t, loaded)
				value, err = loaded.String(ctx, "value")
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

	ctx := context.Background()
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	for history := range histories {
		base, err := automerge.New(ctx, actor(byte(190+history)))
		require.NoError(t, err)
		closeDocument(t, base)
		baseMap, err := base.Root().CreateObject(
			ctx,
			"map",
			automerge.ObjectTypeMap,
		)
		require.NoError(t, err)

		for index, key := range keys {
			require.NoError(
				t,
				baseMap.PutScalar(
					ctx,
					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
		}

		_, err = base.Commit(ctx, "base map", commitTime)
		require.NoError(t, err)
		baseData, err := base.Save(ctx)
		require.NoError(t, err)

		left, err := automerge.Load(ctx, baseData, actor(byte(210+history)))
		require.NoError(t, err)
		closeDocument(t, left)
		leftMap, err := left.Root().Object(ctx, "map")
		require.NoError(t, err)
		right, err := automerge.Load(ctx, baseData, actor(byte(230+history)))
		require.NoError(t, err)
		closeDocument(t, right)
		rightMap, err := right.Root().Object(ctx, "map")
		require.NoError(t, err)
		referenceLeft, err := automerge.LoadReference(
			ctx,
			baseData,
			actor(byte(210+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceLeft)
		referenceLeftMap, err := referenceLeft.Root().Object(ctx, "map")
		require.NoError(t, err)
		referenceRight, err := automerge.LoadReference(
			ctx,
			baseData,
			actor(byte(230+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceRight)
		referenceRightMap, err := referenceRight.Root().Object(ctx, "map")
		require.NoError(t, err)

		leftSeed := int64(40_000 + history)
		mutateRandomMap(
			t,
			ctx,
			rand.New(rand.NewSource(leftSeed)),
			left,
			leftMap,
			keys,
			steps,
			"left",
		)
		mutateRandomMap(
			t,
			ctx,
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
			ctx,
			rand.New(rand.NewSource(rightSeed)),
			right,
			rightMap,
			keys,
			steps,
			"right",
		)
		mutateRandomMap(
			t,
			ctx,
			rand.New(rand.NewSource(rightSeed)),
			referenceRight,
			referenceRightMap,
			keys,
			steps,
			"right",
		)

		_, err = left.Merge(ctx, right)
		require.NoError(t, err)
		_, err = right.Merge(ctx, left)
		require.NoError(t, err)
		_, err = referenceLeft.Merge(ctx, referenceRight)
		require.NoError(t, err)
		_, err = referenceRight.Merge(ctx, referenceLeft)
		require.NoError(t, err)
		leftMap, err = left.Root().Object(ctx, "map")
		require.NoError(t, err)
		rightMap, err = right.Root().Object(ctx, "map")
		require.NoError(t, err)
		assertMapParity(t, ctx, keys, leftMap, rightMap)
		referenceLeftMap, err = referenceLeft.Root().Object(ctx, "map")
		require.NoError(t, err)
		referenceRightMap, err = referenceRight.Root().Object(ctx, "map")
		require.NoError(t, err)
		assertMapParity(t, ctx, keys, referenceLeftMap, referenceRightMap)
		assertMapParity(t, ctx, keys, leftMap, referenceLeftMap)
	}
}

func TestDocument_RandomConcurrentListParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 20
		steps     = 30
	)

	ctx := context.Background()
	for history := range histories {
		base, err := automerge.New(ctx, actor(byte(10+history)))
		require.NoError(t, err)
		closeDocument(t, base)
		baseList, err := base.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)

		for index := range 8 {
			require.NoError(
				t,
				baseList.InsertScalar(
					ctx,
					uint64(index),
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  int64(index),
					},
				),
			)
		}

		_, err = base.Commit(ctx, "base list", commitTime)
		require.NoError(t, err)
		baseData, err := base.Save(ctx)
		require.NoError(t, err)

		left, err := automerge.Load(ctx, baseData, actor(byte(30+history)))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object(ctx, "list")
		require.NoError(t, err)
		right, err := automerge.Load(ctx, baseData, actor(byte(50+history)))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object(ctx, "list")
		require.NoError(t, err)
		referenceLeft, err := automerge.LoadReference(
			ctx,
			baseData,
			actor(byte(30+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceLeft)
		referenceLeftList, err := referenceLeft.Root().Object(ctx, "list")
		require.NoError(t, err)
		referenceRight, err := automerge.LoadReference(
			ctx,
			baseData,
			actor(byte(50+history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceRight)
		referenceRightList, err := referenceRight.Root().Object(ctx, "list")
		require.NoError(t, err)

		leftSeed := int64(60_000 + history)
		mutateRandomList(
			t,
			ctx,
			rand.New(rand.NewSource(leftSeed)),
			left,
			leftList,
			steps,
			"left",
		)
		mutateRandomList(
			t,
			ctx,
			rand.New(rand.NewSource(leftSeed)),
			referenceLeft,
			referenceLeftList,
			steps,
			"left",
		)

		rightSeed := int64(70_000 + history)
		mutateRandomList(
			t,
			ctx,
			rand.New(rand.NewSource(rightSeed)),
			right,
			rightList,
			steps,
			"right",
		)
		mutateRandomList(
			t,
			ctx,
			rand.New(rand.NewSource(rightSeed)),
			referenceRight,
			referenceRightList,
			steps,
			"right",
		)

		_, err = left.Merge(ctx, right)
		require.NoError(t, err)
		_, err = right.Merge(ctx, left)
		require.NoError(t, err)
		_, err = referenceLeft.Merge(ctx, referenceRight)
		require.NoError(t, err)
		_, err = referenceRight.Merge(ctx, referenceLeft)
		require.NoError(t, err)
		leftList, err = left.Root().Object(ctx, "list")
		require.NoError(t, err)
		rightList, err = right.Root().Object(ctx, "list")
		require.NoError(t, err)
		assertListParity(t, ctx, leftList, rightList)
		referenceLeftList, err = referenceLeft.Root().Object(ctx, "list")
		require.NoError(t, err)
		referenceRightList, err = referenceRight.Root().Object(ctx, "list")
		require.NoError(t, err)
		assertListParity(t, ctx, referenceLeftList, referenceRightList)
		assertListParity(t, ctx, leftList, referenceLeftList)
	}
}

func TestDocument_ConcurrentListOrderingMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	results := make(map[string][]string)

	for name, factory := range factories {
		base, err := factory(ctx, actor(80))
		require.NoError(t, err)
		closeDocument(t, base)
		list, err := base.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertScalar(
				ctx,
				0,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "A"},
			),
		)
		_, err = base.Commit(ctx, "base", commitTime)
		require.NoError(t, err)

		left, err := base.Fork(ctx, actor(81))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(
			t,
			leftList.InsertValues(
				ctx,
				1,
				[]automerge.Value{
					hydratedString("L1"),
					hydratedString("L2"),
				},
			),
		)
		_, err = left.Commit(ctx, "left", commitTime.Add(time.Second))
		require.NoError(t, err)

		right, err := base.Fork(ctx, actor(82))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(
			t,
			rightList.InsertValues(
				ctx,
				1,
				[]automerge.Value{
					hydratedString("R1"),
					hydratedString("R2"),
				},
			),
		)
		_, err = right.Commit(ctx, "right", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = left.Merge(ctx, right)
		require.NoError(t, err)
		leftList, err = left.Root().Object(ctx, "list")
		require.NoError(t, err)
		values := listStrings(t, ctx, leftList)
		assert.Equal(t, []string{"A", "R1", "R2", "L1", "L2"}, values)
		results[name] = values
	}

	assert.Equal(t, results["reference"], results["native"])
}

func TestDocument_InsertAfterConcurrentDeleteMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	results := make(map[string][]string)

	for name, factory := range factories {
		base, err := factory(ctx, actor(83))
		require.NoError(t, err)
		closeDocument(t, base)
		list, err := base.Root().CreateObject(
			ctx,
			"list",
			automerge.ObjectTypeList,
		)
		require.NoError(t, err)
		require.NoError(
			t,
			list.InsertValues(
				ctx,
				0,
				[]automerge.Value{
					hydratedString("A"),
					hydratedString("B"),
				},
			),
		)
		_, err = base.Commit(ctx, "base", commitTime)
		require.NoError(t, err)

		left, err := base.Fork(ctx, actor(84))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(t, leftList.DeleteIndex(ctx, 0))
		_, err = left.Commit(ctx, "delete", commitTime.Add(time.Second))
		require.NoError(t, err)

		right, err := base.Fork(ctx, actor(85))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(
			t,
			rightList.InsertScalar(
				ctx,
				1,
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "X"},
			),
		)
		_, err = right.Commit(ctx, "insert", commitTime.Add(time.Second))
		require.NoError(t, err)

		_, err = left.Merge(ctx, right)
		require.NoError(t, err)
		leftList, err = left.Root().Object(ctx, "list")
		require.NoError(t, err)
		results[name] = listStrings(t, ctx, leftList)
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
	ctx context.Context,
	list *automerge.Object,
) []string {
	t.Helper()

	length, err := list.Len(ctx)
	require.NoError(t, err)

	values := make([]string, length)
	for index := range length {
		value, err := list.ScalarAt(ctx, index)
		require.NoError(t, err)

		values[index] = value.String
	}

	return values
}

func mutateRandomMap(
	t *testing.T,
	ctx context.Context,
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
			require.NoError(t, object.DeleteKey(ctx, key))
			present[key] = false
		} else {
			require.NoError(
				t,
				object.PutScalar(
					ctx,
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
				ctx,
				fmt.Sprintf("%s map step %d", side, step),
				commitTime.Add(time.Duration(step+1)*time.Second),
			)
			require.NoError(t, err)
		}
	}
}

func mutateRandomList(
	t *testing.T,
	ctx context.Context,
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
					ctx,
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
			require.NoError(t, object.DeleteIndex(ctx, uint64(index)))

			length--
		default:
			index := random.Intn(length)
			require.NoError(
				t,
				object.PutScalarAt(
					ctx,
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
				ctx,
				fmt.Sprintf("%s list step %d", side, step),
				commitTime.Add(time.Duration(step+1)*time.Second),
			)
			require.NoError(t, err)
		}
	}
}

func assertMapParity(
	t *testing.T,
	ctx context.Context,
	keys []string,
	left *automerge.Object,
	right *automerge.Object,
) {
	t.Helper()

	for _, key := range keys {
		leftValues, leftErr := left.Scalars(ctx, key)
		rightValues, rightErr := right.Scalars(ctx, key)
		assert.Equal(t, leftErr != nil, rightErr != nil, "key %q", key)

		if leftErr == nil && rightErr == nil {
			assert.ElementsMatch(t, leftValues, rightValues, "key %q", key)
		}
	}
}

func assertListParity(
	t *testing.T,
	ctx context.Context,
	left *automerge.Object,
	right *automerge.Object,
) {
	t.Helper()

	leftLength, err := left.Len(ctx)
	require.NoError(t, err)
	rightLength, err := right.Len(ctx)
	require.NoError(t, err)
	require.Equal(t, leftLength, rightLength)

	for index := range leftLength {
		leftValue, err := left.ScalarAt(ctx, index)
		require.NoError(t, err)
		rightValue, err := right.ScalarAt(ctx, index)
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
