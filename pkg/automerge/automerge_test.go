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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

var commitTime = time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)

func actor(value byte) automerge.ActorID {
	var actorID automerge.ActorID
	actorID[0] = value
	return actorID
}

func closeDocument(t *testing.T, document *automerge.Document) {
	t.Helper()
	t.Cleanup(
		func() {
			require.NoError(t, document.Close(context.Background()))
		},
	)
}

func closeSyncState(t *testing.T, state *automerge.SyncState) {
	t.Helper()
	t.Cleanup(
		func() {
			require.NoError(t, state.Close(context.Background()))
		},
	)
}

func synchronize(t *testing.T, left, right *automerge.SyncState) {
	t.Helper()

	ctx := context.Background()
	for range 100 {
		progressed := false

		message, ok, err := left.GenerateMessage(ctx)
		require.NoError(t, err)
		if ok {
			require.NoError(t, right.ReceiveMessage(ctx, message))
			progressed = true
		}

		message, ok, err = right.GenerateMessage(ctx)
		require.NoError(t, err)
		if ok {
			require.NoError(t, left.ReceiveMessage(ctx, message))
			progressed = true
		}

		if !progressed {
			return
		}
	}

	require.Fail(t, "sync did not quiesce")
}

func newBaseDocument(t *testing.T) []byte {
	t.Helper()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Hello"))
	_, err = document.Commit(ctx, "Create document", commitTime)
	require.NoError(t, err)

	data, err := document.Save(ctx)
	require.NoError(t, err)
	return data
}

func TestDocument_SaveAndLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	require.NoError(t, document.PutString(ctx, "title", "Policy"))
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Hello"))

	hash, err := document.Commit(ctx, "Create policy", commitTime)
	require.NoError(t, err)
	assert.Len(t, hash.String(), 64)

	data, err := document.Save(ctx)
	require.NoError(t, err)

	loaded, err := automerge.Load(ctx, data, actor(2))
	require.NoError(t, err)
	closeDocument(t, loaded)

	loadedText, err := loaded.Text(ctx, "body")
	require.NoError(t, err)
	value, err := loadedText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello", value)

	heads, err := loaded.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{hash}, heads)
}

func TestDocument_ConcurrentChangesConverge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := newBaseDocument(t)

	left, err := automerge.Load(ctx, base, actor(2))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 5, 0, " left"))
	_, err = left.Commit(ctx, "Edit left", commitTime.Add(time.Second))
	require.NoError(t, err)

	right, err := automerge.Load(ctx, base, actor(3))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(ctx, 5, 0, " right"))
	_, err = right.Commit(ctx, "Edit right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	leftData, err := left.Save(ctx)
	require.NoError(t, err)
	rightData, err := right.Save(ctx)
	require.NoError(t, err)

	leftFirst, err := automerge.Load(ctx, leftData, actor(4))
	require.NoError(t, err)
	closeDocument(t, leftFirst)
	rightFirst, err := automerge.Load(ctx, rightData, actor(5))
	require.NoError(t, err)
	closeDocument(t, rightFirst)

	_, err = leftFirst.Merge(ctx, right)
	require.NoError(t, err)
	_, err = rightFirst.Merge(ctx, left)
	require.NoError(t, err)

	leftMergedText, err := leftFirst.Text(ctx, "body")
	require.NoError(t, err)
	rightMergedText, err := rightFirst.Text(ctx, "body")
	require.NoError(t, err)
	leftValue, err := leftMergedText.String(ctx)
	require.NoError(t, err)
	rightValue, err := rightMergedText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)

	leftHeads, err := leftFirst.Heads(ctx)
	require.NoError(t, err)
	rightHeads, err := rightFirst.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)
}

func TestText_SpliceUsesUTF16Offsets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A😀B"))
	require.NoError(t, text.Splice(ctx, 1, 2, ""))

	value, err := text.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AB", value)
}

func TestText_CursorTracksConcurrentEdits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.Load(ctx, newBaseDocument(t), actor(2))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.Text(ctx, "body")
	require.NoError(t, err)

	cursor, err := text.Cursor(ctx, 4)
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A"))
	_, err = document.Commit(ctx, "Insert prefix", commitTime.Add(time.Second))
	require.NoError(t, err)

	position, err := text.CursorPosition(ctx, cursor)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), position)
}

func TestSyncState_ExchangesConcurrentChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.Load(ctx, newBaseDocument(t), actor(2))
	require.NoError(t, err)
	closeDocument(t, left)
	right, err := automerge.New(ctx, actor(3))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, leftSync)
	rightSync, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	leftText, err := left.Text(ctx, "body")
	require.NoError(t, err)
	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 5, 0, " left"))
	_, err = left.Commit(ctx, "Edit left", commitTime.Add(time.Second))
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(ctx, 5, 0, " right"))
	_, err = right.Commit(ctx, "Edit right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	synchronize(t, leftSync, rightSync)

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

func TestSyncState_SaveAndLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	state, err := document.NewSyncState(ctx)
	require.NoError(t, err)
	data, err := state.Save(ctx)
	require.NoError(t, err)
	require.NoError(t, state.Close(ctx))

	loaded, err := document.LoadSyncState(ctx, data)
	require.NoError(t, err)
	closeSyncState(t, loaded)
	_, _, err = loaded.GenerateMessage(ctx)
	require.NoError(t, err)
}

func TestDocument_CloseIsIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)

	require.NoError(t, document.Close(ctx))
	require.NoError(t, document.Close(ctx))

	_, err = document.Save(ctx)
	assert.ErrorIs(t, err, automerge.ErrClosed)
}

func TestLoad_InvalidDocument(t *testing.T) {
	t.Parallel()

	document, err := automerge.Load(context.Background(), []byte("invalid"), actor(1))
	assert.Nil(t, document)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, automerge.ErrClosed))
}
