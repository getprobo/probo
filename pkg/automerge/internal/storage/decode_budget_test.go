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

	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestDecode_ReleasesColumnScratch(t *testing.T) {
	t.Parallel()

	const size = 20_000

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
