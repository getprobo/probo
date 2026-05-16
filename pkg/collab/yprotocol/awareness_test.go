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

func TestEncodeAwareness_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		update []byte
		want   []byte
	}{
		{"empty-update", []byte{}, []byte{0x01, 0x00}},
		{"single-byte-update", []byte{0x42}, []byte{0x01, 0x01, 0x42}},
		{"deadbeef-update", []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x01, 0x04, 0xde, 0xad, 0xbe, 0xef}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, yprotocol.EncodeAwareness(tc.update))
			},
		)
	}
}

func TestReadMessage_Awareness(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		frame       []byte
		wantPayload []byte
	}{
		{"empty", []byte{0x01, 0x00}, []byte{}},
		{"deadbeef", []byte{0x01, 0x04, 0xde, 0xad, 0xbe, 0xef}, []byte{0xde, 0xad, 0xbe, 0xef}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				msg, err := yprotocol.ReadMessage(tc.frame)
				require.NoError(t, err)
				assert.Equal(t, yprotocol.MessageAwareness, msg.Type)
				assert.Equal(t, tc.wantPayload, msg.Payload)
				assert.NotNil(t, msg.Payload, "Payload should be non-nil even when empty")
				assert.Equal(t, byte(0), msg.Subtype, "Subtype is not meaningful for awareness and should be left zero")
			},
		)
	}
}

func TestReadMessage_AwarenessRejectsTrailing(t *testing.T) {
	t.Parallel()

	frame := []byte{0x01, 0x00, 0xff}
	_, err := yprotocol.ReadMessage(frame)
	require.ErrorIs(t, err, yprotocol.ErrTrailingBytes)
}

func TestReadMessage_AwarenessRejectsMissingPayload(t *testing.T) {
	t.Parallel()

	// MessageAwareness with no payload length varuint at all.
	frame := []byte{0x01}
	_, err := yprotocol.ReadMessage(frame)
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadMessage_AwarenessRejectsTruncatedPayload(t *testing.T) {
	t.Parallel()

	// MessageAwareness, length=4, but only 2 bytes follow.
	frame := []byte{0x01, 0x04, 0xab, 0xcd}
	_, err := yprotocol.ReadMessage(frame)
	require.Error(t, err)
	assert.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestAwarenessRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload []byte
	}{
		{"empty", []byte{}},
		{"single-byte", []byte{0x42}},
		{"large", makeBytes(300)},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				frame := yprotocol.EncodeAwareness(tc.payload)
				msg, err := yprotocol.ReadMessage(frame)
				require.NoError(t, err)
				assert.Equal(t, yprotocol.MessageAwareness, msg.Type)
				assert.Equal(t, tc.payload, msg.Payload)
			},
		)
	}
}
