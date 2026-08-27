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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

var commitTime = time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)

func actor(value byte) [16]byte {
	var actorID [16]byte

	actorID[0] = value

	return actorID
}

func closeDocument(t *testing.T, document interface{ Close() error }) {
	t.Helper()
	t.Cleanup(
		func() {
			require.NoError(t, document.Close())
		},
	)
}

func closeSyncState(t *testing.T, state interface{ Close() error }) {
	t.Helper()
	t.Cleanup(
		func() {
			require.NoError(t, state.Close())
		},
	)
}

func synchronize(
	t *testing.T,
	left interface {
		GenerateMessage() ([]byte, bool, error)
		ReceiveMessage([]byte) error
	},
	right interface {
		GenerateMessage() ([]byte, bool, error)
		ReceiveMessage([]byte) error
	},
) {
	t.Helper()

	for range 100 {
		progressed := false

		message, ok, err := left.GenerateMessage()
		require.NoError(t, err)

		if ok {
			require.NoError(t, right.ReceiveMessage(message))

			progressed = true
		}

		message, ok, err = right.GenerateMessage()
		require.NoError(t, err)

		if ok {
			require.NoError(t, left.ReceiveMessage(message))

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

	document, err := automerge.New(actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Hello"))

	_, err = document.Commit("Create document", commitTime)
	require.NoError(t, err)

	data, err := document.Save()
	require.NoError(t, err)

	return data
}

func TestDocument_SaveAndLoad(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	require.NoError(t, document.PutString("title", "Policy"))
	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "Hello"))

	hash, err := document.Commit("Create policy", commitTime)
	require.NoError(t, err)
	assert.Len(t, hash.String(), 64)

	data, err := document.Save()
	require.NoError(t, err)

	loaded, err := automerge.Load(data, actor(2))
	require.NoError(t, err)
	closeDocument(t, loaded)

	loadedText, err := loaded.Text("body")
	require.NoError(t, err)
	value, err := loadedText.String()
	require.NoError(t, err)
	assert.Equal(t, "Hello", value)

	heads, err := loaded.Heads()
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{hash}, heads)
}

func TestDocument_ConcurrentChangesConverge(t *testing.T) {
	t.Parallel()

	base := newBaseDocument(t)

	left, err := automerge.Load(base, actor(2))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text("body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(5, 0, " left"))

	_, err = left.Commit("Edit left", commitTime.Add(time.Second))
	require.NoError(t, err)

	right, err := automerge.Load(base, actor(3))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.Text("body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(5, 0, " right"))

	_, err = right.Commit("Edit right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	leftData, err := left.Save()
	require.NoError(t, err)
	rightData, err := right.Save()
	require.NoError(t, err)

	leftFirst, err := automerge.Load(leftData, actor(4))
	require.NoError(t, err)
	closeDocument(t, leftFirst)

	rightFirst, err := automerge.Load(rightData, actor(5))
	require.NoError(t, err)
	closeDocument(t, rightFirst)

	_, err = leftFirst.Merge(right)
	require.NoError(t, err)
	_, err = rightFirst.Merge(left)
	require.NoError(t, err)

	leftMergedText, err := leftFirst.Text("body")
	require.NoError(t, err)
	rightMergedText, err := rightFirst.Text("body")
	require.NoError(t, err)
	leftValue, err := leftMergedText.String()
	require.NoError(t, err)
	rightValue, err := rightMergedText.String()
	require.NoError(t, err)
	assert.Equal(t, leftValue, rightValue)

	leftHeads, err := leftFirst.Heads()
	require.NoError(t, err)
	rightHeads, err := rightFirst.Heads()
	require.NoError(t, err)
	assert.ElementsMatch(t, leftHeads, rightHeads)
}

func TestText_SpliceUsesUTF16Offsets(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "A😀B"))
	require.NoError(t, text.Splice(1, 2, ""))

	value, err := text.String()
	require.NoError(t, err)
	assert.Equal(t, "AB", value)
}

func TestText_CursorTracksConcurrentEdits(t *testing.T) {
	t.Parallel()

	document, err := automerge.Load(newBaseDocument(t), actor(2))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.Text("body")
	require.NoError(t, err)

	cursor, err := text.Cursor(4)
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "A"))

	_, err = document.Commit("Insert prefix", commitTime.Add(time.Second))
	require.NoError(t, err)

	position, err := text.CursorPosition(cursor)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), position)
}

func TestSyncState_ExchangesConcurrentChanges(t *testing.T) {
	t.Parallel()

	left, err := automerge.Load(newBaseDocument(t), actor(2))
	require.NoError(t, err)
	closeDocument(t, left)

	right, err := automerge.New(actor(3))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, leftSync)

	rightSync, err := right.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, rightSync)

	synchronize(t, leftSync, rightSync)

	leftText, err := left.Text("body")
	require.NoError(t, err)
	rightText, err := right.Text("body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(5, 0, " left"))

	_, err = left.Commit("Edit left", commitTime.Add(time.Second))
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(5, 0, " right"))

	_, err = right.Commit("Edit right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	synchronize(t, leftSync, rightSync)

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

func TestSyncState_SaveAndLoad(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	state, err := document.NewSyncState()
	require.NoError(t, err)
	data, err := state.Save()
	require.NoError(t, err)
	require.NoError(t, state.Close())

	loaded, err := document.LoadSyncState(data)
	require.NoError(t, err)
	closeSyncState(t, loaded)
	_, _, err = loaded.GenerateMessage()
	require.NoError(t, err)
}

func TestDocument_CloseIsIdempotent(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(1))
	require.NoError(t, err)

	require.NoError(t, document.Close())
	require.NoError(t, document.Close())

	_, err = document.Save()
	assert.ErrorIs(t, err, automerge.ErrClosed)
}

func TestLoad_InvalidDocument(t *testing.T) {
	t.Parallel()

	document, err := automerge.Load([]byte("invalid"), actor(1))
	assert.Nil(t, document)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, automerge.ErrClosed))
}
