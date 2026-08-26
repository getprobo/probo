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

package prosemirror_test

import (
	"fmt"
	"strings"
	"testing"

	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

func BenchmarkRender(b *testing.B) {
	benchmarks := []struct {
		name  string
		spans []automerge.Span
		bytes int64
	}{
		{
			name: "PlainText10KiB",
			spans: []automerge.Span{
				benchmarkBlock("paragraph", nil),
				{Type: automerge.SpanTypeText, Text: strings.Repeat("a", 10*1024)},
			},
			bytes: 10 * 1024,
		},
		{
			name:  "ThousandParagraphs100KiB",
			spans: benchmarkParagraphs(1_000, 100),
			bytes: 100_000,
		},
		{
			name:  "Table100x10x20B",
			spans: benchmarkTable(100, 10, 20),
			bytes: 20_000,
		},
		{
			name:  "MalformedHierarchy100KiB",
			spans: benchmarkMalformed(1_000, 100),
			bytes: 100_000,
		},
	}

	for _, benchmark := range benchmarks {
		b.Run(
			benchmark.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(benchmark.bytes)

				for b.Loop() {
					if _, err := automergeprosemirror.Render(benchmark.spans); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func benchmarkParagraphs(count, textLength int) []automerge.Span {
	spans := make([]automerge.Span, 0, count*2)

	for range count {
		spans = append(
			spans,
			benchmarkBlock("paragraph", nil),
			automerge.Span{
				Type: automerge.SpanTypeText,
				Text: strings.Repeat("a", textLength),
			},
		)
	}

	return spans
}

func benchmarkTable(rows, columns, textLength int) []automerge.Span {
	spans := []automerge.Span{benchmarkBlock("table", nil)}

	for row := range rows {
		spans = append(spans, benchmarkBlock("table-row", []any{"table"}))

		for column := range columns {
			spans = append(
				spans,
				benchmarkBlock("table-cell", []any{"table", "table-row"}),
				automerge.Span{
					Type: automerge.SpanTypeText,
					Text: fmt.Sprintf(
						"%04d:%02d:%s",
						row,
						column,
						strings.Repeat("a", textLength),
					),
				},
			)
		}
	}

	return spans
}

func benchmarkMalformed(count, textLength int) []automerge.Span {
	spans := make([]automerge.Span, 0, count*2)

	for index := range count {
		spans = append(
			spans,
			benchmarkBlock(
				"paragraph",
				[]any{"table", "table-row", fmt.Sprintf("missing-%d", index)},
			),
			automerge.Span{
				Type: automerge.SpanTypeText,
				Text: strings.Repeat("a", textLength),
			},
		)
	}

	return spans
}

func benchmarkBlock(blockType string, parents []any) automerge.Span {
	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockType,
			"parents": parents,
			"attrs":   map[string]any{},
		},
	}
}
