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

func loadTestDocument(t *testing.T) Node {
	t.Helper()

	data, err := os.ReadFile("testdata/document.json")
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal(data, &doc))
	return doc
}

func TestUnmarshalDocument(t *testing.T) {
	t.Parallel()

	doc := loadTestDocument(t)

	assert.Equal(t, NodeDoc, doc.Type)
	require.Len(t, doc.Content, 14)

	t.Run(
		"heading level 1",
		func(t *testing.T) {
			t.Parallel()
			h1 := doc.Content[0]
			assert.Equal(t, NodeHeading, h1.Type)

			attrs, err := h1.HeadingAttrs()
			require.NoError(t, err)
			assert.Equal(t, 1, attrs.Level)

			require.Len(t, h1.Content, 1)
			assert.Equal(t, NodeText, h1.Content[0].Type)
			require.NotNil(t, h1.Content[0].Text)
			assert.Equal(t, "Heading 1", *h1.Content[0].Text)
		},
	)

	t.Run(
		"paragraph with mixed marks",
		func(t *testing.T) {
			t.Parallel()
			p := doc.Content[1]
			assert.Equal(t, NodeParagraph, p.Type)
			require.True(t, len(p.Content) > 5)

			// Bold text
			boldNode := p.Content[1]
			require.Len(t, boldNode.Marks, 1)
			assert.Equal(t, MarkStrong, boldNode.Marks[0].Type)
			require.NotNil(t, boldNode.Text)
			assert.Equal(t, "with some bold", *boldNode.Text)

			// Italic text
			italicNode := p.Content[3]
			require.Len(t, italicNode.Marks, 1)
			assert.Equal(t, MarkEm, italicNode.Marks[0].Type)

			// Underline text
			underlineNode := p.Content[5]
			require.Len(t, underlineNode.Marks, 1)
			assert.Equal(t, MarkUnderline, underlineNode.Marks[0].Type)

			// Hard break
			assert.Equal(t, NodeHardBreak, p.Content[7].Type)

			// Strikethrough
			strikeNode := p.Content[9]
			require.Len(t, strikeNode.Marks, 1)
			assert.Equal(t, MarkStrike, strikeNode.Marks[0].Type)

			// Inline code
			codeNode := p.Content[12]
			require.Len(t, codeNode.Marks, 1)
			assert.Equal(t, MarkCode, codeNode.Marks[0].Type)

			// Link
			linkNode := p.Content[14]
			require.Len(t, linkNode.Marks, 1)
			assert.Equal(t, MarkLink, linkNode.Marks[0].Type)

			linkAttrs, err := linkNode.Marks[0].LinkAttrs()
			require.NoError(t, err)
			assert.Equal(t, "https://getprobo.com", linkAttrs.Href)
			require.NotNil(t, linkAttrs.Target)
			assert.Equal(t, "_blank", *linkAttrs.Target)
			require.NotNil(t, linkAttrs.Rel)
			assert.Equal(t, "noopener noreferrer nofollow", *linkAttrs.Rel)
			assert.Nil(t, linkAttrs.Class)
			assert.Nil(t, linkAttrs.Title)
		},
	)

	t.Run(
		"heading level 2",
		func(t *testing.T) {
			t.Parallel()
			h2 := doc.Content[2]
			assert.Equal(t, NodeHeading, h2.Type)

			attrs, err := h2.HeadingAttrs()
			require.NoError(t, err)
			assert.Equal(t, 2, attrs.Level)
		},
	)

	t.Run(
		"code block",
		func(t *testing.T) {
			t.Parallel()
			cb := doc.Content[4]
			assert.Equal(t, NodeCodeBlock, cb.Type)

			attrs, err := cb.CodeBlockAttrs()
			require.NoError(t, err)
			assert.Nil(t, attrs.Language)

			require.Len(t, cb.Content, 1)
			require.NotNil(t, cb.Content[0].Text)
			assert.Equal(t, "code block", *cb.Content[0].Text)
		},
	)

	t.Run(
		"heading level 3",
		func(t *testing.T) {
			t.Parallel()
			h3 := doc.Content[5]
			attrs, err := h3.HeadingAttrs()
			require.NoError(t, err)
			assert.Equal(t, 3, attrs.Level)
		},
	)

	t.Run(
		"bullet list",
		func(t *testing.T) {
			t.Parallel()
			bl := doc.Content[6]
			assert.Equal(t, NodeBulletList, bl.Type)
			require.Len(t, bl.Content, 3)

			for _, item := range bl.Content {
				assert.Equal(t, NodeListItem, item.Type)
				require.Len(t, item.Content, 1)
				assert.Equal(t, NodeParagraph, item.Content[0].Type)
			}
		},
	)

	t.Run(
		"horizontal rule",
		func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, NodeHorizontalRule, doc.Content[7].Type)
			assert.Equal(t, NodeHorizontalRule, doc.Content[9].Type)
			assert.Equal(t, NodeHorizontalRule, doc.Content[11].Type)
		},
	)

	t.Run(
		"ordered list",
		func(t *testing.T) {
			t.Parallel()
			ol := doc.Content[8]
			assert.Equal(t, NodeOrderedList, ol.Type)
			require.Len(t, ol.Content, 3)

			attrs, err := ol.OrderedListAttrs()
			require.NoError(t, err)
			assert.Equal(t, 1, attrs.Start)
			assert.Nil(t, attrs.Type)
		},
	)

	t.Run(
		"blockquote",
		func(t *testing.T) {
			t.Parallel()
			bq := doc.Content[10]
			assert.Equal(t, NodeBlockquote, bq.Type)
			require.Len(t, bq.Content, 1)
			assert.Equal(t, NodeParagraph, bq.Content[0].Type)

			// Verify hard break inside blockquote
			bqPara := bq.Content[0]
			require.True(t, len(bqPara.Content) >= 3)
			assert.Equal(t, NodeHardBreak, bqPara.Content[1].Type)
		},
	)

	t.Run(
		"table",
		func(t *testing.T) {
			t.Parallel()
			table := doc.Content[12]
			assert.Equal(t, NodeTable, table.Type)
			require.Len(t, table.Content, 3)

			// Header row
			headerRow := table.Content[0]
			assert.Equal(t, NodeTableRow, headerRow.Type)
			require.Len(t, headerRow.Content, 4)
			for _, cell := range headerRow.Content {
				assert.Equal(t, NodeTableHeader, cell.Type)
			}

			// Last header has colwidth
			th4 := headerRow.Content[3]
			thAttrs, err := th4.TableCellAttrs()
			require.NoError(t, err)
			assert.Equal(t, 1, thAttrs.Colspan)
			assert.Equal(t, 1, thAttrs.Rowspan)
			assert.Equal(t, []int{61}, thAttrs.Colwidth)

			// First header has null colwidth
			th1 := headerRow.Content[0]
			th1Attrs, err := th1.TableCellAttrs()
			require.NoError(t, err)
			assert.Nil(t, th1Attrs.Colwidth)

			// Data rows
			for _, row := range table.Content[1:] {
				assert.Equal(t, NodeTableRow, row.Type)
				require.Len(t, row.Content, 4)
				for _, cell := range row.Content {
					assert.Equal(t, NodeTableCell, cell.Type)
				}
			}

			// Nested bullet list in table cell
			nestedBL := table.Content[1].Content[3]
			require.Len(t, nestedBL.Content, 1)
			assert.Equal(t, NodeBulletList, nestedBL.Content[0].Type)

			// Nested ordered list in table cell
			nestedOL := table.Content[2].Content[3]
			require.Len(t, nestedOL.Content, 1)
			assert.Equal(t, NodeOrderedList, nestedOL.Content[0].Type)
		},
	)

	t.Run(
		"trailing empty paragraph",
		func(t *testing.T) {
			t.Parallel()
			emptyP := doc.Content[13]
			assert.Equal(t, NodeParagraph, emptyP.Type)
			assert.Empty(t, emptyP.Content)
		},
	)
}

func TestMarshalRoundtrip(t *testing.T) {
	t.Parallel()

	doc := loadTestDocument(t)

	marshaled, err := json.Marshal(doc)
	require.NoError(t, err)

	var roundtripped Node
	require.NoError(t, json.Unmarshal(marshaled, &roundtripped))

	// Re-marshal both to compact JSON for comparison, since the original
	// testdata file is pretty-printed and RawMessage preserves whitespace.
	expected, err := json.Marshal(roundtripped)
	require.NoError(t, err)
	assert.JSONEq(t, string(marshaled), string(expected))
}

func TestHeadingAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"heading","attrs":{"level":3},"content":[{"type":"text","text":"Hello"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	attrs, err := n.HeadingAttrs()
	require.NoError(t, err)
	assert.Equal(t, 3, attrs.Level)
}

func TestCodeBlockAttrs(t *testing.T) {
	t.Parallel()

	t.Run(
		"with language",
		func(t *testing.T) {
			t.Parallel()
			raw := `{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println()"}]}`
			var n Node
			require.NoError(t, json.Unmarshal([]byte(raw), &n))

			attrs, err := n.CodeBlockAttrs()
			require.NoError(t, err)
			require.NotNil(t, attrs.Language)
			assert.Equal(t, "go", *attrs.Language)
		},
	)

	t.Run(
		"with null language",
		func(t *testing.T) {
			t.Parallel()
			raw := `{"type":"codeBlock","attrs":{"language":null}}`
			var n Node
			require.NoError(t, json.Unmarshal([]byte(raw), &n))

			attrs, err := n.CodeBlockAttrs()
			require.NoError(t, err)
			assert.Nil(t, attrs.Language)
		},
	)
}

func TestOrderedListAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"orderedList","attrs":{"start":5,"type":null}}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	attrs, err := n.OrderedListAttrs()
	require.NoError(t, err)
	assert.Equal(t, 5, attrs.Start)
	assert.Nil(t, attrs.Type)
}

func TestImageAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"An image","title":null}}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	attrs, err := n.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "An image", *attrs.Alt)
	assert.Nil(t, attrs.Title)
}

func TestTableCellAttrs(t *testing.T) {
	t.Parallel()

	t.Run(
		"with colwidth",
		func(t *testing.T) {
			t.Parallel()
			raw := `{"type":"tableCell","attrs":{"colspan":2,"rowspan":1,"colwidth":[100,200]}}`
			var n Node
			require.NoError(t, json.Unmarshal([]byte(raw), &n))

			attrs, err := n.TableCellAttrs()
			require.NoError(t, err)
			assert.Equal(t, 2, attrs.Colspan)
			assert.Equal(t, 1, attrs.Rowspan)
			assert.Equal(t, []int{100, 200}, attrs.Colwidth)
		},
	)

	t.Run(
		"with null colwidth",
		func(t *testing.T) {
			t.Parallel()
			raw := `{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1,"colwidth":null}}`
			var n Node
			require.NoError(t, json.Unmarshal([]byte(raw), &n))

			attrs, err := n.TableCellAttrs()
			require.NoError(t, err)
			assert.Equal(t, 1, attrs.Colspan)
			assert.Equal(t, 1, attrs.Rowspan)
			assert.Nil(t, attrs.Colwidth)
		},
	)
}

func TestLinkAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":"noopener","class":null,"title":"Example"}}`
	var m Mark
	require.NoError(t, json.Unmarshal([]byte(raw), &m))

	attrs, err := m.LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", attrs.Href)
	require.NotNil(t, attrs.Target)
	assert.Equal(t, "_blank", *attrs.Target)
	require.NotNil(t, attrs.Rel)
	assert.Equal(t, "noopener", *attrs.Rel)
	assert.Nil(t, attrs.Class)
	require.NotNil(t, attrs.Title)
	assert.Equal(t, "Example", *attrs.Title)
}

func TestNodeWithNoAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"paragraph","content":[{"type":"text","text":"Hello"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	assert.Equal(t, NodeParagraph, n.Type)
	assert.Nil(t, n.Attrs)
	require.Len(t, n.Content, 1)
	require.NotNil(t, n.Content[0].Text)
	assert.Equal(t, "Hello", *n.Content[0].Text)
}
