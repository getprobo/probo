// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package yprotocol implements the wire-format helpers from y-protocols
// (https://github.com/yjs/y-protocols) that the collaboration relay needs.
//
// Two protocols are supported:
//   - sync (messageSync = 0): SyncStep1, SyncStep2, Update sub-messages.
//   - awareness (messageAwareness = 1): opaque awareness payload.
package yprotocol

import (
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/collab/lib0"
)

const (
	// MessageSync wraps a sync sub-message (SyncStep1, SyncStep2, Update).
	MessageSync byte = 0

	// MessageAwareness carries an opaque awareness update payload.
	MessageAwareness byte = 1
)

const (
	// SyncStep1 contains the sender's state vector. The receiver replies
	// with a SyncStep2 carrying the missing updates.
	SyncStep1 byte = 0

	// SyncStep2 contains an update binary computed against the peer's
	// state vector previously received in a SyncStep1.
	SyncStep2 byte = 1

	// SyncUpdate carries an unsolicited update binary that the receiver
	// should apply directly.
	SyncUpdate byte = 2
)

var (
	// ErrUnknownMessageType is returned when an incoming frame has a
	// top-level message type that this package does not understand.
	ErrUnknownMessageType = errors.New("unknown message type")

	// ErrUnknownSyncSubtype is returned when an incoming sync frame has
	// an unknown sync sub-message tag.
	ErrUnknownSyncSubtype = errors.New("unknown sync subtype")

	// ErrTrailingBytes is returned when a frame contains data after the
	// last expected field. The relay treats this as a hard error so a
	// malformed peer cannot silently corrupt state.
	ErrTrailingBytes = errors.New("trailing bytes after message")
)

type (
	// Message is a parsed top-level wire frame.
	Message struct {
		// Type is the top-level message tag (MessageSync or MessageAwareness).
		Type byte

		// Subtype is the sync sub-message tag (SyncStep1, SyncStep2,
		// SyncUpdate). Only meaningful when Type == MessageSync.
		Subtype byte

		// Payload is the opaque body bytes (state vector for SyncStep1,
		// update bytes for SyncStep2/SyncUpdate, awareness blob for
		// awareness). Never nil for a valid frame; may be empty.
		Payload []byte
	}
)

// ReadMessage parses a single y-protocol frame.
//
// The frame must contain exactly one message: trailing bytes are reported
// as ErrTrailingBytes. Frames are produced by the WriteSyncStep1/2/Update
// or WriteAwareness functions in this package, or by upstream y-protocols
// peers.
func ReadMessage(frame []byte) (Message, error) {
	dec := lib0.NewDecoder(frame)

	mt, err := lib0.ReadVarUint(dec)
	if err != nil {
		return Message{}, fmt.Errorf("cannot read message type: %w", err)
	}

	switch mt {
	case uint64(MessageSync):
		st, err := lib0.ReadVarUint(dec)
		if err != nil {
			return Message{}, fmt.Errorf("cannot read sync subtype: %w", err)
		}

		switch st {
		case uint64(SyncStep1), uint64(SyncStep2), uint64(SyncUpdate):
		default:
			return Message{}, fmt.Errorf("%w: %d", ErrUnknownSyncSubtype, st)
		}

		payload, err := lib0.ReadVarUint8Array(dec)
		if err != nil {
			return Message{}, fmt.Errorf("cannot read sync payload: %w", err)
		}

		if dec.HasMore() {
			return Message{}, ErrTrailingBytes
		}

		return Message{Type: MessageSync, Subtype: byte(st), Payload: payload}, nil

	case uint64(MessageAwareness):
		payload, err := lib0.ReadVarUint8Array(dec)
		if err != nil {
			return Message{}, fmt.Errorf("cannot read awareness payload: %w", err)
		}

		if dec.HasMore() {
			return Message{}, ErrTrailingBytes
		}

		return Message{Type: MessageAwareness, Payload: payload}, nil

	default:
		return Message{}, fmt.Errorf("%w: %d", ErrUnknownMessageType, mt)
	}
}
