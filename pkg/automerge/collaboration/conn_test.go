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

package collaboration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedSync is a fake SyncSession: it yields a scripted sequence of outbound
// messages and records the inbound messages applied to it.
type scriptedSync struct {
	outbound [][]byte
	received [][]byte
}

func (s *scriptedSync) GenerateMessage() ([]byte, bool, error) {
	if len(s.outbound) == 0 {
		return nil, false, nil
	}

	next := s.outbound[0]
	s.outbound = s.outbound[1:]

	return next, true, nil
}

func (s *scriptedSync) ReceiveMessage(message []byte) error {
	s.received = append(s.received, message)

	return nil
}

func newConn(t *testing.T, sync SyncSession) *ServerConn {
	t.Helper()

	conn, err := NewServerConn(ServerConfig{ServerPeerID: "server"}, "doc-1", sync)
	require.NoError(t, err)

	return conn
}

func joinFixture(t *testing.T) []byte {
	t.Helper()

	return decodeBase64(t, readFixture[wireFixture](t, "wire-join.json").FrameCBORBase64)
}

// TestServerConn_StartAnnouncesInitialSync sends the peer reply then a sync frame
// for each initial sync message.
func TestServerConn_StartAnnouncesInitialSync(t *testing.T) {
	t.Parallel()

	sync := &scriptedSync{outbound: [][]byte{{1, 1}, {2, 2}}}
	conn := newConn(t, sync)

	out, accepted, err := conn.Start(joinFixture(t))
	require.NoError(t, err)
	require.True(t, accepted)
	require.Len(t, out, 3) // peer reply + two sync frames

	peer, err := DecodePeerFrame(out[0])
	require.NoError(t, err)
	assert.Equal(t, ProtocolV1, peer.SelectedProtocolVersion)

	for _, frame := range out[1:] {
		message, err := DecodeMessage(frame)
		require.NoError(t, err)
		assert.Equal(t, MessageSync, message.Type)
		assert.Equal(t, "server", message.SenderID)
		assert.Equal(t, "peer-a", message.TargetID)
		assert.Equal(t, "doc-1", message.DocumentID)
	}
}

// TestServerConn_RejectsUnsupportedVersion returns an error frame and does not
// start.
func TestServerConn_RejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	join, err := marshal(
		JoinFrame{
			Type:                      FrameJoin,
			SenderID:                  "peer-a",
			SupportedProtocolVersions: []string{"999"},
		},
	)
	require.NoError(t, err)

	conn := newConn(t, &scriptedSync{})
	out, accepted, err := conn.Start(join)
	require.NoError(t, err)
	assert.False(t, accepted)
	require.Len(t, out, 1)

	_, err = DecodeErrorFrame(out[0])
	require.NoError(t, err)
}

// TestServerConn_AppliesInboundSyncAndReplies applies an inbound sync message
// and answers with the generated sync frames.
func TestServerConn_AppliesInboundSyncAndReplies(t *testing.T) {
	t.Parallel()

	sync := &scriptedSync{}
	conn := newConn(t, sync)

	_, accepted, err := conn.Start(joinFixture(t))
	require.NoError(t, err)
	require.True(t, accepted)

	// The next generate call yields one reply message.
	sync.outbound = [][]byte{{9, 9}}

	inboundSync, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc-1", Data: []byte{7, 7},
		},
	)
	require.NoError(t, err)

	reply, fanout, err := conn.Receive(inboundSync)
	require.NoError(t, err)
	assert.Nil(t, fanout)
	require.Len(t, reply, 1)

	assert.Equal(t, [][]byte{{7, 7}}, sync.received)

	message, err := DecodeMessage(reply[0])
	require.NoError(t, err)
	assert.Equal(t, MessageSync, message.Type)
	assert.Equal(t, []byte{9, 9}, message.Data)
}

// TestServerConn_FansOutEphemeralOnce forwards a fresh ephemeral and drops a
// duplicate.
func TestServerConn_FansOutEphemeralOnce(t *testing.T) {
	t.Parallel()

	conn := newConn(t, &scriptedSync{})

	_, _, err := conn.Start(joinFixture(t))
	require.NoError(t, err)

	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	ephemeral, err := EncodeMessage(
		Message{
			Type: MessageEphemeral, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc-1", SessionID: "s", Count: 1, Data: payload,
		},
	)
	require.NoError(t, err)

	reply, fanout, err := conn.Receive(ephemeral)
	require.NoError(t, err)
	assert.Nil(t, reply)
	assert.Equal(t, ephemeral, fanout, "a fresh ephemeral is forwarded unchanged")

	_, fanoutAgain, err := conn.Receive(ephemeral)
	require.NoError(t, err)
	assert.Nil(t, fanoutAgain, "a duplicate ephemeral is dropped")
}

// TestServerConn_SyncChangedDrains produces sync frames when the document
// advanced from another source.
func TestServerConn_SyncChangedDrains(t *testing.T) {
	t.Parallel()

	sync := &scriptedSync{}
	conn := newConn(t, sync)

	_, _, err := conn.Start(joinFixture(t))
	require.NoError(t, err)

	sync.outbound = [][]byte{{3, 3}}
	frames, err := conn.SyncChanged()
	require.NoError(t, err)
	require.Len(t, frames, 1)

	message, err := DecodeMessage(frames[0])
	require.NoError(t, err)
	assert.Equal(t, MessageSync, message.Type)
	assert.Equal(t, []byte{3, 3}, message.Data)
}

// TestAdoptingServerConn_LearnsDocumentIDFromClient starts without a document id
// (so it announces nothing) and adopts the id from the client's first sync
// frame, then answers for that same id.
func TestAdoptingServerConn_LearnsDocumentIDFromClient(t *testing.T) {
	t.Parallel()

	sync := &scriptedSync{}

	conn, err := NewAdoptingServerConn(ServerConfig{ServerPeerID: "server"}, sync)
	require.NoError(t, err)

	out, accepted, err := conn.Start(joinFixture(t))
	require.NoError(t, err)
	require.True(t, accepted)
	require.Len(t, out, 1, "an adopting connection announces nothing until the client asks")
	assert.Empty(t, conn.DocumentID())

	sync.outbound = [][]byte{{9, 9}}

	inboundSync, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "client-chosen-doc", Data: []byte{7, 7},
		},
	)
	require.NoError(t, err)

	reply, _, err := conn.Receive(inboundSync)
	require.NoError(t, err)
	require.Len(t, reply, 1)
	assert.Equal(t, "client-chosen-doc", conn.DocumentID())

	message, err := DecodeMessage(reply[0])
	require.NoError(t, err)
	assert.Equal(
		t,
		"client-chosen-doc",
		message.DocumentID,
		"the server answers for the id the client requested",
	)
}

func TestAdoptingServerConn_LearnsDocumentIDFromEphemeral(t *testing.T) {
	t.Parallel()

	conn, err := NewAdoptingServerConn(ServerConfig{ServerPeerID: "server"}, &scriptedSync{})
	require.NoError(t, err)

	_, _, err = conn.Start(joinFixture(t))
	require.NoError(t, err)

	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	ephemeral, err := EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "ephemeral-doc",
			SessionID:  "session",
			Count:      0,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	_, fanout, err := conn.Receive(ephemeral)
	require.NoError(t, err)
	assert.Equal(t, "ephemeral-doc", conn.DocumentID())
	assert.Equal(t, ephemeral, fanout)
}

// TestAdoptingServerConn_RejectsSecondDocument holds the connection to the first
// document id it adopts.
func TestAdoptingServerConn_RejectsSecondDocument(t *testing.T) {
	t.Parallel()

	conn, err := NewAdoptingServerConn(ServerConfig{ServerPeerID: "server"}, &scriptedSync{})
	require.NoError(t, err)

	_, _, err = conn.Start(joinFixture(t))
	require.NoError(t, err)

	first, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc-a", Data: []byte{1},
		},
	)
	require.NoError(t, err)
	_, _, err = conn.Receive(first)
	require.NoError(t, err)

	second, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc-b", Data: []byte{2},
		},
	)
	require.NoError(t, err)
	_, _, err = conn.Receive(second)
	require.Error(t, err)
}

// TestServerConn_RejectsForeignDocument keeps a fixed-id connection from serving
// a different document than it was constructed for.
func TestServerConn_RejectsForeignDocument(t *testing.T) {
	t.Parallel()

	conn := newConn(t, &scriptedSync{})

	_, _, err := conn.Start(joinFixture(t))
	require.NoError(t, err)

	foreign, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "other-doc", Data: []byte{1},
		},
	)
	require.NoError(t, err)

	_, _, err = conn.Receive(foreign)
	require.Error(t, err)
}

func TestServerConn_RejectsForeignEphemeralBeforeDedup(t *testing.T) {
	t.Parallel()

	conn := newConn(t, &scriptedSync{})
	_, _, err := conn.Start(joinFixture(t))
	require.NoError(t, err)

	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	encode := func(documentID string) []byte {
		frame, err := EncodeMessage(
			Message{
				Type:       MessageEphemeral,
				SenderID:   "peer-a",
				TargetID:   "server",
				DocumentID: documentID,
				SessionID:  "session",
				Count:      1,
				Data:       payload,
			},
		)
		require.NoError(t, err)

		return frame
	}

	_, _, err = conn.Receive(encode("other-doc"))
	require.Error(t, err)

	_, fanout, err := conn.Receive(encode("doc-1"))
	require.NoError(t, err)
	assert.NotNil(t, fanout, "the rejected frame must not poison deduplication")
}

// TestServerConn_RequiresStart refuses frames before the handshake.
func TestServerConn_RequiresStart(t *testing.T) {
	t.Parallel()

	conn := newConn(t, &scriptedSync{})
	_, _, err := conn.Receive([]byte{0xa0})
	assert.Error(t, err)
}

func TestClientConn_RejectsForeignDocumentFrames(t *testing.T) {
	t.Parallel()

	sync := &scriptedSync{}
	conn, err := NewClientConn(
		ClientConfig{
			ClientPeerID: "client",
			DocumentID:   "doc-1",
		},
		sync,
	)
	require.NoError(t, err)

	peer, err := EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                "server",
			TargetID:                "client",
			SelectedProtocolVersion: ProtocolV1,
		},
	)
	require.NoError(t, err)

	_, err = conn.Receive(peer)
	require.NoError(t, err)

	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	for _, message := range []Message{
		{
			Type:       MessageSync,
			SenderID:   "server",
			TargetID:   "client",
			DocumentID: "other-doc",
			Data:       []byte{1},
		},
		{
			Type:       MessageRequest,
			SenderID:   "server",
			TargetID:   "client",
			DocumentID: "other-doc",
			Data:       []byte{1},
		},
		{
			Type:       MessageEphemeral,
			SenderID:   "server",
			TargetID:   "client",
			DocumentID: "other-doc",
			SessionID:  "session",
			Count:      1,
			Data:       payload,
		},
		{
			Type:       MessageDocUnavailable,
			SenderID:   "server",
			TargetID:   "client",
			DocumentID: "other-doc",
		},
	} {
		frame, err := EncodeMessage(message)
		require.NoError(t, err)

		_, err = conn.Receive(frame)
		assert.Error(t, err)
	}

	assert.Empty(t, sync.received, "foreign sync data must not be applied")
}

func TestClientConn_RejectsMalformedDocUnavailable(t *testing.T) {
	t.Parallel()

	conn, err := NewClientConn(
		ClientConfig{
			ClientPeerID: "client",
			DocumentID:   "doc-1",
		},
		&scriptedSync{},
	)
	require.NoError(t, err)

	peer, err := EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                "server",
			TargetID:                "client",
			SelectedProtocolVersion: ProtocolV1,
		},
	)
	require.NoError(t, err)

	_, err = conn.Receive(peer)
	require.NoError(t, err)

	malformed, err := marshal(
		map[string]any{
			"type":       MessageDocUnavailable,
			"senderId":   "server",
			"documentId": "doc-1",
		},
	)
	require.NoError(t, err)

	_, err = conn.Receive(malformed)
	require.Error(t, err)
}

func TestClientConn_SurfacesDocUnavailable(t *testing.T) {
	t.Parallel()

	conn, err := NewClientConn(
		ClientConfig{
			ClientPeerID: "client",
			DocumentID:   "doc-1",
		},
		&scriptedSync{},
	)
	require.NoError(t, err)

	peer, err := EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                "server",
			TargetID:                "client",
			SelectedProtocolVersion: ProtocolV1,
		},
	)
	require.NoError(t, err)

	_, err = conn.Receive(peer)
	require.NoError(t, err)

	frame, err := EncodeMessage(
		Message{
			Type:       MessageDocUnavailable,
			SenderID:   "server",
			TargetID:   "client",
			DocumentID: "doc-1",
		},
	)
	require.NoError(t, err)

	inbound, err := conn.Receive(frame)
	require.NoError(t, err)
	assert.True(t, inbound.Unavailable)
}
