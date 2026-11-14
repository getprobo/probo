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
