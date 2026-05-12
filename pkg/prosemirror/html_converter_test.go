// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package prosemirror

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMarkdown_BlockHTML(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("<div>block</div>\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeParagraph, doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, NodeText, doc.Content[0].Content[0].Type)
	require.NotNil(t, doc.Content[0].Content[0].Text)
	assert.Equal(t, "block", *doc.Content[0].Content[0].Text)
}

func TestParseMarkdown_BlockHTMLDivPreservesInlineMarks(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("<div>hello <strong>world</strong></div>\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	require.Equal(t, NodeParagraph, p.Type)

	var foundStrong bool
	for _, ch := range p.Content {
		if ch.Type != NodeText || ch.Text == nil {
			continue
		}
		if *ch.Text != "world" {
			continue
		}
		require.Len(t, ch.Marks, 1)
		assert.Equal(t, MarkStrong, ch.Marks[0].Type)
		foundStrong = true
	}
	assert.True(t, foundStrong, "expected bold mark on 'world' inside a single paragraph")
}

func TestParseMarkdown_BlockHTMLDivAroundSectionWithParagraphs(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("<div><section><p>a</p><p>b</p></section></div>\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 2)
	assert.Equal(t, NodeParagraph, doc.Content[0].Type)
	assert.Equal(t, NodeParagraph, doc.Content[1].Type)
	assert.Equal(t, "a", *doc.Content[0].Content[0].Text)
	assert.Equal(t, "b", *doc.Content[1].Content[0].Text)
}

func TestParseMarkdown_BlockHTMLWithClosureLine(t *testing.T) {
	t.Parallel()

	// Type 1 HTML block: closing tag is stored on ClosureLine, not in Lines.
	// Script is stripped by the HTML sanitizer; nothing safe remains.
	md := "<script>\nconsole.log(1)\n</script>\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	assert.Empty(t, doc.Content)
}

func TestParseMarkdown_BlockHTMLParagraphAndHeading(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("<p>a</p>\n<h2>Title</h2>\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 2)
	assert.Equal(t, NodeParagraph, doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, "a", *doc.Content[0].Content[0].Text)
	assert.Equal(t, NodeHeading, doc.Content[1].Type)
	attrs, err := doc.Content[1].HeadingAttrs()
	require.NoError(t, err)
	assert.Equal(t, 2, attrs.Level)
}

func TestParseMarkdown_BlockHTMLListAndBlockquote(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("<ul><li>one</li></ul>\n<blockquote><p>q</p></blockquote>\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 2)
	assert.Equal(t, NodeBulletList, doc.Content[0].Type)
	assert.Equal(t, NodeBlockquote, doc.Content[1].Type)
}

func TestParseMarkdown_BlockHTMLTable(t *testing.T) {
	t.Parallel()

	md := "<table><tr><th>A</th><td>B</td></tr></table>\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeTable, doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, NodeTableRow, doc.Content[0].Content[0].Type)
	require.Len(t, doc.Content[0].Content[0].Content, 2)
	assert.Equal(t, NodeTableHeader, doc.Content[0].Content[0].Content[0].Type)
	assert.Equal(t, NodeTableCell, doc.Content[0].Content[0].Content[1].Type)

	thAttrs, err := doc.Content[0].Content[0].Content[0].TableCellAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, thAttrs.Colspan)
	assert.Equal(t, 1, thAttrs.Rowspan)

	tdAttrs, err := doc.Content[0].Content[0].Content[1].TableCellAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, tdAttrs.Colspan)
	assert.Equal(t, 1, tdAttrs.Rowspan)
}

func TestParseMarkdown_BlockHTMLNestedTableDoesNotHoistInnerRows(t *testing.T) {
	t.Parallel()

	md := "<table><tr><td><table><tr><td>inner</td></tr></table></td></tr></table>\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeTable, doc.Content[0].Type)
	// Outer table must have exactly one row; inner <tr> must not become a second outer row.
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, NodeTableRow, doc.Content[0].Content[0].Type)
}

func TestParseMarkdown_BlockHTMLTableCellSpans(t *testing.T) {
	t.Parallel()

	md := `<table><tr><td colspan="3" rowspan="2">X</td></tr></table>` + "\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	cell := doc.Content[0].Content[0].Content[0]
	require.Equal(t, NodeTableCell, cell.Type)
	attrs, err := cell.TableCellAttrs()
	require.NoError(t, err)
	assert.Equal(t, 3, attrs.Colspan)
	assert.Equal(t, 2, attrs.Rowspan)
}

func TestParseMarkdown_MarkdownTableCellWithBlockList(t *testing.T) {
	t.Parallel()

	md := "| Header |\n| --- |\n| <ul><li>one</li><li>two</li></ul> |\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeTable, doc.Content[0].Type)

	dataRow := doc.Content[0].Content[1]
	assert.Equal(t, NodeTableRow, dataRow.Type)
	cell := dataRow.Content[0]
	assert.Equal(t, NodeTableCell, cell.Type)

	require.Len(t, cell.Content, 1)
	assert.Equal(t, NodeBulletList, cell.Content[0].Type)
	require.Len(t, cell.Content[0].Content, 2)
	assert.Equal(t, NodeListItem, cell.Content[0].Content[0].Type)
	assert.Equal(t, NodeListItem, cell.Content[0].Content[1].Type)

	li1Para := cell.Content[0].Content[0].Content[0]
	require.Equal(t, NodeParagraph, li1Para.Type)
	assert.Equal(t, "one", *li1Para.Content[0].Text)

	li2Para := cell.Content[0].Content[1].Content[0]
	require.Equal(t, NodeParagraph, li2Para.Type)
	assert.Equal(t, "two", *li2Para.Content[0].Text)
}

func TestParseMarkdown_MarkdownTableCellWithOrderedList(t *testing.T) {
	t.Parallel()

	md := "| Header |\n| --- |\n| <ol><li>first</li><li>second</li></ol> |\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeTable, doc.Content[0].Type)

	dataRow := doc.Content[0].Content[1]
	cell := dataRow.Content[0]
	assert.Equal(t, NodeTableCell, cell.Type)

	require.Len(t, cell.Content, 1)
	assert.Equal(t, NodeOrderedList, cell.Content[0].Type)
	require.Len(t, cell.Content[0].Content, 2)
	assert.Equal(t, NodeListItem, cell.Content[0].Content[0].Type)
	assert.Equal(t, NodeListItem, cell.Content[0].Content[1].Type)
}

func TestParseMarkdown_BlockHTMLScriptRemovedKeepsSafeContent(t *testing.T) {
	t.Parallel()

	md := "<p>ok</p>\n<script>bad()</script>\n"
	doc, err := ParseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeParagraph, doc.Content[0].Type)
	assert.Equal(t, "ok", *doc.Content[0].Content[0].Text)
}

func TestParseMarkdown_InlineRawHTML(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("before <span>x</span> after")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	require.Equal(t, NodeParagraph, p.Type)

	var joined strings.Builder
	for _, ch := range p.Content {
		require.Equal(t, NodeText, ch.Type)
		require.NotNil(t, ch.Text)
		joined.WriteString(*ch.Text)
	}
	// Sanitized HTML: span is unwrapped to plain text content.
	assert.Equal(t, "before x after", joined.String())
}

func TestParseMarkdown_InlineRawHTMLStrong(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown(`a <strong>b</strong> c`)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	require.Len(t, p.Content, 3)
	assert.Equal(t, "a ", *p.Content[0].Text)
	assert.Equal(t, "b", *p.Content[1].Text)
	require.Len(t, p.Content[1].Marks, 1)
	assert.Equal(t, MarkStrong, p.Content[1].Marks[0].Type)
	assert.Equal(t, " c", *p.Content[2].Text)
}

func TestParseMarkdown_InlineRawHTMLScriptStripped(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown(`hi <script>evil()</script> there`)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	var joined strings.Builder
	for _, ch := range p.Content {
		if ch.Type == NodeText && ch.Text != nil {
			joined.WriteString(*ch.Text)
		}
	}
	assert.NotContains(t, joined.String(), "script")
	assert.NotContains(t, joined.String(), "evil")
	assert.Contains(t, joined.String(), "hi")
	assert.Contains(t, joined.String(), "there")
}

func TestParseMarkdown_InlineRawHTMLWithOuterBold(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown(`**a <em>b</em> c**`)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	require.GreaterOrEqual(t, len(p.Content), 3)

	var joined strings.Builder
	for _, ch := range p.Content {
		require.Equal(t, NodeText, ch.Type)
		require.NotNil(t, ch.Text)
		joined.WriteString(*ch.Text)
	}
	assert.Equal(t, "a b c", joined.String())

	var mid *Node
	for i := range p.Content {
		if p.Content[i].Text != nil && *p.Content[i].Text == "b" {
			mid = &p.Content[i]
			break
		}
	}
	require.NotNil(t, mid, "expected inner <em> as text node b")
	require.GreaterOrEqual(t, len(mid.Marks), 2)
	assert.Equal(t, MarkStrong, mid.Marks[0].Type)
	assert.Equal(t, MarkEm, mid.Marks[1].Type)
}
