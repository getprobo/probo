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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// ParseMarkdown converts a markdown string into a ProseMirror Node tree.
func ParseMarkdown(markdown string) (Node, error) {
	source := []byte(markdown)

	md := goldmark.New(
		goldmark.WithExtensions(
			goldmarkext.Strikethrough,
			goldmarkext.Table,
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	doc := md.Parser().Parse(text.NewReader(source))

	c := &converter{source: source}

	nodes, err := c.convertChildren(doc)
	if err != nil {
		return Node{}, fmt.Errorf("cannot convert markdown to prosemirror: %w", err)
	}

	return Node{
		Type:    NodeDoc,
		Content: nodes,
	}, nil
}

// normalizeCodeBlockContent strips one trailing newline when it is only the
// line terminator goldmark attaches to the last content line. That avoids an
// extra visible blank line in editors (e.g. TipTap) while preserving a
// trailing blank line in the source, which ends with two newlines.
func normalizeCodeBlockContent(content string) string {
	if content == "" {
		return content
	}
	if strings.HasSuffix(content, "\n") && !strings.HasSuffix(content, "\n\n") {
		return strings.TrimSuffix(content, "\n")
	}
	return content
}

type converter struct {
	source []byte
	marks  []Mark
}

func (c *converter) convertChildren(n ast.Node) ([]Node, error) {
	var nodes []Node

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		converted, err := c.convertNode(child)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, converted...)
	}

	return nodes, nil
}

// convertInlineChildren walks inline content and merges adjacent RawHTML + Text
// segments so tags split by goldmark (e.g. <strong>, b, </strong>) form one HTML
// fragment for sanitization and conversion.
func (c *converter) convertInlineChildren(n ast.Node) ([]Node, error) {
	var nodes []Node

	for ch := n.FirstChild(); ch != nil; {
		if ch.Kind() == ast.KindRawHTML {
			run, next := c.collectRawHTMLRun(ch)
			inodes, err := convertProseMirrorFromInlineHTML(run)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, prependOuterMarks(copyMarks(c.marks), inodes)...)
			ch = next
			continue
		}

		converted, err := c.convertNode(ch)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, converted...)
		ch = ch.NextSibling()
	}

	return nodes, nil
}

// collectRawHTMLRun concatenates a leading RawHTML node and following Text/String
// and RawHTML siblings until a different node kind is seen. next is the first
// sibling not consumed (or nil).
func (c *converter) collectRawHTMLRun(start ast.Node) (run string, next ast.Node) {
	var buf bytes.Buffer
	ch := start
	for ch != nil {
		switch ch.Kind() {
		case ast.KindRawHTML:
			raw := ch.(*ast.RawHTML)
			for i := 0; i < raw.Segments.Len(); i++ {
				seg := raw.Segments.At(i)
				buf.Write(seg.Value(c.source))
			}
		case ast.KindText:
			t := ch.(*ast.Text)
			buf.Write(t.Segment.Value(c.source))
			if t.SoftLineBreak() {
				buf.WriteByte(' ')
			}
		case ast.KindString:
			buf.Write(ch.(*ast.String).Value)
		default:
			return buf.String(), ch
		}
		ch = ch.NextSibling()
	}
	return buf.String(), nil
}

func (c *converter) convertNode(n ast.Node) ([]Node, error) {
	switch n.Kind() {
	case ast.KindHeading:
		return c.convertHeading(n.(*ast.Heading))
	case ast.KindParagraph:
		return c.convertParagraph(n)
	case ast.KindBlockquote:
		return c.convertBlockquote(n)
	case ast.KindFencedCodeBlock:
		return c.convertFencedCodeBlock(n.(*ast.FencedCodeBlock))
	case ast.KindCodeBlock:
		return c.convertCodeBlock(n.(*ast.CodeBlock))
	case ast.KindList:
		return c.convertList(n.(*ast.List))
	case ast.KindListItem:
		return c.convertListItem(n)
	case ast.KindThematicBreak:
		return []Node{{Type: NodeHorizontalRule}}, nil
	case ast.KindImage:
		return c.convertImage(n.(*ast.Image))
	case ast.KindTextBlock:
		return c.convertParagraph(n)
	case ast.KindText:
		return c.convertText(n.(*ast.Text))
	case ast.KindString:
		return c.convertString(n.(*ast.String))
	case ast.KindEmphasis:
		return c.convertEmphasis(n.(*ast.Emphasis))
	case ast.KindCodeSpan:
		return c.convertCodeSpan(n)
	case ast.KindLink:
		return c.convertLink(n.(*ast.Link))
	case ast.KindAutoLink:
		return c.convertAutoLink(n.(*ast.AutoLink))
	case ast.KindRawHTML:
		return c.convertRawHTML(n)
	case ast.KindHTMLBlock:
		return c.convertHTMLBlock(n.(*ast.HTMLBlock))
	default:
		switch n.Kind() {
		case goldmarkast.KindStrikethrough:
			return c.convertStrikethrough(n)
		case goldmarkast.KindTable:
			return c.convertTable(n)
		case goldmarkast.KindTableHeader:
			return c.convertTableHeaderRow(n.(*goldmarkast.TableHeader))
		case goldmarkast.KindTableRow:
			return c.convertTableDataRow(n)
		case goldmarkast.KindTableCell:
			return nil, fmt.Errorf("cannot convert table cell outside of a table row")
		}
		return nil, fmt.Errorf("cannot convert markdown node of kind %s", n.Kind())
	}
}

func (c *converter) convertHeading(n *ast.Heading) ([]Node, error) {
	children, err := c.convertInlineChildren(n)
	if err != nil {
		return nil, err
	}

	attrs, err := json.Marshal(HeadingAttrs{Level: n.Level})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal heading attrs: %w", err)
	}

	return []Node{{
		Type:    NodeHeading,
		Content: children,
		Attrs:   attrs,
	}}, nil
}

func (c *converter) convertParagraph(n ast.Node) ([]Node, error) {
	children, err := c.convertInlineChildren(n)
	if err != nil {
		return nil, err
	}

	return []Node{{
		Type:    NodeParagraph,
		Content: children,
	}}, nil
}

func (c *converter) convertBlockquote(n ast.Node) ([]Node, error) {
	children, err := c.convertChildren(n)
	if err != nil {
		return nil, err
	}

	return []Node{{
		Type:    NodeBlockquote,
		Content: children,
	}}, nil
}

func (c *converter) convertFencedCodeBlock(n *ast.FencedCodeBlock) ([]Node, error) {
	var buf bytes.Buffer

	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(c.source))
	}

	content := normalizeCodeBlockContent(buf.String())

	var lang *string
	if n.Language(c.source) != nil {
		l := string(n.Language(c.source))
		lang = &l
	}

	attrs, err := json.Marshal(CodeBlockAttrs{Language: lang})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []Node
	if content != "" {
		textNodes = []Node{{
			Type: NodeText,
			Text: &content,
		}}
	}

	return []Node{{
		Type:    NodeCodeBlock,
		Content: textNodes,
		Attrs:   attrs,
	}}, nil
}

func (c *converter) convertCodeBlock(n *ast.CodeBlock) ([]Node, error) {
	var buf bytes.Buffer

	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(c.source))
	}

	content := normalizeCodeBlockContent(buf.String())

	attrs, err := json.Marshal(CodeBlockAttrs{Language: nil})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []Node
	if content != "" {
		textNodes = []Node{{
			Type: NodeText,
			Text: &content,
		}}
	}

	return []Node{{
		Type:    NodeCodeBlock,
		Content: textNodes,
		Attrs:   attrs,
	}}, nil
}

func (c *converter) convertList(n *ast.List) ([]Node, error) {
	children, err := c.convertChildren(n)
	if err != nil {
		return nil, err
	}

	if n.IsOrdered() {
		attrs, err := json.Marshal(OrderedListAttrs{Start: n.Start})
		if err != nil {
			return nil, fmt.Errorf("cannot marshal ordered list attrs: %w", err)
		}
		return []Node{{
			Type:    NodeOrderedList,
			Content: children,
			Attrs:   attrs,
		}}, nil
	}

	return []Node{{
		Type:    NodeBulletList,
		Content: children,
	}}, nil
}

func (c *converter) convertListItem(n ast.Node) ([]Node, error) {
	children, err := c.convertChildren(n)
	if err != nil {
		return nil, err
	}

	return []Node{{
		Type:    NodeListItem,
		Content: children,
	}}, nil
}

func (c *converter) convertImage(n *ast.Image) ([]Node, error) {
	imgAttrs := ImageAttrs{
		Src: string(n.Destination),
	}

	if n.Title != nil {
		t := string(n.Title)
		imgAttrs.Title = &t
	}

	if alt := c.extractText(n); alt != "" {
		imgAttrs.Alt = &alt
	}

	attrs, err := json.Marshal(imgAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal image attrs: %w", err)
	}

	return []Node{{
		Type:  NodeImage,
		Attrs: attrs,
	}}, nil
}

func (c *converter) convertText(n *ast.Text) ([]Node, error) {
	content := string(n.Segment.Value(c.source))
	if n.SoftLineBreak() {
		content += " "
	}
	if content == "" {
		return nil, nil
	}

	nodes := []Node{{
		Type:  NodeText,
		Text:  &content,
		Marks: copyMarks(c.marks),
	}}

	if n.HardLineBreak() {
		nodes = append(nodes, Node{Type: NodeHardBreak})
	}

	return nodes, nil
}

func (c *converter) convertString(n *ast.String) ([]Node, error) {
	content := string(n.Value)
	if content == "" {
		return nil, nil
	}

	return []Node{{
		Type:  NodeText,
		Text:  &content,
		Marks: copyMarks(c.marks),
	}}, nil
}

func (c *converter) convertEmphasis(n *ast.Emphasis) ([]Node, error) {
	var mark Mark
	if n.Level == 2 {
		mark = Mark{Type: MarkStrong}
	} else {
		mark = Mark{Type: MarkEm}
	}

	c.marks = append(c.marks, mark)
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertCodeSpan(n ast.Node) ([]Node, error) {
	var buf bytes.Buffer

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			buf.Write(t.Segment.Value(c.source))
		}
	}

	content := buf.String()
	if content == "" {
		return nil, nil
	}

	marks := copyMarks(c.marks)
	marks = append(marks, Mark{Type: MarkCode})

	return []Node{{
		Type:  NodeText,
		Text:  &content,
		Marks: marks,
	}}, nil
}

func (c *converter) convertLink(n *ast.Link) ([]Node, error) {
	linkAttrs := LinkAttrs{
		Href: safeLinkHref(string(n.Destination)),
	}

	if n.Title != nil {
		t := string(n.Title)
		linkAttrs.Title = &t
	}

	attrs, err := json.Marshal(linkAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}

	c.marks = append(c.marks, Mark{Type: MarkLink, Attrs: attrs})
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertAutoLink(n *ast.AutoLink) ([]Node, error) {
	url := string(n.URL(c.source))

	linkAttrs := LinkAttrs{Href: safeLinkHref(url)}
	attrs, err := json.Marshal(linkAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}

	label := string(n.Label(c.source))

	return []Node{{
		Type:  NodeText,
		Text:  &label,
		Marks: append(copyMarks(c.marks), Mark{Type: MarkLink, Attrs: attrs}),
	}}, nil
}

func (c *converter) convertRawHTML(n ast.Node) ([]Node, error) {
	raw, ok := n.(*ast.RawHTML)
	if !ok {
		return nil, fmt.Errorf("cannot convert raw html: unexpected node type %T", n)
	}
	var buf bytes.Buffer
	for i := 0; i < raw.Segments.Len(); i++ {
		seg := raw.Segments.At(i)
		buf.Write(seg.Value(c.source))
	}
	run := buf.String()
	if run == "" {
		return nil, nil
	}
	nodes, err := convertProseMirrorFromInlineHTML(run)
	if err != nil {
		return nil, err
	}
	return prependOuterMarks(copyMarks(c.marks), nodes), nil
}

func (c *converter) convertHTMLBlock(n *ast.HTMLBlock) ([]Node, error) {
	var buf bytes.Buffer

	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(c.source))
	}
	if n.HasClosure() {
		buf.Write(n.ClosureLine.Value(c.source))
	}

	raw := buf.String()
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	return convertProseMirrorFromHTMLBlock(raw)
}

func (c *converter) convertStrikethrough(n ast.Node) ([]Node, error) {
	c.marks = append(c.marks, Mark{Type: MarkStrike})
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertTable(n ast.Node) ([]Node, error) {
	var rows []Node

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		converted, err := c.convertNode(child)
		if err != nil {
			return nil, err
		}
		rows = append(rows, converted...)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	return []Node{{Type: NodeTable, Content: rows}}, nil
}

func (c *converter) convertTableHeaderRow(n *goldmarkast.TableHeader) ([]Node, error) {
	cells, err := c.convertTableCells(n, NodeTableHeader)
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return nil, nil
	}

	return []Node{{Type: NodeTableRow, Content: cells}}, nil
}

func (c *converter) convertTableDataRow(n ast.Node) ([]Node, error) {
	cells, err := c.convertTableCells(n, NodeTableCell)
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return nil, nil
	}

	return []Node{{Type: NodeTableRow, Content: cells}}, nil
}

func (c *converter) convertTableCells(row ast.Node, cellType NodeType) ([]Node, error) {
	var cells []Node

	for child := row.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != goldmarkast.KindTableCell {
			continue
		}

		content, err := c.convertTableCellContent(child)
		if err != nil {
			return nil, err
		}

		cellAttrs := TableCellAttrs{
			Colspan: 1,
			Rowspan: 1,
		}

		attrs, err := json.Marshal(cellAttrs)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal table cell attrs: %w", err)
		}

		cells = append(cells, Node{
			Type:    cellType,
			Attrs:   attrs,
			Content: content,
		})
	}

	return cells, nil
}

func (c *converter) convertTableCellContent(cell ast.Node) ([]Node, error) {
	if c.cellHasBlockHTML(cell) {
		raw := c.collectCellRawContent(cell)
		nodes, err := convertProseMirrorFromHTMLBlock(raw)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			return nodes, nil
		}
	}

	inlineContent, err := c.convertInlineChildren(cell)
	if err != nil {
		return nil, err
	}
	return []Node{{Type: NodeParagraph, Content: inlineContent}}, nil
}

func (c *converter) cellHasBlockHTML(cell ast.Node) bool {
	for ch := cell.FirstChild(); ch != nil; ch = ch.NextSibling() {
		if ch.Kind() != ast.KindRawHTML {
			continue
		}
		raw := ch.(*ast.RawHTML)
		for i := 0; i < raw.Segments.Len(); i++ {
			seg := raw.Segments.At(i)
			val := strings.ToLower(string(seg.Value(c.source)))
			if containsBlockOpenTag(val) {
				return true
			}
		}
	}
	return false
}

func containsBlockOpenTag(s string) bool {
	for _, tag := range []string{
		"<ul", "<ol", "<table", "<blockquote", "<pre", "<div",
		"<h1", "<h2", "<h3", "<h4", "<h5", "<h6", "<hr",
	} {
		if strings.Contains(s, tag) {
			return true
		}
	}
	return false
}

func (c *converter) collectCellRawContent(cell ast.Node) string {
	var buf bytes.Buffer
	for ch := cell.FirstChild(); ch != nil; ch = ch.NextSibling() {
		switch ch.Kind() {
		case ast.KindRawHTML:
			raw := ch.(*ast.RawHTML)
			for i := 0; i < raw.Segments.Len(); i++ {
				seg := raw.Segments.At(i)
				buf.Write(seg.Value(c.source))
			}
		case ast.KindText:
			t := ch.(*ast.Text)
			buf.Write(t.Segment.Value(c.source))
			if t.SoftLineBreak() {
				buf.WriteByte(' ')
			}
		case ast.KindString:
			buf.Write(ch.(*ast.String).Value)
		default:
			buf.WriteString(c.extractText(ch))
		}
	}
	return buf.String()
}

// extractText recursively collects the text content of all descendant nodes.
func (c *converter) extractText(n ast.Node) string {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind() {
		case ast.KindText:
			buf.Write(child.(*ast.Text).Segment.Value(c.source))
		case ast.KindString:
			buf.Write(child.(*ast.String).Value)
		default:
			buf.WriteString(c.extractText(child))
		}
	}
	return buf.String()
}

func copyMarks(marks []Mark) []Mark {
	if len(marks) == 0 {
		return nil
	}

	cp := make([]Mark, len(marks))
	copy(cp, marks)

	return cp
}

// prependOuterMarks applies markdown inline context marks (e.g. emphasis around
// raw HTML) to nodes produced from sanitized HTML. Images are left unchanged.
func prependOuterMarks(outer []Mark, nodes []Node) []Node {
	if len(outer) == 0 {
		return nodes
	}
	for i := range nodes {
		if nodes[i].Type == NodeImage {
			continue
		}
		nodes[i].Marks = append(copyMarks(outer), nodes[i].Marks...)
	}
	return nodes
}
