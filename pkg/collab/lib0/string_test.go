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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/collab/lib0"
)

func TestWriteVarString_GoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  []byte
	}{
		{"empty", "", []byte{0x00}},
		{"ascii", "hi", []byte{0x02, 'h', 'i'}},
		{"unicode-multibyte", "é", []byte{0x02, 0xc3, 0xa9}},
		{"emoji", "🚀", []byte{0x04, 0xf0, 0x9f, 0x9a, 0x80}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				enc := lib0.NewEncoder()
				lib0.WriteVarString(enc, tc.value)
				assert.Equal(t, tc.want, enc.Bytes())
			},
		)
	}
}

func TestVarString_RoundTrip(t *testing.T) {
	t.Parallel()

	values := []string{
		"",
		"a",
		"hello world",
		"é",
		"🚀 to the moon",
		strings.Repeat("x", 200),
		strings.Repeat("é", 200),
	}

	for _, v := range values {
		v := v
		enc := lib0.NewEncoder()
		lib0.WriteVarString(enc, v)

		dec := lib0.NewDecoder(enc.Bytes())
		got, err := lib0.ReadVarString(dec)
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, 0, dec.Remaining())
	}
}

func TestReadVarString_TruncatedBody(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder([]byte{0x05, 'h', 'i'}) // claims 5 bytes, only 2 available
	_, err := lib0.ReadVarString(dec)
	require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}

func TestReadVarString_TruncatedLength(t *testing.T) {
	t.Parallel()

	dec := lib0.NewDecoder([]byte{0x80}) // varuint with continuation, no follow-up
	_, err := lib0.ReadVarString(dec)
	require.ErrorIs(t, err, lib0.ErrUnexpectedEOF)
}
