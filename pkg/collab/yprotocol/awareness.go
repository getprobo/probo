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

package yprotocol

import "go.probo.inc/probo/pkg/collab/lib0"

// WriteAwareness writes a complete awareness frame
// (MessageAwareness + opaque update) into enc.
func WriteAwareness(enc *lib0.Encoder, update []byte) {
	lib0.WriteVarUint(enc, uint64(MessageAwareness))
	lib0.WriteVarUint8Array(enc, update)
}

// EncodeAwareness returns a fresh frame containing an awareness message.
func EncodeAwareness(update []byte) []byte {
	enc := lib0.NewEncoder()
	WriteAwareness(enc, update)
	return enc.Bytes()
}
