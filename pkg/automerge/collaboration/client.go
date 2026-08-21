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
	"context"
	"fmt"
)

// ClientConfig configures the client side of a collaboration connection, used by
// Go agents that participate as repo peers.
type ClientConfig struct {
	// ClientPeerID is the peer id the client advertises. It is the client's own
	// identifier, chosen by the client.
	ClientPeerID string
	// PeerMetadata is the metadata the client presents. Optional.
	PeerMetadata PeerMetadata
	// DocumentID is the repo document id this connection syncs.
	DocumentID string
	// StartsEmpty reports whether the local document has no changes yet, which
	// selects the first outbound sync message's type: a peer that does not have
	// the document sends a request, one that already has it sends a sync. This
	// mirrors the repo synchronizer's isNew rule.
	StartsEmpty bool
}

// ClientConn is the synchronous, deterministic driver for the client side of one
// collaboration connection. Like ServerConn it owns no socket and starts no
// goroutines: the caller sends the join frame it returns, then feeds inbound
// frames and sends the frames it returns. A Go agent wraps this with a real
// socket and its own document.
//
// A ClientConn is not safe for concurrent use.
type ClientConn struct {
	config       ClientConfig
	sync         SyncSession
	serverPeerID string
	joined       bool
	sentFirst    bool
	highestCount map[string]uint64
}

// NewClientConn creates a client connection driver for one document.
func NewClientConn(config ClientConfig, sync SyncSession) (*ClientConn, error) {
	if config.ClientPeerID == "" {
		return nil, fmt.Errorf("client connection requires a client peer id")
	}

	if config.DocumentID == "" {
		return nil, fmt.Errorf("client connection requires a document id")
	}

	if sync == nil {
		return nil, fmt.Errorf("client connection requires a sync session")
	}

	return &ClientConn{
		config:       config,
		sync:         sync,
		highestCount: make(map[string]uint64),
	}, nil
}

// Start returns the join frame the client sends first.
func (c *ClientConn) Start() ([]byte, error) {
	return EncodeJoinFrame(NewJoinFrame(c.config.ClientPeerID, c.config.PeerMetadata))
}

// ClientInbound is the result of handling one inbound server frame.
type ClientInbound struct {
	// Outgoing frames to send to the server (sync or, for the first message, a
	// request).
	Outgoing [][]byte
	// Ephemeral is a non-duplicate ephemeral message received from the server
	// (gossiped from another peer), or nil. Its Data is a payload for the
	// application, for example a presence message.
	Ephemeral *Message
	// Unavailable is true when the server reported the document unavailable.
	Unavailable bool
	// ServerError carries the message of a server error frame, which precedes the
	// server closing the connection.
	ServerError string
}

// Receive handles one inbound server frame. On the peer reply it completes the
// handshake and emits the client's initial sync; on a sync or request it applies
// the payload and emits the resulting sync; on an ephemeral it de-duplicates and
// surfaces the payload; it reports an error or doc-unavailable frame.
func (c *ClientConn) Receive(ctx context.Context, frame []byte) (ClientInbound, error) {
	kind, err := FrameKind(frame)
	if err != nil {
		return ClientInbound{}, err
	}

	switch kind {
	case FramePeer:
		peer, err := DecodePeerFrame(frame)
		if err != nil {
			return ClientInbound{}, err
		}

		if peer.SelectedProtocolVersion != ProtocolV1 {
			return ClientInbound{}, fmt.Errorf("server selected unsupported protocol version %q",
				peer.SelectedProtocolVersion)
		}

		c.joined = true
		c.serverPeerID = peer.SenderID

		outgoing, err := c.drainSync(ctx)
		if err != nil {
			return ClientInbound{}, err
		}

		return ClientInbound{Outgoing: outgoing}, nil
	case FrameError:
		errorFrame, err := DecodeErrorFrame(frame)
		if err != nil {
			return ClientInbound{}, err
		}

		return ClientInbound{ServerError: errorFrame.Message}, nil
	}

	if !c.joined {
		return ClientInbound{}, fmt.Errorf("received %q before the handshake completed", kind)
	}

	switch MessageType(kind) {
	case MessageSync, MessageRequest:
		message, err := DecodeMessage(frame)
		if err != nil {
			return ClientInbound{}, err
		}

		if err := c.sync.ReceiveMessage(ctx, message.Data); err != nil {
			return ClientInbound{}, fmt.Errorf("cannot apply inbound sync message: %w", err)
		}

		outgoing, err := c.drainSync(ctx)
		if err != nil {
			return ClientInbound{}, err
		}

		return ClientInbound{Outgoing: outgoing}, nil
	case MessageEphemeral:
		message, err := DecodeMessage(frame)
		if err != nil {
			return ClientInbound{}, err
		}

		if c.seenEphemeral(message) {
			return ClientInbound{}, nil
		}

		return ClientInbound{Ephemeral: &message}, nil
	case MessageDocUnavailable:
		return ClientInbound{Unavailable: true}, nil
	default:
		return ClientInbound{}, nil
	}
}

// SyncChanged drains sync frames after the local document changed, so local
// edits propagate to the server.
func (c *ClientConn) SyncChanged(ctx context.Context) ([][]byte, error) {
	if !c.joined {
		return nil, fmt.Errorf("cannot sync before the handshake completed")
	}

	return c.drainSync(ctx)
}

// Ephemeral builds an ephemeral frame carrying an application payload (such as a
// presence message) for the server to gossip. The caller supplies the session id
// chosen at startup and a per-session count that must strictly increase.
func (c *ClientConn) Ephemeral(sessionID string, count uint64, payload []byte) ([]byte, error) {
	if !c.joined {
		return nil, fmt.Errorf("cannot send ephemeral before the handshake completed")
	}

	return EncodeMessage(Message{
		Type:       MessageEphemeral,
		SenderID:   c.config.ClientPeerID,
		TargetID:   c.serverPeerID,
		DocumentID: c.config.DocumentID,
		SessionID:  sessionID,
		Count:      count,
		Data:       payload,
	})
}

// drainSync generates sync frames until the client is up to date. The first
// outbound message is a request when the local document started empty, matching
// the repo synchronizer; every later message is a sync.
func (c *ClientConn) drainSync(ctx context.Context) ([][]byte, error) {
	var frames [][]byte

	for {
		message, ok, err := c.sync.GenerateMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot generate sync message: %w", err)
		}

		if !ok {
			return frames, nil
		}

		messageType := MessageSync
		if !c.sentFirst && c.config.StartsEmpty {
			messageType = MessageRequest
		}

		c.sentFirst = true

		frame, err := EncodeMessage(Message{
			Type:       messageType,
			SenderID:   c.config.ClientPeerID,
			TargetID:   c.serverPeerID,
			DocumentID: c.config.DocumentID,
			Data:       message,
		})
		if err != nil {
			return nil, err
		}

		frames = append(frames, frame)
	}
}

func (c *ClientConn) seenEphemeral(message Message) bool {
	highest, ok := c.highestCount[message.SessionID]
	if ok && message.Count <= highest {
		return true
	}

	c.highestCount[message.SessionID] = message.Count

	return false
}

// ServerPeerID returns the server's peer id once the handshake has completed.
func (c *ClientConn) ServerPeerID() string {
	return c.serverPeerID
}
