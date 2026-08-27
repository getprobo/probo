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
	"time"

	"github.com/fxamacker/cbor/v2"
)

// PresenceMarker is the single envelope key automerge-repo's Presence wraps
// every message in before it is CBOR-encoded into an ephemeral payload.
const PresenceMarker = "__presence"

// Presence heartbeat and expiry defaults, matching the upstream constants.
const (
	DefaultHeartbeatInterval = 15 * time.Second
	DefaultPeerTTL           = 3 * DefaultHeartbeatInterval
)

// PresenceType is the discriminator of a presence message.
type PresenceType string

const (
	// PresenceUpdate carries one channel's new value.
	PresenceUpdate PresenceType = "update"
	// PresenceSnapshot carries the full multi-channel state, sent on start and
	// to newly seen peers.
	PresenceSnapshot PresenceType = "snapshot"
	// PresenceHeartbeat signals liveness when nothing has changed.
	PresenceHeartbeat PresenceType = "heartbeat"
	// PresenceGoodbye tells peers to forget the sender immediately.
	PresenceGoodbye PresenceType = "goodbye"
)

// PresenceMessage is one decoded presence event. Value (for an update) and
// State (for a snapshot) are application-defined and kept as raw CBOR so a
// caller decodes them into its own type; use Set/Get helpers or the cbor
// package directly. They are nil for heartbeat and goodbye.
type PresenceMessage struct {
	Type    PresenceType
	Channel string
	Value   cbor.RawMessage
	State   cbor.RawMessage
}

// presenceEnvelope is the CBOR shape on the wire: a single-key map whose key is
// the presence marker.
type presenceEnvelope struct {
	Presence *presenceBody `cbor:"__presence"`
}

type presenceBody struct {
	Type    string          `cbor:"type"`
	Channel string          `cbor:"channel,omitempty"`
	Value   cbor.RawMessage `cbor:"value,omitempty"`
	State   cbor.RawMessage `cbor:"state,omitempty"`
}

// EncodePresence encodes a presence message into the CBOR bytes carried as an
// ephemeral message's data. It validates that the fields present match the type.
func EncodePresence(message PresenceMessage) ([]byte, error) {
	if err := message.validate(); err != nil {
		return nil, err
	}

	body := &presenceBody{
		Type:    string(message.Type),
		Channel: message.Channel,
		Value:   message.Value,
		State:   message.State,
	}

	data, err := marshal(presenceEnvelope{Presence: body})
	if err != nil {
		return nil, fmt.Errorf("cannot encode presence message: %w", err)
	}

	if err := validateApplicationSize(data); err != nil {
		return nil, fmt.Errorf("invalid presence message: %w", err)
	}

	return data, nil
}

// DecodePresence decodes the CBOR bytes of an ephemeral message's data into a
// presence message, rejecting a payload that is not a presence envelope or whose
// fields do not match its type.
func DecodePresence(data []byte) (PresenceMessage, error) {
	if err := validateApplicationSize(data); err != nil {
		return PresenceMessage{}, fmt.Errorf("invalid presence message: %w", err)
	}

	var envelope presenceEnvelope
	if err := unmarshal(data, &envelope); err != nil {
		return PresenceMessage{}, fmt.Errorf("cannot decode presence envelope: %w", err)
	}

	if envelope.Presence == nil {
		return PresenceMessage{}, fmt.Errorf("payload is missing the %q presence marker", PresenceMarker)
	}

	message := PresenceMessage{
		Type:    PresenceType(envelope.Presence.Type),
		Channel: envelope.Presence.Channel,
		Value:   envelope.Presence.Value,
		State:   envelope.Presence.State,
	}

	if err := message.validate(); err != nil {
		return PresenceMessage{}, err
	}

	return message, nil
}

// validate enforces that only the fields meaningful for a type are set, so a
// malformed cross-type payload (an update with no channel, a heartbeat carrying
// state) is rejected on both encode and decode.
func (m PresenceMessage) validate() error {
	switch m.Type {
	case PresenceUpdate:
		if m.Channel == "" {
			return fmt.Errorf("presence update requires a channel")
		}

		if m.Value == nil {
			return fmt.Errorf("presence update requires a value")
		}

		if m.State != nil {
			return fmt.Errorf("presence update must not carry snapshot state")
		}
	case PresenceSnapshot:
		if m.Channel != "" || m.Value != nil {
			return fmt.Errorf("presence snapshot must not carry a channel or value")
		}

		if m.State == nil {
			return fmt.Errorf("presence snapshot requires state")
		}
	case PresenceHeartbeat, PresenceGoodbye:
		if m.Channel != "" || m.Value != nil || m.State != nil {
			return fmt.Errorf("presence %s must not carry a channel, value, or state", m.Type)
		}
	default:
		return fmt.Errorf("unknown presence type %q", m.Type)
	}

	return nil
}

// MarshalPresenceValue encodes an application value into the raw CBOR carried by
// an update's Value or a snapshot's State, using the shared deterministic mode.
func MarshalPresenceValue(value any) (cbor.RawMessage, error) {
	data, err := marshal(value)
	if err != nil {
		return nil, fmt.Errorf("cannot encode presence value: %w", err)
	}

	if err := validateApplicationSize(data); err != nil {
		return nil, fmt.Errorf("invalid presence value: %w", err)
	}

	return data, nil
}

// UnmarshalValue decodes an update's Value into the given destination.
func (m PresenceMessage) UnmarshalValue(destination any) error {
	if m.Value == nil {
		return fmt.Errorf("presence message has no value")
	}

	return unmarshal(m.Value, destination)
}

// UnmarshalState decodes a snapshot's State into the given destination.
func (m PresenceMessage) UnmarshalState(destination any) error {
	if m.State == nil {
		return fmt.Errorf("presence message has no state")
	}

	return unmarshal(m.State, destination)
}
