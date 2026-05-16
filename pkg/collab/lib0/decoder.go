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

package lib0

import "errors"

var (
	ErrUnexpectedEOF   = errors.New("lib0: unexpected end of input")
	ErrVarUintOverflow = errors.New("lib0: varuint overflow")
	ErrVarIntOverflow  = errors.New("lib0: varint overflow")
)

type (
	Decoder struct {
		buf []byte
		pos int
	}
)

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{buf: buf}
}

func (d *Decoder) Pos() int {
	return d.pos
}

func (d *Decoder) Remaining() int {
	return len(d.buf) - d.pos
}

func (d *Decoder) HasMore() bool {
	return d.pos < len(d.buf)
}

func (d *Decoder) ReadByte() (byte, error) {
	if d.pos >= len(d.buf) {
		return 0, ErrUnexpectedEOF
	}
	b := d.buf[d.pos]
	d.pos++
	return b, nil
}

// ReadBytes copies and returns the next n bytes from the underlying buffer.
// A copy is returned so callers cannot accidentally mutate the decoder's
// backing storage.
func (d *Decoder) ReadBytes(n int) ([]byte, error) {
	if n < 0 || n > d.Remaining() {
		return nil, ErrUnexpectedEOF
	}

	out := make([]byte, n)
	copy(out, d.buf[d.pos:d.pos+n])
	d.pos += n
	return out, nil
}
