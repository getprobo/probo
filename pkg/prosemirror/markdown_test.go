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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMarkdown_EmptyInput(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("")
	require.NoError(t, err)
	assert.Equal(t, NodeDoc, doc.Type)
	assert.Empty(t, doc.Content)
}

func TestParseMarkdown_Paragraph(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("Hello world")
	require.NoError(t, err)
	assert.Equal(t, NodeDoc, doc.Type)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	assert.Equal(t, NodeParagraph, p.Type)
	require.Len(t, p.Content, 1)
	assert.Equal(t, NodeText, p.Content[0].Type)
	assert.Equal(t, "Hello world", *p.Content[0].Text)
}

func TestParseMarkdown_Headings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		markdown string
		level    int
	}{
		{"h1", "# Heading 1", 1},
		{"h2", "## Heading 2", 2},
		{"h3", "### Heading 3", 3},
		{"h4", "#### Heading 4", 4},
		{"h5", "##### Heading 5", 5},
		{"h6", "###### Heading 6", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := ParseMarkdown(tt.markdown)
			require.NoError(t, err)
			require.Len(t, doc.Content, 1)

			h := doc.Content[0]
			assert.Equal(t, NodeHeading, h.Type)

			attrs, err := h.HeadingAttrs()
			require.NoError(t, err)
			assert.Equal(t, tt.level, attrs.Level)

			require.Len(t, h.Content, 1)
			assert.Equal(t, NodeText, h.Content[0].Type)
		})
	}
}

func TestParseMarkdown_Bold(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("**bold text**")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "bold text", *txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, MarkStrong, txt.Marks[0].Type)
}

func TestParseMarkdown_Italic(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("*italic text*")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "italic text", *txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, MarkEm, txt.Marks[0].Type)
}

func TestParseMarkdown_Strikethrough(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("~~deleted~~")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "deleted", *txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, MarkStrike, txt.Marks[0].Type)
}

func TestParseMarkdown_InlineCode(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("`code`")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "code", *txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, MarkCode, txt.Marks[0].Type)
}

func TestParseMarkdown_CodeBlock(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		t.Parallel()

		doc, err := ParseMarkdown("```go\nfmt.Println(\"hello\")\n```")
		require.NoError(t, err)
		require.Len(t, doc.Content, 1)

		cb := doc.Content[0]
		assert.Equal(t, NodeCodeBlock, cb.Type)

		attrs, err := cb.CodeBlockAttrs()
		require.NoError(t, err)
		require.NotNil(t, attrs.Language)
		assert.Equal(t, "go", *attrs.Language)

		require.Len(t, cb.Content, 1)
		assert.Equal(t, "fmt.Println(\"hello\")\n", *cb.Content[0].Text)
	})

	t.Run("without language", func(t *testing.T) {
		t.Parallel()

		doc, err := ParseMarkdown("```\nsome code\n```")
		require.NoError(t, err)
		require.Len(t, doc.Content, 1)

		cb := doc.Content[0]
		assert.Equal(t, NodeCodeBlock, cb.Type)

		attrs, err := cb.CodeBlockAttrs()
		require.NoError(t, err)
		assert.Nil(t, attrs.Language)
	})
}

func TestParseMarkdown_Link(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("[click here](https://example.com)")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "click here", *txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, MarkLink, txt.Marks[0].Type)

	linkAttrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", linkAttrs.Href)
}

func TestParseMarkdown_Image(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("![alt text](https://example.com/img.png \"title\")")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	img := p.Content[0]
	assert.Equal(t, NodeImage, img.Type)

	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "alt text", *attrs.Alt)
	require.NotNil(t, attrs.Title)
	assert.Equal(t, "title", *attrs.Title)
}

func TestParseMarkdown_BulletList(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("- item 1\n- item 2\n- item 3")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, NodeBulletList, list.Type)
	require.Len(t, list.Content, 3)

	for i, item := range list.Content {
		assert.Equal(t, NodeListItem, item.Type, "item %d", i)
		require.Len(t, item.Content, 1)
		assert.Equal(t, NodeParagraph, item.Content[0].Type)
	}
}

func TestParseMarkdown_OrderedList(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("1. first\n2. second\n3. third")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, NodeOrderedList, list.Type)

	attrs, err := list.OrderedListAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, attrs.Start)

	require.Len(t, list.Content, 3)
}

func TestParseMarkdown_NestedList(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("- parent\n  - child\n  - child 2\n- parent 2")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, NodeBulletList, list.Type)
	require.Len(t, list.Content, 2)

	// First item should have a paragraph and a nested bullet list.
	firstItem := list.Content[0]
	assert.Equal(t, NodeListItem, firstItem.Type)
	require.Len(t, firstItem.Content, 2)
	assert.Equal(t, NodeParagraph, firstItem.Content[0].Type)
	assert.Equal(t, NodeBulletList, firstItem.Content[1].Type)
	require.Len(t, firstItem.Content[1].Content, 2)
}

func TestParseMarkdown_Blockquote(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("> quoted text")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	bq := doc.Content[0]
	assert.Equal(t, NodeBlockquote, bq.Type)
	require.Len(t, bq.Content, 1)
	assert.Equal(t, NodeParagraph, bq.Content[0].Type)
}

func TestParseMarkdown_HorizontalRule(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("---")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, NodeHorizontalRule, doc.Content[0].Type)
}

func TestParseMarkdown_HardBreak(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("line one\\\nline two")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	assert.Equal(t, NodeParagraph, p.Type)

	// Should contain: text("line one"), hardBreak, text("line two")
	var hasHardBreak bool
	for _, child := range p.Content {
		if child.Type == NodeHardBreak {
			hasHardBreak = true
		}
	}
	assert.True(t, hasHardBreak, "expected hard break node")
}

func TestParseMarkdown_NestedMarks(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("***bold and italic***")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "bold and italic", *txt.Text)
	require.Len(t, txt.Marks, 2)

	markTypes := make(map[MarkType]bool)
	for _, m := range txt.Marks {
		markTypes[m.Type] = true
	}
	assert.True(t, markTypes[MarkStrong])
	assert.True(t, markTypes[MarkEm])
}

func TestParseMarkdown_MixedContent(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown("Normal **bold** and *italic* text")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.True(t, len(p.Content) >= 4, "expected at least 4 inline nodes")

	// Verify first text node is plain.
	assert.Equal(t, "Normal ", *p.Content[0].Text)
	assert.Empty(t, p.Content[0].Marks)

	// Verify bold text node.
	assert.Equal(t, "bold", *p.Content[1].Text)
	require.Len(t, p.Content[1].Marks, 1)
	assert.Equal(t, MarkStrong, p.Content[1].Marks[0].Type)
}

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

	var joined string
	for _, ch := range p.Content {
		require.Equal(t, NodeText, ch.Type)
		require.NotNil(t, ch.Text)
		joined += *ch.Text
	}
	// Sanitized HTML: span is unwrapped to plain text content.
	assert.Equal(t, "before x after", joined)
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
	var joined string
	for _, ch := range p.Content {
		if ch.Type == NodeText && ch.Text != nil {
			joined += *ch.Text
		}
	}
	assert.NotContains(t, joined, "script")
	assert.NotContains(t, joined, "evil")
	assert.Contains(t, joined, "hi")
	assert.Contains(t, joined, "there")
}

func TestParseMarkdown_InlineRawHTMLWithOuterBold(t *testing.T) {
	t.Parallel()

	doc, err := ParseMarkdown(`**a <em>b</em> c**`)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	p := doc.Content[0]
	require.GreaterOrEqual(t, len(p.Content), 3)

	var joined string
	for _, ch := range p.Content {
		require.Equal(t, NodeText, ch.Type)
		require.NotNil(t, ch.Text)
		joined += *ch.Text
	}
	assert.Equal(t, "a b c", joined)

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

func TestParseMarkdown_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	md := `# Title

A paragraph with **bold**, *italic*, and ` + "`code`" + `.

- item 1
- item 2

> blockquote

---

` + "```go\nfmt.Println()\n```"

	doc, err := ParseMarkdown(md)
	require.NoError(t, err)

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	// Verify we can parse it back.
	parsed, err := Parse(string(data))
	require.NoError(t, err)
	assert.Equal(t, NodeDoc, parsed.Type)
	assert.True(t, len(parsed.Content) > 0)
}
