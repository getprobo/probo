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

// WriteVarString writes a UTF-8 string prefixed by its byte length encoded
// as a varuint. Note that the prefix is the byte length, not the rune count
// (matching lib0's `writeVarString`, which feeds the string through
// TextEncoder before length-prefixing).
func WriteVarString(e *Encoder, s string) {
	WriteVarUint(e, uint64(len(s)))
	_, _ = e.Write([]byte(s))
}

// ReadVarString reads a UTF-8 string written by WriteVarString.
func ReadVarString(d *Decoder) (string, error) {
	n, err := ReadVarUint(d)
	if err != nil {
		return "", err
	}
	if n > uint64(d.Remaining()) {
		return "", ErrUnexpectedEOF
	}

	out := string(d.buf[d.pos : d.pos+int(n)])
	d.pos += int(n)
	return out, nil
}
