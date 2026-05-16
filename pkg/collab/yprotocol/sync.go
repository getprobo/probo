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

// WriteSyncStep1 writes a complete sync.Step1 frame
// (MessageSync + SyncStep1 + state vector) into enc.
func WriteSyncStep1(enc *lib0.Encoder, stateVector []byte) {
	lib0.WriteVarUint(enc, uint64(MessageSync))
	lib0.WriteVarUint(enc, uint64(SyncStep1))
	lib0.WriteVarUint8Array(enc, stateVector)
}

// WriteSyncStep2 writes a complete sync.Step2 frame
// (MessageSync + SyncStep2 + update) into enc.
func WriteSyncStep2(enc *lib0.Encoder, update []byte) {
	lib0.WriteVarUint(enc, uint64(MessageSync))
	lib0.WriteVarUint(enc, uint64(SyncStep2))
	lib0.WriteVarUint8Array(enc, update)
}

// WriteSyncUpdate writes a complete sync.Update frame
// (MessageSync + SyncUpdate + update) into enc.
func WriteSyncUpdate(enc *lib0.Encoder, update []byte) {
	lib0.WriteVarUint(enc, uint64(MessageSync))
	lib0.WriteVarUint(enc, uint64(SyncUpdate))
	lib0.WriteVarUint8Array(enc, update)
}

// EncodeSyncStep1 returns a fresh frame containing a sync.Step1 message.
func EncodeSyncStep1(stateVector []byte) []byte {
	enc := lib0.NewEncoder()
	WriteSyncStep1(enc, stateVector)
	return enc.Bytes()
}

// EncodeSyncStep2 returns a fresh frame containing a sync.Step2 message.
func EncodeSyncStep2(update []byte) []byte {
	enc := lib0.NewEncoder()
	WriteSyncStep2(enc, update)
	return enc.Bytes()
}

// EncodeSyncUpdate returns a fresh frame containing a sync.Update message.
func EncodeSyncUpdate(update []byte) []byte {
	enc := lib0.NewEncoder()
	WriteSyncUpdate(enc, update)
	return enc.Bytes()
}
