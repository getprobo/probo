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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/collab/lib0"
)

func TestNewDecoder_StartsAtZero(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder([]byte{0x01, 0x02, 0x03})
	assert.Equal(t, 0, dec.Pos())
	assert.Equal(t, 3, dec.Remaining())
	assert.True(t, dec.HasMore())
}

func TestDecoder_EmptyBuffer(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder(nil)
	assert.Equal(t, 0, dec.Pos())
	assert.Equal(t, 0, dec.Remaining())
	assert.False(t, dec.HasMore())
}

func TestDecoder_ReadByte(t *testing.T) {
	t.Parallel()

	t.Run(
		"advances position",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder([]byte{0x01, 0x02})

			got, err := dec.ReadByte()
			require.NoError(t, err)
			assert.Equal(t, byte(0x01), got)
			assert.Equal(t, 1, dec.Pos())
			assert.Equal(t, 1, dec.Remaining())
			assert.True(t, dec.HasMore())

			got, err = dec.ReadByte()
			require.NoError(t, err)
			assert.Equal(t, byte(0x02), got)
			assert.Equal(t, 2, dec.Pos())
			assert.Equal(t, 0, dec.Remaining())
			assert.False(t, dec.HasMore())
		},
	)

	t.Run(
		"returns EOF when exhausted",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder(nil)
			_, err := dec.ReadByte()
			require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
		},
	)
}

func TestDecoder_ReadBytes(t *testing.T) {
	t.Parallel()

	t.Run(
		"reads requested length",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder([]byte{0x01, 0x02, 0x03, 0x04})

			got, err := dec.ReadBytes(3)
			require.NoError(t, err)
			assert.Equal(t, []byte{0x01, 0x02, 0x03}, got)
			assert.Equal(t, 3, dec.Pos())
			assert.Equal(t, 1, dec.Remaining())
		},
	)

	t.Run(
		"returns a copy",
		func(t *testing.T) {
			t.Parallel()

			src := []byte{0xde, 0xad, 0xbe, 0xef}
			dec := lib0.NewDecoder(src)

			got, err := dec.ReadBytes(4)
			require.NoError(t, err)

			got[0] = 0x00
			assert.Equal(t, byte(0xde), src[0], "mutating the returned slice must not affect the decoder buffer")
		},
	)

	t.Run(
		"zero length",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder([]byte{0x01, 0x02})
			got, err := dec.ReadBytes(0)
			require.NoError(t, err)
			assert.Empty(t, got)
			assert.Equal(t, 0, dec.Pos())
		},
	)

	t.Run(
		"negative length is rejected",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder([]byte{0x01, 0x02})
			_, err := dec.ReadBytes(-1)
			require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
			assert.Equal(t, 0, dec.Pos(), "position must not advance on error")
		},
	)

	t.Run(
		"over-read is rejected",
		func(t *testing.T) {
			t.Parallel()

			dec := lib0.NewDecoder([]byte{0x01, 0x02})
			_, err := dec.ReadBytes(3)
			require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
			assert.Equal(t, 0, dec.Pos(), "position must not advance on error")
		},
	)
}
