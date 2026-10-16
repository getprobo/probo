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
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/fxamacker/cbor/v2"
)

const (
	selectionPresenceKind     = "selection"
	departurePresenceKind     = "departure"
	maxSelectionDisplayRunes  = 128
	maxSelectionAvatarRunes   = 2048
	selectionDisplayNameField = "displayName"
	selectionIdentityIDField  = "identityID"
	selectionAvatarURLField   = "avatarURL"
	selectionKindField        = "kind"
)

// EncodeDepartureEphemeral builds an ephemeral frame announcing that senderID
// left the document. A fresh sessionID keeps the frame newer than any caret
// that sender already gossiped.
func EncodeDepartureEphemeral(
	senderID string,
	targetID string,
	documentID string,
	sessionID string,
) ([]byte, error) {
	payload, err := marshal(map[string]string{
		selectionKindField: departurePresenceKind,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot encode departure ephemeral: %w", err)
	}

	frame, err := EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   senderID,
			TargetID:   targetID,
			DocumentID: documentID,
			SessionID:  sessionID,
			Count:      1,
			Data:       payload,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot encode departure frame: %w", err)
	}

	return frame, nil
}

// IsDepartureEphemeral reports whether message is a departure announcement.
func IsDepartureEphemeral(message Message) bool {
	if message.Type != MessageEphemeral {
		return false
	}

	kind, ok := ephemeralPayloadKind(message.Data)
	if !ok {
		return false
	}

	return kind == departurePresenceKind
}

func ephemeralPayloadKind(data []byte) (string, bool) {
	var payload map[string]cbor.RawMessage
	if err := unmarshal(data, &payload); err != nil {
		return "", false
	}

	kindRaw, ok := payload[selectionKindField]
	if !ok {
		return "", false
	}

	var kind string
	if err := unmarshal(kindRaw, &kind); err != nil {
		return "", false
	}

	return kind, true
}

// StampSelectionIdentity overwrites the authenticated person's name, identity
// id, and avatar URL on a selection ephemeral frame. Non-selection frames and
// empty display names are returned unchanged so gossip of other payloads stays
// intact. An empty avatarURL removes any client-supplied picture so it cannot
// be spoofed.
func StampSelectionIdentity(frame []byte, displayName, identityID, avatarURL string) []byte {
	displayName = sanitizeDisplayName(displayName)
	if displayName == "" {
		return frame
	}

	stamped, err := stampSelectionIdentity(
		frame,
		displayName,
		identityID,
		sanitizeAvatarURL(avatarURL),
	)
	if err != nil {
		return frame
	}

	return stamped
}

func stampSelectionIdentity(
	frame []byte,
	displayName string,
	identityID string,
	avatarURL string,
) ([]byte, error) {
	message, err := DecodeMessage(frame)
	if err != nil {
		return nil, err
	}

	if message.Type != MessageEphemeral {
		return frame, nil
	}

	var payload map[string]cbor.RawMessage
	if err := unmarshal(message.Data, &payload); err != nil {
		return nil, fmt.Errorf("cannot decode selection ephemeral: %w", err)
	}

	kindRaw, ok := payload[selectionKindField]
	if !ok {
		return frame, nil
	}

	var kind string
	if err := unmarshal(kindRaw, &kind); err != nil {
		return nil, fmt.Errorf("cannot decode selection kind: %w", err)
	}

	if kind != selectionPresenceKind {
		return frame, nil
	}

	nameRaw, err := marshal(displayName)
	if err != nil {
		return nil, fmt.Errorf("cannot encode display name: %w", err)
	}

	payload[selectionDisplayNameField] = nameRaw

	if identityID != "" {
		identityRaw, err := marshal(identityID)
		if err != nil {
			return nil, fmt.Errorf("cannot encode identity id: %w", err)
		}

		payload[selectionIdentityIDField] = identityRaw
	}

	if avatarURL != "" {
		avatarRaw, err := marshal(avatarURL)
		if err != nil {
			return nil, fmt.Errorf("cannot encode avatar url: %w", err)
		}

		payload[selectionAvatarURLField] = avatarRaw
	} else {
		delete(payload, selectionAvatarURLField)
	}

	data, err := marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot encode selection ephemeral: %w", err)
	}

	if err := validateApplicationSize(data); err != nil {
		return nil, err
	}

	message.Data = data

	stamped, err := EncodeMessage(message)
	if err != nil {
		return nil, fmt.Errorf("cannot encode stamped selection frame: %w", err)
	}

	return stamped, nil
}

func sanitizeAvatarURL(avatarURL string) string {
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return ""
	}

	if !utf8.ValidString(avatarURL) {
		return ""
	}

	if utf8.RuneCountInString(avatarURL) > maxSelectionAvatarRunes {
		return ""
	}

	for _, character := range avatarURL {
		if unicode.IsControl(character) {
			return ""
		}
	}

	return avatarURL
}

func sanitizeDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(name))

	runes := 0
	for _, character := range name {
		if unicode.IsControl(character) {
			if character == '\n' || character == '\r' || character == '\t' {
				if builder.Len() > 0 && builder.String()[builder.Len()-1] != ' ' {
					builder.WriteByte(' ')
				}
			}

			continue
		}

		builder.WriteRune(character)
		runes++
		if runes >= maxSelectionDisplayRunes {
			break
		}
	}

	return strings.TrimSpace(builder.String())
}

// SelectionDisplayName reports the display name stamped onto a selection
// ephemeral payload, if any. It is used by tests and debugging tooling.
func SelectionDisplayName(data []byte) (string, bool) {
	return selectionStringField(data, selectionDisplayNameField)
}

// SelectionAvatarURL reports the avatar URL stamped onto a selection ephemeral
// payload, if any. It is used by tests and debugging tooling.
func SelectionAvatarURL(data []byte) (string, bool) {
	return selectionStringField(data, selectionAvatarURLField)
}

func selectionStringField(data []byte, field string) (string, bool) {
	var payload map[string]cbor.RawMessage
	if err := unmarshal(data, &payload); err != nil {
		return "", false
	}

	raw, ok := payload[field]
	if !ok {
		return "", false
	}

	var value string
	if err := unmarshal(raw, &value); err != nil {
		return "", false
	}

	if !utf8.ValidString(value) {
		return "", false
	}

	return value, value != ""
}
