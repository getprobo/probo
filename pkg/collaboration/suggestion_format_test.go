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

const formatSuggestionID = "suggestion-format"

func TestApplySuggestionFormat_AddsBoldUntilItIsAccepted(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWith("Hello"))
	change := FormatSuggestion{Mark: FormatSuggestionBold, Action: FormatSuggestionAdd, Quote: "Hello"}

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, change))

	spans := formatSpans(t, document)
	assert.Equal(t, true, textMarks(spans, "Hello")["strong"])
	assert.Contains(t, textMarks(spans, "Hello")["suggestion-format-strong"], formatSuggestionID)
	assert.False(t, markPresent(ProjectAcceptedSpans(spans)[1].Marks, "strong"))

	changed, err := ApplySuggestionDecision(document, formatSuggestionID, true)
	require.NoError(t, err)
	require.True(t, changed)

	accepted := formatSpans(t, document)
	assert.Equal(t, true, textMarks(accepted, "Hello")["strong"])
	assert.NotContains(t, textMarks(accepted, "Hello"), "suggestion-format-strong")
	assert.False(t, HasOpenSuggestions(accepted))
}

func TestApplySuggestionFormat_RejectRestoresBold(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWithBold("Hello"))
	change := FormatSuggestion{Mark: FormatSuggestionBold, Action: FormatSuggestionRemove, Quote: "Hello"}

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, change))
	assert.False(t, markPresent(textMarks(formatSpans(t, document), "Hello"), "strong"))

	changed, err := ApplySuggestionDecision(document, formatSuggestionID, false)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, true, textMarks(formatSpans(t, document), "Hello")["strong"])
	assert.False(t, HasOpenSuggestions(formatSpans(t, document)))
}

func TestApplySuggestionFormat_FormatsOnlyTheQuotedSlice(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWith("Hello"))

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionItalic,
		Action: FormatSuggestionAdd,
		Quote:  "ell",
	}))

	spans := formatSpans(t, document)
	assert.Equal(t, []string{"H", "ell", "o"}, textPieces(spans))
	assert.False(t, markPresent(textMarks(spans, "H"), "em"))
	assert.Equal(t, true, textMarks(spans, "ell")["em"])
	assert.False(t, markPresent(textMarks(spans, "o"), "em"))
}

func TestApplySuggestionFormat_SuggestsALinkAndRejectsIt(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWith("Hello"))

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionLink,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
		Href:   "https://example.com",
	}))
	assert.Contains(t, linkHrefOf(textMarks(formatSpans(t, document), "Hello")), "https://example.com")
	assert.Equal(t, "", linkHrefOf(ProjectAcceptedSpans(formatSpans(t, document))[1].Marks))

	rejected := suggestionDocument(t, docWith("Hello"))
	require.NoError(t, ApplySuggestionFormat(rejected, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionLink,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
		Href:   "https://example.com",
	}))

	changed, err := ApplySuggestionDecision(rejected, formatSuggestionID, false)
	require.NoError(t, err)
	require.True(t, changed)
	assert.False(t, markPresent(textMarks(formatSpans(t, rejected), "Hello"), "link"))
}

func TestApplySuggestionFormat_RejectsAnUnsafeLink(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWith("Hello"))
	err := ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionLink,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
		Href:   "javascript:alert(1)",
	})
	require.ErrorIs(t, err, ErrSuggestionHref)
	assert.False(t, markPresent(textMarks(formatSpans(t, document), "Hello"), "link"))
}

func TestApplySuggestionFormat_CodeReplacesBoldUntilRejected(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWithBold("Hello"))

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionCode,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
	}))

	marks := textMarks(formatSpans(t, document), "Hello")
	assert.Equal(t, true, marks["code"])
	assert.False(t, markPresent(marks, "strong"))
	assert.Contains(t, marks["suggestion-format-strong"], formatActionRemove)

	changed, err := ApplySuggestionDecision(document, formatSuggestionID, false)
	require.NoError(t, err)
	require.True(t, changed)

	restored := textMarks(formatSpans(t, document), "Hello")
	assert.Equal(t, true, restored["strong"])
	assert.False(t, markPresent(restored, "code"))
}

func TestApplySuggestionFormat_RefusesAnExistingRedline(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, docWith("Hello"))

	require.NoError(t, ApplySuggestionEdit(document, formatSuggestionID, "Hello", "Hi", 0))
	err := ApplySuggestionFormat(document, "other", FormatSuggestion{
		Mark:   FormatSuggestionBold,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
	})
	require.ErrorIs(t, err, ErrSuggestionOverlaps)

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionBold,
		Action: FormatSuggestionAdd,
		Quote:  "i",
	}))
	marks := textMarks(formatSpans(t, document), "i")
	assert.Equal(t, true, marks["strong"])
	assert.NotContains(t, marks, "suggestion-format-strong")
}

func TestApplySuggestionFormat_KeepsTheAcceptedDocumentStable(t *testing.T) {
	t.Parallel()

	before := formatSpans(t, suggestionDocument(t, docWith("Hello")))
	document := suggestionDocument(t, docWith("Hello"))

	require.NoError(t, ApplySuggestionFormat(document, formatSuggestionID, FormatSuggestion{
		Mark:   FormatSuggestionBold,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
	}))

	after := formatSpans(t, document)
	open := map[string]bool{formatSuggestionID: true}

	assert.True(t, CommentChangeAllowed(before, after, open))
	assert.False(t, CommentChangeAllowed(before, after, map[string]bool{}))
	assert.False(t, HasOpenSuggestions(ProjectAcceptedSpans(after)))

	bold := formatSpans(t, suggestionDocument(t, docWithBold("Hello")))
	assert.False(t, CommentChangeAllowed(before, bold, open))
}

func docWith(text string) string {
	return `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + text + `"}]}]}`
}

func docWithBold(text string) string {
	return `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + text + `","marks":[{"type":"bold"}]}]}]}`
}

func formatSpans(t *testing.T, document *automerge.Document) []automerge.Span {
	t.Helper()

	spans, err := BodySpans(document)
	require.NoError(t, err)

	return spans
}

func textMarks(spans []automerge.Span, text string) map[string]any {
	for _, span := range spans {
		if span.Type == automerge.SpanTypeText && span.Text == text {
			return span.Marks
		}
	}

	return nil
}

func textPieces(spans []automerge.Span) []string {
	pieces := make([]string, 0)

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText && span.Text != "" {
			pieces = append(pieces, span.Text)
		}
	}

	return pieces
}
