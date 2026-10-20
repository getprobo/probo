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

package collaboration

import "testing"

// FuzzDecodePresence checks that decoding arbitrary bytes never panics and only
// ever returns a valid message or an error.
func FuzzDecodePresence(f *testing.F) {
	for _, seed := range [][]byte{
		nil,
		{0xa0},
		{0xff},
		[]byte("not cbor"),
	} {
		f.Add(seed)
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			message, err := DecodePresence(data)
			if err != nil {
				return
			}

			// A returned message must satisfy the type invariants, and re-encoding it
			// must succeed and decode back to the same type.
			if err := message.validate(); err != nil {
				t.Fatalf("decoded presence message is invalid: %v", err)
			}

			encoded, err := EncodePresence(message)
			if err != nil {
				t.Fatalf("cannot re-encode a decoded presence message: %v", err)
			}

			again, err := DecodePresence(encoded)
			if err != nil {
				t.Fatalf("cannot decode a re-encoded presence message: %v", err)
			}

			if again.Type != message.Type {
				t.Fatalf("re-encoded type %q != %q", again.Type, message.Type)
			}
		},
	)
}

// FuzzDecodeMessage checks that decoding arbitrary bytes as a repo message never
// panics.
func FuzzDecodeMessage(f *testing.F) {
	f.Add([]byte{0xa0})
	f.Add([]byte("garbage"))

	f.Fuzz(
		func(t *testing.T, data []byte) {
			message, err := DecodeMessage(data)
			if err != nil {
				return
			}

			if err := message.validate(); err != nil {
				t.Fatalf("decoded repo message is invalid: %v", err)
			}
		},
	)
}
