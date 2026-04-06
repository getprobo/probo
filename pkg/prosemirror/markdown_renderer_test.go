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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdown_Document(t *testing.T) {
	t.Parallel()

	expected, err := os.ReadFile("testdata/document.md")
	require.NoError(t, err)

	doc := loadTestDocument(t)
	got, err := RenderMarkdown(doc)
	require.NoError(t, err)
	assert.Equal(t, string(expected), got)
}

func TestRenderMarkdown_EmptyDoc(t *testing.T) {
	t.Parallel()

	got, err := RenderMarkdown(Node{Type: NodeDoc})
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRenderMarkdown_EmptyParagraph(t *testing.T) {
	t.Parallel()

	got, err := RenderMarkdown(Node{
		Type:    NodeDoc,
		Content: []Node{{Type: NodeParagraph}},
	})
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestRenderMarkdown_Paragraph(t *testing.T) {
	t.Parallel()

	text := "Hello world"
	got, err := RenderMarkdown(Node{
		Type: NodeDoc,
		Content: []Node{
			{
				Type: NodeParagraph,
				Content: []Node{
					{Type: NodeText, Text: &text},
				},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello world\n", got)
}

func TestRenderMarkdown_HeadingLevels(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		level int
		want  string
	}{
		{1, "# X\n"},
		{2, "## X\n"},
		{3, "### X\n"},
		{4, "#### X\n"},
		{5, "##### X\n"},
		{6, "###### X\n"},
	} {
		t.Run(
			"level "+string(rune('0'+tc.level)),
			func(t *testing.T) {
				t.Parallel()
				raw := `{"type":"doc","content":[{"type":"heading","attrs":{"level":` + string(rune('0'+tc.level)) + `},"content":[{"type":"text","text":"X"}]}]}`
				var n Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := RenderMarkdown(n)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			},
		)
	}
}

func TestRenderMarkdown_HeadingInvalidLevel(t *testing.T) {
	t.Parallel()

	raw := `{"type":"heading","attrs":{"level":7},"content":[{"type":"text","text":"X"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	_, err := RenderMarkdown(n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestRenderMarkdown_Bold(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"bold"}],"text":"bold"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "**bold**\n", got)
}

func TestRenderMarkdown_Italic(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"italic"}],"text":"italic"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "*italic*\n", got)
}

func TestRenderMarkdown_ItalicTrailingSpace(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"italic"}],"text":"italic "},{"type":"text","text":"rest"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "*italic* rest\n", got)
}

func TestRenderMarkdown_Strikethrough(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"strike"}],"text":"deleted"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "~~deleted~~\n", got)
}

func TestRenderMarkdown_Underline(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"underlined"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "<u>underlined</u>\n", got)
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"code"}],"text":"code"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "`code`\n", got)
}

func TestRenderMarkdown_InlineCodeWithBacktick(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"code"}],"text":"a ` + "`" + ` b"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "`` a ` b ``\n", got)
}

func TestRenderMarkdown_InlineCodeWithDoubleBacktickRun(t *testing.T) {
	t.Parallel()

	// Two consecutive backticks in content need a 3+ backtick fence.
	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"code"}],"text":"` + "``" + `x"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "``` ``x ```\n", got)
}

func TestRenderMarkdown_InlineCodeWithTripleBacktickRun(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"code"}],"text":"` + "```" + `"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "```` ``` ````\n", got)
}

func TestRenderMarkdown_Link(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":null,"rel":null,"class":null,"title":null}}],"text":"click"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "[click](https://example.com)\n", got)
}

func TestRenderMarkdown_LinkWithTitle(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":null,"rel":null,"class":null,"title":"My Title"}}],"text":"click"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "[click](https://example.com \"My Title\")\n", got)
}

func TestRenderMarkdown_LinkSanitizesDangerousHrefs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		href     string
		wantHref string
	}{
		{name: "javascript scheme", href: `javascript:alert(1)`, wantHref: `#`},
		{name: "data html", href: `data:text/html,<script>alert(1)</script>`, wantHref: `#`},
		{name: "protocol-relative", href: `//evil.example/phish`, wantHref: `#`},
		{name: "empty href", href: ``, wantHref: `#`},
		{name: "https preserved", href: `https://example.com/x`, wantHref: `https://example.com/x`},
		{name: "mailto", href: `mailto:user@example.com`, wantHref: `mailto:user@example.com`},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				hrefJSON, err := json.Marshal(tc.href)
				require.NoError(t, err)
				raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":` + string(hrefJSON) + `,"target":null,"rel":null,"class":null,"title":null}}],"text":"x"}]}]}`
				var n Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := RenderMarkdown(n)
				require.NoError(t, err)
				assert.Equal(t, "[x]("+tc.wantHref+")\n", got)
			},
		)
	}
}

func TestRenderMarkdown_MultipleMarks(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"bold"},{"type":"italic"}],"text":"hello"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "***hello***\n", got)
}

func TestRenderMarkdown_CodeBlockWithLanguage(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println()"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "```go\nfmt.Println()\n```\n", got)
}

func TestRenderMarkdown_CodeBlockWithoutLanguage(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"codeBlock","attrs":{"language":null},"content":[{"type":"text","text":"hello"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "```\nhello\n```\n", got)
}

func TestRenderMarkdown_CodeBlockWithTripleBackticks(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"codeBlock","attrs":{"language":null},"content":[{"type":"text","text":"` + "```" + `\nsome code"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Contains(t, got, "````")
	assert.Contains(t, got, "```\nsome code")
}

func TestRenderMarkdown_HorizontalRule(t *testing.T) {
	t.Parallel()

	got, err := RenderMarkdown(Node{
		Type: NodeDoc,
		Content: []Node{
			{Type: NodeHorizontalRule},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "---\n", got)
}

func TestRenderMarkdown_HardBreak(t *testing.T) {
	t.Parallel()

	line1 := "line one"
	line2 := "line two"
	got, err := RenderMarkdown(Node{
		Type: NodeDoc,
		Content: []Node{
			{
				Type: NodeParagraph,
				Content: []Node{
					{Type: NodeText, Text: &line1},
					{Type: NodeHardBreak},
					{Type: NodeText, Text: &line2},
				},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "line one\\\nline two\n", got)
}

func TestRenderMarkdown_Image(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":"My image"}}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "![A photo](https://example.com/img.png \"My image\")\n", got)
}

func TestRenderMarkdown_ImageWithoutTitle(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":null}}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "![A photo](https://example.com/img.png)\n", got)
}

func TestRenderMarkdown_ImageSanitizesDangerousSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"javascript:alert(1)","alt":null,"title":null}}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "![]()\n", got)
}

func TestRenderMarkdown_BulletList(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"one"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"two"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"three"}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "- one\n- two\n- three\n", got)
}

func TestRenderMarkdown_OrderedList(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"orderedList","attrs":{"start":1,"type":null},"content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"first"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"second"}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "1. first\n2. second\n", got)
}

func TestRenderMarkdown_OrderedListWithStart(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"orderedList","attrs":{"start":5,"type":null},"content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"item"}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "5. item\n", got)
}

func TestRenderMarkdown_NestedList(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"parent"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"child"}]}]}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "- parent\n  \n  - child\n", got)
}

func TestRenderMarkdown_Blockquote(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"quoted"}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "> quoted\n", got)
}

func TestRenderMarkdown_BlockquoteMultipleParagraphs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"first"}]},{"type":"paragraph","content":[{"type":"text","text":"second"}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "> first\n> \n> second\n", got)
}

func TestRenderMarkdown_GFMTable(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"Name"}]}]},{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"Age"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"Alice"}]}]},{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"30"}]}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n", got)
}

func TestRenderMarkdown_TableWithBlockContent(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"Header"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"item"}]}]}]}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "| Header |\n| --- |\n| <ul><li><p>item</p></li></ul> |\n", got)
}

func TestRenderMarkdown_MarkdownEscaping(t *testing.T) {
	t.Parallel()

	text := `*bold* _italic_ ` + "`code`" + ` [link] ~strike~ |pipe| <html>`
	got, err := RenderMarkdown(Node{
		Type: NodeDoc,
		Content: []Node{
			{
				Type: NodeParagraph,
				Content: []Node{
					{Type: NodeText, Text: &text},
				},
			},
		},
	})
	require.NoError(t, err)
	assert.NotContains(t, got, "*bold*")
	assert.Contains(t, got, `\*bold\*`)
	assert.Contains(t, got, `\_italic\_`)
	assert.Contains(t, got, "\\`code\\`")
	assert.Contains(t, got, `\[link\]`)
	assert.Contains(t, got, `\~strike\~`)
	assert.Contains(t, got, `\|pipe\|`)
	assert.Contains(t, got, `\<html>`)
}

func TestRenderMarkdown_UnknownNodeType(t *testing.T) {
	t.Parallel()

	node := Node{Type: NodeType("unknownWidget")}
	_, err := RenderMarkdown(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestRenderMarkdown_UnknownMarkType(t *testing.T) {
	t.Parallel()

	text := "hello"
	node := Node{
		Type: NodeDoc,
		Content: []Node{
			{
				Type: NodeParagraph,
				Content: []Node{
					{
						Type: NodeText,
						Text: &text,
						Marks: []Mark{
							{Type: MarkType("superscript")},
						},
					},
				},
			},
		},
	}
	_, err := RenderMarkdown(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestRenderMarkdown_MixedContent(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Normal "},{"type":"text","marks":[{"type":"bold"}],"text":"bold"},{"type":"text","text":" and "},{"type":"text","marks":[{"type":"italic"}],"text":"italic"},{"type":"text","text":" text"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "Normal **bold** and *italic* text\n", got)
}

func TestRenderMarkdown_BlockquoteWithHardBreak(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"line one"},{"type":"hardBreak"},{"type":"text","text":"line two"}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "> line one\\\n> line two\n", got)
}

func TestRenderMarkdown_GFMTableWithMarks(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"Header"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"bold"}],"text":"bold"},{"type":"text","text":" and "},{"type":"text","marks":[{"type":"italic"}],"text":"italic"}]}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "| Header |\n| --- |\n| **bold** and *italic* |\n", got)
}

func TestRenderMarkdown_TextNil(t *testing.T) {
	t.Parallel()

	node := Node{
		Type: NodeDoc,
		Content: []Node{
			{
				Type: NodeParagraph,
				Content: []Node{
					{Type: NodeText, Text: nil},
				},
			},
		},
	}
	_, err := RenderMarkdown(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text is nil")
}

func TestRenderMarkdown_CodeBlockInBlockquote(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println()"}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Equal(t, "> ```go\n> fmt.Println()\n> ```\n", got)
}

func TestRenderMarkdown_ListItemOutsideList(t *testing.T) {
	t.Parallel()

	_, err := RenderMarkdown(Node{Type: NodeListItem})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list item outside of list context")
}

func TestRenderMarkdown_TableRowOutsideTable(t *testing.T) {
	t.Parallel()

	_, err := RenderMarkdown(Node{Type: NodeTableRow})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside of table context")
}

func TestRenderMarkdown_TableBlockCellEscapesPipes(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"H"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":null},"content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"a | b"}]}]}]}]}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderMarkdown(n)
	require.NoError(t, err)
	assert.Contains(t, got, `a \| b`)
	assert.NotContains(t, got, "<table>")
}
