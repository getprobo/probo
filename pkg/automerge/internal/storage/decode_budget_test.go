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
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestDecode_ReleasesColumnScratch(t *testing.T) {
	t.Parallel()

	// Past the size that failed in "cannot reserve snapshot change 0" when the
	// decoder kept a one-kilobyte reservation per operation.
	const size = 300_000

	document, operations := benchmarkSnapshot(size)
	for index := range document.Changes {
		_, err := EncodeChange(&document.Changes[index])
		require.NoError(t, err)
	}

	document.Heads = []opset.ChangeHash{*document.Changes[0].Hash}
	document.OperationOrder = make([]opset.OpID, len(operations))
	for index := range operations {
		document.OperationOrder[index] = operations[index].ID
	}

	data, err := EncodePreparedDocument(document, operations, true)
	require.NoError(t, err)

	decoded, err := Decode(data)
	require.NoError(t, err)
	require.Len(t, decoded.OperationOrder, size+1)
}

func TestGrowCharged_ChargesShortRunCapacity(t *testing.T) {
	t.Parallel()

	var values []optional[uint64]
	budget := &decodeBudget{}
	elemSize := int(unsafe.Sizeof(optional[uint64]{}))

	for range 40 {
		bound, err := sliceCapacityBound(cap(values), len(values), 1, elemSize)
		require.NoError(t, err)

		values, err = growCharged(values, 1, budget)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, bound, cap(values))

		values = append(values, optional[uint64]{value: 1, valid: true})
	}

	assert.Greater(t, cap(values), len(values))
	assert.Equal(t, uint64(cap(values))*uint64(elemSize), budget.used)
}

func TestGrowCharged_ChargesLongRunNearItsLength(t *testing.T) {
	t.Parallel()

	const count = 10_000

	elemSize := int(unsafe.Sizeof(optional[uint64]{}))
	bound, err := sliceCapacityBound(0, 0, count, elemSize)
	require.NoError(t, err)

	budget := &decodeBudget{}
	values, err := growCharged([]optional[uint64](nil), count, budget)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, bound, cap(values))

	for range count {
		values = append(values, optional[uint64]{value: 1, valid: true})
	}

	itemSize := uint64(elemSize)
	assert.Equal(t, uint64(cap(values))*itemSize, budget.used)
	assert.Less(t, budget.used, uint64(count)*itemSize*2)
}

func TestDecodeRLE_ShortRunsChargeCapacity(t *testing.T) {
	t.Parallel()

	var data []byte
	for range 40 {
		data = appendLEB(data, 1)
		data = appendULEB(data, 1)
	}

	budget := &decodeBudget{}
	values, err := decodeULEBColumnWithBudget(data, budget)
	require.NoError(t, err)
	require.Greater(t, cap(values), len(values))

	itemSize := uint64(unsafe.Sizeof(optional[uint64]{}))
	assert.Equal(t, uint64(cap(values))*itemSize, budget.used)
	assert.Greater(t, budget.used, uint64(len(values))*itemSize)
}

func TestGrowCharged_CoversStringHeaderSlack(t *testing.T) {
	t.Parallel()

	var values []optional[string]
	budget := &decodeBudget{}
	elemSize := int(unsafe.Sizeof(optional[string]{}))

	for range 40 {
		bound, err := sliceCapacityBound(cap(values), len(values), 1, elemSize)
		require.NoError(t, err)

		values, err = growCharged(values, 1, budget)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, bound, cap(values))

		values = append(values, optional[string]{value: "a", valid: true})
	}

	assert.Equal(t, uint64(cap(values))*uint64(elemSize), budget.used)
}
