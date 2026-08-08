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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestPureGoDocument_ReferenceLoadsNativeHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.NewPureGo(ctx, actor(40))
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
	nativeDocument, err := automerge.LoadPureGo(ctx, newBaseDocument(t), actor(42))
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
	nativeDocument, err := automerge.LoadPureGo(ctx, newBaseDocument(t), actor(46))
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
	nativeDocument, err := automerge.LoadPureGo(ctx, newBaseDocument(t), actor(48))
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
	nativeDocument, err := automerge.NewPureGo(ctx, actor(44))
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
	document, err := automerge.NewPureGo(ctx, actorID)
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A"))
	_, err = document.Commit(ctx, "First change", commitTime)
	require.NoError(t, err)
	data, err := document.Save(ctx)
	require.NoError(t, err)

	loaded, err := automerge.LoadPureGo(ctx, data, actorID)
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
	base, err := automerge.NewPureGo(ctx, actor(50))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(ctx, 0, 0, "A"))
	_, err = base.Commit(ctx, "Create body", commitTime)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	left, err := automerge.LoadPureGo(ctx, baseData, actor(51))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 1, 0, "L"))
	_, err = left.Commit(ctx, "Edit left", commitTime)
	require.NoError(t, err)

	right, err := automerge.LoadPureGo(ctx, baseData, actor(52))
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
	nativeDocument, err := automerge.LoadPureGo(ctx, baseData, actor(60))
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
	nativeDocument, err := automerge.LoadPureGo(ctx, baseData, actor(62))
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

func TestPureGoDocument_SynchronizesWithNativePeer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.NewPureGo(ctx, actor(70))
	require.NoError(t, err)
	closeDocument(t, left)
	text, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Native sync"))
	_, err = left.Commit(ctx, "Create sync body", commitTime)
	require.NoError(t, err)

	right, err := automerge.NewPureGo(ctx, actor(71))
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
	document, err := automerge.NewPureGo(ctx, actor(78))
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
	left, err := automerge.NewPureGo(ctx, actor(72))
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

	right, err := automerge.NewPureGo(ctx, actor(75))
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
	nativeDocument, err := automerge.NewPureGo(ctx, actor(76))
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

	server, err := automerge.NewPureGo(ctx, actor(80))
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
