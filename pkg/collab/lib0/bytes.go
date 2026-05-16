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

// WriteVarUint8Array writes a byte slice prefixed by its length encoded as
// a varuint. This matches lib0's `writeVarUint8Array`.
func WriteVarUint8Array(e *Encoder, b []byte) {
	WriteVarUint(e, uint64(len(b)))
	_, _ = e.Write(b)
}

// ReadVarUint8Array reads a byte slice written by WriteVarUint8Array.
// The returned slice is a copy: callers may mutate it without affecting
// the decoder's underlying buffer.
func ReadVarUint8Array(d *Decoder) ([]byte, error) {
	n, err := ReadVarUint(d)
	if err != nil {
		return nil, err
	}
	if n > uint64(d.Remaining()) {
		return nil, ErrUnexpectedEOF
	}

	return d.ReadBytes(int(n))
}
