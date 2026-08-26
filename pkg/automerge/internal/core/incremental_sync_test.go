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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendSync_SendsOnlyChangesSinceRemoteHeads(t *testing.T) {
	t.Parallel()

	backend, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, backend.SetActor([]byte{1}))
	text, err := backend.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, backend.SpliceText(text, 0, 0, "A"))
	_, err = backend.Commit("first", time.Unix(1, 0))
	require.NoError(t, err)

	syncHandle, err := backend.NewSyncState()
	require.NoError(t, err)
	firstMessage, ok, err := backend.GenerateSyncMessage(syncHandle)
	require.NoError(t, err)
	assert.True(t, ok)

	first, err := ParseSyncMessage(firstMessage)
	require.NoError(t, err)
	require.Len(t, first.Changes, 1)
	firstDocument, err := Decode(first.Changes[0])
	require.NoError(t, err)
	assert.Equal(t, []ChunkType{ChunkChange}, firstDocument.ChunkTypes)

	ackMessage := SyncMessage{
		Version: SyncMessageVersion2,
		Heads:   first.Heads,
	}
	ack, err := ackMessage.Encode()
	require.NoError(t, err)
	require.NoError(t, backend.ReceiveSyncMessage(syncHandle, ack))

	require.NoError(t, backend.SpliceText(text, 1, 0, "B"))
	_, err = backend.Commit("second", time.Unix(2, 0))
	require.NoError(t, err)
	secondMessage, ok, err := backend.GenerateSyncMessage(syncHandle)
	require.NoError(t, err)
	assert.True(t, ok)

	second, err := ParseSyncMessage(secondMessage)
	require.NoError(t, err)
	require.Len(t, second.Changes, 1)

	secondDocument, err := DecodePartial(second.Changes[0])
	require.NoError(t, err)
	require.Len(t, secondDocument.Changes, 1)
	assert.Equal(t, "second", secondDocument.Changes[0].Message)

	fullDocument, err := backend.Save(true, true)
	require.NoError(t, err)
	assert.Less(t, len(secondMessage), len(fullDocument))
}
