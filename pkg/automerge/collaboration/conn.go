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

// SyncSession is the subset of an Automerge sync state the connection driver
// needs. *automerge.SyncState satisfies it, so the driver never touches the
// CRDT engine directly.
type SyncSession interface {
	// GenerateMessage returns the next outbound sync message for this peer, or
	// ok=false when the peer is up to date.
	GenerateMessage(ctx context.Context) ([]byte, bool, error)
	// ReceiveMessage applies an inbound sync message from this peer.
	ReceiveMessage(ctx context.Context, message []byte) error
}

// ServerConn is the synchronous, deterministic driver for one collaboration
// connection on the server side. It owns no socket and spawns no goroutines: the
// adapter reads a frame, calls the matching method, and sends the frames
// returned. This keeps the protocol behavior unit-testable and leaves I/O,
// rooms, and authentication to the server wiring.
//
// A ServerConn is not safe for concurrent use; the adapter drives it from one
// connection goroutine.
type ServerConn struct {
	session      *ServerSession
	sync         SyncSession
	documentID   string
	adoptDocID   bool
	serverPeerID string
	remotePeerID string
	started      bool
}

// NewServerConn creates a connection driver for a known document. documentID is
// the automerge-repo document id this connection is scoped to, and sync is the
// peer's sync state. Use this when the server already knows which document the
// connection serves and both peers agree on the id; the server announces the
// document proactively after the handshake.
//
// When the id the client will request is not known ahead of time (the common
// production case, where the frontend picks the automerge:<id> URL), use
// NewAdoptingServerConn instead.
func NewServerConn(config ServerConfig, documentID string, sync SyncSession) (*ServerConn, error) {
	if documentID == "" {
		return nil, fmt.Errorf("server connection requires a document id")
	}

	return newServerConn(config, documentID, false, sync)
}

// NewAdoptingServerConn creates a connection driver that learns its document id
// from the client's first sync or request frame. Because the id is unknown until
// then, the server does not announce the document in Start; it answers the
// client once the client asks for a specific document. Every subsequent frame on
// the connection must reference the same id.
func NewAdoptingServerConn(config ServerConfig, sync SyncSession) (*ServerConn, error) {
	return newServerConn(config, "", true, sync)
}

func newServerConn(config ServerConfig, documentID string, adopt bool, sync SyncSession) (*ServerConn, error) {
	session, err := NewServerSession(config)
	if err != nil {
		return nil, err
	}

	if sync == nil {
		return nil, fmt.Errorf("server connection requires a sync session")
	}

	return &ServerConn{
		session:      session,
		sync:         sync,
		documentID:   documentID,
		adoptDocID:   adopt,
		serverPeerID: config.ServerPeerID,
	}, nil
}

// Start processes the client's join frame and returns the frames to send: the
// handshake reply, then, when accepted, the initial sync frames that announce
// the document. When accepted is false the single reply is an error frame and
// the socket should be closed after sending it.
func (c *ServerConn) Start(ctx context.Context, joinFrame []byte) (out [][]byte, accepted bool, err error) {
	if c.started {
		return nil, false, fmt.Errorf("server connection already started")
	}

	handshake, err := c.session.Accept(joinFrame)
	if err != nil {
		return nil, false, err
	}

	out = append(out, handshake.Reply)

	if !handshake.Accepted {
		return out, false, nil
	}

	c.started = true
	c.remotePeerID = handshake.RemotePeerID

	// In adopt mode the document id is unknown until the client asks for it, so
	// there is nothing to announce yet: the client drives with a request frame.
	if c.adoptDocID {
		return out, true, nil
	}

	// Announce: drain the initial sync messages the server has for this peer.
	syncFrames, err := c.drainSync(ctx)
	if err != nil {
		return nil, false, err
	}

	return append(out, syncFrames...), true, nil
}

// Receive handles one post-handshake frame. A sync or request frame is applied
// and answered with the resulting sync frames (reply). A non-duplicate ephemeral
// frame is returned in fanout to publish to the room, unchanged. Duplicate
// ephemerals, doc-unavailable, and unrecognised control frames yield nothing.
func (c *ServerConn) Receive(ctx context.Context, frame []byte) (reply [][]byte, fanout []byte, err error) {
	if !c.started {
		return nil, nil, fmt.Errorf("received a frame before the connection was started")
	}

	inbound, err := c.session.Receive(frame)
	if err != nil {
		return nil, nil, err
	}

	switch inbound.Kind {
	case InboundSync:
		if err := c.adoptDocumentID(inbound.Message.DocumentID); err != nil {
			return nil, nil, err
		}

		if err := c.sync.ReceiveMessage(ctx, inbound.Message.Data); err != nil {
			return nil, nil, fmt.Errorf("cannot apply inbound sync message: %w", err)
		}

		frames, err := c.drainSync(ctx)
		if err != nil {
			return nil, nil, err
		}

		return frames, nil, nil
	case InboundEphemeral:
		if inbound.Duplicate {
			return nil, nil, nil
		}

		return nil, frame, nil
	default:
		// InboundDocUnavailable and InboundIgnored need no server action: the
		// gateway is the document authority.
		return nil, nil, nil
	}
}

// SyncChanged drains any sync messages produced because the document changed
// from another source (a peer's merge, a server-side edit). The adapter calls it
// when the room signals the document advanced.
func (c *ServerConn) SyncChanged(ctx context.Context) ([][]byte, error) {
	if !c.started {
		return nil, fmt.Errorf("cannot sync a connection before it is started")
	}

	return c.drainSync(ctx)
}

// adoptDocumentID binds the connection to the document id the client asked for.
// A fixed-id connection rejects a mismatching id; an adopting one records the
// first non-empty id it sees and then holds the peer to it. This guarantees one
// connection only ever serves a single document.
func (c *ServerConn) adoptDocumentID(id string) error {
	if id == "" {
		return nil
	}

	if c.documentID == "" {
		c.documentID = id

		return nil
	}

	if c.documentID != id {
		return fmt.Errorf(
			"connection scoped to document %q received a frame for document %q",
			c.documentID,
			id,
		)
	}

	return nil
}

// drainSync generates sync frames until the peer is up to date. The server is
// always the document authority, so every generated message is a sync frame,
// never a request. Before the document id is known (adopt mode, prior to the
// client's first request) there is nothing to send.
func (c *ServerConn) drainSync(ctx context.Context) ([][]byte, error) {
	if c.documentID == "" {
		return nil, nil
	}

	var frames [][]byte

	for {
		message, ok, err := c.sync.GenerateMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot generate sync message: %w", err)
		}

		if !ok {
			return frames, nil
		}

		frame, err := EncodeMessage(
			Message{
				Type:       MessageSync,
				SenderID:   c.serverPeerID,
				TargetID:   c.remotePeerID,
				DocumentID: c.documentID,
				Data:       message,
			},
		)
		if err != nil {
			return nil, err
		}

		frames = append(frames, frame)
	}
}

// RemotePeerID returns the connected client's peer id after Start.
func (c *ServerConn) RemotePeerID() string {
	return c.remotePeerID
}

// DocumentID returns the document id this connection serves. For an adopting
// connection it is empty until the client's first sync or request frame binds
// it.
func (c *ServerConn) DocumentID() string {
	return c.documentID
}
