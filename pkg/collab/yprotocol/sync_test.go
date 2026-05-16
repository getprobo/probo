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

package yprotocol_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/collab/lib0"
	"go.probo.inc/probo/pkg/collab/yprotocol"
)

func TestEncodeSyncStep1_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		stateVector []byte
		want        []byte
	}{
		{"empty-state-vector", []byte{}, []byte{0x00, 0x00, 0x00}},
		{"nil-state-vector", nil, []byte{0x00, 0x00, 0x00}},
		{"two-byte-state-vector", []byte{0xab, 0xcd}, []byte{0x00, 0x00, 0x02, 0xab, 0xcd}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, yprotocol.EncodeSyncStep1(tc.stateVector))
			},
		)
	}
}

func TestEncodeSyncStep2_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		update []byte
		want   []byte
	}{
		{"empty-update", []byte{}, []byte{0x00, 0x01, 0x00}},
		{"deadbeef-update", []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x00, 0x01, 0x04, 0xde, 0xad, 0xbe, 0xef}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, yprotocol.EncodeSyncStep2(tc.update))
			},
		)
	}
}

func TestEncodeSyncUpdate_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		update []byte
		want   []byte
	}{
		{"empty-update", []byte{}, []byte{0x00, 0x02, 0x00}},
		{"deadbeef-update", []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x00, 0x02, 0x04, 0xde, 0xad, 0xbe, 0xef}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, yprotocol.EncodeSyncUpdate(tc.update))
			},
		)
	}
}

func TestReadMessage_Sync(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		frame       []byte
		wantSubtype byte
		wantPayload []byte
	}{
		{"step1-empty", []byte{0x00, 0x00, 0x00}, yprotocol.SyncStep1, []byte{}},
		{"step1-payload", []byte{0x00, 0x00, 0x02, 0xab, 0xcd}, yprotocol.SyncStep1, []byte{0xab, 0xcd}},
		{"step2-payload", []byte{0x00, 0x01, 0x04, 0xde, 0xad, 0xbe, 0xef}, yprotocol.SyncStep2, []byte{0xde, 0xad, 0xbe, 0xef}},
		{"update-payload", []byte{0x00, 0x02, 0x04, 0xde, 0xad, 0xbe, 0xef}, yprotocol.SyncUpdate, []byte{0xde, 0xad, 0xbe, 0xef}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				msg, err := yprotocol.ReadMessage(tc.frame)
				require.NoError(t, err)
				assert.Equal(t, yprotocol.MessageSync, msg.Type)
				assert.Equal(t, tc.wantSubtype, msg.Subtype)
				assert.Equal(t, tc.wantPayload, msg.Payload)
				assert.NotNil(t, msg.Payload, "Payload should be non-nil even when empty")
			},
		)
	}
}

func TestReadMessage_RejectsEmptyFrame(t *testing.T) {
	t.Parallel()

	_, err := yprotocol.ReadMessage(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)

	_, err = yprotocol.ReadMessage([]byte{})
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadMessage_RejectsSyncMissingSubtype(t *testing.T) {
	t.Parallel()

	frame := []byte{0x00}
	_, err := yprotocol.ReadMessage(frame)
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadMessage_RejectsTruncatedSyncPayloadLength(t *testing.T) {
	t.Parallel()

	// MessageSync, SyncStep1, but no payload length varuint at all.
	frame := []byte{0x00, 0x00}
	_, err := yprotocol.ReadMessage(frame)
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestSyncRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		encode      func([]byte) []byte
		wantSubtype byte
		payload     []byte
	}{
		{"step1-empty", yprotocol.EncodeSyncStep1, yprotocol.SyncStep1, []byte{}},
		{"step1-large", yprotocol.EncodeSyncStep1, yprotocol.SyncStep1, makeBytes(300)},
		{"step2-empty", yprotocol.EncodeSyncStep2, yprotocol.SyncStep2, []byte{}},
		{"step2-large", yprotocol.EncodeSyncStep2, yprotocol.SyncStep2, makeBytes(300)},
		{"update-empty", yprotocol.EncodeSyncUpdate, yprotocol.SyncUpdate, []byte{}},
		{"update-large", yprotocol.EncodeSyncUpdate, yprotocol.SyncUpdate, makeBytes(300)},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				frame := tc.encode(tc.payload)
				msg, err := yprotocol.ReadMessage(frame)
				require.NoError(t, err)
				assert.Equal(t, yprotocol.MessageSync, msg.Type)
				assert.Equal(t, tc.wantSubtype, msg.Subtype)
				assert.Equal(t, tc.payload, msg.Payload)
			},
		)
	}
}

func makeBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

func TestReadMessage_RejectsTrailingBytes(t *testing.T) {
	t.Parallel()

	// Valid SyncStep1 frame followed by an extra byte.
	frame := []byte{0x00, 0x00, 0x00, 0xff}
	_, err := yprotocol.ReadMessage(frame)
	require.ErrorIs(t, err, yprotocol.ErrTrailingBytes)
}

func TestReadMessage_RejectsUnknownMessageType(t *testing.T) {
	t.Parallel()

	frame := []byte{0x05, 0x00, 0x00}
	_, err := yprotocol.ReadMessage(frame)
	require.ErrorIs(t, err, yprotocol.ErrUnknownMessageType)
}

func TestReadMessage_RejectsUnknownSyncSubtype(t *testing.T) {
	t.Parallel()

	frame := []byte{0x00, 0x09, 0x00}
	_, err := yprotocol.ReadMessage(frame)
	require.ErrorIs(t, err, yprotocol.ErrUnknownSyncSubtype)
}

func TestReadMessage_RejectsTruncatedSyncFrame(t *testing.T) {
	t.Parallel()

	// MessageSync, SyncStep1, then claims a 4-byte payload but provides 0.
	frame := []byte{0x00, 0x00, 0x04}
	_, err := yprotocol.ReadMessage(frame)
	require.Error(t, err)
}
