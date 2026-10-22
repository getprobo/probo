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

package collaboration_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/collaboration"
)

func TestStampSelectionIdentity_ReplacesClientName(t *testing.T) {
	t.Parallel()

	payload, err := collaboration.MarshalPresenceValue(
		map[string]any{
			"kind": "selection",
			"selection": map[string]any{
				"field":  "body",
				"anchor": "cursor-a",
				"head":   "cursor-a",
			},
			"displayName": "spoofed",
			"identityID":  "not-real",
			"avatarURL":   "https://evil.example/pic.png",
		},
	)
	require.NoError(t, err)

	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      1,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	stamped := collaboration.StampSelectionIdentity(
		frame,
		"Ada Lovelace",
		"idnt_ada",
		"https://app.example.com/api/files/v1/public/file_ada",
	)
	require.NotEqual(t, frame, stamped)

	message, err := collaboration.DecodeMessage(stamped)
	require.NoError(t, err)
	assert.Equal(t, collaboration.MessageEphemeral, message.Type)
	assert.Equal(t, "peer-a", message.SenderID)
	assert.Equal(t, "session-1", message.SessionID)
	assert.Equal(t, uint64(1), message.Count)

	name, ok := collaboration.SelectionDisplayName(message.Data)
	require.True(t, ok)
	assert.Equal(t, "Ada Lovelace", name)

	avatarURL, ok := collaboration.SelectionAvatarURL(message.Data)
	require.True(t, ok)
	assert.Equal(t, "https://app.example.com/api/files/v1/public/file_ada", avatarURL)

	var decoded struct {
		Kind        string `cbor:"kind"`
		DisplayName string `cbor:"displayName"`
		IdentityID  string `cbor:"identityID"`
		AvatarURL   string `cbor:"avatarURL"`
		Selection   struct {
			Field  string `cbor:"field"`
			Anchor string `cbor:"anchor"`
			Head   string `cbor:"head"`
		} `cbor:"selection"`
	}
	require.NoError(t, decodeApplication(t, message.Data, &decoded))
	assert.Equal(t, "selection", decoded.Kind)
	assert.Equal(t, "Ada Lovelace", decoded.DisplayName)
	assert.Equal(t, "idnt_ada", decoded.IdentityID)
	assert.Equal(t, "https://app.example.com/api/files/v1/public/file_ada", decoded.AvatarURL)
	assert.Equal(t, "body", decoded.Selection.Field)
	assert.Equal(t, "cursor-a", decoded.Selection.Anchor)
	assert.Equal(t, "cursor-a", decoded.Selection.Head)
}

func TestStampSelectionIdentity_StripsSpoofedAvatarWithoutServerAvatar(t *testing.T) {
	t.Parallel()

	payload, err := collaboration.MarshalPresenceValue(
		map[string]any{
			"kind":      "selection",
			"avatarURL": "https://evil.example/pic.png",
			"selection": map[string]any{
				"field":  "body",
				"anchor": "cursor-a",
				"head":   "cursor-a",
			},
		},
	)
	require.NoError(t, err)

	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      5,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	stamped := collaboration.StampSelectionIdentity(frame, "Ada Lovelace", "idnt_ada", "")
	message, err := collaboration.DecodeMessage(stamped)
	require.NoError(t, err)

	_, ok := collaboration.SelectionAvatarURL(message.Data)
	assert.False(t, ok)

	var decoded map[string]any
	require.NoError(t, decodeApplication(t, message.Data, &decoded))
	_, hasAvatar := decoded["avatarURL"]
	assert.False(t, hasAvatar)
}

func TestStampSelectionIdentity_LeavesNonSelectionUnchanged(t *testing.T) {
	t.Parallel()

	payload, err := collaboration.EncodePresence(
		collaboration.PresenceMessage{Type: collaboration.PresenceHeartbeat},
	)
	require.NoError(t, err)

	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      2,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	assert.Equal(
		t,
		frame,
		collaboration.StampSelectionIdentity(frame, "Ada Lovelace", "idnt_ada", ""),
	)
}

func TestStampSelectionIdentity_EmptyNameIsNoOp(t *testing.T) {
	t.Parallel()

	payload, err := collaboration.MarshalPresenceValue(
		map[string]any{
			"kind":      "selection",
			"selection": map[string]any{"field": "body", "anchor": "a", "head": "a"},
		},
	)
	require.NoError(t, err)

	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      3,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	assert.Equal(
		t,
		frame,
		collaboration.StampSelectionIdentity(frame, "  ", "idnt_ada", "https://example.com/a.png"),
	)
}

func TestStampSelectionIdentity_TruncatesLongNames(t *testing.T) {
	t.Parallel()

	payload, err := collaboration.MarshalPresenceValue(
		map[string]any{
			"kind":      "selection",
			"selection": map[string]any{"field": "body", "anchor": "a", "head": "a"},
		},
	)
	require.NoError(t, err)

	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      4,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	longName := strings.Repeat("あ", 200)
	stamped := collaboration.StampSelectionIdentity(frame, longName+"\nextra", "idnt_ada", "")
	message, err := collaboration.DecodeMessage(stamped)
	require.NoError(t, err)

	name, ok := collaboration.SelectionDisplayName(message.Data)
	require.True(t, ok)
	assert.Equal(t, 128, utf8.RuneCountInString(name))
}

func TestStampSelectionIdentity_DropsFrameThatCannotFitIdentity(t *testing.T) {
	t.Parallel()

	payload := largestSelectionPayload(t)
	frame, err := collaboration.EncodeMessage(
		collaboration.Message{
			Type:       collaboration.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "server",
			DocumentID: "doc-1",
			SessionID:  "session-1",
			Count:      6,
			Data:       payload,
		},
	)
	require.NoError(t, err)

	message, err := collaboration.DecodeMessage(frame)
	require.NoError(t, err)
	name, ok := collaboration.SelectionDisplayName(message.Data)
	require.True(t, ok)
	assert.Equal(t, "spoofed", name)

	stamped := collaboration.StampSelectionIdentity(
		frame,
		strings.Repeat("A", 128),
		strings.Repeat("i", 64),
		"",
	)
	assert.Nil(t, stamped)
	assert.NotEqual(t, frame, stamped)
}

func largestSelectionPayload(t *testing.T) []byte {
	t.Helper()

	low := 0
	high := 2 << 20
	var payload []byte
	for low <= high {
		mid := low + (high-low)/2
		encoded, err := collaboration.MarshalPresenceValue(selectionWithPadding(mid))
		if err != nil {
			high = mid - 1

			continue
		}

		payload = encoded
		low = mid + 1
	}

	require.NotEmpty(t, payload)

	return payload
}

func selectionWithPadding(padding int) map[string]any {
	return map[string]any{
		"kind":        "selection",
		"displayName": "spoofed",
		"identityID":  "attacker",
		"selection": map[string]any{
			"field":  "body",
			"anchor": "cursor-a",
			"head":   "cursor-a",
		},
		"padding": strings.Repeat("x", padding),
	}
}

func decodeApplication(t *testing.T, data []byte, destination any) error {
	t.Helper()

	wrapper := collaboration.PresenceMessage{
		Type:  collaboration.PresenceUpdate,
		Value: data,
	}

	return wrapper.UnmarshalValue(destination)
}
