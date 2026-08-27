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
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

// storedOperationOrder reports the operation-set order a document chunk was
// written in, which is the order the encoder has to be given to reproduce it.
func storedOperationOrder(t *testing.T, data []byte) []opset.OpID {
	t.Helper()

	r := &reader{data: data}
	budget := &decodeBudget{}

	chunk, err := decodeChunk(r, budget, true)
	require.NoError(t, err)
	require.Equal(t, opset.ChunkDocument, chunk.kind)

	content := &reader{data: chunk.content}

	actors, err := decodeActorArray(content, true, budget)
	require.NoError(t, err)

	_, err = decodeHashArray(content, true, budget)
	require.NoError(t, err)

	changeMetadata, err := parseColumnMetadata(content, true, budget)
	require.NoError(t, err)

	operationMetadata, err := parseColumnMetadata(content, true, budget)
	require.NoError(t, err)

	_, err = readColumns(content, changeMetadata, budget)
	require.NoError(t, err)

	operationColumns, err := readColumns(content, operationMetadata, budget)
	require.NoError(t, err)

	operations, _, err := decodeOperations(operationColumns, actors, false, nil, &decodeBudget{})
	require.NoError(t, err)

	order := make([]opset.OpID, len(operations))
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

	encoded, err := EncodeDocument(document, storedOperationOrder(t, data), true)
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

	encoded, err := EncodeDocument(document, storedOperationOrder(t, data), true)
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

func TestEncodeDocument_RejectsDependencyCycle(t *testing.T) {
	t.Parallel()

	first := opset.ChangeHash{1}
	second := opset.ChangeHash{2}
	document := &opset.Document{
		Changes: []opset.Change{
			{Hash: &first, Sequence: 1, Dependencies: []opset.ChangeHash{second}},
			{Hash: &second, Sequence: 1, Dependencies: []opset.ChangeHash{first}},
		},
	}

	_, err := EncodeDocument(document, nil, false)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cycle")
}

func TestEncodeDocument_RejectsValuesOutsideUint32Domain(t *testing.T) {
	t.Parallel()

	hash := opset.ChangeHash{1}
	document := &opset.Document{
		Changes: []opset.Change{
			{
				Hash:     &hash,
				Sequence: uint64(math.MaxUint32) + 1,
			},
		},
	}

	_, err := EncodeDocument(document, nil, false)
	require.Error(t, err)
	assert.ErrorContains(t, err, "uint32")
}

func TestEncodeDocument_NonByteExtraReconstructsChangeHash(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	property := "value"
	identifier := opset.OpID{Actor: actor, Counter: 1}
	change := opset.Change{
		Actor:      actor,
		Sequence:   1,
		StartOp:    1,
		MaxOp:      1,
		Extra:      &opset.Scalar{Type: opset.ScalarString, String: "extra"},
		ExtraBytes: []byte("extra"),
		Operations: []opset.Operation{
			{
				ID:     identifier,
				Object: opset.RootObject(),
				Key:    opset.Key{Property: &property},
				Action: opset.ActionSet,
				Value:  &opset.Scalar{Type: opset.ScalarString, String: "value"},
			},
		},
	}

	_, err = EncodeChange(&change)
	require.NoError(t, err)

	document := &opset.Document{
		Heads:   []opset.ChangeHash{*change.Hash},
		Changes: []opset.Change{change},
	}
	encoded, err := EncodeDocument(document, []opset.OpID{identifier}, false)
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	require.Len(t, decoded.Changes, 1)
	assert.Equal(t, change.Hash, decoded.Changes[0].Hash)
	assert.Equal(t, []byte("extra"), decoded.Changes[0].ExtraBytes)
	require.NotNil(t, decoded.Changes[0].Extra)
	assert.Equal(t, opset.ScalarString, decoded.Changes[0].Extra.Type)
}

func TestDocumentChangeOrder_HandlesDeepGraphIteratively(t *testing.T) {
	t.Parallel()

	const count = 100_000

	hashes := make([]opset.ChangeHash, count)
	changes := make([]opset.Change, count)
	for i := range changes {
		binary.LittleEndian.PutUint64(hashes[i][:], uint64(i+1))
		changes[i].Hash = &hashes[i]
		if i > 0 {
			changes[i].Dependencies = []opset.ChangeHash{hashes[i-1]}
		}
	}

	ordered, err := documentChangeOrder(&opset.Document{Changes: changes})
	require.NoError(t, err)
	require.Len(t, ordered, count)
	assert.Equal(t, hashes[0], *ordered[0].Hash)
	assert.Equal(t, hashes[count-1], *ordered[count-1].Hash)
}
