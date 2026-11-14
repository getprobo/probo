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

package collaboration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestWritePreservingSuggestions_EditsTheOtherParagraph(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Keep")+`,`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(document, "s-keep", "Keep", "Kept", 0))

	replacement := `{"type":"doc","content":[` + paragraphJSON("Keep") + `,` + paragraphJSON("range") + `]}`
	require.NoError(t, writePreservingSuggestions(document, replacement))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.Equal(t, "pt", markedText(spans, SuggestionInsertMark, "s-keep"))
	assert.Equal(t, "ep", markedText(spans, SuggestionDeleteMark, "s-keep"))
	assert.Contains(t, joinedSpanText(spans), "range")
	assert.NotContains(t, joinedSpanText(spans), "scope")
}

func TestWritePreservingSuggestions_KeepsATypeChangeWithTheSameText(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Keep")+`,`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(document, "s-keep", "Keep", "Kept", 0))

	heading := `{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"scope"}]}`
	replacement := `{"type":"doc","content":[` + paragraphJSON("Keep") + `,` + heading + `]}`
	require.NoError(t, writePreservingSuggestions(document, replacement))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.Equal(t, "pt", markedText(spans, SuggestionInsertMark, "s-keep"))

	types := make([]string, 0, len(spans))
	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			types = append(types, blockType(span))
		}
	}

	assert.Equal(t, []string{"paragraph", "heading"}, types)
}

func markedText(spans []automerge.Span, mark string, id string) string {
	text := ""

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText && markID(span.Marks, mark) == id {
			text += span.Text
		}
	}

	return text
}

func TestWritePreservingSuggestions_RefusesTheSuggestedParagraph(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Keep")+`,`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(document, "s-keep", "Keep", "Kept", 0))

	replacement := `{"type":"doc","content":[` + paragraphJSON("Drop") + `,` + paragraphJSON("scope") + `]}`
	err := writePreservingSuggestions(document, replacement)
	require.ErrorIs(t, err, ErrSuggestionOverlaps)
}

func joinedSpanText(spans []automerge.Span) string {
	text := ""

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText {
			text += span.Text
		}
	}

	return text
}
