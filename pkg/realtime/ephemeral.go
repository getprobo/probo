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

package realtime

import (
	"encoding/json"
	"fmt"
)

// MaxCollaborationEphemeralBytes bounds the encoded NOTIFY payload for an
// ephemeral gossip event. PostgreSQL caps a NOTIFY payload at 8000 bytes; this
// leaves headroom for the JSON envelope and base64 expansion of the frame.
// Presence and cursor frames are far smaller, so an oversized frame is dropped
// from cross-instance gossip (local fan-out still delivers it) rather than
// risking a failed NOTIFY.
const MaxCollaborationEphemeralBytes = 7000

// ephemeralKind marks a collaboration-changed NOTIFY payload as an ephemeral
// envelope rather than the bare document-version id that signals a persisted
// change. A bare id is not valid JSON, so the two never collide.
const ephemeralKind = "ephemeral"

// CollaborationEphemeral is an opaque automerge-repo gossip frame carried across
// server instances over the collaboration NOTIFY channel, so presence and
// cursors reach peers connected to other instances.
type CollaborationEphemeral struct {
	// VersionID is the document-version GID string the frame belongs to.
	VersionID string
	// InstanceID identifies the server instance that published the frame, so the
	// publisher can ignore its own echo (it already delivered the frame to its
	// local peers directly).
	InstanceID string
	// Frame is the opaque repo message to relay unchanged.
	Frame []byte
}

type ephemeralEnvelope struct {
	Kind       string `json:"k"`
	VersionID  string `json:"v"`
	InstanceID string `json:"i"`
	Frame      []byte `json:"e"`
}

// EncodeCollaborationEphemeral encodes an ephemeral event into a NOTIFY payload.
// It returns an error when the encoded payload would exceed the NOTIFY size
// budget, so the caller can fall back to local-only fan-out.
func EncodeCollaborationEphemeral(event CollaborationEphemeral) (string, error) {
	payload, err := json.Marshal(ephemeralEnvelope{
		Kind:       ephemeralKind,
		VersionID:  event.VersionID,
		InstanceID: event.InstanceID,
		Frame:      event.Frame,
	})
	if err != nil {
		return "", fmt.Errorf("cannot encode collaboration ephemeral: %w", err)
	}

	if len(payload) > MaxCollaborationEphemeralBytes {
		return "", fmt.Errorf(
			"collaboration ephemeral payload is %d bytes, over the %d-byte limit",
			len(payload), MaxCollaborationEphemeralBytes,
		)
	}

	return string(payload), nil
}

// DecodeCollaborationEphemeral decodes a NOTIFY payload as an ephemeral event. It
// reports ok=false for a bare document-version id (the persisted-change signal)
// or any payload that is not a well-formed ephemeral envelope, so callers can
// treat the two payload kinds apart.
func DecodeCollaborationEphemeral(payload string) (CollaborationEphemeral, bool) {
	// A bare document-version id is base64url and never starts with '{', so this
	// keeps the change-signal path free of a JSON decode.
	if len(payload) == 0 || payload[0] != '{' {
		return CollaborationEphemeral{}, false
	}

	var envelope ephemeralEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return CollaborationEphemeral{}, false
	}

	if envelope.Kind != ephemeralKind || envelope.VersionID == "" {
		return CollaborationEphemeral{}, false
	}

	return CollaborationEphemeral{
		VersionID:  envelope.VersionID,
		InstanceID: envelope.InstanceID,
		Frame:      envelope.Frame,
	}, true
}
