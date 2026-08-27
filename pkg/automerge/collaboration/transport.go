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
	"slices"
)

// ProtocolV1 is the only automerge-repo WebSocket protocol version this package
// speaks.
const ProtocolV1 = "1"

// Handshake frame types exchanged before document messages flow.
const (
	// FrameJoin is the client's first frame, announcing its peer id and the
	// protocol versions it supports.
	FrameJoin = "join"
	// FramePeer is the server's reply, selecting a protocol version and
	// advertising the server peer id.
	FramePeer = "peer"
	// FrameError is sent by either side to report a fatal error immediately
	// before closing the socket.
	FrameError = "error"
)

// PeerMetadata is the optional metadata a peer presents in the handshake. It is
// not identity: the peer id and metadata are peer-chosen, so a gateway
// authenticates the connection out of band and never trusts these as a user.
type PeerMetadata struct {
	StorageID   string `cbor:"storageId,omitempty"`
	IsEphemeral bool   `cbor:"isEphemeral,omitempty"`
}

// JoinFrame is the client handshake frame. It has no target id because it is
// sent before the client knows the server's peer id.
type JoinFrame struct {
	Type                      string       `cbor:"type"`
	SenderID                  string       `cbor:"senderId"`
	PeerMetadata              PeerMetadata `cbor:"peerMetadata"`
	SupportedProtocolVersions []string     `cbor:"supportedProtocolVersions"`
}

// PeerFrame is the server's reply to a join frame.
type PeerFrame struct {
	Type                    string       `cbor:"type"`
	SenderID                string       `cbor:"senderId"`
	TargetID                string       `cbor:"targetId"`
	PeerMetadata            PeerMetadata `cbor:"peerMetadata"`
	SelectedProtocolVersion string       `cbor:"selectedProtocolVersion"`
}

// ErrorFrame reports a fatal error; the sender closes the socket after it.
type ErrorFrame struct {
	Type     string `cbor:"type"`
	SenderID string `cbor:"senderId"`
	TargetID string `cbor:"targetId"`
	Message  string `cbor:"message"`
}

// NewJoinFrame builds a client join frame advertising ProtocolV1.
func NewJoinFrame(senderID string, metadata PeerMetadata) JoinFrame {
	return JoinFrame{
		Type:                      FrameJoin,
		SenderID:                  senderID,
		PeerMetadata:              metadata,
		SupportedProtocolVersions: []string{ProtocolV1},
	}
}

// EncodeJoinFrame encodes a client join frame.
func EncodeJoinFrame(frame JoinFrame) ([]byte, error) {
	if frame.Type != FrameJoin {
		return nil, fmt.Errorf("join frame has type %q", frame.Type)
	}

	if frame.SenderID == "" {
		return nil, fmt.Errorf("join frame is missing a sender id")
	}

	if len(frame.SupportedProtocolVersions) == 0 {
		return nil, fmt.Errorf("join frame lists no supported protocol versions")
	}

	return marshal(frame)
}

// EncodePeerFrame encodes a server peer reply frame.
func EncodePeerFrame(frame PeerFrame) ([]byte, error) {
	if frame.Type != FramePeer {
		return nil, fmt.Errorf("peer frame has type %q", frame.Type)
	}

	if frame.SenderID == "" || frame.TargetID == "" {
		return nil, fmt.Errorf("peer frame is missing a sender or target id")
	}

	if frame.SelectedProtocolVersion != ProtocolV1 {
		return nil, fmt.Errorf(
			"peer frame selected unsupported protocol version %q",
			frame.SelectedProtocolVersion,
		)
	}

	return marshal(frame)
}

// EncodeErrorFrame encodes an error frame.
func EncodeErrorFrame(frame ErrorFrame) ([]byte, error) {
	if frame.Type != FrameError {
		return nil, fmt.Errorf("error frame has type %q", frame.Type)
	}

	if frame.SenderID == "" || frame.TargetID == "" {
		return nil, fmt.Errorf("error frame is missing a sender or target id")
	}

	if frame.Message == "" {
		return nil, fmt.Errorf("error frame is missing a message")
	}

	return marshal(frame)
}

// frameType peeks only the discriminator so a reader can route a frame to the
// right decoder without decoding it twice into the wrong shape.
type frameType struct {
	Type string `cbor:"type"`
}

// FrameKind returns the type discriminator of a raw WebSocket frame.
func FrameKind(data []byte) (string, error) {
	var peek frameType
	if err := unmarshal(data, &peek); err != nil {
		return "", fmt.Errorf("cannot read frame type: %w", err)
	}

	if peek.Type == "" {
		return "", fmt.Errorf("frame is missing a type")
	}

	return peek.Type, nil
}

// DecodeJoinFrame decodes and validates a client join frame, returning whether
// it supports ProtocolV1.
func DecodeJoinFrame(data []byte) (JoinFrame, error) {
	var frame JoinFrame
	if err := unmarshal(data, &frame); err != nil {
		return JoinFrame{}, fmt.Errorf("cannot decode join frame: %w", err)
	}

	if frame.Type != FrameJoin {
		return JoinFrame{}, fmt.Errorf("expected a join frame, got %q", frame.Type)
	}

	if frame.SenderID == "" {
		return JoinFrame{}, fmt.Errorf("join frame is missing a sender id")
	}

	return frame, nil
}

// SupportsV1 reports whether the join frame offers ProtocolV1.
func (f JoinFrame) SupportsV1() bool {
	return slices.Contains(f.SupportedProtocolVersions, ProtocolV1)
}

// DecodePeerFrame decodes and validates a server peer frame.
func DecodePeerFrame(data []byte) (PeerFrame, error) {
	var frame PeerFrame
	if err := unmarshal(data, &frame); err != nil {
		return PeerFrame{}, fmt.Errorf("cannot decode peer frame: %w", err)
	}

	if frame.Type != FramePeer {
		return PeerFrame{}, fmt.Errorf("expected a peer frame, got %q", frame.Type)
	}

	if frame.SenderID == "" || frame.TargetID == "" {
		return PeerFrame{}, fmt.Errorf("peer frame is missing a sender or target id")
	}

	if frame.SelectedProtocolVersion != ProtocolV1 {
		return PeerFrame{}, fmt.Errorf(
			"peer frame selected unsupported protocol version %q",
			frame.SelectedProtocolVersion,
		)
	}

	return frame, nil
}

// DecodeErrorFrame decodes an error frame.
func DecodeErrorFrame(data []byte) (ErrorFrame, error) {
	var frame ErrorFrame
	if err := unmarshal(data, &frame); err != nil {
		return ErrorFrame{}, fmt.Errorf("cannot decode error frame: %w", err)
	}

	if frame.Type != FrameError {
		return ErrorFrame{}, fmt.Errorf("expected an error frame, got %q", frame.Type)
	}

	if frame.SenderID == "" || frame.TargetID == "" {
		return ErrorFrame{}, fmt.Errorf("error frame is missing a sender or target id")
	}

	if frame.Message == "" {
		return ErrorFrame{}, fmt.Errorf("error frame is missing a message")
	}

	return frame, nil
}
