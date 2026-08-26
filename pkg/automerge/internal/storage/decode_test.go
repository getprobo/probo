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
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/automerge/internal/reference"
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

func scalarOperations(document *Document) map[string]Scalar {
	result := make(map[string]Scalar)

	for _, operation := range document.Changes[0].Operations {
		if operation.Action == ActionSet && operation.Key.Property != nil && operation.Value != nil {
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
	assert.Equal(t, ScalarNull, values["nil"].Type)
	assert.Equal(t, ScalarFalse, values["no"].Type)
	assert.Equal(t, ScalarTrue, values["yes"].Type)
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
	data = append(data, byte(ChunkCompressedChange))
	data = appendULEB(data, uint64(compressed.Len()))
	data = append(data, compressed.Bytes()...)

	document, err := Decode(data)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	assert.Equal(t, ChunkCompressedChange, document.ChunkTypes[0])
	assert.Equal(t, "99c38e85f3aae8af5fc91b50329124c399d11a23eb834fe148b237280e4ba8a7", document.Heads[0].String())
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

	ctx := context.Background()
	backend, err := reference.New(ctx)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, backend.Close(ctx))
		},
	)

	actor := []byte{1, 3, 3, 7}
	require.NoError(t, backend.SetActor(ctx, actor))
	require.NoError(t, backend.PutString(ctx, 0, "policy", "approved"))
	_, err = backend.Commit(ctx, "reference fixture", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)
	data, err := backend.Save(ctx, true, true)
	require.NoError(t, err)

	document, err := Decode(data)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	assert.Equal(t, "01030307", document.Changes[0].Actor.String())
	assert.Equal(t, "reference fixture", document.Changes[0].Message)
}

func TestDecode_ReferenceBackendConcurrentGraph(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := reference.New(ctx)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, base.Close(ctx))
		},
	)
	require.NoError(t, base.SetActor(ctx, []byte{1}))
	require.NoError(t, base.PutString(ctx, 0, "base", "value"))
	_, err = base.Commit(ctx, "base", time.Unix(1, 0))
	require.NoError(t, err)
	baseData, err := base.Save(ctx, true, true)
	require.NoError(t, err)

	left, err := reference.Load(ctx, baseData)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, left.Close(ctx))
		},
	)
	require.NoError(t, left.SetActor(ctx, []byte{2}))
	require.NoError(t, left.PutString(ctx, 0, "left", "value"))
	_, err = left.Commit(ctx, "left", time.Unix(2, 0))
	require.NoError(t, err)

	right, err := reference.Load(ctx, baseData)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			assert.NoError(t, right.Close(ctx))
		},
	)
	require.NoError(t, right.SetActor(ctx, []byte{3}))
	require.NoError(t, right.PutString(ctx, 0, "right", "value"))
	_, err = right.Commit(ctx, "right", time.Unix(3, 0))
	require.NoError(t, err)
	rightData, err := right.Save(ctx, true, true)
	require.NoError(t, err)

	_, err = left.Merge(ctx, rightData)
	require.NoError(t, err)
	mergedData, err := left.Save(ctx, true, true)
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

	actor, err := NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []Change{
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

	err = validateSnapshotGraph(changes, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "dependency cycle")
}

func TestValidateSnapshotGraph_RejectsSequenceGap(t *testing.T) {
	t.Parallel()

	actor, err := NewActorID([]byte{1})
	require.NoError(t, err)

	changes := []Change{
		{Actor: actor, Sequence: 1, MaxOp: 1},
		{Actor: actor, Sequence: 3, MaxOp: 2, DependencyIndexes: []uint64{0}},
	}

	err = validateSnapshotGraph(changes, []uint64{1})
	require.Error(t, err)
	assert.ErrorContains(t, err, "sequence 3, expected 2")
}
