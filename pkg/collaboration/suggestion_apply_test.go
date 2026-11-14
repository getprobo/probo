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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestApplyDirectText_BesideOpenSuggestion(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Keep")+`,`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(document, "s-keep", "Keep", "Kept", 0))
	require.NoError(t, ApplyDirectText(document, "scope", "range", 0))

	accepted, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, accepted, "Ke")
	assert.Contains(t, accepted, "ep")
	assert.Contains(t, accepted, "range")
	assert.NotContains(t, accepted, "scope")

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.True(t, HasOpenSuggestions(spans))
}

func TestApplyDirectTable_RefusesAQuoteInsideASuggestion(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())
	require.NoError(t, ApplySuggestionFormat(document, "s-bold", FormatSuggestion{
		Mark:   FormatSuggestionBold,
		Action: FormatSuggestionAdd,
		Quote:  "Read",
	}))

	original := bodySpans(t, document)

	err := ApplyDirectTable(document, TableSuggestion{
		Operation: TableSuggestionInsertColumn,
		Quote:     "Read",
	})
	require.ErrorIs(t, err, ErrSuggestionOverlaps)
	assert.True(t, spansEqual(original, bodySpans(t, document)))

	err = ApplyDirectTable(document, TableSuggestion{
		Operation: TableSuggestionDeleteRow,
		Quote:     "Read",
	})
	require.ErrorIs(t, err, ErrSuggestionOverlaps)
	assert.True(t, spansEqual(original, bodySpans(t, document)))

	require.NoError(t, ApplyDirectTable(document, TableSuggestion{
		Operation: TableSuggestionInsertRow,
		Quote:     "Right",
	}))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.True(t, HasOpenSuggestions(spans))
	assert.Contains(t, textMarks(spans, "Read")["suggestion-format-strong"], "s-bold")
	assert.Equal(t, 3, countBlocks(spans, blockTypeTableRow))

	accepted, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, accepted, "Read")
	assert.Contains(t, accepted, "Right")
	assert.Equal(t, 3, strings.Count(accepted, `"type":"tableRow"`))
}

func TestApplyDirectText_RefusesOverlap(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(document, "s-keep", "scope", "range", 0))

	err := ApplyDirectText(document, "rang", "area", 0)
	require.ErrorIs(t, err, ErrSuggestionOverlaps)
}

func TestApplyDirectText_HardBreak(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("HelloWorld")+`]}`)
	require.NoError(t, ApplyDirectText(document, "HelloWorld", "Hello\nWorld", 0))

	spans, err := BodySpans(document)
	require.NoError(t, err)

	found := false

	for _, span := range spans {
		if blockType(span) == blockTypeHardBreak {
			found = true
		}
	}

	assert.True(t, found)
}

func TestApplyDirectText_HardBreakOnlyInTheEditedBlock(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[`+paragraphJSON("a\nb")+`,`+paragraphJSON("HelloWorld")+`]}`,
	)
	require.NoError(t, ApplyDirectText(document, "HelloWorld", "Hello\nWorld", 0))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.Equal(t, 1, countBlocks(spans, blockTypeHardBreak))
	assert.Contains(t, joinedSpanText(spans), "a\nb")
}

func TestApplyDirectText_KeepsNewlinesInCode(t *testing.T) {
	t.Parallel()

	code := `{"type":"codeBlock","content":[{"type":"text","text":"ab"}]}`
	document := suggestionDocument(t, `{"type":"doc","content":[`+code+`]}`)
	require.NoError(t, ApplyDirectText(document, "ab", "a\nb", 0))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.Equal(t, 0, countBlocks(spans, blockTypeHardBreak))
	assert.Contains(t, joinedSpanText(spans), "a\nb")
}

func countBlocks(spans []automerge.Span, kind string) int {
	count := 0

	for _, span := range spans {
		if blockType(span) == kind {
			count++
		}
	}

	return count
}

func TestApplyDirectFormat_WritesTheStyle(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`]}`)
	require.NoError(t, ApplyDirectFormat(document, FormatSuggestion{
		Mark:   FormatSuggestionBold,
		Action: FormatSuggestionAdd,
		Quote:  "Hello",
	}))

	spans, err := BodySpans(document)
	require.NoError(t, err)
	assert.False(t, HasOpenSuggestions(spans))

	accepted, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, accepted, "bold")
}

func TestApplyBlock_HeadingSuggestionRestores(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`]}`)
	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation:    BlockSet,
		Quote:        "Hello",
		Block:        BlockKindHeading,
		Level:        2,
		SuggestionID: "s-head",
	}))

	readers, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.NotContains(t, readers, "heading")
	assert.Contains(t, readers, "Hello")

	require.NoError(t, WriteBodySpans(document, decidedSpans(t, document, "s-head", true)))

	live, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, live, "heading")
}

func TestApplyBlock_RejectHeadingRestoresParagraph(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`]}`)
	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation:    BlockSet,
		Quote:        "Hello",
		Block:        BlockKindHeading,
		Level:        2,
		SuggestionID: "s-head",
	}))

	require.NoError(t, WriteBodySpans(document, decidedSpans(t, document, "s-head", false)))

	live, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.NotContains(t, live, "heading")
	assert.Contains(t, live, "Hello")
}

func TestApplyBlock_ListAndImage(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`]}`)
	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation: BlockSet,
		Quote:     "Hello",
		Block:     BlockKindBulletList,
	}))

	listed, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, listed, "bulletList")

	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation: BlockInsertImage,
		Quote:     "Hello",
		Src:       "https://example.com/diagram.png",
		Alt:       "Diagram",
	}))

	withImage, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, withImage, "image")
	assert.Contains(t, withImage, "https://example.com/diagram.png")
}

func TestApplyBlock_RetypesNestedItems(t *testing.T) {
	t.Parallel()

	child := `{"type":"bulletList","content":[{"type":"listItem","content":[` + paragraphJSON("Child") + `]}]}`
	parent := `{"type":"bulletList","content":[{"type":"listItem","content":[` + paragraphJSON("Parent") + `,` + child + `]}]}`
	document := suggestionDocument(t, `{"type":"doc","content":[`+parent+`]}`)

	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation:    BlockSet,
		Quote:        "Parent",
		Block:        BlockKindOrderedList,
		SuggestionID: "s-list",
	}))
	assert.Equal(t, []string{blockTypeOrderedListItem}, childParents(t, document))

	require.NoError(t, WriteBodySpans(document, decidedSpans(t, document, "s-list", false)))
	assert.Equal(t, []string{blockTypeUnorderedListItem}, childParents(t, document))
}

func childParents(t *testing.T, document *automerge.Document) []string {
	t.Helper()

	spans, err := BodySpans(document)
	require.NoError(t, err)

	for index, span := range spans {
		if span.Type == automerge.SpanTypeText && span.Text == "Child" {
			for cursor := index - 1; cursor >= 0; cursor-- {
				if spans[cursor].Type == automerge.SpanTypeBlock {
					return blockParents(spans[cursor])
				}
			}
		}
	}

	return nil
}

func TestApplyBlock_MentionRefusedInsideBlockSuggestion(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`,`+paragraphJSON("World")+`]}`)
	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation:    BlockSet,
		Quote:        "Hello",
		Block:        BlockKindHeading,
		Level:        2,
		SuggestionID: "s-head",
	}))

	err := ApplyBlock(document, BlockChange{
		Operation: BlockInsertMention,
		Quote:     "Hello",
		ProfileID: "profile",
		Label:     "Ada",
	})
	require.ErrorIs(t, err, ErrSuggestionOverlaps)

	require.NoError(t, ApplyBlock(document, BlockChange{
		Operation: BlockInsertMention,
		Quote:     "World",
		ProfileID: "profile",
		Label:     "Ada",
	}))
}

func TestApplyBlock_UnsafeImage(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("Hello")+`]}`)
	err := ApplyBlock(document, BlockChange{
		Operation: BlockInsertImage,
		Quote:     "Hello",
		Src:       "javascript:alert(1)",
	})
	require.ErrorIs(t, err, ErrSuggestionHref)
}

func decidedSpans(t *testing.T, document *automerge.Document, id string, accept bool) []automerge.Span {
	t.Helper()

	spans, err := BodySpans(document)
	require.NoError(t, err)

	next, changed := decideSuggestionSpans(spans, id, accept)
	require.True(t, changed)

	return next
}
