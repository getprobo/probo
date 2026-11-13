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

package markdown_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
	"go.probo.inc/probo/pkg/richtext"
	"go.probo.inc/probo/pkg/richtext/attachment"
	"go.probo.inc/probo/pkg/richtext/mention"
)

// editorSchema matches the product document schema closely enough for import tests.
var editorSchema = prosemirror.NewSchema(
	[]prosemirror.NodeSpec{
		prosemirror.Doc(),
		prosemirror.Paragraph(),
		prosemirror.Text(),
		prosemirror.Heading(),
		prosemirror.Blockquote(),
		prosemirror.CodeBlock(),
		prosemirror.HorizontalRule(),
		prosemirror.HardBreak(),
		attachment.Image(),
		attachment.File(),
		mention.Document(),
		prosemirror.BulletList(),
		prosemirror.OrderedList(),
		prosemirror.ListItem(),
		prosemirror.TaskList(),
		prosemirror.TaskItem(),
		prosemirror.Table(),
		prosemirror.TableRow(),
		prosemirror.TableCell(),
		prosemirror.TableHeader(),
	},
	[]prosemirror.MarkSpec{
		prosemirror.Bold(),
		prosemirror.Italic(),
		prosemirror.Underline(),
		prosemirror.Strike(),
		prosemirror.Code(),
		prosemirror.Link(),
	},
)

func attachmentPath(fileID string) string {
	return attachment.Path(fileID)
}

func parseMarkdown(src string) (prosemirror.Node, error) {
	node, err := markdown.Parse(src)
	if err != nil {
		return prosemirror.Node{}, err
	}

	markdown.ApplyRules(&node, attachment.ImportRule())

	return node, nil
}

func TestParseMarkdown_EmptyInput(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("")
	require.NoError(t, err)
	assert.Equal(t, prosemirror.NodeDoc, doc.Type)
	assert.Empty(t, doc.Content)
}

func TestParseMarkdown_Paragraph(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("Hello world")
	require.NoError(t, err)
	assert.Equal(t, prosemirror.NodeDoc, doc.Type)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	assert.Equal(t, prosemirror.NodeParagraph, p.Type)
	require.Len(t, p.Content, 1)
	assert.Equal(t, prosemirror.NodeText, p.Content[0].Type)
	assert.Equal(t, "Hello world", p.Content[0].Text)
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

			doc, err := parseMarkdown(tt.markdown)
			require.NoError(t, err)
			require.Len(t, doc.Content, 1)

			h := doc.Content[0]
			assert.Equal(t, prosemirror.NodeHeading, h.Type)

			attrs, err := h.HeadingAttrs()
			require.NoError(t, err)
			assert.Equal(t, tt.level, attrs.Level)

			require.Len(t, h.Content, 1)
			assert.Equal(t, prosemirror.NodeText, h.Content[0].Type)
		})
	}
}

func TestParseMarkdown_Bold(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("**bold text**")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "bold text", txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, prosemirror.MarkStrong, txt.Marks[0].Type)
}

func TestParseMarkdown_Italic(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("*italic text*")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "italic text", txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, prosemirror.MarkEm, txt.Marks[0].Type)
}

func TestParseMarkdown_Strikethrough(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("~~deleted~~")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "deleted", txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, prosemirror.MarkStrike, txt.Marks[0].Type)
}

func TestParseMarkdown_InlineCode(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("`code`")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "code", txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, prosemirror.MarkCode, txt.Marks[0].Type)
}

func TestParseMarkdown_CodeBlock(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		t.Parallel()

		doc, err := parseMarkdown("```go\nfmt.Println(\"hello\")\n```")
		require.NoError(t, err)
		require.Len(t, doc.Content, 1)

		cb := doc.Content[0]
		assert.Equal(t, prosemirror.NodeCodeBlock, cb.Type)

		attrs, err := cb.CodeBlockAttrs()
		require.NoError(t, err)
		require.NotNil(t, attrs.Language)
		assert.Equal(t, "go", *attrs.Language)

		require.Len(t, cb.Content, 1)
		assert.Equal(t, "fmt.Println(\"hello\")", cb.Content[0].Text)
	})

	t.Run("without language", func(t *testing.T) {
		t.Parallel()

		doc, err := parseMarkdown("```\nsome code\n```")
		require.NoError(t, err)
		require.Len(t, doc.Content, 1)

		cb := doc.Content[0]
		assert.Equal(t, prosemirror.NodeCodeBlock, cb.Type)

		attrs, err := cb.CodeBlockAttrs()
		require.NoError(t, err)
		assert.Nil(t, attrs.Language)

		require.Len(t, cb.Content, 1)
		assert.Equal(t, "some code", cb.Content[0].Text)
	})

	t.Run("trailing blank line preserved", func(t *testing.T) {
		t.Parallel()

		doc, err := parseMarkdown("```\nline\n\n```")
		require.NoError(t, err)
		require.Len(t, doc.Content, 1)

		cb := doc.Content[0]
		require.Len(t, cb.Content, 1)
		assert.Equal(t, "line\n\n", cb.Content[0].Text)
	})
}

func TestParseMarkdown_Link(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("[click here](https://example.com)")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "click here", txt.Text)
	require.Len(t, txt.Marks, 1)
	assert.Equal(t, prosemirror.MarkLink, txt.Marks[0].Type)

	linkAttrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", linkAttrs.Href)
}

func TestParseMarkdown_LinkSanitizesDangerousHrefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		markdown string
		wantHref string
	}{
		{name: "javascript scheme", markdown: `[x](javascript:alert(1))`, wantHref: `#`},
		{name: "data html", markdown: `[x](data:text/html,<script>alert(1)</script>)`, wantHref: `#`},
		{name: "protocol-relative", markdown: `[x](//evil.example/phish)`, wantHref: `#`},
		{name: "https preserved", markdown: `[x](https://example.com/y)`, wantHref: `https://example.com/y`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parseMarkdown(tt.markdown)
			require.NoError(t, err)

			txt := doc.Content[0].Content[0]
			linkAttrs, err := txt.Marks[0].LinkAttrs()
			require.NoError(t, err)
			assert.Equal(t, tt.wantHref, linkAttrs.Href)
		})
	}
}

func TestParseMarkdown_ImageOnlyHeading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
	}{
		{name: "markdown", source: "# ![diagram](https://example.com/diagram.png)"},
		{name: "html", source: `<h1><img src="https://example.com/diagram.png" alt="diagram"></h1>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parseMarkdown(tt.source)
			require.NoError(t, err)
			require.NoError(t, editorSchema.Validate(doc))
			require.Len(t, doc.Content, 1)

			image := doc.Content[0]
			assert.Equal(t, prosemirror.NodeImage, image.Type)

			imageAttrs, err := image.ImageAttrs()
			require.NoError(t, err)
			assert.Equal(t, "https://example.com/diagram.png", imageAttrs.Src)
			require.NotNil(t, imageAttrs.Alt)
			assert.Equal(t, "diagram", *imageAttrs.Alt)
		})
	}
}

func TestParseMarkdown_HeadingStartsWithImage(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("# ![diagram](https://example.com/diagram.png) Title")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 3)

	heading := doc.Content[0]
	assert.Equal(t, prosemirror.NodeHeading, heading.Type)
	assert.Empty(t, heading.Content)

	attrs, err := heading.HeadingAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, attrs.Level)

	assert.Equal(t, prosemirror.NodeImage, doc.Content[1].Type)
	assert.Equal(t, prosemirror.NodeParagraph, doc.Content[2].Type)
	require.Len(t, doc.Content[2].Content, 1)
	assert.Equal(t, " Title", doc.Content[2].Content[0].Text)
}

func TestParseMarkdown_HeadingImageThenText(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("# Hello ![diagram](https://example.com/diagram.png) world")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 3)

	heading := doc.Content[0]
	assert.Equal(t, prosemirror.NodeHeading, heading.Type)
	require.Len(t, heading.Content, 1)
	assert.Equal(t, "Hello ", heading.Content[0].Text)

	attrs, err := heading.HeadingAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, attrs.Level)

	assert.Equal(t, prosemirror.NodeImage, doc.Content[1].Type)
	assert.Equal(t, prosemirror.NodeParagraph, doc.Content[2].Type)
	require.Len(t, doc.Content[2].Content, 1)
	assert.Equal(t, " world", doc.Content[2].Content[0].Text)
}

func TestParseMarkdown_ImagesSeparatedBySpace(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("before ![one](https://example.com/one.png) ![two](https://example.com/two.png) after")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 4)
	assert.Equal(t, prosemirror.NodeParagraph, doc.Content[0].Type)
	assert.Equal(t, "before ", doc.Content[0].Content[0].Text)
	assert.Equal(t, prosemirror.NodeImage, doc.Content[1].Type)
	assert.Equal(t, prosemirror.NodeImage, doc.Content[2].Type)
	assert.Equal(t, prosemirror.NodeParagraph, doc.Content[3].Type)
	assert.Equal(t, " after", doc.Content[3].Content[0].Text)
}

func TestParseMarkdown_Image(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("![alt text](https://example.com/img.png \"title\")")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	img := doc.Content[0]
	assert.Equal(t, prosemirror.NodeImage, img.Type)

	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "alt text", *attrs.Alt)
	require.NotNil(t, attrs.Title)
	assert.Equal(t, "title", *attrs.Title)
}

func TestParseMarkdown_ImageWidthRoundTrip(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	markdown := `<img src="/api/files/v1/attachments/` + fileID + `" alt="diagram" width="50%">`

	doc, err := parseMarkdown(markdown)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.FileID)
	assert.Equal(t, fileID, *attrs.FileID)
	assert.Empty(t, attrs.Src)
	require.NotNil(t, attrs.Width)
	assert.Equal(t, 50, *attrs.Width)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "diagram", *attrs.Alt)

	rendered, err := richtext.DocumentMarkdown.Render(doc)
	require.NoError(t, err)
	assert.Equal(t, `<img src="/api/files/v1/`+fileID+`" alt="diagram" width="50%">`+"\n", rendered)
}

func TestParseMarkdown_ImageFormattedAltText(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("![**bold** and *italic*](https://example.com/img.png)")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	img := doc.Content[0]
	assert.Equal(t, prosemirror.NodeImage, img.Type)

	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "bold and italic", *attrs.Alt)
}

func TestParseMarkdown_EmptyListItem(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("-\n- item\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeBulletList, list.Type)
	require.Len(t, list.Content, 2)

	empty := list.Content[0]
	assert.Equal(t, prosemirror.NodeListItem, empty.Type)
	require.Len(t, empty.Content, 1)
	assert.Equal(t, prosemirror.NodeParagraph, empty.Content[0].Type)
	assert.Empty(t, empty.Content[0].Content)
}

func TestParseMarkdown_BulletList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- item 1\n- item 2\n- item 3")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeBulletList, list.Type)
	require.Len(t, list.Content, 3)

	for i, item := range list.Content {
		assert.Equal(t, prosemirror.NodeListItem, item.Type, "item %d", i)
		require.Len(t, item.Content, 1)
		assert.Equal(t, prosemirror.NodeParagraph, item.Content[0].Type)
	}
}

func TestParseMarkdown_OrderedList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("1. first\n2. second\n3. third")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeOrderedList, list.Type)

	attrs, err := list.OrderedListAttrs()
	require.NoError(t, err)
	assert.Equal(t, 1, attrs.Start)

	require.Len(t, list.Content, 3)
}

func TestParseMarkdown_OrderedListStartingAtZero(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("0. first\n1. second\n")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeOrderedList, list.Type)

	attrs, err := list.OrderedListAttrs()
	require.NoError(t, err)
	assert.Equal(t, 0, attrs.Start)
}

func TestParseMarkdown_NestedList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- parent\n  - child\n  - child 2\n- parent 2")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeBulletList, list.Type)
	require.Len(t, list.Content, 2)

	// First item should have a paragraph and a nested bullet list.
	firstItem := list.Content[0]
	assert.Equal(t, prosemirror.NodeListItem, firstItem.Type)
	require.Len(t, firstItem.Content, 2)
	assert.Equal(t, prosemirror.NodeParagraph, firstItem.Content[0].Type)
	assert.Equal(t, prosemirror.NodeBulletList, firstItem.Content[1].Type)
	require.Len(t, firstItem.Content[1].Content, 2)
}

func TestParseMarkdown_Blockquote(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("> quoted text")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	bq := doc.Content[0]
	assert.Equal(t, prosemirror.NodeBlockquote, bq.Type)
	require.Len(t, bq.Content, 1)
	assert.Equal(t, prosemirror.NodeParagraph, bq.Content[0].Type)
}

func TestParseMarkdown_HorizontalRule(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("---")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, prosemirror.NodeHorizontalRule, doc.Content[0].Type)
}

func TestParseMarkdown_HardBreak(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("line one\\\nline two")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	assert.Equal(t, prosemirror.NodeParagraph, p.Type)

	// Should contain: text("line one"), hardBreak, text("line two")
	var hasHardBreak bool

	for _, child := range p.Content {
		if child.Type == prosemirror.NodeHardBreak {
			hasHardBreak = true
		}
	}

	assert.True(t, hasHardBreak, "expected hard break node")
}

func TestParseMarkdown_SoftLineBreak(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("line one\nand line two")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Equal(t, prosemirror.NodeParagraph, p.Type)

	var joined strings.Builder

	for _, child := range p.Content {
		if child.Type == prosemirror.NodeText && child.Text != "" {
			joined.WriteString(child.Text)
		}
	}

	assert.Equal(t, "line one and line two", joined.String())
}

func TestParseMarkdown_NestedMarks(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("***bold and italic***")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.Len(t, p.Content, 1)

	txt := p.Content[0]
	assert.Equal(t, "bold and italic", txt.Text)
	require.Len(t, txt.Marks, 2)

	markTypes := make(map[prosemirror.MarkType]bool)
	for _, m := range txt.Marks {
		markTypes[m.Type] = true
	}

	assert.True(t, markTypes[prosemirror.MarkStrong])
	assert.True(t, markTypes[prosemirror.MarkEm])
}

func TestParseMarkdown_MixedContent(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("Normal **bold** and *italic* text")
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)

	p := doc.Content[0]
	require.True(t, len(p.Content) >= 4, "expected at least 4 inline nodes")

	// Verify first text node is plain.
	assert.Equal(t, "Normal ", p.Content[0].Text)
	assert.Empty(t, p.Content[0].Marks)

	// Verify bold text node.
	assert.Equal(t, "bold", p.Content[1].Text)
	require.Len(t, p.Content[1].Marks, 1)
	assert.Equal(t, prosemirror.MarkStrong, p.Content[1].Marks[0].Type)
}

func TestParseMarkdown_MarkdownTable(t *testing.T) {
	t.Parallel()

	md := "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |\n"
	doc, err := parseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, prosemirror.NodeTable, doc.Content[0].Type)

	table := doc.Content[0]
	require.Len(t, table.Content, 3)

	headerRow := table.Content[0]
	assert.Equal(t, prosemirror.NodeTableRow, headerRow.Type)
	require.Len(t, headerRow.Content, 2)
	assert.Equal(t, prosemirror.NodeTableHeader, headerRow.Content[0].Type)
	assert.Equal(t, prosemirror.NodeTableHeader, headerRow.Content[1].Type)

	require.Len(t, headerRow.Content[0].Content, 1)
	require.Equal(t, prosemirror.NodeParagraph, headerRow.Content[0].Content[0].Type)
	require.Len(t, headerRow.Content[0].Content[0].Content, 1)
	assert.Equal(t, "Name", headerRow.Content[0].Content[0].Content[0].Text)

	require.Len(t, headerRow.Content[1].Content, 1)
	require.Equal(t, prosemirror.NodeParagraph, headerRow.Content[1].Content[0].Type)
	require.Len(t, headerRow.Content[1].Content[0].Content, 1)
	assert.Equal(t, "Age", headerRow.Content[1].Content[0].Content[0].Text)

	for i := 1; i <= 2; i++ {
		row := table.Content[i]
		assert.Equal(t, prosemirror.NodeTableRow, row.Type)
		require.Len(t, row.Content, 2)
		assert.Equal(t, prosemirror.NodeTableCell, row.Content[0].Type)
		assert.Equal(t, prosemirror.NodeTableCell, row.Content[1].Type)

		attrs, err := row.Content[0].TableCellAttrs()
		require.NoError(t, err)
		assert.Equal(t, 1, attrs.Colspan)
		assert.Equal(t, 1, attrs.Rowspan)
	}

	assert.Equal(t, "Alice", table.Content[1].Content[0].Content[0].Content[0].Text)
	assert.Equal(t, "30", table.Content[1].Content[1].Content[0].Content[0].Text)
	assert.Equal(t, "Bob", table.Content[2].Content[0].Content[0].Content[0].Text)
	assert.Equal(t, "25", table.Content[2].Content[1].Content[0].Content[0].Text)
}

func TestParseMarkdown_ImageInTableCell(t *testing.T) {
	t.Parallel()

	md := "| Label |\n| --- |\n| before ![diagram](https://example.com/diagram.png) after |\n"
	doc, err := parseMarkdown(md)
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))

	cell := doc.Content[0].Content[1].Content[0]
	assert.Equal(t, prosemirror.NodeTableCell, cell.Type)
	require.Len(t, cell.Content, 3)
	assert.Equal(t, prosemirror.NodeParagraph, cell.Content[0].Type)
	assert.Equal(t, "before ", cell.Content[0].Content[0].Text)
	assert.Equal(t, prosemirror.NodeImage, cell.Content[1].Type)
	assert.Equal(t, prosemirror.NodeParagraph, cell.Content[2].Type)
	assert.Equal(t, " after", cell.Content[2].Content[0].Text)

	attrs, err := cell.Content[1].ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/diagram.png", attrs.Src)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "diagram", *attrs.Alt)
}

func TestParseMarkdown_MarkdownTableWithInlineMarks(t *testing.T) {
	t.Parallel()

	md := "| Header |\n| --- |\n| **bold** and *italic* |\n"
	doc, err := parseMarkdown(md)
	require.NoError(t, err)
	require.Len(t, doc.Content, 1)
	assert.Equal(t, prosemirror.NodeTable, doc.Content[0].Type)

	dataRow := doc.Content[0].Content[1]
	assert.Equal(t, prosemirror.NodeTableRow, dataRow.Type)
	cell := dataRow.Content[0]
	assert.Equal(t, prosemirror.NodeTableCell, cell.Type)

	p := cell.Content[0]
	require.Equal(t, prosemirror.NodeParagraph, p.Type)
	require.GreaterOrEqual(t, len(p.Content), 3)

	assert.Equal(t, "bold", p.Content[0].Text)
	require.Len(t, p.Content[0].Marks, 1)
	assert.Equal(t, prosemirror.MarkStrong, p.Content[0].Marks[0].Type)

	assert.Equal(t, "italic", p.Content[2].Text)
	require.Len(t, p.Content[2].Marks, 1)
	assert.Equal(t, prosemirror.MarkEm, p.Content[2].Marks[0].Type)
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

	doc, err := parseMarkdown(md)
	require.NoError(t, err)

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	// Verify we can parse it back.
	parsed, err := prosemirror.Parse(string(data))
	require.NoError(t, err)
	assert.Equal(t, prosemirror.NodeDoc, parsed.Type)
	assert.True(t, len(parsed.Content) > 0)
}

func TestParseMarkdown_TaskList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- [ ] open\n- [x] done\n- [X] also\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)

	list := doc.Content[0]
	assert.Equal(t, prosemirror.NodeTaskList, list.Type)
	require.Len(t, list.Content, 3)

	for i, want := range []bool{false, true, true} {
		item := list.Content[i]
		assert.Equal(t, prosemirror.NodeTaskItem, item.Type)

		attrs, err := item.TaskItemAttrs()
		require.NoError(t, err)
		assert.Equal(t, want, attrs.Checked)
		require.Len(t, item.Content, 1)
		assert.Equal(t, prosemirror.NodeParagraph, item.Content[0].Type)
	}

	rendered, err := richtext.DocumentMarkdown.Render(doc)
	require.NoError(t, err)
	assert.Equal(t, "- [ ] open\n- [x] done\n- [x] also\n", rendered)
}

func TestParseMarkdown_TaskListWithMark(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- [ ] **bold** item")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)

	paragraph := doc.Content[0].Content[0].Content[0]
	require.GreaterOrEqual(t, len(paragraph.Content), 1)
	assert.Equal(t, "bold", paragraph.Content[0].Text)
	require.Len(t, paragraph.Content[0].Marks, 1)
	assert.Equal(t, prosemirror.MarkStrong, paragraph.Content[0].Marks[0].Type)
}

func TestParseMarkdown_NestedTaskList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- [ ] parent\n  - [x] child\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)

	parent := doc.Content[0].Content[0]
	assert.Equal(t, prosemirror.NodeTaskItem, parent.Type)
	require.Len(t, parent.Content, 2)
	assert.Equal(t, prosemirror.NodeTaskList, parent.Content[1].Type)

	child := parent.Content[1].Content[0]
	attrs, err := child.TaskItemAttrs()
	require.NoError(t, err)
	assert.True(t, attrs.Checked)
}

func TestParseMarkdown_MixedTaskList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("- plain\n- [ ] task\n- after\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 3)
	assert.Equal(t, prosemirror.NodeBulletList, doc.Content[0].Type)
	assert.Equal(t, prosemirror.NodeTaskList, doc.Content[1].Type)
	assert.Equal(t, prosemirror.NodeBulletList, doc.Content[2].Type)

	rendered, err := richtext.DocumentMarkdown.Render(doc)
	require.NoError(t, err)
	assert.Equal(t, "- plain\n\n- [ ] task\n\n- after\n", rendered)
}

func TestParseMarkdown_OrderedTaskList(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("1. [ ] first\n2. [x] second\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)
	assert.Equal(t, prosemirror.NodeTaskList, doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 2)
}

func TestParseMarkdown_TaskListHTML(t *testing.T) {
	t.Parallel()

	doc, err := parseMarkdown("<ul><li><input type=\"checkbox\"> open</li><li><input checked type=\"checkbox\"> done</li></ul>\n")
	require.NoError(t, err)
	require.NoError(t, editorSchema.Validate(doc))
	require.Len(t, doc.Content, 1)
	assert.Equal(t, prosemirror.NodeTaskList, doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 2)

	open, err := doc.Content[0].Content[0].TaskItemAttrs()
	require.NoError(t, err)
	assert.False(t, open.Checked)

	done, err := doc.Content[0].Content[1].TaskItemAttrs()
	require.NoError(t, err)
	assert.True(t, done.Checked)
	require.NotEmpty(t, doc.Content[0].Content[0].Content[0].Content[0].Text)
	assert.Equal(t, "open", doc.Content[0].Content[0].Content[0].Content[0].Text)
	require.NotEmpty(t, doc.Content[0].Content[1].Content[0].Content[0].Text)
	assert.Equal(t, "done", doc.Content[0].Content[1].Content[0].Content[0].Text)
}
