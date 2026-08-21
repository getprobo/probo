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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestPureGoDocument_ReferenceLoadsNativeHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(40))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Hello 😀"))
	_, err = nativeDocument.Commit(ctx, "Create native document", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, data, actor(41))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	value, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello 😀", value)
}

func TestPureGoDocument_ExtendsReferenceSnapshot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.Load(ctx, newBaseDocument(t), actor(42))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 5, 0, " native"))
	_, err = nativeDocument.Commit(ctx, "Extend in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, data, actor(43))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	value, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello native", value)
}

func TestPureGoDocument_InsertsInsideReferenceText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.Load(ctx, newBaseDocument(t), actor(46))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 2, 0, "X"))
	_, err = nativeDocument.Commit(ctx, "Insert in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, data, actor(47))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	value, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "HeXllo", value)
}

func TestPureGoDocument_DeletesReferenceText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.Load(ctx, newBaseDocument(t), actor(48))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 2, 2, ""))
	_, err = nativeDocument.Commit(ctx, "Delete in Go", commitTime)
	require.NoError(t, err)
	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, data, actor(49))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	value, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Heo", value)
}

func TestPureGoDocument_EmptySnapshotLoadsInReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(44))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, data, actor(45))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	heads, err := referenceDocument.Heads(ctx)
	require.NoError(t, err)
	assert.Empty(t, heads)
}

func TestPureGoDocument_ReusesActorAfterLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	actorID := actor(55)
	document, err := automerge.New(ctx, actorID)
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A"))
	_, err = document.Commit(ctx, "First change", commitTime)
	require.NoError(t, err)
	data, err := document.Save(ctx)
	require.NoError(t, err)

	loaded, err := automerge.Load(ctx, data, actorID)
	require.NoError(t, err)
	closeDocument(t, loaded)
	loadedText, err := loaded.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, loadedText.Splice(ctx, 1, 0, "B"))
	_, err = loaded.Commit(ctx, "Second change", commitTime)
	require.NoError(t, err)
	data, err = loaded.Save(ctx)
	require.NoError(t, err)

	reference, err := automerge.LoadReference(ctx, data, actor(56))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceText, err := reference.Text(ctx, "body")
	require.NoError(t, err)
	value, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AB", value)
}

func TestPureGoDocument_ConcurrentChangesConverge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := automerge.New(ctx, actor(50))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(ctx, 0, 0, "A"))
	_, err = base.Commit(ctx, "Create body", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	left, err := automerge.Load(ctx, baseData, actor(51))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 1, 0, "L"))
	_, err = left.Commit(ctx, "Edit left", commitTime)
	require.NoError(t, err)

	right, err := automerge.Load(ctx, baseData, actor(52))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(ctx, 1, 0, "R"))
	_, err = right.Commit(ctx, "Edit right", commitTime)
	require.NoError(t, err)

	_, err = left.Merge(ctx, right)
	require.NoError(t, err)
	_, err = right.Merge(ctx, left)
	require.NoError(t, err)

	leftValue, err := leftText.String(ctx)
	require.NoError(t, err)
	rightValue, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)

	leftHeads, err := left.Heads(ctx)
	require.NoError(t, err)
	rightHeads, err := right.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)
}

func TestPureGoDocument_CursorMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseData := newBaseDocument(t)
	nativeDocument, err := automerge.Load(ctx, baseData, actor(60))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	nativeCursor, err := nativeText.Cursor(ctx, 2)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, baseData, actor(61))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	referenceCursor, err := referenceText.Cursor(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, referenceCursor, nativeCursor)

	require.NoError(t, nativeText.Splice(ctx, 0, 0, "X"))
	position, err := nativeText.CursorPosition(ctx, nativeCursor)
	require.NoError(t, err)
	assert.Equal(t, uint32(3), position)
}

func TestPureGoDocument_DeletedCursorMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseData := newBaseDocument(t)
	nativeDocument, err := automerge.Load(ctx, baseData, actor(62))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	cursor, err := nativeText.Cursor(ctx, 2)
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(ctx, 2, 1, ""))
	_, err = nativeDocument.Commit(ctx, "Delete cursor target", commitTime)
	require.NoError(t, err)
	nativePosition, err := nativeText.CursorPosition(ctx, cursor)
	require.NoError(t, err)

	data, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceDocument, err := automerge.LoadReference(ctx, data, actor(63))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)
	referencePosition, err := referenceText.CursorPosition(ctx, cursor)
	require.NoError(t, err)
	assert.Equal(t, referencePosition, nativePosition)
}

func TestPureGoDocument_UTF16CursorBoundariesMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := automerge.NewReference(ctx, actor(122))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(ctx, 0, 0, "A😀B"))
	_, err = base.Commit(ctx, "Create emoji text", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(ctx, baseData, actor(123))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(ctx, baseData, actor(124))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)

	nativeInside, nativeErr := nativeText.Cursor(ctx, 2)
	referenceInside, referenceErr := referenceText.Cursor(ctx, 2)

	require.NoError(t, nativeErr)
	require.NoError(t, referenceErr)
	require.Equal(t, referenceInside, nativeInside)
	nativeInsidePosition, err := nativeText.CursorPosition(ctx, nativeInside)
	require.NoError(t, err)
	referenceInsidePosition, err := referenceText.CursorPosition(
		ctx,
		referenceInside,
	)
	require.NoError(t, err)
	assert.Equal(t, referenceInsidePosition, nativeInsidePosition)

	for _, index := range []uint32{1, 3} {
		nativeCursor, err := nativeText.Cursor(ctx, index)
		require.NoError(t, err)
		referenceCursor, err := referenceText.Cursor(ctx, index)
		require.NoError(t, err)
		require.Equal(t, referenceCursor, nativeCursor)

		require.NoError(t, nativeText.Splice(ctx, 0, 0, "X"))
		require.NoError(t, referenceText.Splice(ctx, 0, 0, "X"))
		nativePosition, err := nativeText.CursorPosition(ctx, nativeCursor)
		require.NoError(t, err)
		referencePosition, err := referenceText.CursorPosition(
			ctx,
			referenceCursor,
		)
		require.NoError(t, err)
		assert.Equal(t, referencePosition, nativePosition)

		require.NoError(t, nativeText.Splice(ctx, 0, 1, ""))
		require.NoError(t, referenceText.Splice(ctx, 0, 1, ""))
	}
}

func TestPureGoDocument_CursorModesMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := automerge.NewReference(ctx, actor(146))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(ctx, 0, 0, "A😀B"))
	_, err = base.Commit(ctx, "cursor base", commitTime)
	require.NoError(t, err)
	data, err := base.Save(ctx)
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(ctx, data, actor(147))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	referenceDocument, err := automerge.LoadReference(ctx, data, actor(148))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)

	for _, index := range []int64{-1, 0, 1, 2, 3, 4, 100} {
		for _, movement := range []automerge.CursorMove{
			automerge.CursorMoveBefore,
			automerge.CursorMoveAfter,
		} {
			nativeCursor, err := nativeText.CursorFor(ctx, index, movement)
			require.NoError(t, err)
			referenceCursor, err := referenceText.CursorFor(
				ctx,
				index,
				movement,
			)
			require.NoError(t, err)
			assert.Equal(t, referenceCursor, nativeCursor)
			nativePosition, err := nativeText.CursorPosition(
				ctx,
				nativeCursor,
			)
			require.NoError(t, err)
			referencePosition, err := referenceText.CursorPosition(
				ctx,
				referenceCursor,
			)
			require.NoError(t, err)
			assert.Equal(t, referencePosition, nativePosition)
		}
	}

	nativeCursor, err := nativeText.CursorFor(
		ctx,
		1,
		automerge.CursorMoveAfter,
	)
	require.NoError(t, err)
	referenceCursor, err := referenceText.CursorFor(
		ctx,
		1,
		automerge.CursorMoveAfter,
	)
	require.NoError(t, err)
	nativeBefore, err := nativeText.CursorFor(
		ctx,
		1,
		automerge.CursorMoveBefore,
	)
	require.NoError(t, err)
	referenceBefore, err := referenceText.CursorFor(
		ctx,
		1,
		automerge.CursorMoveBefore,
	)
	require.NoError(t, err)
	require.NoError(t, nativeText.SpliceCursor(ctx, nativeCursor, 2, "X"))
	require.NoError(t, referenceText.SpliceCursor(ctx, referenceCursor, 2, "X"))
	nativeValue, err := nativeText.String(ctx)
	require.NoError(t, err)
	referenceValue, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, referenceValue, nativeValue)
	assert.Equal(t, "AXB", nativeValue)

	for _, cursors := range [][2]automerge.Cursor{
		{nativeBefore, referenceBefore},
		{nativeCursor, referenceCursor},
	} {
		nativePosition, err := nativeText.CursorPosition(ctx, cursors[0])
		require.NoError(t, err)
		referencePosition, err := referenceText.CursorPosition(ctx, cursors[1])
		require.NoError(t, err)
		assert.Equal(t, referencePosition, nativePosition)
	}
}

func TestPureGoDocument_MarkAuthoringMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(153))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(153))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeText, err := nativeDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	referenceText, err := referenceDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(ctx, 0, 0, "ABCD"))
	require.NoError(t, referenceText.Splice(ctx, 0, 0, "ABCD"))
	_, err = nativeDocument.Commit(ctx, "create text", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit(ctx, "create text", commitTime)
	require.NoError(t, err)

	strong := automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}
	require.NoError(t, nativeText.Mark(
		ctx,
		0,
		4,
		"strong",
		strong,
		automerge.MarkExpandBoth,
	))
	require.NoError(t, referenceText.Mark(
		ctx,
		0,
		4,
		"strong",
		strong,
		automerge.MarkExpandBoth,
	))
	_, err = nativeDocument.Commit(ctx, "mark", commitTime.Add(time.Second))
	require.NoError(t, err)
	_, err = referenceDocument.Commit(ctx, "mark", commitTime.Add(time.Second))
	require.NoError(t, err)
	nativeMarkedSpans, err := nativeText.Spans(ctx)
	require.NoError(t, err)
	referenceMarkedSpans, err := referenceText.Spans(ctx)
	require.NoError(t, err)
	assert.Equal(t, referenceMarkedSpans, nativeMarkedSpans)

	require.NoError(t, nativeText.Unmark(
		ctx,
		1,
		3,
		"strong",
		automerge.MarkExpandNone,
	))
	require.NoError(t, referenceText.Unmark(
		ctx,
		1,
		3,
		"strong",
		automerge.MarkExpandNone,
	))
	_, err = nativeDocument.Commit(ctx, "unmark", commitTime.Add(2*time.Second))
	require.NoError(t, err)
	_, err = referenceDocument.Commit(
		ctx,
		"unmark",
		commitTime.Add(2*time.Second),
	)
	require.NoError(t, err)

	nativeSpans, err := nativeText.Spans(ctx)
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(ctx)
	require.NoError(t, err)
	assert.Equal(t, referenceSpans, nativeSpans)

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(154),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	referenceFromNativeText, err := referenceFromNative.Text(ctx, "body")
	require.NoError(t, err)
	referenceFromNativeSpans, err := referenceFromNativeText.Spans(ctx)
	require.NoError(t, err)
	assert.Equal(t, nativeSpans, referenceFromNativeSpans)

	referenceData, err := referenceDocument.Save(ctx)
	require.NoError(t, err)
	nativeFromReference, err := automerge.Load(
		ctx,
		referenceData,
		actor(155),
	)
	require.NoError(t, err)
	closeDocument(t, nativeFromReference)
	nativeFromReferenceText, err := nativeFromReference.Text(ctx, "body")
	require.NoError(t, err)
	nativeFromReferenceSpans, err := nativeFromReferenceText.Spans(ctx)
	require.NoError(t, err)
	assert.Equal(t, referenceSpans, nativeFromReferenceSpans)
}

func TestPureGoDocument_BlockAuthoringMatchesReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(167))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(167))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeText, err := nativeDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	referenceText, err := referenceDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, nativeText.Splice(ctx, 0, 0, "AB"))
	require.NoError(t, referenceText.Splice(ctx, 0, 0, "AB"))

	nativeFirst, err := nativeText.SplitBlock(ctx, 0)
	require.NoError(t, err)
	referenceFirst, err := referenceText.SplitBlock(ctx, 0)
	require.NoError(t, err)
	setBlockAttributes(t, ctx, nativeFirst, "paragraph")
	setBlockAttributes(t, ctx, referenceFirst, "paragraph")
	nativeSecond, err := nativeText.SplitBlock(ctx, 2)
	require.NoError(t, err)
	referenceSecond, err := referenceText.SplitBlock(ctx, 2)
	require.NoError(t, err)
	setBlockAttributes(t, ctx, nativeSecond, "heading")
	setBlockAttributes(t, ctx, referenceSecond, "heading")
	_, err = nativeDocument.Commit(ctx, "blocks", commitTime)
	require.NoError(t, err)
	_, err = referenceDocument.Commit(ctx, "blocks", commitTime)
	require.NoError(t, err)
	assertTextSpansEqual(t, ctx, nativeText, referenceText)

	nativeReplacement, err := nativeText.ReplaceBlock(ctx, 2)
	require.NoError(t, err)
	referenceReplacement, err := referenceText.ReplaceBlock(ctx, 2)
	require.NoError(t, err)
	setBlockAttributes(t, ctx, nativeReplacement, "blockquote")
	setBlockAttributes(t, ctx, referenceReplacement, "blockquote")
	require.NoError(t, nativeText.JoinBlock(ctx, 0))
	require.NoError(t, referenceText.JoinBlock(ctx, 0))
	_, err = nativeDocument.Commit(
		ctx,
		"replace and join",
		commitTime.Add(time.Second),
	)
	require.NoError(t, err)
	_, err = referenceDocument.Commit(
		ctx,
		"replace and join",
		commitTime.Add(time.Second),
	)
	require.NoError(t, err)
	assertTextSpansEqual(t, ctx, nativeText, referenceText)

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(168),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	referenceFromNativeText, err := referenceFromNative.Text(ctx, "body")
	require.NoError(t, err)
	assertTextSpansEqual(t, ctx, nativeText, referenceFromNativeText)
}

func setBlockAttributes(
	t *testing.T,
	ctx context.Context,
	block *automerge.Object,
	blockType string,
) {
	t.Helper()

	require.NoError(t, block.PutScalar(
		ctx,
		"type",
		automerge.Scalar{
			Type:   automerge.ScalarTypeString,
			String: blockType,
		},
	))
	_, err := block.CreateObject(ctx, "parents", automerge.ObjectTypeList)
	require.NoError(t, err)
	_, err = block.CreateObject(ctx, "attrs", automerge.ObjectTypeMap)
	require.NoError(t, err)
	require.NoError(t, block.PutScalar(
		ctx,
		"isEmbed",
		automerge.Scalar{Type: automerge.ScalarTypeBoolean},
	))
}

func assertTextSpansEqual(
	t *testing.T,
	ctx context.Context,
	left *automerge.Text,
	right *automerge.Text,
) {
	t.Helper()

	leftSpans, err := left.Spans(ctx)
	require.NoError(t, err)
	rightSpans, err := right.Spans(ctx)
	require.NoError(t, err)
	assert.Equal(t, rightSpans, leftSpans)
}

func TestPureGoDocument_SynchronizesWithNativePeer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.New(ctx, actor(70))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Native sync"))
	_, err = left.Commit(ctx, "Create sync body", commitTime)
	require.NoError(t, err)

	right, err := automerge.New(ctx, actor(71))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	value, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Native sync", value)
}

func TestPureGoDocument_SyncWaitsForPeerResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(78))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Wait"))
	_, err = document.Commit(ctx, "Create body", commitTime)
	require.NoError(t, err)

	state, err := document.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, state)
	first, ok, err := state.GenerateMessage(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NotEmpty(t, first)

	second, ok, err := state.GenerateMessage(ctx)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, second)
}

func TestPureGoDocument_SynchronizesWithReferencePeer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.New(ctx, actor(72))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Reference sync"))
	_, err = left.Commit(ctx, "Create sync body", commitTime)
	require.NoError(t, err)

	right, err := automerge.NewReference(ctx, actor(73))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	value, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Reference sync", value)
}

func TestPureGoDocument_ReceivesReferenceSync(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.NewReference(ctx, actor(74))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "From reference"))
	_, err = left.Commit(ctx, "Create reference body", commitTime)
	require.NoError(t, err)

	right, err := automerge.New(ctx, actor(75))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	value, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "From reference", value)
}

func TestPureGoDocument_RepeatedMixedPeerSync(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(76))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.CreateText(ctx, "body")
	require.NoError(t, err)
	_, err = nativeDocument.Commit(ctx, "Create native body", commitTime)
	require.NoError(t, err)

	referenceDocument, err := automerge.NewReference(ctx, actor(77))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeSync, err := nativeDocument.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, nativeSync)

	referenceSync, err := referenceDocument.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, referenceSync)
	synchronize(t, nativeSync, referenceSync)

	referenceText, err := referenceDocument.Text(ctx, "body")
	require.NoError(t, err)

	for iteration := range 20 {
		nativeValue, err := nativeText.String(ctx)
		require.NoError(t, err)

		nativeOffset := utf16Offsets([]rune(nativeValue))
		require.NoError(
			t,
			nativeText.Splice(
				ctx,
				nativeOffset[len(nativeOffset)-1],
				0,
				"N",
			),
		)
		_, err = nativeDocument.Commit(ctx, "Native edit", commitTime)
		require.NoError(t, err)

		referenceValue, err := referenceText.String(ctx)
		require.NoError(t, err)

		referenceOffset := utf16Offsets([]rune(referenceValue))
		require.NoError(
			t,
			referenceText.Splice(
				ctx,
				referenceOffset[len(referenceOffset)-1],
				0,
				"R",
			),
		)
		_, err = referenceDocument.Commit(ctx, "Reference edit", commitTime)
		require.NoError(t, err)

		synchronize(t, nativeSync, referenceSync)

		nativeValue, err = nativeText.String(ctx)
		require.NoError(t, err)
		referenceValue, err = referenceText.String(ctx)
		require.NoError(t, err)
		assert.Equal(t, referenceValue, nativeValue, "iteration %d", iteration)
	}
}

func TestPureGoDocument_ReferenceEditsWhileMessageInFlight(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client, err := automerge.NewReference(ctx, actor(79))
	require.NoError(t, err)
	closeDocument(t, client)
	clientText, err := client.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, clientText.Splice(ctx, 0, 0, "A"))
	_, err = client.Commit(ctx, "initial", commitTime)
	require.NoError(t, err)

	server, err := automerge.New(ctx, actor(80))
	require.NoError(t, err)
	closeDocument(t, server)

	clientSync, err := client.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, clientSync)

	serverSync, err := server.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, serverSync)
	synchronize(t, clientSync, serverSync)

	require.NoError(t, clientText.Splice(ctx, 1, 0, "B"))
	_, err = client.Commit(ctx, "first edit", commitTime)
	require.NoError(t, err)
	first, ok, err := clientSync.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	require.NoError(t, clientText.Splice(ctx, 2, 0, "C"))
	_, err = client.Commit(ctx, "second edit", commitTime)
	require.NoError(t, err)

	require.NoError(t, serverSync.ReceiveMessage(ctx, first))
	ack, ok, err := serverSync.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, clientSync.ReceiveMessage(ctx, ack))

	second, ok, err := clientSync.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, serverSync.ReceiveMessage(ctx, second))
	synchronize(t, clientSync, serverSync)

	serverText, err := server.Text(ctx, "body")
	require.NoError(t, err)
	value, err := serverText.String(ctx)
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var (
				source *automerge.Document
				target *automerge.Document
				err    error
			)
			if test.sourceReference {
				source, err = automerge.NewReference(ctx, actor(181))
				require.NoError(t, err)
				target, err = automerge.New(ctx, actor(182))
			} else {
				source, err = automerge.New(ctx, actor(181))
				require.NoError(t, err)
				target, err = automerge.NewReference(ctx, actor(182))
			}

			require.NoError(t, err)
			closeDocument(t, source)
			closeDocument(t, target)

			text, err := source.CreateText(ctx, "body")
			require.NoError(t, err)
			require.NoError(t, text.Splice(ctx, 0, 0, "Published"))
			_, err = source.Commit(ctx, "publish", commitTime)
			require.NoError(t, err)

			sourceState, err := source.NewSyncState(ctx)
			require.NoError(t, err)
			closeSyncState(t, sourceState)

			targetState, err := target.NewSyncState(ctx)
			require.NoError(t, err)
			closeSyncState(t, targetState)
			require.NoError(t, targetState.SetReadOnly(ctx, true))

			synchronize(t, sourceState, targetState)

			peerReadOnly, err := sourceState.PeerReadOnly(ctx)
			require.NoError(t, err)
			assert.True(t, peerReadOnly)

			_, err = target.Text(ctx, "body")
			require.Error(t, err)

			require.NoError(t, targetState.SetReadOnly(ctx, false))
			synchronize(t, sourceState, targetState)

			peerReadOnly, err = sourceState.PeerReadOnly(ctx)
			require.NoError(t, err)
			assert.False(t, peerReadOnly)

			targetText, err := target.Text(ctx, "body")
			require.NoError(t, err)
			value, err := targetText.String(ctx)
			require.NoError(t, err)
			assert.Equal(t, "Published", value)
		})
	}
}

func TestSyncState_ReadOnlyModeOverridesInFlight(t *testing.T) {
	t.Parallel()

	tests := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			document, err := factory(ctx, actor(183))
			require.NoError(t, err)
			closeDocument(t, document)
			text, err := document.CreateText(ctx, "body")
			require.NoError(t, err)
			require.NoError(t, text.Splice(ctx, 0, 0, "A"))
			_, err = document.Commit(ctx, "initial", commitTime)
			require.NoError(t, err)
			state, err := document.NewSyncState(ctx)
			require.NoError(t, err)
			closeSyncState(t, state)

			_, ok, err := state.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
			require.NoError(t, state.SetReadOnly(ctx, true))
			_, ok, err = state.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
			require.NoError(t, state.SetReadOnly(ctx, false))
			_, ok, err = state.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestSyncState_BothReadOnlyResumeConvergence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.New(ctx, actor(184))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 0, 0, "L"))
	_, err = left.Commit(ctx, "left", commitTime)
	require.NoError(t, err)

	right, err := automerge.NewReference(ctx, actor(185))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(ctx, 0, 0, "R"))
	_, err = right.Commit(ctx, "right", commitTime)
	require.NoError(t, err)

	leftState, err := left.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, leftState)

	rightState, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, rightState)
	require.NoError(t, leftState.SetReadOnly(ctx, true))
	require.NoError(t, rightState.SetReadOnly(ctx, true))
	synchronize(t, leftState, rightState)

	leftValue, err := leftText.String(ctx)
	require.NoError(t, err)
	rightValue, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "L", leftValue)
	assert.Equal(t, "R", rightValue)

	require.NoError(t, leftState.SetReadOnly(ctx, false))
	require.NoError(t, rightState.SetReadOnly(ctx, false))
	synchronize(t, leftState, rightState)

	leftHeads, err := left.Heads(ctx)
	require.NoError(t, err)
	rightHeads, err := right.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)

	leftText, err = left.Text(ctx, "body")
	require.NoError(t, err)
	leftValue, err = leftText.String(ctx)
	require.NoError(t, err)
	rightText, err = right.Text(ctx, "body")
	require.NoError(t, err)
	rightValue, err = rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)
}
