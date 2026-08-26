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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func TestPureGoDocument_ReferenceLoadsNativeHistory(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(40))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Hello 😀"))
	_, err = nativeDocument.Commit("Create native document", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save()
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(data, actor(41))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	value, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, "Hello 😀", value)
}

func TestPureGoDocument_ExtendsReferenceSnapshot(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.Load(newBaseDocument(t), actor(42))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(5, 0, " native"))
	_, err = nativeDocument.Commit("Extend in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save()
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(data, actor(43))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	value, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, "Hello native", value)
}

func TestPureGoDocument_InsertsInsideReferenceText(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.Load(newBaseDocument(t), actor(46))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(2, 0, "X"))
	_, err = nativeDocument.Commit("Insert in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save()
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(data, actor(47))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	value, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, "HeXllo", value)
}

func TestPureGoDocument_DeletesReferenceText(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.Load(newBaseDocument(t), actor(48))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(2, 2, ""))
	_, err = nativeDocument.Commit("Delete in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save()
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(data, actor(49))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	value, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, "Heo", value)
}

func TestPureGoDocument_EmptySnapshotLoadsInReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(44))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	data, err := nativeDocument.Save()
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(data, actor(45))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	heads, err := referenceDocument.Heads()
	require.NoError(t, err)
	assert.Empty(t, heads)
}

func TestPureGoDocument_ReusesActorAfterLoad(t *testing.T) {
	t.Parallel()

	actorID := actor(55)
	document, err := automerge.New(actorID)
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "A"))
	_, err = document.Commit("First change", commitTime)
	require.NoError(t, err)
	data, err := document.Save()
	require.NoError(t, err)

	loaded, err := automerge.Load(data, actorID)
	require.NoError(t, err)
	closeDocument(t, loaded)
	loadedText, err := loaded.Text("body")
	require.NoError(t, err)
	require.NoError(t, loadedText.Splice(1, 0, "B"))
	_, err = loaded.Commit("Second change", commitTime)
	require.NoError(t, err)
	data, err = loaded.Save()
	require.NoError(t, err)

	reference, err := automerge.LoadReference(data, actor(56))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceText, err := reference.Text("body")
	require.NoError(t, err)
	value, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, "AB", value)
}

func TestPureGoDocument_ConcurrentChangesConverge(t *testing.T) {
	t.Parallel()

	base, err := automerge.New(actor(50))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(0, 0, "A"))
	_, err = base.Commit("Create body", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save()
	require.NoError(t, err)

	left, err := automerge.Load(baseData, actor(51))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text("body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(1, 0, "L"))
	_, err = left.Commit("Edit left", commitTime)
	require.NoError(t, err)

	right, err := automerge.Load(baseData, actor(52))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.Text("body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(1, 0, "R"))
	_, err = right.Commit("Edit right", commitTime)
	require.NoError(t, err)

	_, err = left.Merge(right)
	require.NoError(t, err)
	_, err = right.Merge(left)
	require.NoError(t, err)

	leftValue, err := leftText.String()
	require.NoError(t, err)
	rightValue, err := rightText.String()
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)

	leftHeads, err := left.Heads()
	require.NoError(t, err)
	rightHeads, err := right.Heads()
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)
}

func TestPureGoDocument_CursorMatchesReference(t *testing.T) {
	t.Parallel()

	baseData := newBaseDocument(t)
	nativeDocument, err := automerge.Load(baseData, actor(60))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text("body")
	require.NoError(t, err)
	nativeCursor, err := nativeText.Cursor(2)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(baseData, actor(61))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	referenceCursor, err := referenceText.Cursor(2)
	require.NoError(t, err)
	assert.Equal(t, referenceCursor, nativeCursor)

	require.NoError(t, nativeText.Splice(0, 0, "X"))
	position, err := nativeText.CursorPosition(nativeCursor)
	require.NoError(t, err)
	assert.Equal(t, uint32(3), position)
}

func TestPureGoDocument_DeletedCursorMatchesReference(t *testing.T) {
	t.Parallel()

	baseData := newBaseDocument(t)
	nativeDocument, err := automerge.Load(baseData, actor(62))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text("body")
	require.NoError(t, err)
	cursor, err := nativeText.Cursor(2)
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(2, 1, ""))
	_, err = nativeDocument.Commit("Delete cursor target", commitTime)
	require.NoError(t, err)
	nativePosition, err := nativeText.CursorPosition(cursor)
	require.NoError(t, err)

	data, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceDocument, err := automerge.LoadReference(data, actor(63))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)
	referencePosition, err := referenceText.CursorPosition(cursor)
	require.NoError(t, err)
	assert.Equal(t, referencePosition, nativePosition)
}

func TestPureGoDocument_UTF16CursorBoundariesMatchReference(t *testing.T) {
	t.Parallel()

	base, err := automerge.NewReference(actor(122))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(0, 0, "A😀B"))
	_, err = base.Commit("Create emoji text", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save()
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(baseData, actor(123))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text("body")
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(baseData, actor(124))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)

	nativeInside, nativeErr := nativeText.Cursor(2)
	referenceInside, referenceErr := referenceText.Cursor(2)

	require.NoError(t, nativeErr)
	require.NoError(t, referenceErr)
	require.Equal(t, referenceInside, nativeInside)
	nativeInsidePosition, err := nativeText.CursorPosition(nativeInside)
	require.NoError(t, err)
	referenceInsidePosition, err := referenceText.CursorPosition(

		referenceInside,
	)
	require.NoError(t, err)
	assert.Equal(t, referenceInsidePosition, nativeInsidePosition)

	for _, index := range []uint32{1, 3} {
		nativeCursor, err := nativeText.Cursor(index)
		require.NoError(t, err)
		referenceCursor, err := referenceText.Cursor(index)
		require.NoError(t, err)
		require.Equal(t, referenceCursor, nativeCursor)

		require.NoError(t, nativeText.Splice(0, 0, "X"))
		require.NoError(t, referenceText.Splice(0, 0, "X"))
		nativePosition, err := nativeText.CursorPosition(nativeCursor)
		require.NoError(t, err)
		referencePosition, err := referenceText.CursorPosition(

			referenceCursor,
		)
		require.NoError(t, err)
		assert.Equal(t, referencePosition, nativePosition)

		require.NoError(t, nativeText.Splice(0, 1, ""))
		require.NoError(t, referenceText.Splice(0, 1, ""))
	}
}

func TestPureGoDocument_CursorModesMatchReference(t *testing.T) {
	t.Parallel()

	base, err := automerge.NewReference(actor(146))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(0, 0, "A😀B"))
	_, err = base.Commit("cursor base", commitTime)
	require.NoError(t, err)
	data, err := base.Save()
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(data, actor(147))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text("body")
	require.NoError(t, err)
	referenceDocument, err := automerge.LoadReference(data, actor(148))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)

	for _, index := range []int64{-1, 0, 1, 2, 3, 4, 100} {
		for _, movement := range []automerge.CursorMove{
			automerge.CursorMoveBefore,
			automerge.CursorMoveAfter,
		} {
			nativeCursor, err := nativeText.CursorFor(index, movement)
			require.NoError(t, err)
			referenceCursor, err := referenceText.CursorFor(

				index,
				movement,
			)
			require.NoError(t, err)
			assert.Equal(t, referenceCursor, nativeCursor)
			nativePosition, err := nativeText.CursorPosition(

				nativeCursor,
			)
			require.NoError(t, err)
			referencePosition, err := referenceText.CursorPosition(

				referenceCursor,
			)
			require.NoError(t, err)
			assert.Equal(t, referencePosition, nativePosition)
		}
	}

	nativeCursor, err := nativeText.CursorFor(

		1,
		automerge.CursorMoveAfter,
	)
	require.NoError(t, err)
	referenceCursor, err := referenceText.CursorFor(

		1,
		automerge.CursorMoveAfter,
	)
	require.NoError(t, err)
	nativeBefore, err := nativeText.CursorFor(

		1,
		automerge.CursorMoveBefore,
	)
	require.NoError(t, err)
	referenceBefore, err := referenceText.CursorFor(

		1,
		automerge.CursorMoveBefore,
	)
	require.NoError(t, err)
	require.NoError(t, nativeText.SpliceCursor(nativeCursor, 2, "X"))
	require.NoError(t, referenceText.SpliceCursor(referenceCursor, 2, "X"))
	nativeValue, err := nativeText.String()
	require.NoError(t, err)
	referenceValue, err := referenceText.String()
	require.NoError(t, err)
	assert.Equal(t, referenceValue, nativeValue)
	assert.Equal(t, "AXB", nativeValue)

	for _, cursors := range [][2]automerge.Cursor{
		{nativeBefore, referenceBefore},
		{nativeCursor, referenceCursor},
	} {
		nativePosition, err := nativeText.CursorPosition(cursors[0])
		require.NoError(t, err)
		referencePosition, err := referenceText.CursorPosition(cursors[1])
		require.NoError(t, err)
		assert.Equal(t, referencePosition, nativePosition)
	}
}

func TestPureGoDocument_MarkAuthoringMatchesReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(153))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(153))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeText, err := nativeDocument.CreateText("body")
	require.NoError(t, err)
	referenceText, err := referenceDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(0, 0, "ABCD"))
	require.NoError(t, referenceText.Splice(0, 0, "ABCD"))
	_, err = nativeDocument.Commit("create text", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit("create text", commitTime)
	require.NoError(t, err)

	strong := automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}
	require.NoError(
		t,
		nativeText.Mark(

			0,
			4,
			"strong",
			strong,
			automerge.MarkExpandBoth,
		),
	)
	require.NoError(
		t,
		referenceText.Mark(

			0,
			4,
			"strong",
			strong,
			automerge.MarkExpandBoth,
		),
	)
	_, err = nativeDocument.Commit("mark", commitTime.Add(time.Second))
	require.NoError(t, err)
	_, err = referenceDocument.Commit("mark", commitTime.Add(time.Second))
	require.NoError(t, err)
	nativeMarkedSpans, err := nativeText.Spans()
	require.NoError(t, err)
	referenceMarkedSpans, err := referenceText.Spans()
	require.NoError(t, err)
	assert.Equal(t, referenceMarkedSpans, nativeMarkedSpans)

	require.NoError(
		t,
		nativeText.Unmark(

			1,
			3,
			"strong",
			automerge.MarkExpandNone,
		),
	)
	require.NoError(
		t,
		referenceText.Unmark(

			1,
			3,
			"strong",
			automerge.MarkExpandNone,
		),
	)
	_, err = nativeDocument.Commit("unmark", commitTime.Add(2*time.Second))
	require.NoError(t, err)
	_, err = referenceDocument.Commit(

		"unmark",
		commitTime.Add(2*time.Second),
	)
	require.NoError(t, err)

	nativeSpans, err := nativeText.Spans()
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans()
	require.NoError(t, err)
	assert.Equal(t, referenceSpans, nativeSpans)

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(154),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	referenceFromNativeText, err := referenceFromNative.Text("body")
	require.NoError(t, err)
	referenceFromNativeSpans, err := referenceFromNativeText.Spans()
	require.NoError(t, err)
	assert.Equal(t, nativeSpans, referenceFromNativeSpans)

	referenceData, err := referenceDocument.Save()
	require.NoError(t, err)
	nativeFromReference, err := automerge.Load(

		referenceData,
		actor(155),
	)
	require.NoError(t, err)
	closeDocument(t, nativeFromReference)
	nativeFromReferenceText, err := nativeFromReference.Text("body")
	require.NoError(t, err)
	nativeFromReferenceSpans, err := nativeFromReferenceText.Spans()
	require.NoError(t, err)
	assert.Equal(t, referenceSpans, nativeFromReferenceSpans)
}

func TestPureGoDocument_BlockAuthoringMatchesReference(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(167))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(actor(167))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeText, err := nativeDocument.CreateText("body")
	require.NoError(t, err)
	referenceText, err := referenceDocument.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(0, 0, "AB"))
	require.NoError(t, referenceText.Splice(0, 0, "AB"))

	nativeFirst, err := nativeText.SplitBlock(0)
	require.NoError(t, err)
	referenceFirst, err := referenceText.SplitBlock(0)
	require.NoError(t, err)
	setBlockAttributes(t, nativeFirst, "paragraph")
	setBlockAttributes(t, referenceFirst, "paragraph")
	nativeSecond, err := nativeText.SplitBlock(2)
	require.NoError(t, err)
	referenceSecond, err := referenceText.SplitBlock(2)
	require.NoError(t, err)
	setBlockAttributes(t, nativeSecond, "heading")
	setBlockAttributes(t, referenceSecond, "heading")
	_, err = nativeDocument.Commit("blocks", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit("blocks", commitTime)
	require.NoError(t, err)
	assertTextSpansEqual(t, nativeText, referenceText)

	nativeReplacement, err := nativeText.ReplaceBlock(2)
	require.NoError(t, err)
	referenceReplacement, err := referenceText.ReplaceBlock(2)
	require.NoError(t, err)
	setBlockAttributes(t, nativeReplacement, "blockquote")
	setBlockAttributes(t, referenceReplacement, "blockquote")
	require.NoError(t, nativeText.JoinBlock(0))
	require.NoError(t, referenceText.JoinBlock(0))
	_, err = nativeDocument.Commit(

		"replace and join",
		commitTime.Add(time.Second),
	)
	require.NoError(t, err)
	_, err = referenceDocument.Commit(

		"replace and join",
		commitTime.Add(time.Second),
	)
	require.NoError(t, err)
	assertTextSpansEqual(t, nativeText, referenceText)

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(168),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	referenceFromNativeText, err := referenceFromNative.Text("body")
	require.NoError(t, err)
	assertTextSpansEqual(t, nativeText, referenceFromNativeText)
}

func setBlockAttributes(
	t *testing.T,
	block *automerge.Object,
	blockType string,
) {
	t.Helper()

	require.NoError(
		t,
		block.PutScalar(

			"type",
			automerge.Scalar{
				Type:   automerge.ScalarTypeString,
				String: blockType,
			},
		),
	)
	_, err := block.CreateObject("parents", automerge.ObjectTypeList)
	require.NoError(t, err)
	_, err = block.CreateObject("attrs", automerge.ObjectTypeMap)
	require.NoError(t, err)
	require.NoError(
		t,
		block.PutScalar(

			"isEmbed",
			automerge.Scalar{Type: automerge.ScalarTypeBoolean},
		),
	)
}

func assertTextSpansEqual(
	t *testing.T,
	left *automerge.Text,
	right *automerge.Text,
) {
	t.Helper()

	leftSpans, err := left.Spans()
	require.NoError(t, err)
	rightSpans, err := right.Spans()
	require.NoError(t, err)
	assert.Equal(t, rightSpans, leftSpans)
}

func TestPureGoDocument_SynchronizesWithNativePeer(t *testing.T) {
	t.Parallel()

	left, err := automerge.New(actor(70))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Native sync"))
	_, err = left.Commit("Create sync body", commitTime)
	require.NoError(t, err)

	right, err := automerge.New(actor(71))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text("body")
	require.NoError(t, err)
	value, err := rightText.String()
	require.NoError(t, err)
	assert.Equal(t, "Native sync", value)
}

func TestPureGoDocument_SyncWaitsForPeerResponse(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(78))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Wait"))
	_, err = document.Commit("Create body", commitTime)
	require.NoError(t, err)

	state, err := document.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, state)
	first, ok, err := state.GenerateMessage()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NotEmpty(t, first)

	second, ok, err := state.GenerateMessage()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, second)
}

func TestPureGoDocument_SynchronizesWithReferencePeer(t *testing.T) {
	t.Parallel()

	left, err := automerge.New(actor(72))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Reference sync"))
	_, err = left.Commit("Create sync body", commitTime)
	require.NoError(t, err)

	right, err := automerge.NewReference(actor(73))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text("body")
	require.NoError(t, err)
	value, err := rightText.String()
	require.NoError(t, err)
	assert.Equal(t, "Reference sync", value)
}

func TestPureGoDocument_ReceivesReferenceSync(t *testing.T) {
	t.Parallel()

	left, err := automerge.NewReference(actor(74))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "From reference"))
	_, err = left.Commit("Create reference body", commitTime)
	require.NoError(t, err)

	right, err := automerge.New(actor(75))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text("body")
	require.NoError(t, err)
	value, err := rightText.String()
	require.NoError(t, err)
	assert.Equal(t, "From reference", value)
}

func TestPureGoDocument_RepeatedMixedPeerSync(t *testing.T) {
	t.Parallel()

	nativeDocument, err := automerge.New(actor(76))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.CreateText("body")
	require.NoError(t, err)
	_, err = nativeDocument.Commit("Create native body", commitTime)
	require.NoError(t, err)

	referenceDocument, err := automerge.NewReference(actor(77))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeSync, err := nativeDocument.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, nativeSync)

	referenceSync, err := referenceDocument.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, referenceSync)
	synchronize(t, nativeSync, referenceSync)

	referenceText, err := referenceDocument.Text("body")
	require.NoError(t, err)

	for iteration := range 20 {
		nativeValue, err := nativeText.String()
		require.NoError(t, err)

		nativeOffset := utf16Offsets([]rune(nativeValue))
		require.NoError(
			t,
			nativeText.Splice(

				nativeOffset[len(nativeOffset)-1],
				0,
				"N",
			),
		)
		_, err = nativeDocument.Commit("Native edit", commitTime)
		require.NoError(t, err)

		referenceValue, err := referenceText.String()
		require.NoError(t, err)

		referenceOffset := utf16Offsets([]rune(referenceValue))
		require.NoError(
			t,
			referenceText.Splice(

				referenceOffset[len(referenceOffset)-1],
				0,
				"R",
			),
		)
		_, err = referenceDocument.Commit("Reference edit", commitTime)
		require.NoError(t, err)

		synchronize(t, nativeSync, referenceSync)

		nativeValue, err = nativeText.String()
		require.NoError(t, err)
		referenceValue, err = referenceText.String()
		require.NoError(t, err)
		assert.Equal(t, referenceValue, nativeValue, "iteration %d", iteration)
	}
}

func TestPureGoDocument_ReferenceEditsWhileMessageInFlight(t *testing.T) {
	t.Parallel()

	client, err := automerge.NewReference(actor(79))
	require.NoError(t, err)
	closeDocument(t, client)
	clientText, err := client.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, clientText.Splice(0, 0, "A"))
	_, err = client.Commit("initial", commitTime)
	require.NoError(t, err)

	server, err := automerge.New(actor(80))
	require.NoError(t, err)
	closeDocument(t, server)

	clientSync, err := client.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, clientSync)

	serverSync, err := server.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, serverSync)
	synchronize(t, clientSync, serverSync)

	require.NoError(t, clientText.Splice(1, 0, "B"))
	_, err = client.Commit("first edit", commitTime)
	require.NoError(t, err)
	first, ok, err := clientSync.GenerateMessage()
	require.NoError(t, err)
	require.True(t, ok)

	require.NoError(t, clientText.Splice(2, 0, "C"))
	_, err = client.Commit("second edit", commitTime)
	require.NoError(t, err)

	require.NoError(t, serverSync.ReceiveMessage(first))
	ack, ok, err := serverSync.GenerateMessage()
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, clientSync.ReceiveMessage(ack))

	second, ok, err := clientSync.GenerateMessage()
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, serverSync.ReceiveMessage(second))
	synchronize(t, clientSync, serverSync)

	serverText, err := server.Text("body")
	require.NoError(t, err)
	value, err := serverText.String()
	require.NoError(t, err)
	assert.Equal(t, "ABC", value)
}

func TestSyncState_ReadOnlyParity(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sourceReference bool
	}{
		"native publisher":    {sourceReference: false},
		"reference publisher": {sourceReference: true},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				var (
					source *automerge.Document
					target *automerge.Document
					err    error
				)
				if test.sourceReference {
					source, err = automerge.NewReference(actor(181))
					require.NoError(t, err)
					target, err = automerge.New(actor(182))
				} else {
					source, err = automerge.New(actor(181))
					require.NoError(t, err)
					target, err = automerge.NewReference(actor(182))
				}

				require.NoError(t, err)
				closeDocument(t, source)
				closeDocument(t, target)

				text, err := source.CreateText("body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(0, 0, "Published"))
				_, err = source.Commit("publish", commitTime)
				require.NoError(t, err)

				sourceState, err := source.NewSyncState()
				require.NoError(t, err)
				closeSyncState(t, sourceState)

				targetState, err := target.NewSyncState()
				require.NoError(t, err)
				closeSyncState(t, targetState)
				require.NoError(t, targetState.SetReadOnly(true))

				synchronize(t, sourceState, targetState)

				peerReadOnly, err := sourceState.PeerReadOnly()
				require.NoError(t, err)
				assert.True(t, peerReadOnly)

				_, err = target.Text("body")
				require.Error(t, err)

				require.NoError(t, targetState.SetReadOnly(false))
				synchronize(t, sourceState, targetState)

				peerReadOnly, err = sourceState.PeerReadOnly()
				require.NoError(t, err)
				assert.False(t, peerReadOnly)

				targetText, err := target.Text("body")
				require.NoError(t, err)
				value, err := targetText.String()
				require.NoError(t, err)
				assert.Equal(t, "Published", value)
			},
		)
	}
}

func TestSyncState_ReadOnlyModeOverridesInFlight(t *testing.T) {
	t.Parallel()

	tests := map[string]func(
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(actor(183))
				require.NoError(t, err)
				closeDocument(t, document)
				text, err := document.CreateText("body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(0, 0, "A"))
				_, err = document.Commit("initial", commitTime)
				require.NoError(t, err)
				state, err := document.NewSyncState()
				require.NoError(t, err)
				closeSyncState(t, state)

				_, ok, err := state.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, state.SetReadOnly(true))
				_, ok, err = state.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, state.SetReadOnly(false))
				_, ok, err = state.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
			},
		)
	}
}

func TestSyncState_BothReadOnlyResumeConvergence(t *testing.T) {
	t.Parallel()

	left, err := automerge.New(actor(184))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(0, 0, "L"))
	_, err = left.Commit("left", commitTime)
	require.NoError(t, err)

	right, err := automerge.NewReference(actor(185))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(0, 0, "R"))
	_, err = right.Commit("right", commitTime)
	require.NoError(t, err)

	leftState, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftState)

	rightState, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightState)
	require.NoError(t, leftState.SetReadOnly(true))
	require.NoError(t, rightState.SetReadOnly(true))
	synchronize(t, leftState, rightState)

	leftValue, err := leftText.String()
	require.NoError(t, err)
	rightValue, err := rightText.String()
	require.NoError(t, err)
	assert.Equal(t, "L", leftValue)
	assert.Equal(t, "R", rightValue)

	require.NoError(t, leftState.SetReadOnly(false))
	require.NoError(t, rightState.SetReadOnly(false))
	synchronize(t, leftState, rightState)

	leftHeads, err := left.Heads()
	require.NoError(t, err)
	rightHeads, err := right.Heads()
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)

	leftText, err = left.Text("body")
	require.NoError(t, err)
	leftValue, err = leftText.String()
	require.NoError(t, err)
	rightText, err = right.Text("body")
	require.NoError(t, err)
	rightValue, err = rightText.String()
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)
}
