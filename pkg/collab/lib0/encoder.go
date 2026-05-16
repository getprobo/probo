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

// Package lib0 implements the binary encoding primitives used by Y.js's
// `lib0` and `y-protocols` packages: little-endian 7-bit-continuation
// variable-length unsigned integers, lib0-flavoured signed variable-length
// integers, length-prefixed UTF-8 strings, and length-prefixed byte arrays.
package lib0

type (
	Encoder struct {
		buf []byte
	}
)

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Bytes() []byte {
	return e.buf
}

func (e *Encoder) Len() int {
	return len(e.buf)
}

func (e *Encoder) Reset() {
	e.buf = e.buf[:0]
}

// WriteByte appends a single byte to the encoder. It always returns nil
// (the signature matches io.ByteWriter so the encoder can be passed to
// stdlib helpers expecting that interface).
func (e *Encoder) WriteByte(b byte) error {
	e.buf = append(e.buf, b)
	return nil
}

// Write appends p to the encoder. It always returns len(p), nil.
func (e *Encoder) Write(p []byte) (int, error) {
	e.buf = append(e.buf, p...)
	return len(p), nil
}
