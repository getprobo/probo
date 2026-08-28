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
	"bytes"
	"compress/flate"
	"encoding/base64"
	"math"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/testsupport/reference"
)

const (
	officialChangeFixture   = "hW9Kg5nDjoUBzgEAEAECAwQFBgcICQoLDA0ODxABAYDiz6oGF29mZmljaWFsIHNjYWxhciBmaXh0dXJlAAoBBAIHEQYTBxU4NAJCCFYQVxtwAgALBAAACwILfgwNAAx/AAACAAt8AAx0AHUDbmlsAm5vA3llcwR1aW50A2ludAVmbG9hdAR0ZXh0BWJ5dGVzBHdoZW4FY291bnQEbGlzdAAECwQKAX8CAgQCAXYAAQITFIUBVjdpGAMAAhYqeQAAAAAAAPg/aGVsbG8A/wf70JX/vDEJYWIPAA=="
	officialDocumentFixture = "hW9Kg3tNcOoAogIBEAECAwQFBgcICQoLDA0ODxABmcOOhfOq6K9fyRtQMpEkw5nRGiPrg0/hSLI3KA5LqKcHAQIDAhMCIwY1GUACVgIMAQQCBxEGEwcVOCECIw80AkIKVhJXG4ABAn8AfwF/D3+A4s+qBn8Xb2ZmaWNpYWwgc2NhbGFyIGZpeHR1cmV/AH8HAAsEAAALAgt+DA0ADH8AAAIAC3wADHQAdQVieXRlcwVjb3VudAVmbG9hdANpbnQEbGlzdANuaWwCbm8EdGV4dAR1aW50BHdoZW4DeWVzAAQPAHQIAnx/BnYBBX0FegkDAQsEBAF/AgYBAgQCAXw3GIUBFAIAewFWE2kCAgACFgD/BwkAAAAAAAD4P3loZWxsbyr70JX/vDFhYg8AAA=="
)

func fixture(t *testing.T, encoded string) []byte {
	t.Helper()

	data, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)

	return data
}

func scalarOperations(document *opset.Document) map[string]opset.Scalar {
	result := make(map[string]opset.Scalar)

	for _, operation := range document.Changes[0].Operations {
		if operation.Action == opset.ActionSet && operation.Key.Property != nil && operation.Value != nil {
			result[*operation.Key.Property] = *operation.Value
		}
	}

	return result
}

func TestDecode_OfficialDocumentFixture(t *testing.T) {
	t.Parallel()

	document, err := Decode(fixture(t, officialDocumentFixture))
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	require.Len(t, document.Heads, 1)

	change := document.Changes[0]
	assert.Equal(t, uint64(1), change.Sequence)
	assert.Equal(t, "official scalar fixture", change.Message)
	assert.Equal(t, document.Heads[0], *change.Hash)
	assert.Equal(t, "99c38e85f3aae8af5fc91b50329124c399d11a23eb834fe148b237280e4ba8a7", change.Hash.String())

	values := scalarOperations(document)
	require.Len(t, values, 10)
	assert.Equal(t, opset.ScalarNull, values["nil"].Type)
	assert.Equal(t, opset.ScalarFalse, values["no"].Type)
	assert.Equal(t, opset.ScalarTrue, values["yes"].Type)
	assert.Equal(t, uint64(42), values["uint"].Uint)
	assert.Equal(t, int64(-7), values["int"].Int)
	assert.Equal(t, 1.5, values["float"].Float)
	assert.Equal(t, "hello", values["text"].String)
	assert.Equal(t, []byte{0, 255, 7}, values["bytes"].Bytes)
	assert.Equal(t, int64(1_700_000_000_123), values["when"].Int)
	assert.Equal(t, int64(9), values["count"].Int)
}

func TestDecode_OfficialChangeFixture(t *testing.T) {
	t.Parallel()

	document, err := Decode(fixture(t, officialChangeFixture))
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	require.Len(t, document.Heads, 1)

	change := document.Changes[0]
	assert.Equal(t, uint64(1), change.Sequence)
	assert.Equal(t, uint64(1), change.StartOp)
	assert.Equal(t, uint64(15), change.MaxOp)
	assert.Equal(t, document.Heads[0], *change.Hash)
	assert.Equal(t, "99c38e85f3aae8af5fc91b50329124c399d11a23eb834fe148b237280e4ba8a7", change.Hash.String())
}

func TestDecode_OfficialEmptyDocumentFixture(t *testing.T) {
	t.Parallel()

	data := []byte{
		0x85, 0x6f, 0x4a, 0x83,
		0xb8, 0x1a, 0x95, 0x44,
		0x00, 0x04,
		0x00, 0x00, 0x00, 0x00,
	}
	document, err := Decode(data)
	require.NoError(t, err)
	assert.Empty(t, document.Actors)
	assert.Empty(t, document.Heads)
	assert.Empty(t, document.Changes)
}

func TestDecode_CompressedOfficialChangeFixture(t *testing.T) {
	t.Parallel()

	uncompressed := fixture(t, officialChangeFixture)
	r := &reader{data: uncompressed, offset: 9}
	length, err := r.uleb()
	require.NoError(t, err)
	content, err := r.bytes(length)
	require.NoError(t, err)

	var compressed bytes.Buffer

	writer, err := flate.NewWriter(&compressed, flate.BestCompression)
	require.NoError(t, err)
	_, err = writer.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	data := append([]byte(nil), uncompressed[:8]...)
	data = append(data, byte(opset.ChunkCompressedChange))
	data = appendULEB(data, uint64(compressed.Len()))
	data = append(data, compressed.Bytes()...)

	document, err := Decode(data)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	assert.Equal(t, opset.ChunkCompressedChange, document.ChunkTypes[0])
	assert.Equal(t, "99c38e85f3aae8af5fc91b50329124c399d11a23eb834fe148b237280e4ba8a7", document.Heads[0].String())

	withCorruptTail := append(append([]byte(nil), data...), magic[:]...)
	incremental, consumed, err := DecodeIncremental(withCorruptTail)
	require.NoError(t, err)
	assert.Equal(t, len(data), consumed)
	require.Len(t, incremental.Changes, 1)
	assert.Equal(t, opset.ChunkCompressedChange, incremental.ChunkTypes[0])
}

func TestDecode_OfficialStorageCorpus(t *testing.T) {
	t.Parallel()

	valid := []string{
		"counter_value_is_ok.automerge",
		"two_change_chunks.automerge",
		"two_change_chunks_compressed.automerge",
		"two_change_chunks_out_of_order.automerge",
	}
	for _, name := range valid {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				data, err := base64.StdEncoding.DecodeString(
					officialStorageFixtures[name],
				)
				require.NoError(t, err)
				_, err = Decode(data)
				require.NoError(t, err)
			},
		)
	}

	invalid := []string{
		"counter_value_has_incorrect_meta.automerge",
		"counter_value_is_overlong.automerge",
	}
	for _, name := range invalid {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				data, err := base64.StdEncoding.DecodeString(
					officialStorageFixtures[name],
				)
				require.NoError(t, err)
				_, err = Decode(data)
				require.Error(t, err)
			},
		)
	}
}

func TestDecode_OfficialFuzzCrashersDoNotPanic(t *testing.T) {
	t.Parallel()

	for name, encoded := range officialStorageFixtures {
		if !strings.HasPrefix(name, "fuzz-") {
			continue
		}

		name := name
		encoded := encoded

		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				data, err := base64.StdEncoding.DecodeString(encoded)
				require.NoError(t, err)

				_, _ = Decode(data)
			},
		)
	}
}

func TestDecode_Official64BitObjectIDs(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"64bit_obj_id_change.automerge",
		"64bit_obj_id_doc.automerge",
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				data, err := base64.StdEncoding.DecodeString(
					officialStorageFixtures[name],
				)
				require.NoError(t, err)

				document, err := Decode(data)
				if err != nil {
					return
				}

				require.NotEmpty(t, document.Changes)

				found := false

				for _, operation := range document.Changes[0].Operations {
					if operation.ID.Counter == 1<<42 {
						found = true
					}
				}

				assert.True(t, found)
			},
		)
	}
}

func TestDecode_ReferenceBackendDocument(t *testing.T) {
	t.Parallel()

	backend, err := reference.New()
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, backend.Close())
		},
	)

	actor := []byte{1, 3, 3, 7}
	require.NoError(t, backend.SetActor(actor))
	require.NoError(t, backend.PutString(0, "policy", "approved"))
	_, err = backend.Commit("reference fixture", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)
	data, err := backend.Save(true, true)
	require.NoError(t, err)

	document, err := Decode(data)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	assert.Equal(t, "01030307", document.Changes[0].Actor.String())
	assert.Equal(t, "reference fixture", document.Changes[0].Message)
}

func TestDecode_ReferenceBackendConcurrentGraph(t *testing.T) {
	t.Parallel()

	base, err := reference.New()
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, base.Close())
		},
	)
	require.NoError(t, base.SetActor([]byte{1}))
	require.NoError(t, base.PutString(0, "base", "value"))
	_, err = base.Commit("base", time.Unix(1, 0))
	require.NoError(t, err)
	baseData, err := base.Save(true, true)
	require.NoError(t, err)

	left, err := reference.Load(baseData)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, left.Close())
		},
	)
	require.NoError(t, left.SetActor([]byte{2}))
	require.NoError(t, left.PutString(0, "left", "value"))
	_, err = left.Commit("left", time.Unix(2, 0))
	require.NoError(t, err)

	right, err := reference.Load(baseData)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, right.Close())
		},
	)
	require.NoError(t, right.SetActor([]byte{3}))
	require.NoError(t, right.PutString(0, "right", "value"))
	_, err = right.Commit("right", time.Unix(3, 0))
	require.NoError(t, err)
	rightData, err := right.Save(true, true)
	require.NoError(t, err)

	_, err = left.Merge(rightData)
	require.NoError(t, err)
	mergedData, err := left.Save(true, true)
	require.NoError(t, err)

	document, err := Decode(mergedData)
	require.NoError(t, err)
	require.Len(t, document.Changes, 3)
	assert.Len(t, document.Heads, 2)
	assert.Empty(t, document.Changes[0].DependencyIndexes)
	assert.Equal(t, []uint64{0}, document.Changes[1].DependencyIndexes)
	assert.Equal(t, []uint64{0}, document.Changes[2].DependencyIndexes)
}

func TestDecode_RejectsCorruptChecksum(t *testing.T) {
	t.Parallel()

	data := fixture(t, officialDocumentFixture)
	data[len(data)-1] ^= 0xff

	_, err := Decode(data)
	require.Error(t, err)
	assert.ErrorContains(t, err, "checksum mismatch")
}

func TestReaderULEB_RejectsNonCanonicalValue(t *testing.T) {
	t.Parallel()

	r := &reader{data: []byte{0x80, 0x00}}
	_, err := r.uleb()
	require.Error(t, err)
	assert.ErrorContains(t, err, "not minimally encoded")
}

func TestValidateSnapshotGraph_RejectsCycle(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []opset.Change{
		{
			Actor:             actor,
			Sequence:          1,
			MaxOp:             1,
			DependencyIndexes: []uint64{1},
		},
		{
			Actor:             actor,
			Sequence:          2,
			MaxOp:             2,
			DependencyIndexes: []uint64{0},
		},
	}

	err = validateSnapshotGraph(changes, nil, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "dependency cycle")
}

func TestValidateSnapshotGraph_RejectsSequenceGap(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []opset.Change{
		{Actor: actor, Sequence: 1, MaxOp: 1},
		{Actor: actor, Sequence: 3, MaxOp: 2, DependencyIndexes: []uint64{0}},
	}

	err = validateSnapshotGraph(changes, []uint64{1}, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "sequence 3, expected 2")
}

func TestAssignOperations_UsesDependencyClock(t *testing.T) {
	t.Parallel()

	firstActor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)
	secondActor, err := opset.NewActorID([]byte{2})
	require.NoError(t, err)

	changes := []opset.Change{
		{Actor: firstActor, Sequence: 1, MaxOp: 1},
		{
			Actor:             secondActor,
			Sequence:          1,
			MaxOp:             2,
			DependencyIndexes: []uint64{0},
		},
		{
			Actor:             firstActor,
			Sequence:          2,
			MaxOp:             3,
			DependencyIndexes: []uint64{1},
		},
	}
	operations := []opset.Operation{
		{ID: opset.OpID{Actor: firstActor, Counter: 1}},
		{ID: opset.OpID{Actor: secondActor, Counter: 2}},
		{ID: opset.OpID{Actor: firstActor, Counter: 3}},
	}

	err = assignOperations(changes, operations, &decodeBudget{})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), changes[0].StartOp)
	assert.Equal(t, uint64(2), changes[1].StartOp)
	assert.Equal(t, uint64(3), changes[2].StartOp)
	require.Len(t, changes[2].Operations, 1)
	assert.Equal(t, uint64(3), changes[2].Operations[0].ID.Counter)
}

func TestAssignOperations_SingleActorManyChanges(t *testing.T) {
	t.Parallel()

	const count = 10_000

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := make([]opset.Change, count)
	operations := make([]opset.Operation, count)
	for i := range count {
		counter := uint64(i + 1)
		changes[i] = opset.Change{
			Actor:    actor,
			Sequence: counter,
			MaxOp:    counter,
		}
		if i > 0 {
			changes[i].DependencyIndexes = []uint64{uint64(i - 1)}
		}

		operations[i].ID = opset.OpID{Actor: actor, Counter: counter}
	}

	err = assignOperations(changes, operations, &decodeBudget{})
	require.NoError(t, err)

	for i := range changes {
		require.Len(t, changes[i].Operations, 1)
		assert.Equal(t, uint64(i+1), changes[i].Operations[0].ID.Counter)
	}
}

func TestAssignOperations_RejectsOverlappingActorRanges(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []opset.Change{
		{Actor: actor, Sequence: 1, MaxOp: 1},
		{Actor: actor, Sequence: 2, MaxOp: 1},
	}
	operations := []opset.Operation{
		{ID: opset.OpID{Actor: actor, Counter: 1}},
	}

	err = assignOperations(changes, operations, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "overlapping operation ranges")
}

func TestAssignOperations_RejectsOrphanOperation(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []opset.Change{{Actor: actor, Sequence: 1, MaxOp: 1}}
	operations := []opset.Operation{
		{ID: opset.OpID{Actor: actor, Counter: 2}},
	}

	err = assignOperations(changes, operations, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "has no containing change")
}

func TestAssignOperations_DoesNotAllocateFromEncodedRange(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []opset.Change{{Actor: actor, Sequence: 1, MaxOp: math.MaxUint32}}

	err = assignOperations(changes, nil, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "expected")
	assert.Empty(t, changes[0].Operations)
}

func TestDecodeColumns_SharesMemoryBudget(t *testing.T) {
	t.Parallel()

	itemSize := uint64(unsafe.Sizeof(optional[uint64]{}))
	budget := &decodeBudget{used: maxDecodedMemoryBytes - itemSize*2}

	_, err := decodeULEBColumnWithBudget([]byte{1, 1}, budget)
	require.NoError(t, err)

	_, err = decodeULEBColumnWithBudget([]byte{1, 1}, budget)
	require.Error(t, err)
	assert.ErrorContains(t, err, "decoded columns exceed")
}

func TestDecodeOptionalColumns_ChargeAbsentDefaults(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		size   uint64
		decode func(*decodeBudget) error
	}{
		"delta": {
			size: uint64(unsafe.Sizeof(optional[uint64]{})),
			decode: func(budget *decodeBudget) error {
				_, err := decodeOptionalDelta(nil, 1, "value", 1, budget)

				return err
			},
		},
		"signed delta": {
			size: uint64(unsafe.Sizeof(optional[int64]{})),
			decode: func(budget *decodeBudget) error {
				_, err := decodeOptionalSignedDelta(nil, 1, "value", 1, budget)

				return err
			},
		},
		"ULEB": {
			size: uint64(unsafe.Sizeof(optional[uint64]{})),
			decode: func(budget *decodeBudget) error {
				_, err := decodeOptionalULEB(nil, 1, "value", 1, budget)

				return err
			},
		},
		"string": {
			size: uint64(unsafe.Sizeof(optional[string]{})),
			decode: func(budget *decodeBudget) error {
				_, err := decodeOptionalStrings(nil, 1, "value", 1, budget)

				return err
			},
		},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				budget := &decodeBudget{
					used: maxDecodedMemoryBytes - test.size + 1,
				}
				err := test.decode(budget)
				require.Error(t, err)
				assert.ErrorContains(t, err, "decoded columns exceed")
			},
		)
	}
}

func TestDecodeStringColumn_ChargesPayloadCopy(t *testing.T) {
	t.Parallel()

	headerSize := uint64(unsafe.Sizeof(optional[string]{})) * 2
	budget := &decodeBudget{
		used: maxDecodedMemoryBytes - headerSize - 2,
	}

	_, err := decodeStringColumnWithBudget([]byte{1, 3, 'a', 'b', 'c'}, budget)
	require.Error(t, err)
	assert.ErrorContains(t, err, "decoded columns exceed")
}

func TestAssignOperations_HandlesDeepDependencyGraphIteratively(t *testing.T) {
	t.Parallel()

	const count = 100_000

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	changes := make([]opset.Change, count)
	for i := range changes {
		changes[i] = opset.Change{
			Actor:    actor,
			Sequence: uint64(i + 1),
		}
		if i > 0 {
			changes[i].DependencyIndexes = []uint64{uint64(i - 1)}
		}
	}

	err = assignOperations(changes, nil, &decodeBudget{})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), changes[len(changes)-1].StartOp)
}

func TestDecodeDocumentChanges_RejectsSequenceOutsideUint32Domain(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	columns := map[uint32]column{
		1:  {specification: 1, data: encodeRLE([]optional[uint64]{some(uint64(0))}, appendULEB)},
		3:  {specification: 3, data: encodeDelta([]optional[int64]{some(int64(math.MaxUint32) + 1)})},
		19: {specification: 19, data: encodeDelta([]optional[int64]{some(int64(0))})},
		64: {specification: 64, data: encodeRLE([]optional[uint64]{some(uint64(0))}, appendULEB)},
	}

	_, _, err = decodeDocumentChanges(columns, []opset.ActorID{actor}, &decodeBudget{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "sequence")
	assert.ErrorContains(t, err, "uint32")
}

func TestDecode_OwnsRetainedPayloads(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	property := "payload"
	identifier := opset.OpID{Actor: actor, Counter: 1}
	change := opset.Change{
		Actor:      actor,
		Sequence:   1,
		StartOp:    1,
		MaxOp:      1,
		ExtraBytes: []byte{4, 5},
		Operations: []opset.Operation{
			{
				ID:     identifier,
				Object: opset.RootObject(),
				Key:    opset.Key{Property: &property},
				Action: opset.ActionSet,
				Value:  &opset.Scalar{Type: opset.ScalarBytes, Bytes: []byte{1, 2, 3}},
			},
		},
	}

	changeData, err := EncodeChange(&change)
	require.NoError(t, err)

	expectedRaw := append([]byte(nil), changeData...)

	decodedChange, err := Decode(changeData)
	require.NoError(t, err)
	clear(changeData)

	require.Len(t, decodedChange.Changes, 1)
	require.Len(t, decodedChange.Changes[0].Operations, 1)
	require.NotNil(t, decodedChange.Changes[0].Operations[0].Value)
	assert.Equal(t, []byte{1, 2, 3}, decodedChange.Changes[0].Operations[0].Value.Bytes)
	assert.Equal(t, []byte{4, 5}, decodedChange.Changes[0].ExtraBytes)
	assert.Equal(t, expectedRaw, decodedChange.Changes[0].Raw)

	document := &opset.Document{
		Heads:   []opset.ChangeHash{*change.Hash},
		Changes: []opset.Change{change},
		UnknownColumns: []opset.RawColumn{
			{Specification: 200, Data: []byte{6, 7}},
		},
	}
	snapshotData, err := EncodeDocument(document, []opset.OpID{identifier}, false)
	require.NoError(t, err)

	decodedSnapshot, err := Decode(snapshotData)
	require.NoError(t, err)
	clear(snapshotData)

	require.Len(t, decodedSnapshot.Changes, 1)
	require.Len(t, decodedSnapshot.Changes[0].Operations, 1)
	require.NotNil(t, decodedSnapshot.Changes[0].Operations[0].Value)
	assert.Equal(t, []byte{1, 2, 3}, decodedSnapshot.Changes[0].Operations[0].Value.Bytes)
	assert.Equal(t, []byte{4, 5}, decodedSnapshot.Changes[0].ExtraBytes)

	foundUnknown := false

	for _, column := range decodedSnapshot.UnknownColumns {
		if bytes.Equal(column.Data, []byte{6, 7}) {
			foundUnknown = true
		}
	}

	assert.True(t, foundUnknown)
}
