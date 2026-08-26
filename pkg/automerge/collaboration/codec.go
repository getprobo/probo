// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package collaboration lets the Go server and Go agents speak the
// automerge-repo protocol against Probo's own pure-Go CRDT engine. This file
// defines the shared CBOR modes used for every protocol payload.
//
// The wire format is produced upstream by cbor-x. Our decoder accepts what that
// encoder emits (including its non-canonical 16-bit map-length prefixes) and our
// encoder produces deterministic CBOR that cbor-x decodes without issue; byte
// identity with cbor-x is deliberately not a goal, semantic interoperability is.
// The protocol and its ground-truth fixtures are described in PROTOCOL.md.
package collaboration

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Resource limits guard the decoder against hostile payloads. Presence values
// are application-defined, so the bounds are generous but finite.
const (
	maxNestedLevels   = 64
	maxMapPairs       = 4096
	maxArrayElements  = 65536
	maxDecodedPayload = 1 << 20 // 1 MiB per ephemeral payload
)

var (
	decMode cbor.DecMode
	encMode cbor.EncMode
)

func init() {
	decoder, err := cbor.DecOptions{
		// Reject streaming payloads: every length must be known up front.
		IndefLength: cbor.IndefLengthForbidden,
		// A repeated map key is a malformed or hostile payload, not a merge.
		DupMapKey: cbor.DupMapKeyEnforcedAPF,
		// The presence and repo protocols use no CBOR tags; cbor-x emits none
		// because it is configured with tagUint8Array:false, so any tag is
		// unexpected and rejected.
		TagsMd:           cbor.TagsForbidden,
		MaxNestedLevels:  maxNestedLevels,
		MaxMapPairs:      maxMapPairs,
		MaxArrayElements: maxArrayElements,
	}.DecMode()
	if err != nil {
		panic(fmt.Sprintf("collaboration: invalid CBOR decode options: %v", err))
	}

	// Deterministic core encoding: shortest-form integers and sorted map keys,
	// so our output is stable and reproducible. cbor-x decodes it regardless of
	// key order.
	encoder, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		panic(fmt.Sprintf("collaboration: invalid CBOR encode options: %v", err))
	}

	decMode = decoder
	encMode = encoder
}

// unmarshal decodes CBOR into value using the strict shared decode mode, after
// bounding the payload size.
func unmarshal(data []byte, value any) error {
	if len(data) > maxDecodedPayload {
		return fmt.Errorf(
			"collaboration payload of %d bytes exceeds the %d byte limit",
			len(data),
			maxDecodedPayload,
		)
	}

	return decMode.Unmarshal(data, value)
}

// marshal encodes value as deterministic CBOR using the shared encode mode.
func marshal(value any) ([]byte, error) {
	return encMode.Marshal(value)
}
