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

func TestNewEncoder_StartsEmpty(t *testing.T) {
	t.Parallel()

	enc := lib0.NewEncoder()
	assert.Empty(t, enc.Bytes())
	assert.Equal(t, 0, enc.Len())
}

func TestEncoder_WriteByte(t *testing.T) {
	t.Parallel()

	enc := lib0.NewEncoder()

	require.NoError(t, enc.WriteByte(0xde))
	require.NoError(t, enc.WriteByte(0xad))

	assert.Equal(t, []byte{0xde, 0xad}, enc.Bytes())
	assert.Equal(t, 2, enc.Len())
}

func TestEncoder_Write(t *testing.T) {
	t.Parallel()

	enc := lib0.NewEncoder()

	n, err := enc.Write([]byte{0x01, 0x02, 0x03})
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = enc.Write(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	assert.Equal(t, []byte{0x01, 0x02, 0x03}, enc.Bytes())
	assert.Equal(t, 3, enc.Len())
}

func TestEncoder_Reset(t *testing.T) {
	t.Parallel()

	t.Run(
		"after writes",
		func(t *testing.T) {
			t.Parallel()

			enc := lib0.NewEncoder()
			_, err := enc.Write([]byte{0x01, 0x02, 0x03})
			require.NoError(t, err)

			enc.Reset()
			assert.Empty(t, enc.Bytes())
			assert.Equal(t, 0, enc.Len())

			require.NoError(t, enc.WriteByte(0xff))
			assert.Equal(t, []byte{0xff}, enc.Bytes())
		},
	)

	t.Run(
		"on fresh encoder",
		func(t *testing.T) {
			t.Parallel()

			enc := lib0.NewEncoder()
			enc.Reset()
			assert.Empty(t, enc.Bytes())
			assert.Equal(t, 0, enc.Len())
		},
	)
}
