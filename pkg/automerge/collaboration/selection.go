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
	"bytes"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// TextSelectionChannel is the conventional presence channel a caret or selection
// update is published on, so peers know where to look for remote collaborators.
// It is only a convention; any channel string is valid.
const TextSelectionChannel = "selection"

// TextSelectionValue is a collaborator's caret or selection carried inside a
// presence update, expressed with stable Automerge text cursors rather than
// integer offsets.
//
// An integer offset is invalidated by any concurrent edit before it: insert one
// character at the document start and every downstream offset is off by one, so
// a remote caret drawn from a stale offset drifts onto the wrong character. A
// cursor is an opaque, stable address into the sequence; resolving it against
// the current (or any later) document yields the position of the very character
// it was created for, which is what keeps a remote caret anchored while other
// people type. The bytes are exactly the output of pkg/automerge Text.Cursor and
// are resolved with Text.CursorPosition. This package only transports them, so
// it stays independent of the CRDT engine.
type TextSelectionValue struct {
	// Field is the Automerge map key of the text object the selection addresses
	// (for example "body"), so a consumer resolves the cursors against the right
	// object when a document holds more than one text.
	Field string `cbor:"field"`
	// Anchor is the stable cursor for the fixed end of the selection (where the
	// selection started).
	Anchor []byte `cbor:"anchor"`
	// Head is the stable cursor for the moving end (the caret). When Head equals
	// Anchor the selection is a collapsed caret.
	Head []byte `cbor:"head"`
}

func (v TextSelectionValue) validate() error {
	if v.Field == "" {
		return fmt.Errorf("text selection requires a field")
	}

	if len(v.Anchor) == 0 {
		return fmt.Errorf("text selection requires an anchor cursor")
	}

	if len(v.Head) == 0 {
		return fmt.Errorf("text selection requires a head cursor")
	}

	return nil
}

// Collapsed reports whether the selection is a single caret, that is its anchor
// and head address the same position.
func (v TextSelectionValue) Collapsed() bool {
	return bytes.Equal(v.Anchor, v.Head)
}

// Encode marshals the selection into the raw CBOR carried by a presence update's
// Value, validating it first.
func (v TextSelectionValue) Encode() (cbor.RawMessage, error) {
	if err := v.validate(); err != nil {
		return nil, err
	}

	return MarshalPresenceValue(v)
}

// DecodeTextSelectionValue decodes a presence update's Value into a selection,
// rejecting one missing a field or either cursor.
func DecodeTextSelectionValue(raw cbor.RawMessage) (TextSelectionValue, error) {
	var value TextSelectionValue
	if err := unmarshal(raw, &value); err != nil {
		return TextSelectionValue{}, fmt.Errorf("cannot decode text selection: %w", err)
	}

	if err := value.validate(); err != nil {
		return TextSelectionValue{}, err
	}

	return value, nil
}

// NewTextSelectionPresence builds a presence update message that publishes a
// caret or selection on the given channel. An empty channel defaults to
// TextSelectionChannel.
func NewTextSelectionPresence(channel string, value TextSelectionValue) (PresenceMessage, error) {
	if channel == "" {
		channel = TextSelectionChannel
	}

	raw, err := value.Encode()
	if err != nil {
		return PresenceMessage{}, err
	}

	return PresenceMessage{Type: PresenceUpdate, Channel: channel, Value: raw}, nil
}

// TextSelection decodes this presence message's value as a text selection. It is
// a convenience over UnmarshalValue that also validates the selection shape.
func (m PresenceMessage) TextSelection() (TextSelectionValue, error) {
	if m.Type != PresenceUpdate {
		return TextSelectionValue{}, fmt.Errorf("presence %s does not carry a selection update", m.Type)
	}

	if m.Value == nil {
		return TextSelectionValue{}, fmt.Errorf("presence update has no value")
	}

	return DecodeTextSelectionValue(m.Value)
}
