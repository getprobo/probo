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

package automerge_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func manyParagraphDocument(tb testing.TB, paragraphs int) *automerge.Document {
	tb.Helper()

	actorID, err := automerge.NewActorID()
	require.NoError(tb, err)

	document, err := automerge.New(actorID)
	require.NoError(tb, err)

	spans := make([]automerge.SpanInput, 0, 2*paragraphs)
	for i := range paragraphs {
		spans = append(
			spans,
			automerge.SpanInput{
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
					"attrs":   map[string]any{},
				},
			},
			automerge.SpanInput{Text: fmt.Sprintf("Paragraph %d.", i)},
		)
	}

	text, err := document.CreateText("body")
	require.NoError(tb, err)
	require.NoError(
		tb,
		text.UpdateSpans(spans, automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}),
	)

	_, err = document.Commit("Seed paragraphs", time.Now())
	require.NoError(tb, err)

	return document
}

func TestTextSpans_ManyBlocks(t *testing.T) {
	t.Parallel()

	const paragraphs = 300

	document := manyParagraphDocument(t, paragraphs)

	text, err := document.Text("body")
	require.NoError(t, err)

	spans, err := text.Spans()
	require.NoError(t, err)

	var blocks, texts int
	for _, span := range spans {
		switch span.Type {
		case automerge.SpanTypeBlock:
			blocks++
			assert.Equal(t, "paragraph", span.Block["type"])
		case automerge.SpanTypeText:
			texts++
		}
	}

	assert.Equal(t, paragraphs, blocks)
	assert.Equal(t, paragraphs, texts)
	require.NotEmpty(t, spans)
	assert.Equal(t, fmt.Sprintf("Paragraph %d.", paragraphs-1), spans[len(spans)-1].Text)
}

func BenchmarkTextSpans_ManyBlocks(b *testing.B) {
	document := manyParagraphDocument(b, 300)

	text, err := document.Text("body")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := text.Spans(); err != nil {
			b.Fatal(err)
		}
	}
}
