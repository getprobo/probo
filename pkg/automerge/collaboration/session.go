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

import "fmt"

// ServerConfig configures one server side of a collaboration connection.
type ServerConfig struct {
	// ServerPeerID is the peer id the server advertises in its peer reply. It is
	// the server's own identifier and is unrelated to any client identity.
	ServerPeerID string
	// PeerMetadata is the metadata the server presents. Optional.
	PeerMetadata PeerMetadata
}

// ServerSession is the transport-agnostic state machine for one collaboration
// connection as seen by the server. It negotiates the handshake, routes inbound
// frames, and de-duplicates gossiped ephemeral messages. It performs no I/O and
// holds no CRDT state, so the server wiring only has to move bytes and act on
// the returned decisions; document authority stays with pkg/automerge and peer
// authentication stays with the server connection.
//
// A ServerSession is not safe for concurrent use; drive it from one connection
// goroutine.
type ServerSession struct {
	config ServerConfig

	joined       bool
	remotePeerID string

	// highestCount tracks, per ephemeral session, the greatest count applied.
	// The protocol guarantees a session's count strictly increases, so a count
	// at or below the highest already seen is a gossip duplicate. This bounds
	// memory to one entry per session rather than one per message.
	highestCount map[string]uint64
}

// NewServerSession creates a server session. ServerPeerID must be set.
func NewServerSession(config ServerConfig) (*ServerSession, error) {
	if config.ServerPeerID == "" {
		return nil, fmt.Errorf("server session requires a server peer id")
	}

	return &ServerSession{
		config:       config,
		highestCount: make(map[string]uint64),
	}, nil
}

// Handshake is the outcome of processing a client join frame.
type Handshake struct {
	// Reply is the CBOR frame to send back: a peer frame when Accepted, or an
	// error frame when not. It is always sent; when not Accepted the socket
	// should then be closed.
	Reply []byte
	// Accepted reports whether the connection may proceed.
	Accepted bool
	// RemotePeerID is the client's peer id from the join frame. It is
	// peer-chosen and must not be treated as an authenticated identity.
	RemotePeerID string
}

// Accept processes the client's join frame and produces the server's reply. A
// malformed frame returns an error. A well-formed join that does not offer a
// supported protocol version returns an accepted=false handshake carrying an
// error frame to send before closing.
func (s *ServerSession) Accept(joinData []byte) (Handshake, error) {
	if s.joined {
		return Handshake{}, fmt.Errorf("session already completed its handshake")
	}

	join, err := DecodeJoinFrame(joinData)
	if err != nil {
		return Handshake{}, err
	}

	if !join.SupportsV1() {
		reply, encodeErr := EncodeErrorFrame(
			ErrorFrame{
				Type:     FrameError,
				SenderID: s.config.ServerPeerID,
				TargetID: join.SenderID,
				Message:  "unsupported protocol version",
			},
		)
		if encodeErr != nil {
			return Handshake{}, encodeErr
		}

		return Handshake{Reply: reply, Accepted: false, RemotePeerID: join.SenderID}, nil
	}

	reply, err := EncodePeerFrame(
		PeerFrame{
			Type:                    FramePeer,
			SenderID:                s.config.ServerPeerID,
			TargetID:                join.SenderID,
			PeerMetadata:            s.config.PeerMetadata,
			SelectedProtocolVersion: ProtocolV1,
		},
	)
	if err != nil {
		return Handshake{}, err
	}

	s.joined = true
	s.remotePeerID = join.SenderID

	return Handshake{Reply: reply, Accepted: true, RemotePeerID: join.SenderID}, nil
}

// InboundKind classifies a routed frame for the server wiring.
type InboundKind int

const (
	// InboundSync is a sync or request message to feed into the peer's SyncState.
	InboundSync InboundKind = iota
	// InboundEphemeral is a gossiped payload to fan out to the room.
	InboundEphemeral
	// InboundDocUnavailable reports the peer does not have the document.
	InboundDocUnavailable
	// InboundIgnored is a frame the gateway does not act on (for example the
	// remote-heads messages a single-authority server does not need).
	InboundIgnored
)

// Inbound is the decision for one received document frame.
type Inbound struct {
	Kind    InboundKind
	Message Message
	// Duplicate is true for an ephemeral message already seen for its session,
	// which must not be re-applied or re-broadcast.
	Duplicate bool
}

// Receive routes a document frame received after the handshake. Sync and request
// frames are returned for the caller to apply to the peer's SyncState; ephemeral
// frames are de-duplicated by session and count; the remote-heads control
// messages are ignored by a single-authority gateway.
func (s *ServerSession) Receive(frameData []byte) (Inbound, error) {
	if !s.joined {
		return Inbound{}, fmt.Errorf("received a document frame before the handshake completed")
	}

	kind, err := FrameKind(frameData)
	if err != nil {
		return Inbound{}, err
	}

	switch MessageType(kind) {
	case MessageSync, MessageRequest:
		message, err := DecodeMessage(frameData)
		if err != nil {
			return Inbound{}, err
		}

		return Inbound{Kind: InboundSync, Message: message}, nil
	case MessageEphemeral:
		message, err := DecodeMessage(frameData)
		if err != nil {
			return Inbound{}, err
		}

		return Inbound{
			Kind:      InboundEphemeral,
			Message:   message,
			Duplicate: s.seenEphemeral(message),
		}, nil
	case MessageDocUnavailable:
		message, err := DecodeMessage(frameData)
		if err != nil {
			return Inbound{}, err
		}

		return Inbound{Kind: InboundDocUnavailable, Message: message}, nil
	default:
		// remote-subscription-change, remote-heads-changed, or anything a newer
		// peer introduces: acknowledged as a valid frame but not acted on.
		return Inbound{Kind: InboundIgnored}, nil
	}
}

// seenEphemeral records an ephemeral message and reports whether it is a gossip
// duplicate. A message whose count is at or below the highest already recorded
// for its session has been seen; the protocol guarantees counts increase.
func (s *ServerSession) seenEphemeral(message Message) bool {
	highest, ok := s.highestCount[message.SessionID]
	if ok && message.Count <= highest {
		return true
	}

	s.highestCount[message.SessionID] = message.Count

	return false
}

// RemotePeerID returns the client's peer id once the handshake has completed.
func (s *ServerSession) RemotePeerID() string {
	return s.remotePeerID
}
