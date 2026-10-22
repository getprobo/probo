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

package automerge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActorsFromSyncMessage_ReturnsDistinctActors(t *testing.T) {
	t.Parallel()

	leftActor := ActorID{1}
	rightActor := ActorID{2}

	left, err := New(leftActor)
	require.NoError(t, err)
	defer func() { _ = left.Close() }()

	right, err := New(rightActor)
	require.NoError(t, err)
	defer func() { _ = right.Close() }()

	leftText, err := left.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(0, 0, "hello"))
	_, err = left.Commit("left", time.Unix(1, 0).UTC())
	require.NoError(t, err)

	rightText, err := right.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, rightText.Splice(0, 0, "world"))
	_, err = right.Commit("right", time.Unix(2, 0).UTC())
	require.NoError(t, err)

	_, err = left.Merge(right)
	require.NoError(t, err)

	empty, err := New(ActorID{9})
	require.NoError(t, err)
	defer func() { _ = empty.Close() }()

	sender, err := left.NewSyncState()
	require.NoError(t, err)
	defer func() { _ = sender.Close() }()

	receiver, err := empty.NewSyncState()
	require.NoError(t, err)
	defer func() { _ = receiver.Close() }()

	// Empty peer opens with a request; the document peer answers with its changes.
	request, ok, err := receiver.GenerateMessage()
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, sender.ReceiveMessage(request))

	message, ok, err := sender.GenerateMessage()
	require.NoError(t, err)
	require.True(t, ok)

	actors, err := ActorsFromSyncMessage(message)
	require.NoError(t, err)
	require.NotEmpty(t, actors)

	seen := make(map[string]struct{}, len(actors))
	for _, actor := range actors {
		seen[string(actor)] = struct{}{}
	}

	assert.Len(t, seen, len(actors))
	assert.Contains(t, seen, string(leftActor[:]))
	assert.Contains(t, seen, string(rightActor[:]))
}
