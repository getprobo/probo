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

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// storedOperationOrder reports the operation-set order a document chunk was
// written in, which is the order the encoder has to be given to reproduce it.
func storedOperationOrder(t *testing.T, data []byte) []OpID {
	t.Helper()

	r := &reader{data: data}

	chunk, err := decodeChunk(r)
	require.NoError(t, err)
	require.Equal(t, ChunkDocument, chunk.kind)

	content := &reader{data: chunk.content}

	actors, err := decodeActorArray(content, true)
	require.NoError(t, err)

	_, err = decodeHashArray(content, true)
	require.NoError(t, err)

	changeMetadata, err := parseColumnMetadata(content, true)
	require.NoError(t, err)

	operationMetadata, err := parseColumnMetadata(content, true)
	require.NoError(t, err)

	_, err = readColumns(content, changeMetadata)
	require.NoError(t, err)

	operationColumns, err := readColumns(content, operationMetadata)
	require.NoError(t, err)

	operations, _, err := decodeOperations(operationColumns, actors, false, nil)
	require.NoError(t, err)

	order := make([]OpID, len(operations))
	for i, operation := range operations {
		order[i] = operation.ID
	}

	return order
}

// TestEncodeDocument_ReproducesOfficialSnapshotBytes is the byte-identity gate
// for snapshot writing. Re-encoding a snapshot the reference implementation
// wrote has to produce that same file, which pins the column layout, the actor
// and head tables, the extra payload and the chunk framing all at once.
func TestEncodeDocument_ReproducesOfficialSnapshotBytes(t *testing.T) {
	t.Parallel()

	data := fixture(t, officialDocumentFixture)

	document, err := Decode(data)
	require.NoError(t, err)

	encoded, err := EncodeDocument(document, storedOperationOrder(t, data))
	require.NoError(t, err)

	assert.Equal(t, data, encoded, "re-encoded snapshot must match the original bytes")
}

// TestEncodeDocument_RoundTripsThroughDecode checks the written snapshot reads
// back as the same history even where byte identity is not the subject.
func TestEncodeDocument_RoundTripsThroughDecode(t *testing.T) {
	t.Parallel()

	data := fixture(t, officialDocumentFixture)

	document, err := Decode(data)
	require.NoError(t, err)

	encoded, err := EncodeDocument(document, storedOperationOrder(t, data))
	require.NoError(t, err)

	reloaded, err := Decode(encoded)
	require.NoError(t, err)

	require.Len(t, reloaded.Changes, len(document.Changes))
	assert.Equal(t, document.Heads, reloaded.Heads)
	assert.Equal(t, document.Actors, reloaded.Actors)

	for i := range document.Changes {
		expected := &document.Changes[i]
		actual := &reloaded.Changes[i]

		assert.Equal(t, expected.Hash, actual.Hash, "change %d hash", i)
		assert.Equal(t, expected.Raw, actual.Raw, "change %d bytes", i)
		assert.Equal(t, expected.Message, actual.Message, "change %d message", i)
		assert.Equal(t, expected.Time, actual.Time, "change %d time", i)
	}
}
