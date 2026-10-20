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

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

var benchmarkDecodedDocument any

func BenchmarkDecodeDocument(b *testing.B) {
	document, operations := benchmarkSnapshot(10_000)
	for index := range document.Changes {
		if _, err := EncodeChange(&document.Changes[index]); err != nil {
			b.Fatal(err)
		}
	}

	document.Heads = []opset.ChangeHash{*document.Changes[0].Hash}

	document.OperationOrder = make([]opset.OpID, len(operations))
	for index := range operations {
		document.OperationOrder[index] = operations[index].ID
	}

	data, err := EncodePreparedDocument(document, operations, true)
	if err != nil {
		b.Fatal(err)
	}

	decoded, err := Decode(data)
	if err != nil {
		b.Fatal(err)
	}

	if len(decoded.OperationOrder) != len(operations) {
		b.Fatalf(
			"decoded %d operations, want %d",
			len(decoded.OperationOrder),
			len(operations),
		)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		decoded, err := Decode(data)
		if err != nil {
			b.Fatal(err)
		}

		benchmarkDecodedDocument = decoded
	}
}
