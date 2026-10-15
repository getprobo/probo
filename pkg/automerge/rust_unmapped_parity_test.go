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

// These tests reproduce the Rust automerge 0.11 cases that were still pending
// in the parity ledger. Each scenario runs on the native engine and the
// Rust/WASM reference, and the observable heads, hydrated values, and patches
// must agree.

package automerge_test

import (
	"bytes"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

type parityView struct {
	Heads   []string
	Hydrate automerge.Value
	Patches []automerge.Patch
	Missing []string
	Saved   automerge.Value
	Message string
	Failed  bool
	Count   uint64
}

func openParity(
	t *testing.T,
	engine rustParityEngine,
	actorID automerge.ActorID,
) *automerge.Document {
	t.Helper()

	document, err := engine.open(actorID)
	require.NoError(t, err)
	closeDocument(t, document)

	return document
}

func loadParity(
	t *testing.T,
	engine rustParityEngine,
	data []byte,
	actorID automerge.ActorID,
) *automerge.Document {
	t.Helper()

	document, err := engine.load(data, actorID)
	require.NoError(t, err)
	closeDocument(t, document)

	return document
}

func commitParity(
	t *testing.T,
	document *automerge.Document,
	message string,
) {
	t.Helper()

	_, err := document.Commit(message, commitTime)
	require.NoError(t, err)
}

func headsOf(t *testing.T, document *automerge.Document) []automerge.Hash {
	t.Helper()

	heads, err := document.Heads()
	require.NoError(t, err)

	return heads
}

func hydrateOf(t *testing.T, document *automerge.Document) automerge.Value {
	t.Helper()

	return materializeObject(t, document.Root())
}

func materializeObject(t *testing.T, object *automerge.Object) automerge.Value {
	t.Helper()

	switch object.Type {
	case automerge.ObjectTypeList:
		length, err := object.Len()
		require.NoError(t, err)

		list := make([]automerge.Value, 0, length)
		for index := range length {
			if child, err := object.ObjectAt(index); err == nil {
				list = append(list, materializeObject(t, child))
				continue
			}

			scalar, err := object.ScalarAt(index)
			require.NoError(t, err)

			list = append(list, automerge.Value{
				Type:   automerge.ValueTypeScalar,
				Scalar: scalar,
			})
		}

		return automerge.Value{Type: automerge.ValueTypeList, List: list}
	case automerge.ObjectTypeText:
		text, err := object.Text()
		require.NoError(t, err)
		rendered, err := text.String()
		require.NoError(t, err)

		return automerge.Value{Type: automerge.ValueTypeText, Text: rendered}
	default:
		keys, err := object.Keys()
		require.NoError(t, err)
		sort.Strings(keys)

		properties := make(map[string]automerge.Value, len(keys))
		for _, key := range keys {
			if child, err := object.Object(key); err == nil {
				properties[key] = materializeObject(t, child)
				continue
			}

			scalar, err := object.Scalar(key)
			require.NoError(t, err)

			properties[key] = automerge.Value{
				Type:   automerge.ValueTypeScalar,
				Scalar: scalar,
			}
		}

		return automerge.Value{Type: automerge.ValueTypeMap, Map: properties}
	}
}

func hashIDs(heads []automerge.Hash) []string {
	ids := make([]string, len(heads))
	for index, head := range heads {
		ids[index] = head.String()
	}

	sort.Strings(ids)

	return ids
}

func agreeView(
	t *testing.T,
	run func(t *testing.T, engine rustParityEngine) parityView,
) parityView {
	t.Helper()

	views := make(map[string]parityView, 2)
	for _, engine := range rustParityEngines() {
		views[engine.name] = run(t, engine)
	}

	assert.Equal(t, views["reference"], views["native"])

	return views["native"]
}

func rootScalar(
	t *testing.T,
	document *automerge.Document,
	key string,
) (automerge.Scalar, bool) {
	t.Helper()

	value, err := document.Root().Scalar(key)
	if err != nil {
		return automerge.Scalar{}, false
	}

	return value, true
}

func changeMessage(t *testing.T, document *automerge.Document) string {
	t.Helper()

	data, err := document.Save(automerge.NoCompress())
	require.NoError(t, err)
	decoded, err := storage.Decode(data)
	require.NoError(t, err)
	require.NotEmpty(t, decoded.Changes)

	return decoded.Changes[len(decoded.Changes)-1].Message
}

func resetSaveCursor(t *testing.T, document *automerge.Document) {
	t.Helper()

	_, err := document.SaveIncremental()
	require.NoError(t, err)
}

func takeIncremental(t *testing.T, document *automerge.Document) []byte {
	t.Helper()

	data, err := document.SaveIncremental()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	return data
}

// TestRustHydrate_Simple reproduces simple_hydrate.
func TestRustHydrate_Simple(t *testing.T) {
	t.Parallel()

	view := agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x11))
		require.NoError(t, document.Root().PutScalar("name", automerge.StringScalar("ada")))
		commitParity(t, document, "seed")

		return parityView{Hydrate: hydrateOf(t, document)}
	})

	assert.Equal(t, "ada", view.Hydrate.Map["name"].Scalar.String)
}

// TestRustIter_DocIter reproduces doc_iter by walking every root key and list
// element the same way a document iterator would.
func TestRustIter_DocIter(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x12))
		require.NoError(t, document.Root().PutScalar("name", automerge.StringScalar("ada")))
		list, err := document.Root().CreateObject("items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(0, automerge.StringScalar("one")))
		require.NoError(t, list.InsertScalar(1, automerge.IntScalar(2)))
		commitParity(t, document, "seed")

		keys, err := document.Root().Keys()
		require.NoError(t, err)
		sort.Strings(keys)

		length, err := list.Len()
		require.NoError(t, err)

		return parityView{
			Heads:   keys,
			Count:   length,
			Hydrate: hydrateOf(t, document),
		}
	})
}

// TestRustTransaction_PutAndGetRoundtrip reproduces put_and_get_roundtrip.
func TestRustTransaction_PutAndGetRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x21))
		require.NoError(t, document.Root().PutScalar("key", automerge.StringScalar("value")))
		commitParity(t, document, "put")
		value, ok := rootScalar(t, document, "key")
		require.True(t, ok)

		return parityView{Hydrate: automerge.Value{Type: automerge.ValueTypeScalar, Scalar: value}}
	})
}

// TestRustTransaction_ReadDuringTransaction reproduces read_during_transaction.
func TestRustTransaction_ReadDuringTransaction(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x22))
		require.NoError(t, document.Root().PutScalar("a", automerge.StringScalar("1")))
		value, ok := rootScalar(t, document, "a")
		require.True(t, ok)
		commitParity(t, document, "read")

		return parityView{Hydrate: automerge.Value{Type: automerge.ValueTypeScalar, Scalar: value}}
	})
}

// TestRustTransaction_NestedObjects reproduces nested_objects.
func TestRustTransaction_NestedObjects(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x23))
		list, err := document.Root().CreateObject("items", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(0, automerge.StringScalar("first")))
		require.NoError(t, list.InsertScalar(1, automerge.StringScalar("second")))
		commitParity(t, document, "nested")

		length, err := list.Len()
		require.NoError(t, err)

		return parityView{Count: length, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTransaction_CommitWithOptions reproduces commit_with_options.
func TestRustTransaction_CommitWithOptions(t *testing.T) {
	t.Parallel()

	view := agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x24))
		require.NoError(t, document.Root().PutScalar("x", automerge.IntScalar(42)))
		commitParity(t, document, "test commit")

		return parityView{Message: changeMessage(t, document), Hydrate: hydrateOf(t, document)}
	})

	assert.Equal(t, "test commit", view.Message)
}

// TestRustTransaction_OwnedTransactionAt reproduces owned_transaction_at.
func TestRustTransaction_OwnedTransactionAt(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x25))
		require.NoError(t, document.Root().PutScalar("v", automerge.IntScalar(1)))
		commitParity(t, document, "v1")
		heads := headsOf(t, document)
		require.NoError(t, document.Root().PutScalar("v", automerge.IntScalar(2)))
		commitParity(t, document, "v2")

		require.NoError(t, document.Isolate(heads))
		value, err := document.Root().Scalar("v")
		require.NoError(t, err)
		require.NoError(t, document.Root().PutScalar("from_v1", automerge.BoolScalar(true)))
		commitParity(t, document, "from v1")
		require.NoError(t, document.Integrate())

		return parityView{
			Count:   uint64(value.Int),
			Hydrate: hydrateOf(t, document),
		}
	})
}

// TestRustTransaction_LogPatches reproduces log_patches.
func TestRustTransaction_LogPatches(t *testing.T) {
	t.Parallel()

	view := agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x26))
		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, document.Root().PutScalar("patched", automerge.StringScalar("yes")))
		commitParity(t, document, "patch")
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Count: uint64(len(patches))}
	})

	assert.NotEmpty(t, view.Patches)
}

// TestRustTransaction_HeadsStayAtPreTransactionHeads reproduces
// get_heads_returns_pre_tx_heads: heads read during an open transaction stay at
// the pre-transaction frontier.
func TestRustTransaction_HeadsStayAtPreTransactionHeads(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x27))
		require.NoError(t, document.Root().PutScalar("a", automerge.IntScalar(1)))
		commitParity(t, document, "a")
		before := hashIDs(headsOf(t, document))
		require.NoError(t, document.Root().PutScalar("b", automerge.IntScalar(2)))
		after := hashIDs(headsOf(t, document))
		require.Equal(t, before, after)
		commitParity(t, document, "b")

		return parityView{Heads: after, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTransaction_PendingOps reproduces pending_ops.
func TestRustTransaction_PendingOps(t *testing.T) {
	t.Parallel()

	view := agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x28))
		require.NoError(t, document.Root().PutScalar("a", automerge.IntScalar(1)))
		require.NoError(t, document.Root().PutScalar("b", automerge.IntScalar(2)))
		cancelled, err := document.Rollback()
		require.NoError(t, err)

		return parityView{Count: cancelled, Hydrate: hydrateOf(t, document)}
	})

	assert.Equal(t, uint64(2), view.Count)
}

// TestRustTransaction_EmptyCommitReturnsNoHash reproduces
// empty_commit_returns_none_hash: a transaction with no operations does not
// create a change.
func TestRustTransaction_EmptyCommitReturnsNoHash(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x29))
		_, err := document.Commit("empty", commitTime)
		require.Error(t, err)

		return parityView{Heads: hashIDs(headsOf(t, document)), Failed: true}
	})
}

// TestRustBatch_InsertWithTransaction reproduces batch_insert_with_transaction.
func TestRustBatch_InsertWithTransaction(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x31))
		require.NoError(t, document.Root().PutValue("data", automerge.Value{
			Type: automerge.ValueTypeMap,
			Map: map[string]automerge.Value{
				"key": {Type: automerge.ValueTypeScalar, Scalar: automerge.StringScalar("from_tx")},
			},
		}))
		pending := hydrateOf(t, document)
		commitParity(t, document, "batch")

		return parityView{Hydrate: hydrateOf(t, document), Saved: pending}
	})
}

// TestRustTest_HistoricalTransactionOmitsUnreachableObject reproduces
// historical_owned_transaction_omits_patches_for_unreachable_object_from_fuzz_trace.
func TestRustTest_HistoricalTransactionOmitsUnreachableObject(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x34))
		require.NoError(t, document.Root().PutScalar("seed", automerge.IntScalar(1)))
		commitParity(t, document, "seed")
		frontier := headsOf(t, document)
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		commitParity(t, document, "text")
		require.NoError(t, document.Isolate(frontier))
		require.Error(t, textObject.InsertScalar(0, automerge.CounterScalar(-104)))

		return parityView{
			Failed:  true,
			Heads:   hashIDs(headsOf(t, document)),
			Hydrate: hydrateOf(t, document),
		}
	})
}
func TestRustTest_ForkAtIncrementOverConflictedCounterSurvivesSaveLoad(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x40))
		counter := openParity(t, engine, actor(0x41))
		require.NoError(t, counter.Root().PutScalar("key", automerge.CounterScalar(34)))
		commitParity(t, counter, "counter")
		scalar := openParity(t, engine, actor(0x42))
		require.NoError(t, scalar.Root().PutScalar("key", automerge.NullScalar()))
		commitParity(t, scalar, "null")

		_, err := document.Merge(counter)
		require.NoError(t, err)
		_, err = document.Merge(scalar)
		require.NoError(t, err)
		require.NoError(t, document.Root().Increment("key", -59))
		commitParity(t, document, "increment")

		fork, err := document.Fork(actor(0x43))
		require.NoError(t, err)
		closeDocument(t, fork)
		before := hydrateOf(t, fork)
		data, err := fork.Save(automerge.NoCompress())
		require.NoError(t, err)
		loaded := loadParity(t, engine, data, actor(0x44))

		return parityView{Hydrate: before, Saved: hydrateOf(t, loaded)}
	})
}
func TestRustTest_DiffPreservesConflictAfterWinningValueIsDeleted(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x46))
		second, err := document.Fork(actor(0x47))
		require.NoError(t, err)
		closeDocument(t, second)

		third, err := document.Fork(actor(0x48))
		require.NoError(t, err)
		closeDocument(t, third)
		require.NoError(t, document.Root().PutScalar("key", automerge.BoolScalar(false)))
		commitParity(t, document, "bool")
		require.NoError(t, second.Root().PutScalar("key", automerge.NullScalar()))
		commitParity(t, second, "null")
		require.NoError(t, third.Root().PutScalar("key", automerge.IntScalar(-88)))
		commitParity(t, third, "int")

		_, err = document.Merge(second)
		require.NoError(t, err)
		_, err = document.Merge(third)
		require.NoError(t, err)
		before := headsOf(t, document)
		require.NoError(t, third.Root().DeleteKey("key"))
		commitParity(t, third, "delete")
		_, err = document.Merge(third)
		require.NoError(t, err)
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}
func TestRustTest_ReverseDiffOfReplacedLists(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x4a))
		_, err := document.Root().CreateObject("first", automerge.ObjectTypeList)
		require.NoError(t, err)
		oldSecond, err := document.Root().CreateObject("second", automerge.ObjectTypeList)
		require.NoError(t, err)
		commitParity(t, document, "seed")
		historical := headsOf(t, document)
		_, err = document.Root().CreateObject("first", automerge.ObjectTypeList)
		require.NoError(t, err)
		_, err = document.Root().CreateObject("second", automerge.ObjectTypeText)
		require.NoError(t, err)
		require.NoError(t, oldSecond.InsertScalar(0, automerge.IntScalar(1)))
		commitParity(t, document, "replace")
		patches, err := document.Diff(headsOf(t, document), historical)
		require.NoError(t, err)
		require.NoError(t, document.Isolate(historical))

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}
func TestRustTest_DiffSpansSkipLosingConflictingText(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x4c))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		require.NoError(t, textObject.InsertScalar(0, automerge.StringScalar("base")))
		commitParity(t, document, "base")
		before := headsOf(t, document)
		left, err := document.Fork(actor(0x4d))
		require.NoError(t, err)
		closeDocument(t, left)
		leftText, err := left.Root().Object("text")
		require.NoError(t, err)
		require.NoError(t, leftText.PutScalarAt(0, automerge.StringScalar("left")))
		commitParity(t, left, "left")

		right, err := document.Fork(actor(0x4e))
		require.NoError(t, err)
		closeDocument(t, right)
		rightText, err := right.Root().Object("text")
		require.NoError(t, err)
		require.NoError(t, rightText.PutScalarAt(0, automerge.StringScalar("right")))
		commitParity(t, right, "right")
		_, err = left.Merge(right)
		require.NoError(t, err)
		patches, err := left.Diff(before, headsOf(t, left))
		require.NoError(t, err)
		visible, err := leftText.Text()
		require.NoError(t, err)
		rendered, err := visible.String()
		require.NoError(t, err)

		spliced := ""

		for _, patch := range patches {
			if patch.Action == automerge.PatchSpliceText {
				spliced += patch.Text
			}
		}

		return parityView{
			Patches: patches,
			Message: spliced + "=" + rendered,
			Hydrate: hydrateOf(t, left),
		}
	})
}

// TestRustTest_FailedMergeDuplicateSequenceDoesNotCorruptSaveLoad reproduces
// failed_merge_with_duplicate_sequence_number_does_not_corrupt_save_load.
func TestRustTest_FailedMergeDuplicateSequenceDoesNotCorruptSaveLoad(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		left := openParity(t, engine, actor(0x50))
		right, err := left.Fork(actor(0x51))
		require.NoError(t, err)
		closeDocument(t, right)
		_, err = right.Root().CreateObject("k", automerge.ObjectTypeMap)
		require.NoError(t, err)
		commitParity(t, right, "map")
		sameActor, err := right.Fork(actor(0x50))
		require.NoError(t, err)
		closeDocument(t, sameActor)
		_, err = sameActor.Root().CreateObject("k", automerge.ObjectTypeMap)
		require.NoError(t, err)
		commitParity(t, sameActor, "again")
		_, err = right.Merge(sameActor)
		require.NoError(t, err)

		require.NoError(t, left.Root().PutScalar("k", automerge.StringScalar("a")))
		commitParity(t, left, "string")
		_, mergeErr := left.Merge(right)
		data, err := left.Save(automerge.NoCompress())
		require.NoError(t, err)
		loaded := loadParity(t, engine, data, actor(0x52))

		return parityView{
			Failed:  mergeErr != nil,
			Hydrate: hydrateOf(t, left),
			Saved:   hydrateOf(t, loaded),
		}
	})
}

// TestRustTest_DuplicateSeqInBatchAbortsWholeBatch reproduces
// duplicate_seq_number_in_batch_change_aborts_whole_batch.
func TestRustTest_DuplicateSeqInBatchAbortsWholeBatch(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		base := openParity(t, engine, actor(0x53))
		require.NoError(t, base.Root().PutScalar("foo", automerge.StringScalar("bar")))
		commitParity(t, base, "bar")
		middle, err := base.Fork(actor(0x53))
		require.NoError(t, err)
		closeDocument(t, middle)
		resetSaveCursor(t, middle)
		require.NoError(t, middle.Root().PutScalar("foo", automerge.StringScalar("baz")))
		commitParity(t, middle, "baz")
		change1 := takeIncremental(t, middle)

		left, err := base.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, left)
		resetSaveCursor(t, left)
		require.NoError(t, left.Root().PutScalar("foo", automerge.StringScalar("qux")))
		commitParity(t, left, "qux")
		change2 := takeIncremental(t, left)

		right, err := middle.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, right)
		resetSaveCursor(t, right)
		require.NoError(t, right.Root().PutScalar("foo", automerge.StringScalar("boz")))
		commitParity(t, right, "boz")
		change3 := takeIncremental(t, right)

		_, err = base.LoadIncremental(change2)
		require.NoError(t, err)

		_, applyErr := base.LoadIncremental(append(append([]byte{}, change1...), change3...))

		return parityView{Failed: applyErr != nil, Hydrate: hydrateOf(t, base)}
	})
}

// TestRustTest_DuplicateSeqWithinIncomingBatchIsRejected reproduces
// duplicate_seq_number_within_incoming_batch_is_rejected.
func TestRustTest_DuplicateSeqWithinIncomingBatchIsRejected(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		base := openParity(t, engine, actor(0x54))
		require.NoError(t, base.Root().PutScalar("foo", automerge.StringScalar("bar")))
		commitParity(t, base, "bar")
		left, err := base.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, left)
		resetSaveCursor(t, left)
		require.NoError(t, left.Root().PutScalar("left", automerge.IntScalar(1)))
		commitParity(t, left, "left")

		right, err := base.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, right)
		resetSaveCursor(t, right)
		require.NoError(t, right.Root().PutScalar("right", automerge.IntScalar(2)))
		commitParity(t, right, "right")
		first := takeIncremental(t, left)
		second := takeIncremental(t, right)
		_, applyErr := base.LoadIncremental(append(append([]byte{}, first...), second...))

		return parityView{Failed: applyErr != nil, Hydrate: hydrateOf(t, base)}
	})
}

// TestRustTest_DuplicateSeqChangesAreRejected reproduces
// duplicate_seq_number_changes_are_rejected.
func TestRustTest_DuplicateSeqChangesAreRejected(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		base := openParity(t, engine, actor(0x55))
		require.NoError(t, base.Root().PutScalar("foo", automerge.StringScalar("bar")))
		commitParity(t, base, "bar")
		left, err := base.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, left)
		resetSaveCursor(t, left)
		require.NoError(t, left.Root().PutScalar("foo", automerge.StringScalar("qux")))
		commitParity(t, left, "qux")

		right, err := base.Fork(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, right)
		resetSaveCursor(t, right)
		require.NoError(t, right.Root().PutScalar("foo", automerge.StringScalar("boz")))
		commitParity(t, right, "boz")
		_, err = base.LoadIncremental(takeIncremental(t, left))
		require.NoError(t, err)
		_, applyErr := base.LoadIncremental(takeIncremental(t, right))

		return parityView{Failed: applyErr != nil, Hydrate: hydrateOf(t, base)}
	})
}

// TestRustTest_QueuedOrphanWithConflictingActorSeqRejectsBatch reproduces
// queued_orphan_with_conflicting_actor_seq_rejects_incoming_batch.
func TestRustTest_QueuedOrphanWithConflictingActorSeqRejectsBatch(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		base := openParity(t, engine, actor(0xa1))
		require.NoError(t, base.Root().PutScalar("base", automerge.UintScalar(0)))
		commitParity(t, base, "base")
		saved, err := base.Save(automerge.NoCompress())
		require.NoError(t, err)

		receiver := loadParity(t, engine, saved, actor(0xa1))
		stale := loadParity(t, engine, saved, actor(0xb1))
		resetSaveCursor(t, stale)
		require.NoError(t, stale.Root().PutScalar("stale_missing", automerge.UintScalar(1)))
		commitParity(t, stale, "missing")
		resetSaveCursor(t, stale)
		require.NoError(t, stale.Root().PutScalar("stale_orphan", automerge.UintScalar(2)))
		commitParity(t, stale, "orphan")
		_, err = receiver.LoadIncremental(takeIncremental(t, stale))
		require.NoError(t, err)

		live := loadParity(t, engine, saved, actor(0xb1))
		resetSaveCursor(t, live)
		require.NoError(t, live.Root().PutScalar("live_1", automerge.UintScalar(3)))
		commitParity(t, live, "live1")
		require.NoError(t, live.Root().PutScalar("live_2", automerge.UintScalar(4)))
		commitParity(t, live, "live2")
		_, applyErr := receiver.LoadIncremental(takeIncremental(t, live))
		missing, err := receiver.MissingDependencies(nil)
		require.NoError(t, err)

		return parityView{
			Failed:  applyErr != nil,
			Heads:   hashIDs(headsOf(t, receiver)),
			Missing: hashIDs(missing),
			Hydrate: hydrateOf(t, receiver),
		}
	})
}

// TestRustTest_ForkAtCurrentHeadsAfterInterleavedActors reproduces
// fork_at_current_heads_after_interleaved_actor_changes.
func TestRustTest_ForkAtCurrentHeadsAfterInterleavedActors(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x02))
		_, err := document.Root().CreateObject("a", automerge.ObjectTypeMap)
		require.NoError(t, err)
		commitParity(t, document, "a")

		middle, err := document.Fork(actor(0x00))
		require.NoError(t, err)
		closeDocument(t, middle)
		require.NoError(t, middle.Root().PutScalar("b", automerge.BoolScalar(false)))
		commitParity(t, middle, "b")
		_, err = document.Merge(middle)
		require.NoError(t, err)
		require.NoError(t, document.Root().PutScalar("c", automerge.IntScalar(-7)))
		commitParity(t, document, "c")

		fork, err := document.Fork(actor(0x05))
		require.NoError(t, err)
		closeDocument(t, fork)

		return parityView{Hydrate: hydrateOf(t, fork), Heads: hashIDs(headsOf(t, fork))}
	})
}

// TestRustTest_ForkAtForeignHeadsErrors reproduces
// fork_at_foreign_heads_errors_instead_of_panicking.
func TestRustTest_ForkAtForeignHeadsErrors(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			first := openParity(t, engine, actor(0x61))
			require.NoError(t, first.Root().PutScalar("k", automerge.IntScalar(1)))
			commitParity(t, first, "k")
			second := openParity(t, engine, actor(0x62))
			require.NoError(t, second.Root().PutScalar("x", automerge.IntScalar(2)))
			commitParity(t, second, "x")
			require.Error(t, second.ForkAt(headsOf(t, first)))
		})
	}
}

// TestRustTest_IncrementPatchClearsResolvedMapConflict reproduces
// increment_patch_clears_resolved_map_conflict.
func TestRustTest_IncrementPatchClearsResolvedMapConflict(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x63))
		require.NoError(t, document.Root().PutScalar("key", automerge.NullScalar()))
		commitParity(t, document, "null")
		concurrent := openParity(t, engine, actor(0x64))
		require.NoError(t, concurrent.Root().PutScalar("key", automerge.CounterScalar(-1000)))
		commitParity(t, concurrent, "counter")
		_, err := document.Merge(concurrent)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())
		writer, err := document.Fork(actor(0x65))
		require.NoError(t, err)
		closeDocument(t, writer)
		require.NoError(t, writer.Root().Increment("key", 28))
		commitParity(t, writer, "increment")
		patches, err := writer.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, writer)}
	})
}
func isolateIntegrateDiff(
	t *testing.T,
	engine rustParityEngine,
	build func(t *testing.T, document *automerge.Document),
) parityView {
	t.Helper()

	document := openParity(t, engine, actor(0x7f))
	build(t, document)
	_, err := document.DiffIncremental()
	require.NoError(t, err)
	require.NoError(t, document.Isolate(nil))
	require.NoError(t, document.Integrate())
	patches, err := document.DiffIncremental()
	require.NoError(t, err)

	return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
}

// TestRustTest_IncrementalDiffSurvivesIsolateIntegrateRoundtrip reproduces
// incremental_diff_survives_isolate_integrate_roundtrip.
func TestRustTest_IncrementalDiffSurvivesIsolateIntegrateRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		return isolateIntegrateDiff(t, engine, func(t *testing.T, document *automerge.Document) {
			list, err := document.Root().CreateObject("k6", automerge.ObjectTypeList)
			require.NoError(t, err)
			commitParity(t, document, "list")
			require.NoError(t, list.InsertScalar(0, automerge.BoolScalar(false)))
			require.NoError(t, list.InsertScalar(1, automerge.IntScalar(1000)))
			commitParity(t, document, "values")
		})
	})
}

// TestRustTest_IncrementalTextDiffSurvivesIsolateIntegrateRoundtrip reproduces
// incremental_text_diff_survives_isolate_integrate_roundtrip.
func TestRustTest_IncrementalTextDiffSurvivesIsolateIntegrateRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		return isolateIntegrateDiff(t, engine, func(t *testing.T, document *automerge.Document) {
			textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
			require.NoError(t, err)
			commitParity(t, document, "text")

			text, err := textObject.Text()
			require.NoError(t, err)
			require.NoError(t, text.Splice(0, 0, "hello"))
			commitParity(t, document, "hello")
		})
	})
}

// TestRustTest_IncrementalCounterDiffSurvivesIsolateIntegrateRoundtrip reproduces
// incremental_counter_diff_survives_isolate_integrate_roundtrip.
func TestRustTest_IncrementalCounterDiffSurvivesIsolateIntegrateRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		return isolateIntegrateDiff(t, engine, func(t *testing.T, document *automerge.Document) {
			container, err := document.Root().CreateObject("container", automerge.ObjectTypeMap)
			require.NoError(t, err)
			require.NoError(t, container.PutScalar("count", automerge.CounterScalar(0)))
			commitParity(t, document, "counter")
			require.NoError(t, container.Increment("count", 1))
			commitParity(t, document, "increment")
		})
	})
}

// TestRustTest_IncrementalDeepTextDiffSurvivesIsolateIntegrateRoundtrip reproduces
// incremental_deep_text_diff_survives_isolate_integrate_roundtrip.
func TestRustTest_IncrementalDeepTextDiffSurvivesIsolateIntegrateRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		return isolateIntegrateDiff(t, engine, func(t *testing.T, document *automerge.Document) {
			outer, err := document.Root().CreateObject("outer", automerge.ObjectTypeMap)
			require.NoError(t, err)
			inner, err := outer.CreateObject("inner", automerge.ObjectTypeMap)
			require.NoError(t, err)
			textObject, err := inner.CreateObject("text", automerge.ObjectTypeText)
			require.NoError(t, err)
			commitParity(t, document, "deep")

			text, err := textObject.Text()
			require.NoError(t, err)
			require.NoError(t, text.Splice(0, 0, "deep"))
			commitParity(t, document, "text")
		})
	})
}

// TestRustTest_IncrementalSiblingTextDiffsSurviveIsolateIntegrateRoundtrip
// reproduces incremental_sibling_text_diffs_survive_isolate_integrate_roundtrip.
func TestRustTest_IncrementalSiblingTextDiffsSurviveIsolateIntegrateRoundtrip(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		return isolateIntegrateDiff(t, engine, func(t *testing.T, document *automerge.Document) {
			leftObject, err := document.Root().CreateObject("left", automerge.ObjectTypeText)
			require.NoError(t, err)
			rightObject, err := document.Root().CreateObject("right", automerge.ObjectTypeText)
			require.NoError(t, err)
			commitParity(t, document, "texts")

			left, err := leftObject.Text()
			require.NoError(t, err)
			right, err := rightObject.Text()
			require.NoError(t, err)
			require.NoError(t, left.Splice(0, 0, "L"))
			require.NoError(t, right.Splice(0, 0, "R"))
			commitParity(t, document, "letters")
		})
	})
}
func TestRustTest_IncrementalLoadPreservesText(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x72))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "preserved"))
		require.NoError(t, document.Root().PutScalar(
			"padding",
			automerge.BytesScalar(bytes.Repeat([]byte{0}, 300)),
		))
		commitParity(t, document, "padded")
		expectedHeads := hashIDs(headsOf(t, document))
		expected := hydrateOf(t, document)

		incremental, err := document.SaveIncremental()
		require.NoError(t, err)
		fromChange := openParity(t, engine, actor(0x73))
		_, err = fromChange.LoadIncremental(incremental)
		require.NoError(t, err)

		compressed, err := document.Save()
		require.NoError(t, err)
		fromCompressed := loadParity(t, engine, compressed, actor(0x74))
		plain, err := document.Save(automerge.NoCompress())
		require.NoError(t, err)
		fromPlain := loadParity(t, engine, plain, actor(0x75))

		return parityView{
			Heads:   expectedHeads,
			Hydrate: expected,
			Saved:   hydrateOf(t, fromChange),
			Message: hashJoin(
				hashIDs(headsOf(t, fromChange)),
				hashIDs(headsOf(t, fromCompressed)),
				hashIDs(headsOf(t, fromPlain)),
			),
		}
	})
}
func TestRustTest_TransactionPatchesSplitBlock(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x82))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "👩‍🚒"))
		commitParity(t, document, "emoji")
		before := headsOf(t, document)
		_, err = text.SplitBlock(3)
		require.NoError(t, err)
		commitParity(t, document, "split")
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches}
	})
}
func TestRustTest_PatchesExposeSurvivingConflictAfterDeletingOtherBranch(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x93))
		require.NoError(t, document.Root().PutValue("value", automerge.Value{
			Type: automerge.ValueTypeText,
			Text: "a",
		}))
		commitParity(t, document, "value")
		left, err := document.Fork(actor(0x93))
		require.NoError(t, err)
		closeDocument(t, left)
		leftText, err := left.Root().Object("value")
		require.NoError(t, err)
		require.NoError(t, leftText.PutScalarAt(0, automerge.StringScalar("🦊🐻")))
		commitParity(t, left, "left")

		right, err := document.Fork(actor(0x94))
		require.NoError(t, err)
		closeDocument(t, right)
		rightText, err := right.Root().Object("value")
		require.NoError(t, err)
		require.NoError(t, rightText.PutScalarAt(0, automerge.BytesScalar(bytes.Repeat([]byte{0}, 32))))
		commitParity(t, right, "right")

		_, err = document.Merge(left)
		require.NoError(t, err)
		_, err = document.Merge(right)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, rightText.DeleteIndex(0))
		commitParity(t, right, "delete")
		_, err = document.Merge(right)
		require.NoError(t, err)
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

func exchangeDocuments(
	t *testing.T,
	left *automerge.Document,
	right *automerge.Document,
) {
	t.Helper()

	leftState, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftState)

	rightState, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightState)
	synchronize(t, leftState, rightState)
}

// TestRustTest_MergePatchesClearConflict reproduces
// merge_patches_clear_conflict_after_losing_list_value_is_deleted_from_fuzz_trace.
func TestRustTest_MergePatchesClearConflict(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x02))
		list, err := document.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(0, automerge.NullScalar()))
		commitParity(t, document, "list")

		left, err := document.Fork(actor(0x03))
		require.NoError(t, err)
		closeDocument(t, left)
		leftList, err := left.Root().Object("list")
		require.NoError(t, err)
		require.NoError(t, leftList.PutScalarAt(0, automerge.IntScalar(10)))
		commitParity(t, left, "ten")

		right, err := document.Fork(actor(0x02))
		require.NoError(t, err)
		closeDocument(t, right)
		rightList, err := right.Root().Object("list")
		require.NoError(t, err)
		require.NoError(t, rightList.PutScalarAt(0, automerge.StringScalar("a")))
		commitParity(t, right, "a")

		_, err = document.Merge(left)
		require.NoError(t, err)
		_, err = document.Merge(right)
		require.NoError(t, err)
		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, rightList.DeleteIndex(0))
		commitParity(t, right, "delete")
		_, err = document.Merge(right)
		require.NoError(t, err)
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_ForkAtRejectsEditOutsideIsolation reproduces
// fork_at_rejects_edit_to_object_outside_isolation_from_fuzz_trace.
func TestRustTest_ForkAtRejectsEditOutsideIsolation(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x21))
		list, err := document.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		commitParity(t, document, "list")
		require.NoError(t, document.Isolate(nil))

		insertErr := list.InsertScalar(0, automerge.IntScalar(1))

		fork, forkErr := document.Fork(actor(0x22))
		if forkErr == nil {
			closeDocument(t, fork)
		}

		return parityView{
			Failed:  insertErr != nil,
			Count:   boolCount(forkErr == nil),
			Hydrate: hydrateOf(t, document),
		}
	})
}

func boolCount(value bool) uint64 {
	if value {
		return 1
	}

	return 0
}

// TestRustTest_DiffIncrementalRespectsIsolatedHeads reproduces
// diff_incremental_respects_isolated_heads_from_fuzz_trace.
func TestRustTest_DiffIncrementalRespectsIsolatedHeads(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x23))
		_, err := document.Root().CreateObject("map", automerge.ObjectTypeMap)
		require.NoError(t, err)
		commitParity(t, document, "map")
		require.NoError(t, document.Isolate(nil))
		patches, diffErr := document.DiffIncremental()

		return parityView{
			Failed:  diffErr != nil,
			Patches: patches,
			Hydrate: hydrateOf(t, document),
		}
	})
}

// TestRustTest_TransactionPatchesReplaceMultiCharacterString reproduces
// transaction_patches_replace_multi_character_string_in_text_from_fuzz_trace.
func TestRustTest_TransactionPatchesReplaceMultiCharacterString(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x24))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		require.NoError(t, textObject.InsertScalar(0, automerge.StringScalar("ab")))
		commitParity(t, document, "ab")
		before := headsOf(t, document)
		require.NoError(t, textObject.PutScalarAt(1, automerge.FloatScalar(math.MaxFloat64)))
		commitParity(t, document, "replace")
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_IncrementalPutSeqInText reproduces
// incremental_put_seq_in_text_applies_to_hydrated_value_from_fuzz_trace.
func TestRustTest_IncrementalPutSeqInText(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x25))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "a"))
		commitParity(t, document, "a")
		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, textObject.PutScalarAt(0, automerge.NullScalar()))
		commitParity(t, document, "null")
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_DiffSpansPreserveMarksOnInsertedText reproduces
// diff_spans_preserve_marks_on_inserted_text.
func TestRustTest_DiffSpansPreserveMarksOnInsertedText(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x26))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		commitParity(t, document, "text")
		before := headsOf(t, document)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "abc"))
		require.NoError(t, text.Mark(0, 3, "bold", automerge.IntScalar(1), automerge.MarkExpandAfter))
		commitParity(t, document, "marked")
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_DiffFromIntegratedToIsolatedHeads reproduces
// diff_from_integrated_to_isolated_heads_reconstructs_target.
func TestRustTest_DiffFromIntegratedToIsolatedHeads(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x7f))
		_, err := document.Root().CreateObject("k8", automerge.ObjectTypeMap)
		require.NoError(t, err)
		commitParity(t, document, "map")
		require.NoError(t, document.Isolate(nil))
		require.NoError(t, document.Root().PutScalar("k8", automerge.IntScalar(1000)))
		commitParity(t, document, "scalar")
		isolated := headsOf(t, document)
		require.NoError(t, document.Integrate())
		patches, err := document.Diff(headsOf(t, document), isolated)
		require.NoError(t, err)

		return parityView{
			Patches: patches,
			Hydrate: hydrateOf(t, document),
			Heads:   hashIDs(isolated),
		}
	})
}

// TestRustTest_IncrementalDiffDiscardsEditsInsideDeletedMap reproduces
// incremental_diff_discards_edits_inside_deleted_map.
func TestRustTest_IncrementalDiffDiscardsEditsInsideDeletedMap(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x27))
		container, err := document.Root().CreateObject("container", automerge.ObjectTypeMap)
		require.NoError(t, err)
		textObject, err := container.CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		commitParity(t, document, "container")
		_, err = document.DiffIncremental()
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "hello"))
		commitParity(t, document, "hello")
		beforeDelete := headsOf(t, document)
		require.NoError(t, document.Root().DeleteKey("container"))
		commitParity(t, document, "delete")
		require.NoError(t, document.Isolate(nil))
		require.NoError(t, document.Isolate(beforeDelete))
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_IncrementalPatchesSurviveNestedEditAndListShift reproduces
// incremental_patches_survive_nested_edit_and_list_shift_across_isolation.
func TestRustTest_IncrementalPatchesSurviveNestedEditAndListShift(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x28))
		items, err := document.Root().CreateObject("items", automerge.ObjectTypeList)
		require.NoError(t, err)
		first, err := items.InsertObject(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(t, first.PutScalar("id", automerge.StringScalar("a")))
		require.NoError(t, first.PutScalar("n", automerge.IntScalar(1)))

		second, err := items.InsertObject(1, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(t, second.PutScalar("id", automerge.StringScalar("b")))
		require.NoError(t, second.PutScalar("n", automerge.IntScalar(2)))
		commitParity(t, document, "items")
		beginning := headsOf(t, document)

		inserted, err := items.InsertObject(0, automerge.ObjectTypeMap)
		require.NoError(t, err)
		require.NoError(t, inserted.PutScalar("id", automerge.StringScalar("z")))
		require.NoError(t, inserted.PutScalar("n", automerge.IntScalar(0)))
		commitParity(t, document, "shift")
		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, document.Isolate(beginning))
		require.NoError(t, first.PutScalar("n", automerge.IntScalar(99)))
		commitParity(t, document, "edit")
		require.NoError(t, document.Integrate())
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_IncrementalListConflictSurvivesIsolateIntegrate reproduces
// incremental_list_conflict_survives_isolate_integrate_roundtrip.
func TestRustTest_IncrementalListConflictSurvivesIsolateIntegrate(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x29))
		list, err := document.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(0, automerge.StringScalar("base")))
		commitParity(t, document, "base")
		beginning := headsOf(t, document)
		first, err := list.PutObjectAt(0, automerge.ObjectTypeText)
		require.NoError(t, err)
		firstText, err := first.Text()
		require.NoError(t, err)
		require.NoError(t, firstText.Splice(0, 0, "first"))
		commitParity(t, document, "first")
		_, err = document.DiffIncremental()
		require.NoError(t, err)
		require.NoError(t, document.Isolate(beginning))

		second, err := list.PutObjectAt(0, automerge.ObjectTypeText)
		require.NoError(t, err)
		secondText, err := second.Text()
		require.NoError(t, err)
		require.NoError(t, secondText.Splice(0, 0, "second"))
		commitParity(t, document, "second")
		require.NoError(t, document.Integrate())
		patches, err := document.DiffIncremental()
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_DiffToIsolatedBatchCreatedText reproduces
// diff_to_isolated_batch_created_text_applies_to_hydrated_value_from_fuzz_trace.
func TestRustTest_DiffToIsolatedBatchCreatedText(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x7f))
		peer, err := document.Fork(actor(0x7f))
		require.NoError(t, err)
		closeDocument(t, peer)
		_, err = peer.Root().CreateObject("k4", automerge.ObjectTypeText)
		require.NoError(t, err)
		commitParity(t, peer, "text")
		exchangeDocuments(t, document, peer)
		require.NoError(t, document.Isolate(nil))
		require.NoError(t, document.Root().PutValue("k4", automerge.Value{
			Type: automerge.ValueTypeText,
			Text: "hello world",
		}))
		commitParity(t, document, "hello")
		isolated := headsOf(t, document)
		require.NoError(t, document.Integrate())
		patches, err := document.Diff(headsOf(t, document), isolated)
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_DiffExposesListObjectWhenConflictWinnerIsRemoved reproduces
// diff_exposes_list_object_when_conflict_winner_is_removed.
func TestRustTest_DiffExposesListObjectWhenConflictWinnerIsRemoved(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x7f))
		list, err := document.Root().CreateObject("list", automerge.ObjectTypeList)
		require.NoError(t, err)
		require.NoError(t, list.InsertScalar(0, automerge.StringScalar("base")))
		commitParity(t, document, "base")
		base := headsOf(t, document)
		peer, err := document.Fork(actor(0x30))
		require.NoError(t, err)
		closeDocument(t, peer)
		peerList, err := peer.Root().Object("list")
		require.NoError(t, err)
		_, err = peerList.PutObjectAt(0, automerge.ObjectTypeText)
		require.NoError(t, err)
		commitParity(t, peer, "object")
		_, err = document.Merge(peer)
		require.NoError(t, err)
		require.NoError(t, document.Isolate(base))
		require.NoError(t, list.PutValueAt(0, automerge.Value{
			Type: automerge.ValueTypeText,
			Text: "hello world",
		}))
		commitParity(t, document, "hello")
		isolated := headsOf(t, document)
		require.NoError(t, document.Integrate())
		patches, err := document.Diff(headsOf(t, document), isolated)
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_TransactionPatchesInsertHiddenScalar reproduces
// transaction_patches_insert_hidden_scalar_into_utf8_text_from_fuzz_trace.
func TestRustTest_TransactionPatchesInsertHiddenScalar(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x31))
		require.NoError(t, document.Root().PutValue("k31", automerge.Value{
			Type: automerge.ValueTypeText,
			Text: "👩‍🚒",
		}))
		commitParity(t, document, "emoji")
		textObject, err := document.Root().Object("k31")
		require.NoError(t, err)
		length, err := textObject.Len()
		require.NoError(t, err)
		before := headsOf(t, document)
		require.NoError(t, textObject.InsertScalar(195%(length+1), automerge.UintScalar(150)))
		commitParity(t, document, "scalar")
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_TransactionPatchesMark reproduces
// transaction_patches_mark_at_normalized_utf8_indexes.
func TestRustTest_TransactionPatchesMark(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		document := openParity(t, engine, actor(0x32))
		textObject, err := document.Root().CreateObject("text", automerge.ObjectTypeText)
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)
		require.NoError(t, text.Splice(0, 0, "👩‍🚒"))
		commitParity(t, document, "emoji")
		before := headsOf(t, document)
		require.NoError(t, text.Mark(
			3,
			5,
			"bold",
			automerge.BoolScalar(true),
			automerge.MarkExpandNone,
		))
		commitParity(t, document, "mark")
		patches, err := document.Diff(before, headsOf(t, document))
		require.NoError(t, err)

		return parityView{Patches: patches, Hydrate: hydrateOf(t, document)}
	})
}

// TestRustTest_QueuedChangeDoesNotCollideWithLaterLocalChange reproduces
// queued_change_does_not_collide_with_later_local_change.
func TestRustTest_QueuedChangeDoesNotCollideWithLaterLocalChange(t *testing.T) {
	t.Parallel()

	agreeView(t, func(t *testing.T, engine rustParityEngine) parityView {
		local := openParity(t, engine, actor(0x7f))
		require.NoError(t, local.Root().PutScalar("base", automerge.IntScalar(0)))
		commitParity(t, local, "base")

		remote, err := local.Fork(actor(0x7f))
		require.NoError(t, err)
		closeDocument(t, remote)
		require.NoError(t, local.Root().PutScalar("local-2", automerge.IntScalar(2)))
		commitParity(t, local, "local-2")
		resetSaveCursor(t, remote)
		require.NoError(t, remote.Root().PutScalar("remote-2", automerge.IntScalar(2)))
		commitParity(t, remote, "remote-2")
		remote2 := takeIncremental(t, remote)
		require.NoError(t, remote.Root().PutScalar("remote-3", automerge.IntScalar(3)))
		commitParity(t, remote, "remote-3")
		remote3 := takeIncremental(t, remote)
		dependent, err := remote.Fork(actor(0x03))
		require.NoError(t, err)
		closeDocument(t, dependent)
		resetSaveCursor(t, dependent)
		require.NoError(t, dependent.Root().PutScalar("dependent", automerge.IntScalar(4)))
		commitParity(t, dependent, "dependent")
		dependentChange := takeIncremental(t, dependent)

		_, err = local.LoadIncremental(append(append([]byte{}, remote3...), dependentChange...))
		require.NoError(t, err)

		_, applyErr := local.LoadIncremental(remote2)
		missing, err := local.MissingDependencies(nil)
		require.NoError(t, err)
		require.NoError(t, local.Root().PutScalar("local-3", automerge.IntScalar(3)))
		commitParity(t, local, "local-3")
		saved, err := local.Save(automerge.NoCompress())
		require.NoError(t, err)
		loaded := loadParity(t, engine, saved, actor(0x33))

		return parityView{
			Failed:  applyErr != nil,
			Missing: hashIDs(missing),
			Hydrate: hydrateOf(t, local),
			Saved:   hydrateOf(t, loaded),
		}
	})
}

func hashJoin(groups ...[]string) string {
	parts := make([]string, len(groups))
	for index, group := range groups {
		parts[index] = strings.Join(group, ",")
	}

	return strings.Join(parts, ";")
}
