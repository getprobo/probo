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
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ephemeralFixture struct {
	Description string `json:"description"`
	Message     struct {
		Type       string `json:"type"`
		SenderID   string `json:"senderId"`
		TargetID   string `json:"targetId"`
		DocumentID string `json:"documentId"`
		SessionID  string `json:"sessionId"`
		Count      uint64 `json:"count"`
		Data       string `json:"data"`
	} `json:"message"`
	PayloadCBORBase64 string `json:"payloadCborBase64"`
}

// TestEphemeralFixtureCarriesPresence confirms the ephemeral message fields the
// JavaScript client sends decode into our Message, and that the payload it
// carries is exactly a presence payload our presence codec reads.
func TestEphemeralFixtureCarriesPresence(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"ephemeral-update.json",
		"ephemeral-snapshot.json",
		"ephemeral-heartbeat.json",
		"ephemeral-goodbye.json",
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				fixture := readFixture[ephemeralFixture](t, name)

				message := Message{
					Type:       MessageType(fixture.Message.Type),
					SenderID:   fixture.Message.SenderID,
					TargetID:   fixture.Message.TargetID,
					DocumentID: fixture.Message.DocumentID,
					SessionID:  fixture.Message.SessionID,
					Count:      fixture.Message.Count,
					Data:       decodeBase64(t, fixture.Message.Data),
				}

				require.NoError(t, message.validate())
				assert.Equal(t, MessageEphemeral, message.Type)
				assert.Equal(t, DedupeKey{SessionID: "session-a", Count: 1}, message.DedupeKey())

				// The ephemeral payload is a presence message the presence codec reads.
				presence, err := DecodePresence(message.Data)
				require.NoError(t, err)
				assert.Contains(
					t,
					[]PresenceType{PresenceUpdate, PresenceSnapshot, PresenceHeartbeat, PresenceGoodbye},
					presence.Type,
				)
			},
		)
	}
}

// TestMessage_RoundTrip encodes and decodes each document-scoped message type.
func TestMessage_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, message := range []Message{
		{Type: MessageSync, SenderID: "a", TargetID: "b", DocumentID: "doc", Data: []byte{1, 2, 3}},
		{Type: MessageRequest, SenderID: "a", TargetID: "b", DocumentID: "doc", Data: []byte{4, 5}},
		{Type: MessageEphemeral, SenderID: "a", TargetID: "b", DocumentID: "doc", SessionID: "s", Count: 7, Data: []byte{6}},
		{Type: MessageDocUnavailable, SenderID: "a", TargetID: "b", DocumentID: "doc"},
	} {
		t.Run(
			string(message.Type),
			func(t *testing.T) {
				t.Parallel()

				encoded, err := EncodeMessage(message)
				require.NoError(t, err)

				decoded, err := DecodeMessage(encoded)
				require.NoError(t, err)
				assert.Equal(t, message.Type, decoded.Type)
				assert.Equal(t, message.DocumentID, decoded.DocumentID)
				assert.Equal(t, message.Data, decoded.Data)
				assert.Equal(t, message.SessionID, decoded.SessionID)
				assert.Equal(t, message.Count, decoded.Count)
			},
		)
	}
}

func TestMessage_WireShapeIncludesOnlyTypeFields(t *testing.T) {
	t.Parallel()

	ephemeral, err := EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   "a",
			TargetID:   "b",
			DocumentID: "doc",
			SessionID:  "session",
			Count:      0,
			Data:       []byte{1},
		},
	)
	require.NoError(t, err)

	var ephemeralFields map[string]cbor.RawMessage
	require.NoError(t, unmarshal(ephemeral, &ephemeralFields))
	assert.Contains(t, ephemeralFields, "count")
	assert.Contains(t, ephemeralFields, "sessionId")

	sync, err := EncodeMessage(
		Message{
			Type:       MessageSync,
			SenderID:   "a",
			TargetID:   "b",
			DocumentID: "doc",
			Data:       []byte{1},
		},
	)
	require.NoError(t, err)

	var syncFields map[string]cbor.RawMessage
	require.NoError(t, unmarshal(sync, &syncFields))
	assert.NotContains(t, syncFields, "count")
	assert.NotContains(t, syncFields, "sessionId")
}

func TestDecodeMessage_RequiresEphemeralCount(t *testing.T) {
	t.Parallel()

	frame, err := marshal(
		map[string]any{
			"type":       MessageEphemeral,
			"senderId":   "a",
			"targetId":   "b",
			"documentId": "doc",
			"sessionId":  "session",
			"data":       []byte{1},
		},
	)
	require.NoError(t, err)

	_, err = DecodeMessage(frame)
	assert.Error(t, err)
}

func TestMessage_AppliesPayloadLimitByType(t *testing.T) {
	t.Parallel()

	large := make([]byte, maxApplicationBytes+1)

	sync, err := EncodeMessage(
		Message{
			Type:       MessageSync,
			SenderID:   "a",
			TargetID:   "b",
			DocumentID: "doc",
			Data:       large,
		},
	)
	require.NoError(t, err)

	decoded, err := DecodeMessage(sync)
	require.NoError(t, err)
	assert.Len(t, decoded.Data, len(large))

	_, err = EncodeMessage(
		Message{
			Type:       MessageEphemeral,
			SenderID:   "a",
			TargetID:   "b",
			DocumentID: "doc",
			SessionID:  "session",
			Data:       large,
		},
	)
	assert.Error(t, err)
}

// TestMessage_Validation rejects messages missing type-required fields.
func TestMessage_Validation(t *testing.T) {
	t.Parallel()

	cases := []Message{
		{Type: MessageSync, SenderID: "a", TargetID: "b", DocumentID: "doc"},      // no data
		{Type: MessageSync, SenderID: "a", TargetID: "b", Data: []byte{1}},        // no doc
		{Type: MessageEphemeral, SenderID: "a", DocumentID: "d", Data: []byte{1}}, // no session
		{Type: MessageType("bogus"), SenderID: "a"},
		{Type: MessageSync},
	}

	for _, message := range cases {
		_, err := EncodeMessage(message)
		assert.Error(t, err)
	}
}

// TestDecodeMessage_RejectsDuplicateKeys confirms the strict decoder refuses a
// duplicate map key rather than silently taking one.
func TestDecodeMessage_RejectsDuplicateKeys(t *testing.T) {
	t.Parallel()

	// { "type":"doc-unavailable", "type":"sync", "senderId":"a", ... } built by
	// hand as CBOR with a duplicated key.
	duplicate := buildDuplicateKeyCBOR(t)

	_, err := DecodeMessage(duplicate)
	assert.Error(t, err)
}

func buildDuplicateKeyCBOR(t *testing.T) []byte {
	t.Helper()

	// Encode two separate maps and splice a duplicate "type" key by hand would be
	// fragile; instead assert the decoder mode rejects duplicates via a crafted
	// map[string] with the same key twice is impossible in Go, so encode a raw
	// CBOR map literal.
	//
	// CBOR: map(4) { "type":"sync", "type":"sync", "senderId":"a", "documentId":"d" }
	// 0xa4 (map,4)
	//   "type"(0x64 74797065) "sync"(0x64 73796e63)
	//   "type"(0x64 74797065) "sync"(0x64 73796e63)
	//   "senderId"(0x68 ...) "a"(0x61 61)
	//   "documentId"(0x6a ...) "d"(0x61 64)
	return []byte{
		0xa4,
		0x64, 't', 'y', 'p', 'e', 0x64, 's', 'y', 'n', 'c',
		0x64, 't', 'y', 'p', 'e', 0x64, 's', 'y', 'n', 'c',
		0x68, 's', 'e', 'n', 'd', 'e', 'r', 'I', 'd', 0x61, 'a',
		0x6a, 'd', 'o', 'c', 'u', 'm', 'e', 'n', 't', 'I', 'd', 0x61, 'd',
	}
}

// Ensure the fixture files are valid JSON we can parse (guards regeneration).
func TestFixturesAreParseable(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"presence-roundtrip.json"} {
		raw := readFixture[map[string]json.RawMessage](t, name)
		assert.Contains(t, raw, "cborBase64")
	}
}
