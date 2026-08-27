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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
)

// clientServerHarness drives a ClientConn against a ServerConn, moving frames
// between them until both are quiescent, so a test can assert convergence.
type clientServerHarness struct {
	t      *testing.T
	client *collaboration.ClientConn
	server *collaboration.ServerConn
}

func TestNewClientConn_NilSyncPanicsOnUse(t *testing.T) {
	t.Parallel()

	conn, err := collaboration.NewClientConn(
		collaboration.ClientConfig{
			ClientPeerID: "agent",
			DocumentID:   "doc-1",
		},
		nil,
	)
	require.NoError(t, err)

	peer, err := collaboration.EncodePeerFrame(
		collaboration.PeerFrame{
			Type:                    collaboration.FramePeer,
			SenderID:                "server",
			TargetID:                "agent",
			SelectedProtocolVersion: collaboration.ProtocolV1,
		},
	)
	require.NoError(t, err)

	require.Panics(t, func() {
		_, _ = conn.Receive(peer)
	})
}

// pump exchanges frames until neither side produces more, starting from the
// client's join. It bounds the rounds so a protocol bug fails instead of hangs.
func (h *clientServerHarness) pump() {
	h.t.Helper()

	join, err := h.client.Start()
	require.NoError(h.t, err)

	serverOut, accepted, err := h.server.Start(join)
	require.NoError(h.t, err)
	require.True(h.t, accepted)

	toClient := serverOut

	for range 50 {
		var nextToServer [][]byte

		for _, frame := range toClient {
			inbound, err := h.client.Receive(frame)
			require.NoError(h.t, err)

			nextToServer = append(nextToServer, inbound.Outgoing...)
		}

		var nextToClient [][]byte

		for _, frame := range nextToServer {
			reply, _, err := h.server.Receive(frame)
			require.NoError(h.t, err)

			nextToClient = append(nextToClient, reply...)
		}

		if len(nextToServer) == 0 && len(nextToClient) == 0 {
			return
		}

		toClient = nextToClient
	}

	h.t.Fatal("client and server did not converge")
}

func newHarness(
	t *testing.T,
	client *automerge.Document,
	clientEmpty bool,
	server *automerge.Document,
) *clientServerHarness {
	t.Helper()

	clientSync, err := client.NewSyncState()
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSync.Close() })

	serverSync, err := server.NewSyncState()
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSync.Close() })

	clientConn, err := collaboration.NewClientConn(
		collaboration.ClientConfig{
			ClientPeerID: "agent",
			DocumentID:   "doc-1",
			StartsEmpty:  clientEmpty,
		},
		clientSync,
	)
	require.NoError(t, err)

	serverConn, err := collaboration.NewServerConn(
		collaboration.ServerConfig{ServerPeerID: "server"},
		"doc-1",
		serverSync,
	)
	require.NoError(t, err)

	return &clientServerHarness{t: t, client: clientConn, server: serverConn}
}

// TestClientConn_LearnsServerDocument syncs an empty Go client from a server
// that holds the document, driver to driver.
func TestClientConn_LearnsServerDocument(t *testing.T) {
	t.Parallel()

	server, err := automerge.New(actor(1))
	require.NoError(t, err)

	defer func() { _ = server.Close() }()

	text, err := server.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "hello world"))

	_, err = server.Commit("seed", commitTime())
	require.NoError(t, err)

	client, err := automerge.New(actor(2))
	require.NoError(t, err)

	defer func() { _ = client.Close() }()

	newHarness(t, client, true, server).pump()

	serverHeads, err := server.Heads()
	require.NoError(t, err)
	clientHeads, err := client.Heads()
	require.NoError(t, err)
	assert.Equal(t, serverHeads, clientHeads)

	clientText, err := client.Text("body")
	require.NoError(t, err)
	value, err := clientText.String()
	require.NoError(t, err)
	assert.Equal(t, "hello world", value)
}

// TestClientConn_PushesLocalEdits syncs a client that already has content into a
// server that starts empty, so the request/sync direction is reversed.
func TestClientConn_PushesLocalEdits(t *testing.T) {
	t.Parallel()

	server, err := automerge.New(actor(1))
	require.NoError(t, err)

	defer func() { _ = server.Close() }()

	client, err := automerge.New(actor(2))
	require.NoError(t, err)

	defer func() { _ = client.Close() }()

	text, err := client.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "from the agent"))

	_, err = client.Commit("edit", commitTime())
	require.NoError(t, err)

	newHarness(t, client, false, server).pump()

	serverText, err := server.Text("body")
	require.NoError(t, err)
	value, err := serverText.String()
	require.NoError(t, err)
	assert.Equal(t, "from the agent", value)
}

// TestClientConn_FirstMessageIsRequestWhenEmpty checks the request/sync
// selection an empty client uses for its first outbound message.
func TestClientConn_FirstMessageIsRequestWhenEmpty(t *testing.T) {
	t.Parallel()

	client, err := automerge.New(actor(2))
	require.NoError(t, err)

	defer func() { _ = client.Close() }()

	sync, err := client.NewSyncState()
	require.NoError(t, err)

	defer func() { _ = sync.Close() }()

	conn, err := collaboration.NewClientConn(
		collaboration.ClientConfig{
			ClientPeerID: "agent", DocumentID: "doc-1", StartsEmpty: true,
		},
		sync,
	)
	require.NoError(t, err)

	// Complete the handshake with a peer frame so the client emits its first sync.
	peer, err := collaboration.EncodePeerFrame(
		collaboration.PeerFrame{
			Type: collaboration.FramePeer, SenderID: "server", TargetID: "agent",
			SelectedProtocolVersion: collaboration.ProtocolV1,
		},
	)
	require.NoError(t, err)

	inbound, err := conn.Receive(peer)
	require.NoError(t, err)
	require.NotEmpty(t, inbound.Outgoing)

	message, err := collaboration.DecodeMessage(inbound.Outgoing[0])
	require.NoError(t, err)
	assert.Equal(
		t,
		collaboration.MessageRequest,
		message.Type,
		"an empty client's first message is a request",
	)
}

// TestClientConn_SurfacesServerError reports an error frame to the caller.
func TestClientConn_SurfacesServerError(t *testing.T) {
	t.Parallel()

	client, err := automerge.New(actor(2))
	require.NoError(t, err)

	defer func() { _ = client.Close() }()

	sync, err := client.NewSyncState()
	require.NoError(t, err)

	defer func() { _ = sync.Close() }()

	conn, err := collaboration.NewClientConn(
		collaboration.ClientConfig{
			ClientPeerID: "agent", DocumentID: "doc-1", StartsEmpty: true,
		},
		sync,
	)
	require.NoError(t, err)

	errorFrame, err := collaboration.EncodeErrorFrame(
		collaboration.ErrorFrame{
			Type: collaboration.FrameError, SenderID: "server", TargetID: "agent",
			Message: "unauthorized",
		},
	)
	require.NoError(t, err)

	inbound, err := conn.Receive(errorFrame)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", inbound.ServerError)
}
