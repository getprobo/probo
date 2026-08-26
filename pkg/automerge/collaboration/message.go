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

// MessageType is the discriminator shared by every repo message.
type MessageType string

const (
	// MessageSync carries an Automerge sync message for a document.
	MessageSync MessageType = "sync"
	// MessageRequest is the initial sync that also asks whether the peer has the
	// document at all.
	MessageRequest MessageType = "request"
	// MessageEphemeral carries a gossiped, non-persisted payload (such as
	// presence) for a document.
	MessageEphemeral MessageType = "ephemeral"
	// MessageDocUnavailable reports that neither the peer nor its peers hold the
	// document.
	MessageDocUnavailable MessageType = "doc-unavailable"
)

// Message is one document-scoped repo message carried inside the WebSocket
// adapter framing. For sync and request Data is an Automerge sync message our
// engine owns, and for ephemeral Data is a CBOR presence payload.
//
// Sync bytes are intentionally left opaque here so this package never
// re-implements the CRDT wire format.
type Message struct {
	Type       MessageType `cbor:"type"`
	SenderID   string      `cbor:"senderId"`
	TargetID   string      `cbor:"targetId"`
	DocumentID string      `cbor:"documentId"`

	// Data is the Automerge sync message for sync and request messages, and the
	// CBOR presence payload for ephemeral messages.
	Data []byte `cbor:"data,omitempty"`

	// SessionID and Count identify an ephemeral message for gossip
	// de-duplication and are unset for other types.
	SessionID string `cbor:"sessionId,omitempty"`
	Count     uint64 `cbor:"count,omitempty"`
}

// validate checks that a message carries exactly the fields its type requires,
// which guards both directions of the codec.
func (m Message) validate() error {
	if m.SenderID == "" {
		return fmt.Errorf("repo message is missing a sender id")
	}

	switch m.Type {
	case MessageSync, MessageRequest:
		if m.DocumentID == "" {
			return fmt.Errorf("%s message is missing a document id", m.Type)
		}

		if len(m.Data) == 0 {
			return fmt.Errorf("%s message is missing sync data", m.Type)
		}
	case MessageEphemeral:
		if m.DocumentID == "" {
			return fmt.Errorf("ephemeral message is missing a document id")
		}

		if m.SessionID == "" {
			return fmt.Errorf("ephemeral message is missing a session id")
		}

		if len(m.Data) == 0 {
			return fmt.Errorf("ephemeral message is missing a payload")
		}
	case MessageDocUnavailable:
		if m.DocumentID == "" {
			return fmt.Errorf("doc-unavailable message is missing a document id")
		}
	default:
		return fmt.Errorf("unknown repo message type %q", m.Type)
	}

	return nil
}

// EncodeMessage encodes the repo message object framed by the WebSocket
// adapter.
func EncodeMessage(message Message) ([]byte, error) {
	if err := message.validate(); err != nil {
		return nil, err
	}

	data, err := marshal(message)
	if err != nil {
		return nil, fmt.Errorf("cannot encode repo message: %w", err)
	}

	return data, nil
}

// DecodeMessage decodes a CBOR repo message and validates it.
func DecodeMessage(data []byte) (Message, error) {
	var message Message
	if err := unmarshal(data, &message); err != nil {
		return Message{}, fmt.Errorf("cannot decode repo message: %w", err)
	}

	if err := message.validate(); err != nil {
		return Message{}, err
	}

	return message, nil
}

// DedupeKey identifies an ephemeral message for gossip de-duplication: a
// receiver discards a (session, count) pair it has already seen, which breaks
// forwarding loops. It is only meaningful for ephemeral messages.
type DedupeKey struct {
	SessionID string
	Count     uint64
}

// DedupeKey returns the de-duplication key for an ephemeral message.
func (m Message) DedupeKey() DedupeKey {
	return DedupeKey{SessionID: m.SessionID, Count: m.Count}
}
