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

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestEncodedBodySize_MatchesEncodeChange(t *testing.T) {
	t.Parallel()

	document, _ := benchmarkSnapshot(64)
	change := document.Changes[0]
	change.Hash = nil

	assertBodySize(t, &change)

	other, err := opset.NewActorID([]byte{2})
	require.NoError(t, err)

	property := "title"
	mark := "em"
	expand := true
	change.Message = "edit"
	change.Operations[2].Predecessors = []opset.OpID{{
		Actor:   other,
		Counter: 4,
	}}
	change.Operations[3].Key = opset.Key{Property: &property}
	change.Operations[3].Insert = false
	change.Operations[3].Value = &opset.Scalar{
		Type:   opset.ScalarString,
		String: "heading",
	}
	change.Operations[4].Action = opset.ActionMark
	change.Operations[4].MarkName = &mark
	change.Operations[4].MarkExpand = &expand
	// Implicit operation IDs must still line up after the edits above.
	for i := range change.Operations {
		change.Operations[i].ID = opset.OpID{
			Actor:   change.Actor,
			Counter: change.StartOp + uint64(i),
		}
	}

	assertBodySize(t, &change)
}

func TestRLESize_MatchesEncoder(t *testing.T) {
	t.Parallel()

	patterns := [][]optional[uint64]{
		nil,
		{some(uint64(1))},
		{some(uint64(1)), some(uint64(1)), some(uint64(1))},
		{some(uint64(1)), some(uint64(2)), some(uint64(3))},
		{some(uint64(1)), some(uint64(2)), some(uint64(2))},
		{{}, {}, some(uint64(4))},
		{some(uint64(4)), {}, {}},
		{some(uint64(1)), some(uint64(1)), {}, some(uint64(2)), some(uint64(3)), some(uint64(3))},
	}

	for _, pattern := range patterns {
		encoded := encodeRLE(pattern, appendULEB)
		sized, err := rleSize(
			len(pattern),
			func(i int) optional[uint64] {
				return pattern[i]
			},
			ulebSize,
		)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(encoded)), sized)
	}
}

func TestReconstructionScratch_BelowFlatReservation(t *testing.T) {
	t.Parallel()

	document, _ := benchmarkSnapshot(500)
	change := document.Changes[0]
	change.Hash = nil

	scratch, err := reconstructionScratch(&change)
	require.NoError(t, err)

	flat := uint64(len(change.Operations)+1) * 1024
	assert.Less(t, scratch, flat)

	budget := &decodeBudget{
		used: maxDecodedMemoryBytes - scratch - 64,
	}
	require.Less(t, uint64(maxDecodedMemoryBytes)-budget.used, flat)

	changes := []opset.Change{change}
	start := budget.used
	err = reconstructSnapshotChanges(changes, budget)
	require.NoError(t, err)
	assert.Less(t, budget.used-start, scratch/2)
	require.NotNil(t, changes[0].Hash)
	assert.NotEmpty(t, changes[0].Raw)
}

func assertBodySize(t *testing.T, change *opset.Change) {
	t.Helper()

	size, err := encodedBodySize(change)
	require.NoError(t, err)

	raw, err := EncodeChange(change)
	require.NoError(t, err)

	reader := &reader{data: raw}
	_, err = reader.bytes(4 + 4 + 1)
	require.NoError(t, err)

	length, err := reader.uleb()
	require.NoError(t, err)
	assert.Equal(t, size, length)
	assert.Equal(t, int(length), reader.remaining())
}
