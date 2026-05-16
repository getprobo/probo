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

package lib0_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/collab/lib0"
)

func TestWriteVarUint_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value uint64
		want  []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"max-single-byte", 127, []byte{0x7f}},
		{"min-two-bytes", 128, []byte{0x80, 0x01}},
		{"max-two-bytes", 16383, []byte{0xff, 0x7f}},
		{"min-three-bytes", 16384, []byte{0x80, 0x80, 0x01}},
		{"max-three-bytes", 2097151, []byte{0xff, 0xff, 0x7f}},
		{"min-four-bytes", 2097152, []byte{0x80, 0x80, 0x80, 0x01}},
		{"32-bit-value", 1 << 31, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
		{"max-uint64", math.MaxUint64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				enc := lib0.NewEncoder()
				lib0.WriteVarUint(enc, tc.value)
				assert.Equal(t, tc.want, enc.Bytes())
			},
		)
	}
}

func TestReadVarUint_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input []byte
		want  uint64
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x01}, 1},
		{"max-single-byte", []byte{0x7f}, 127},
		{"min-two-bytes", []byte{0x80, 0x01}, 128},
		{"max-two-bytes", []byte{0xff, 0x7f}, 16383},
		{"max-uint64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, math.MaxUint64},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				dec := lib0.NewDecoder(tc.input)
				got, err := lib0.ReadVarUint(dec)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, 0, dec.Remaining())
			},
		)
	}
}

func TestVarUint_RoundTrip(t *testing.T) {
	t.Parallel()

	values := []uint64{
		0, 1, 2, 63, 64, 127, 128, 255, 256, 16383, 16384,
		1 << 20, 1 << 32, 1 << 53, math.MaxInt64, math.MaxUint64,
	}

	for _, v := range values {
		v := v
		enc := lib0.NewEncoder()
		lib0.WriteVarUint(enc, v)

		dec := lib0.NewDecoder(enc.Bytes())
		got, err := lib0.ReadVarUint(dec)
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, 0, dec.Remaining())
	}
}

func TestReadVarUint_TruncatedInput(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder([]byte{0x80}) // continuation bit set, no follow-up byte
	_, err := lib0.ReadVarUint(dec)
	require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadVarUint_Overflow(t *testing.T) {
	t.Parallel()

	// 11 bytes all 0xff is well past the uint64 boundary.
	overflow := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	dec := lib0.NewDecoder(overflow)
	_, err := lib0.ReadVarUint(dec)
	require.ErrorIs(t, err, lib0.ErrVarUintOverflow)
}

func TestWriteVarInt_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value int64
		want  []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"positive-one", 1, []byte{0x01}},
		{"negative-one", -1, []byte{0x41}},
		{"max-single-byte-positive", 63, []byte{0x3f}},
		{"max-single-byte-negative", -63, []byte{0x7f}},
		{"min-two-bytes-positive", 64, []byte{0x80, 0x01}},
		{"min-two-bytes-negative", -64, []byte{0xc0, 0x01}},
		{"larger-positive", 8191, []byte{0xbf, 0x7f}},
		{"larger-negative", -8191, []byte{0xff, 0x7f}},
		{
			"max-int64",
			math.MaxInt64,
			[]byte{0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		},
		{
			"min-int64",
			math.MinInt64,
			[]byte{0xc0, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x02},
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				enc := lib0.NewEncoder()
				lib0.WriteVarInt(enc, tc.value)
				assert.Equal(t, tc.want, enc.Bytes())
			},
		)
	}
}

func TestVarInt_RoundTrip(t *testing.T) {
	t.Parallel()

	values := []int64{
		0, 1, -1, 63, -63, 64, -64, 127, -127, 128, -128,
		8191, -8191, 1 << 20, -(1 << 20), 1 << 32, -(1 << 32),
		1<<53 - 1, -(1<<53 - 1), math.MaxInt64, math.MinInt64,
	}

	for _, v := range values {
		v := v
		enc := lib0.NewEncoder()
		lib0.WriteVarInt(enc, v)

		dec := lib0.NewDecoder(enc.Bytes())
		got, err := lib0.ReadVarInt(dec)
		require.NoError(t, err, "value=%d", v)
		assert.Equal(t, v, got)
		assert.Equal(t, 0, dec.Remaining())
	}
}

func TestReadVarInt_TruncatedInput(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder(nil)
	_, err := lib0.ReadVarInt(dec)
	require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)

	dec = lib0.NewDecoder([]byte{0x80}) // continuation set, no follow-up
	_, err = lib0.ReadVarInt(dec)
	require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadVarInt_Overflow(t *testing.T) {
	t.Parallel()

	// In ReadVarInt the first byte carries 6 data bits and each
	// continuation byte adds 7 more. The 10th byte (9th iteration of
	// the loop) is therefore the boundary where mult reaches 2^62 and
	// the various overflow guards become reachable.

	cases := []struct {
		name  string
		input []byte
	}{
		{
			// (b & 0x7f) * mult overflows uint64 on iter 9
			// because mult = 2^62 and b & 0x7f = 4 yields 2^64.
			name:  "multiply-overflow",
			input: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x04},
		},
		{
			// mult << 7 would jump from 2^62 to 2^69; an extra
			// continuation byte after byte 10 forces that shift.
			name:  "mult-shift-overflow",
			input: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00},
		},
		{
			// Magnitude exceeds 2^63 for a negative value: sign
			// bit on byte 0 and 0x03 on byte 10 yields num =
			// 3 * 2^62 = 2^63 + 2^62.
			name:  "negative-magnitude-exceeds-int64",
			input: []byte{0xc0, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x03},
		},
		{
			// Positive value exceeding MaxInt64: 0x02 on byte 10
			// puts the value at exactly 2^63.
			name:  "positive-exceeds-max-int64",
			input: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x02},
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				dec := lib0.NewDecoder(tc.input)
				_, err := lib0.ReadVarInt(dec)
				require.ErrorIs(t, err, lib0.ErrVarIntOverflow)
			},
		)
	}
}
