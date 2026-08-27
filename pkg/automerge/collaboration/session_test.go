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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T) *ServerSession {
	t.Helper()

	session, err := NewServerSession(ServerConfig{ServerPeerID: "server"})
	require.NoError(t, err)

	return session
}

// TestServerSession_AcceptsJavaScriptJoin drives the handshake with the exact
// join frame the JavaScript client emits and checks the peer reply decodes.
func TestServerSession_AcceptsJavaScriptJoin(t *testing.T) {
	t.Parallel()

	join := decodeBase64(t, readFixture[wireFixture](t, "wire-join.json").FrameCBORBase64)

	session := newTestSession(t)
	handshake, err := session.Accept(join)
	require.NoError(t, err)

	assert.True(t, handshake.Accepted)
	assert.Equal(t, "peer-a", handshake.RemotePeerID)
	assert.Equal(t, "peer-a", session.RemotePeerID())

	peer, err := DecodePeerFrame(handshake.Reply)
	require.NoError(t, err)
	assert.Equal(t, "server", peer.SenderID)
	assert.Equal(t, "peer-a", peer.TargetID)
	assert.Equal(t, ProtocolV1, peer.SelectedProtocolVersion)
}

// TestServerSession_RejectsUnsupportedVersion returns an error frame, not an
// error, and does not complete the handshake.
func TestServerSession_RejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	join, err := marshal(
		JoinFrame{
			Type:                      FrameJoin,
			SenderID:                  "peer-a",
			SupportedProtocolVersions: []string{"999"},
		},
	)
	require.NoError(t, err)

	session := newTestSession(t)
	handshake, err := session.Accept(join)
	require.NoError(t, err)
	assert.False(t, handshake.Accepted)

	errorFrame, err := DecodeErrorFrame(handshake.Reply)
	require.NoError(t, err)
	assert.Equal(t, "peer-a", errorFrame.TargetID)
	assert.NotEmpty(t, errorFrame.Message)
}

// TestServerSession_RequiresHandshakeFirst refuses document frames before join.
func TestServerSession_RequiresHandshakeFirst(t *testing.T) {
	t.Parallel()

	sync, err := EncodeMessage(
		Message{
			Type: MessageSync, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc", Data: []byte{1},
		},
	)
	require.NoError(t, err)

	session := newTestSession(t)
	_, err = session.Receive(sync)
	assert.Error(t, err)
}

func acceptedSession(t *testing.T) *ServerSession {
	t.Helper()

	session := newTestSession(t)
	join := decodeBase64(t, readFixture[wireFixture](t, "wire-join.json").FrameCBORBase64)
	_, err := session.Accept(join)
	require.NoError(t, err)

	return session
}

// TestServerSession_RoutesSync classifies a sync frame for the SyncState.
func TestServerSession_RoutesSync(t *testing.T) {
	t.Parallel()

	session := acceptedSession(t)

	sync := decodeBase64(t, readFixture[wireFixture](t, "wire-sync.json").FrameCBORBase64)
	inbound, err := session.Receive(sync)
	require.NoError(t, err)
	assert.Equal(t, InboundSync, inbound.Kind)
	assert.Equal(t, "4NMNnkMhL2wRfvHYuG1uxN", inbound.Message.DocumentID)
	assert.Equal(t, []byte{0, 1, 2, 3}, inbound.Message.Data)
}

// TestServerSession_DeduplicatesEphemeral drops a repeated (session,count).
func TestServerSession_DeduplicatesEphemeral(t *testing.T) {
	t.Parallel()

	session := acceptedSession(t)

	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	frame := func(count uint64) []byte {
		data, err := EncodeMessage(
			Message{
				Type: MessageEphemeral, SenderID: "peer-a", TargetID: "server",
				DocumentID: "doc", SessionID: "s", Count: count, Data: payload,
			},
		)
		require.NoError(t, err)

		return data
	}

	first, err := session.Receive(frame(1))
	require.NoError(t, err)
	assert.Equal(t, InboundEphemeral, first.Kind)
	assert.False(t, first.Duplicate)

	second, err := session.Receive(frame(2))
	require.NoError(t, err)
	assert.False(t, second.Duplicate)

	replay, err := session.Receive(frame(2))
	require.NoError(t, err)
	assert.True(t, replay.Duplicate, "a repeated count is a gossip duplicate")

	older, err := session.Receive(frame(1))
	require.NoError(t, err)
	assert.True(t, older.Duplicate, "a lower count is already covered")

	// A different session with the same count is not a duplicate.
	otherSession, err := EncodeMessage(
		Message{
			Type: MessageEphemeral, SenderID: "peer-a", TargetID: "server",
			DocumentID: "doc", SessionID: "s2", Count: 1, Data: payload,
		},
	)
	require.NoError(t, err)

	other, err := session.Receive(otherSession)
	require.NoError(t, err)
	assert.False(t, other.Duplicate)
}

func TestServerSession_ValidatesRouteBeforeDeduplication(t *testing.T) {
	t.Parallel()

	session := acceptedSession(t)
	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	encode := func(senderID, targetID string) []byte {
		frame, err := EncodeMessage(
			Message{
				Type:       MessageEphemeral,
				SenderID:   senderID,
				TargetID:   targetID,
				DocumentID: "doc",
				SessionID:  "session",
				Count:      1,
				Data:       payload,
			},
		)
		require.NoError(t, err)

		return frame
	}

	_, err = session.Receive(encode("attacker", "server"))
	require.Error(t, err)

	_, err = session.Receive(encode("peer-a", "other-server"))
	require.Error(t, err)

	inbound, err := session.Receive(encode("peer-a", "server"))
	require.NoError(t, err)
	assert.False(t, inbound.Duplicate)
}

func TestServerSession_CapsEphemeralSessions(t *testing.T) {
	t.Parallel()

	session := acceptedSession(t)
	payload, err := EncodePresence(PresenceMessage{Type: PresenceHeartbeat})
	require.NoError(t, err)

	for i := range maxEphemeralSessions {
		frame, err := EncodeMessage(
			Message{
				Type:       MessageEphemeral,
				SenderID:   "peer-a",
				TargetID:   "server",
				DocumentID: "doc",
				SessionID:  fmt.Sprintf("session-%d", i),
				Count:      1,
				Data:       payload,
			},
		)
		require.NoError(t, err)

		_, err = session.Receive(frame)
		require.NoError(t, err)
	}

	overflow, err := EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc",
			SessionID:  "overflow",
			Count:      1,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	_, err = session.Receive(overflow)
	assert.Error(t, err)

	existing, err := EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc",
			SessionID:  "session-0",
			Count:      2,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	inbound, err := session.Receive(existing)
	require.NoError(t, err)
	assert.False(t, inbound.Duplicate, "existing sessions remain usable at the cap")
}

// TestServerSession_IgnoresRemoteHeads treats unrecognised control frames as
// valid-but-ignored rather than errors.
func TestServerSession_IgnoresRemoteHeads(t *testing.T) {
	t.Parallel()

	session := acceptedSession(t)

	frame, err := marshal(
		map[string]any{
			"type":     "remote-heads-changed",
			"senderId": "peer-a",
			"targetId": "server",
		},
	)
	require.NoError(t, err)

	inbound, err := session.Receive(frame)
	require.NoError(t, err)
	assert.Equal(t, InboundIgnored, inbound.Kind)
}
