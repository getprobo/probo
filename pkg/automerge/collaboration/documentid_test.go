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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// realRepoDocumentID is a genuine @automerge/automerge-repo document id (the one
// the interop client uses). Decoding it verifies our base58check implementation
// against real upstream output rather than against itself: a wrong checksum or
// alphabet would fail here.
const realRepoDocumentID = "34YWzjYt5gPJpq5RfXAkPfPcUj1r"

// TestDocumentID_DecodesRealRepoID proves the base58check codec matches the
// upstream format: a real repo id decodes to exactly 16 bytes and re-encodes to
// the identical string.
func TestDocumentID_DecodesRealRepoID(t *testing.T) {
	t.Parallel()

	id, err := DecodeDocumentID(realRepoDocumentID)
	require.NoError(t, err)

	assert.Equal(t, realRepoDocumentID, EncodeDocumentID(id),
		"a real repo id must round-trip byte-identically")
	assert.True(t, ValidDocumentID(realRepoDocumentID))
}

// TestDocumentID_RoundTrip encodes and decodes arbitrary 16-byte ids, including
// ones with leading zero bytes (which base58 encodes specially).
func TestDocumentID_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := map[string][DocumentIDByteLength]byte{
		"zero":          {},
		"leading zeros": {0, 0, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0},
		"all ones":      {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		"max":           {255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		"mixed":         {0xde, 0xad, 0xbe, 0xef, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb},
	}

	for name, id := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoded := EncodeDocumentID(id)
			assert.True(t, ValidDocumentID(encoded))

			decoded, err := DecodeDocumentID(encoded)
			require.NoError(t, err)
			assert.Equal(t, id, decoded)
		})
	}
}

// TestDocumentID_RejectsCorruption rejects a bad checksum, a bad character, and
// the wrong decoded length.
func TestDocumentID_RejectsCorruption(t *testing.T) {
	t.Parallel()

	// Flip the last character of a valid id to break its checksum.
	corrupted := realRepoDocumentID[:len(realRepoDocumentID)-1]
	if realRepoDocumentID[len(realRepoDocumentID)-1] == 'r' {
		corrupted += "s"
	} else {
		corrupted += "r"
	}

	_, err := DecodeDocumentID(corrupted)
	assert.Error(t, err, "a flipped character must fail the checksum")

	_, err = DecodeDocumentID("0OIl") // characters outside the base58 alphabet
	assert.Error(t, err)

	assert.False(t, ValidDocumentID(""))
}

// TestDeriveDocumentID_StableAndValid derives ids from seeds (a version GID
// stands in) and checks they are deterministic, valid, and seed-specific.
func TestDeriveDocumentID_StableAndValid(t *testing.T) {
	t.Parallel()

	const seed = "document_version_2Abc123"

	first := DeriveDocumentID(seed)
	second := DeriveDocumentID(seed)

	assert.Equal(t, first, second, "derivation must be deterministic")
	assert.True(t, ValidDocumentID(first))
	assert.NotEqual(t, first, DeriveDocumentID(seed+"x"),
		"different seeds must derive different ids")
}

// TestAutomergeURL_RoundTrip wraps and unwraps the automerge: scheme and
// rejects malformed URLs.
func TestAutomergeURL_RoundTrip(t *testing.T) {
	t.Parallel()

	url := DeriveAutomergeURL("document_version_9")
	assert.True(t, len(url) > len(AutomergeURLPrefix))

	documentID, err := ParseAutomergeURL(url)
	require.NoError(t, err)
	assert.Equal(t, AutomergeURL(documentID), url)
	assert.True(t, ValidDocumentID(documentID))

	_, err = ParseAutomergeURL(realRepoDocumentID) // no scheme
	assert.Error(t, err)

	_, err = ParseAutomergeURL(AutomergeURLPrefix + "not-a-valid-id!!")
	assert.Error(t, err)
}
