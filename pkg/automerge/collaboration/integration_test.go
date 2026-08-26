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

package collaboration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
)

// The production sync state must satisfy the driver's interface with no adapter.
var _ collaboration.SyncSession = (*automerge.SyncState)(nil)

func actor(value byte) automerge.ActorID {
	var actorID automerge.ActorID
	actorID[0] = value

	return actorID
}

func commitTime() time.Time {
	return time.Unix(1786147200, 0).UTC()
}

// TestServerConn_ConvergesRealDocument drives a real Automerge client sync state
// through the ServerConn loop and confirms the client reconstructs the server's
// document. This exercises the interop linchpin end to end in Go: the repo sync
// frame payloads are exactly our engine's sync messages.
func TestServerConn_ConvergesRealDocument(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Server document: the authority, holding "hello" in a text object.
	server, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	defer func() { _ = server.Close(ctx) }()

	text, err := server.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "hello"))
	_, err = server.Commit(ctx, "seed", commitTime())
	require.NoError(t, err)

	serverSync, err := server.NewSyncState(ctx)
	require.NoError(t, err)
	defer func() { _ = serverSync.Close(ctx) }()

	conn, err := collaboration.NewServerConn(
		collaboration.ServerConfig{ServerPeerID: "server"},
		"doc-1",
		serverSync,
	)
	require.NoError(t, err)

	// Client document: empty, learning the document over the connection.
	client, err := automerge.New(ctx, actor(2))
	require.NoError(t, err)
	defer func() { _ = client.Close(ctx) }()

	clientSync, err := client.NewSyncState(ctx)
	require.NoError(t, err)
	defer func() { _ = clientSync.Close(ctx) }()

	join, err := collaboration.EncodeJoinFrame(
		collaboration.NewJoinFrame("peer-a", collaboration.PeerMetadata{}),
	)
	require.NoError(t, err)

	// Frames the client still has to process; seeded with the server's announce.
	toClient, accepted, err := conn.Start(ctx, join)
	require.NoError(t, err)
	require.True(t, accepted)

	deliverToClient := func(frames [][]byte) {
		for _, frame := range frames {
			kind, err := collaboration.FrameKind(frame)
			require.NoError(t, err)

			if kind == collaboration.FramePeer {
				continue
			}

			message, err := collaboration.DecodeMessage(frame)
			require.NoError(t, err)
			require.Equal(t, collaboration.MessageSync, message.Type)
			require.NoError(t, clientSync.ReceiveMessage(ctx, message.Data))
		}
	}

	// Pump until neither side has anything more to send. A generous bound stops a
	// protocol bug from hanging the test.
	for round := 0; round < 20; round++ {
		deliverToClient(toClient)
		toClient = nil

		message, ok, err := clientSync.GenerateMessage(ctx)
		require.NoError(t, err)

		if !ok {
			break
		}

		frame, err := collaboration.EncodeMessage(
			collaboration.Message{
				Type:       collaboration.MessageSync,
				SenderID:   "peer-a",
				TargetID:   "server",
				DocumentID: "doc-1",
				Data:       message,
			},
		)
		require.NoError(t, err)

		reply, fanout, err := conn.Receive(ctx, frame)
		require.NoError(t, err)
		assert.Nil(t, fanout)

		toClient = reply
	}

	serverHeads, err := server.Heads(ctx)
	require.NoError(t, err)

	clientHeads, err := client.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, serverHeads, clientHeads, "client must converge to the server frontier")

	clientText, err := client.Text(ctx, "body")
	require.NoError(t, err)

	value, err := clientText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}
