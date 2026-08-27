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

type wireFixture struct {
	Description     string `json:"description"`
	FrameCBORBase64 string `json:"frameCborBase64"`
}

// TestDecodeHandshakeFrames_MatchJavaScriptBytes decodes the exact handshake
// frames the pinned WebSocket adapter emits.
func TestDecodeHandshakeFrames_MatchJavaScriptBytes(t *testing.T) {
	t.Parallel()

	t.Run(
		"join",
		func(t *testing.T) {
			t.Parallel()

			data := decodeBase64(t, readFixture[wireFixture](t, "wire-join.json").FrameCBORBase64)

			kind, err := FrameKind(data)
			require.NoError(t, err)
			assert.Equal(t, FrameJoin, kind)

			frame, err := DecodeJoinFrame(data)
			require.NoError(t, err)
			assert.Equal(t, "peer-a", frame.SenderID)
			assert.True(t, frame.SupportsV1())
			assert.False(t, frame.PeerMetadata.IsEphemeral)
		},
	)

	t.Run(
		"peer",
		func(t *testing.T) {
			t.Parallel()

			data := decodeBase64(t, readFixture[wireFixture](t, "wire-peer.json").FrameCBORBase64)

			frame, err := DecodePeerFrame(data)
			require.NoError(t, err)
			assert.Equal(t, "server", frame.SenderID)
			assert.Equal(t, "peer-a", frame.TargetID)
			assert.Equal(t, ProtocolV1, frame.SelectedProtocolVersion)
		},
	)

	t.Run(
		"error",
		func(t *testing.T) {
			t.Parallel()

			data := decodeBase64(t, readFixture[wireFixture](t, "wire-error.json").FrameCBORBase64)

			frame, err := DecodeErrorFrame(data)
			require.NoError(t, err)
			assert.Equal(t, "unauthorized", frame.Message)
		},
	)
}

// TestDecodeFramedMessages_MatchJavaScriptBytes decodes the framed document
// messages the adapter emits and confirms the ephemeral frame carries a
// presence payload our presence codec reads.
func TestDecodeFramedMessages_MatchJavaScriptBytes(t *testing.T) {
	t.Parallel()

	t.Run(
		"sync",
		func(t *testing.T) {
			t.Parallel()

			data := decodeBase64(t, readFixture[wireFixture](t, "wire-sync.json").FrameCBORBase64)

			kind, err := FrameKind(data)
			require.NoError(t, err)
			assert.Equal(t, string(MessageSync), kind)

			message, err := DecodeMessage(data)
			require.NoError(t, err)
			assert.Equal(t, MessageSync, message.Type)
			assert.Equal(t, "4NMNnkMhL2wRfvHYuG1uxN", message.DocumentID)
			assert.Equal(t, []byte{0, 1, 2, 3}, message.Data)
		},
	)

	t.Run(
		"ephemeral carries presence",
		func(t *testing.T) {
			t.Parallel()

			data := decodeBase64(t, readFixture[wireFixture](t, "wire-ephemeral.json").FrameCBORBase64)

			message, err := DecodeMessage(data)
			require.NoError(t, err)
			assert.Equal(t, MessageEphemeral, message.Type)

			presence, err := DecodePresence(message.Data)
			require.NoError(t, err)
			assert.Equal(t, PresenceHeartbeat, presence.Type)
		},
	)
}

// TestHandshakeFrames_RoundTrip encodes then decodes each handshake frame.
func TestHandshakeFrames_RoundTrip(t *testing.T) {
	t.Parallel()

	join, err := EncodeJoinFrame(NewJoinFrame("peer-a", PeerMetadata{IsEphemeral: true}))
	require.NoError(t, err)

	decodedJoin, err := DecodeJoinFrame(join)
	require.NoError(t, err)
	assert.True(t, decodedJoin.SupportsV1())
	assert.True(t, decodedJoin.PeerMetadata.IsEphemeral)

	peer, err := EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                "server",
			TargetID:                "peer-a",
			SelectedProtocolVersion: ProtocolV1,
		},
	)
	require.NoError(t, err)

	decodedPeer, err := DecodePeerFrame(peer)
	require.NoError(t, err)
	assert.Equal(t, ProtocolV1, decodedPeer.SelectedProtocolVersion)
}

// TestEncodeFrames_Validation rejects malformed handshake frames.
func TestEncodeFrames_Validation(t *testing.T) {
	t.Parallel()

	_, err := EncodeJoinFrame(JoinFrame{Type: FrameJoin, SupportedProtocolVersions: []string{ProtocolV1}})
	assert.Error(t, err, "join without a sender must be rejected")

	_, err = EncodePeerFrame(PeerFrame{Type: FramePeer, SenderID: "s", TargetID: "c"})
	assert.Error(t, err, "peer without a selected version must be rejected")

	_, err = EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                "s",
			TargetID:                "c",
			SelectedProtocolVersion: "2",
		},
	)
	assert.Error(t, err, "peer selecting a protocol other than v1 must be rejected")

	_, err = EncodeErrorFrame(ErrorFrame{Type: FrameError, SenderID: "s", TargetID: "c"})
	assert.Error(t, err, "error without a message must be rejected")

	_, err = EncodeErrorFrame(ErrorFrame{Type: FrameError, Message: "failed"})
	assert.Error(t, err, "error without sender and target ids must be rejected")
}
