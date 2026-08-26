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
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type presenceFixture struct {
	Description string          `json:"description"`
	Marker      string          `json:"marker"`
	Envelope    json.RawMessage `json:"envelope"`
	CBORBase64  string          `json:"cborBase64"`
}

func readFixture[T any](t *testing.T, name string) T {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	var fixture T
	require.NoError(t, json.Unmarshal(data, &fixture))

	return fixture
}

func decodeBase64(t *testing.T, value string) []byte {
	t.Helper()

	data, err := base64.StdEncoding.DecodeString(value)
	require.NoError(t, err)

	return data
}

// TestDecodePresence_MatchesJavaScriptFixtures decodes the exact CBOR the
// JavaScript client produces for every presence message type.
func TestDecodePresence_MatchesJavaScriptFixtures(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		fixture string
		expect  func(t *testing.T, message PresenceMessage)
	}{
		{
			fixture: "presence-update.json",
			expect: func(t *testing.T, message PresenceMessage) {
				assert.Equal(t, PresenceUpdate, message.Type)
				assert.Equal(t, "cursor", message.Channel)

				var value map[string]string
				require.NoError(t, message.UnmarshalValue(&value))
				assert.Equal(t, map[string]string{"anchor": "a", "head": "b"}, value)
			},
		},
		{
			fixture: "presence-snapshot.json",
			expect: func(t *testing.T, message PresenceMessage) {
				assert.Equal(t, PresenceSnapshot, message.Type)

				var state map[string]map[string]string
				require.NoError(t, message.UnmarshalState(&state))
				assert.Equal(
					t,
					map[string]map[string]string{
						"cursor": {"anchor": "a", "head": "b"},
					},
					state,
				)
			},
		},
		{
			fixture: "presence-heartbeat.json",
			expect: func(t *testing.T, message PresenceMessage) {
				assert.Equal(t, PresenceHeartbeat, message.Type)
				assert.Nil(t, message.Value)
				assert.Nil(t, message.State)
			},
		},
		{
			fixture: "presence-goodbye.json",
			expect: func(t *testing.T, message PresenceMessage) {
				assert.Equal(t, PresenceGoodbye, message.Type)
			},
		},
	} {
		t.Run(
			testCase.fixture,
			func(t *testing.T) {
				t.Parallel()

				fixture := readFixture[presenceFixture](t, testCase.fixture)
				assert.Equal(t, PresenceMarker, fixture.Marker)

				message, err := DecodePresence(decodeBase64(t, fixture.CBORBase64))
				require.NoError(t, err)

				testCase.expect(t, message)
			},
		)
	}
}

// TestPresence_RoundTripsThroughJavaScriptBytes proves our re-encoding of a
// decoded JavaScript payload decodes back to the same message, so a Go peer and
// a JS peer agree on meaning even though the byte encodings need not be identical.
func TestPresence_RoundTripsThroughJavaScriptBytes(t *testing.T) {
	t.Parallel()

	fixture := readFixture[presenceFixture](t, "presence-update.json")
	original := decodeBase64(t, fixture.CBORBase64)

	message, err := DecodePresence(original)
	require.NoError(t, err)

	reEncoded, err := EncodePresence(message)
	require.NoError(t, err)

	roundTripped, err := DecodePresence(reEncoded)
	require.NoError(t, err)
	assert.Equal(t, message.Type, roundTripped.Type)
	assert.Equal(t, message.Channel, roundTripped.Channel)

	var first, second map[string]string
	require.NoError(t, message.UnmarshalValue(&first))
	require.NoError(t, roundTripped.UnmarshalValue(&second))
	assert.Equal(t, first, second)
}

// TestEncodePresence_BuildsUpdate constructs an update from scratch and confirms
// it decodes to the same value, the path a Go agent takes when broadcasting.
func TestEncodePresence_BuildsUpdate(t *testing.T) {
	t.Parallel()

	value, err := MarshalPresenceValue(map[string]string{"anchor": "x", "head": "y"})
	require.NoError(t, err)

	data, err := EncodePresence(
		PresenceMessage{
			Type:    PresenceUpdate,
			Channel: "cursor",
			Value:   value,
		},
	)
	require.NoError(t, err)

	message, err := DecodePresence(data)
	require.NoError(t, err)
	assert.Equal(t, PresenceUpdate, message.Type)

	var decoded map[string]string
	require.NoError(t, message.UnmarshalValue(&decoded))
	assert.Equal(t, map[string]string{"anchor": "x", "head": "y"}, decoded)
}

// TestEncodePresence_RejectsCrossTypeFields guards the type/field invariants.
func TestEncodePresence_RejectsCrossTypeFields(t *testing.T) {
	t.Parallel()

	_, err := EncodePresence(PresenceMessage{Type: PresenceUpdate})
	assert.Error(t, err, "update without a channel must be rejected")

	state, err := MarshalPresenceValue(map[string]int{"n": 1})
	require.NoError(t, err)

	_, err = EncodePresence(PresenceMessage{Type: PresenceHeartbeat, State: state})
	assert.Error(t, err, "heartbeat carrying state must be rejected")

	_, err = EncodePresence(PresenceMessage{Type: PresenceType("bogus")})
	assert.Error(t, err, "unknown type must be rejected")
}

// TestDecodePresence_RejectsNonPresencePayload ensures a well-formed CBOR map
// that is not a presence envelope is refused rather than silently accepted.
func TestDecodePresence_RejectsNonPresencePayload(t *testing.T) {
	t.Parallel()

	other, err := marshal(map[string]string{"hello": "world"})
	require.NoError(t, err)

	_, err = DecodePresence(other)
	assert.Error(t, err)
}
