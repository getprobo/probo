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

package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/internal/urlsafe"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var markdownParser = goldmark.New(
	goldmark.WithExtensions(
		goldmarkext.Strikethrough,
		goldmarkext.Table,
		goldmarkext.TaskList,
	),
)

// Parse converts markdown into a ProseMirror node tree with no product-specific rewrites.
func Parse(src string) (prosemirror.Node, error) {
	source := []byte(src)
	doc := markdownParser.Parser().Parse(text.NewReader(source))
	c := &converter{source: source}

	nodes, err := c.convertChildren(doc)
	if err != nil {
		return prosemirror.Node{}, fmt.Errorf("cannot convert markdown to prosemirror: %w", err)
	}

	return prosemirror.Node{
		Type:    prosemirror.NodeDoc,
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
	marks  []prosemirror.Mark
}

func (c *converter) convertChildren(n ast.Node) ([]prosemirror.Node, error) {
	var nodes []prosemirror.Node

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
func (c *converter) convertInlineChildren(n ast.Node) ([]prosemirror.Node, error) {
	var nodes []prosemirror.Node

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

func (c *converter) convertNode(n ast.Node) ([]prosemirror.Node, error) {
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
		return []prosemirror.Node{{Type: prosemirror.NodeHorizontalRule}}, nil
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
		case goldmarkast.KindTaskCheckBox:
			return nil, nil
		}

		return nil, fmt.Errorf("cannot convert markdown node of kind %s", n.Kind())
	}
}

func (c *converter) convertHeading(n *ast.Heading) ([]prosemirror.Node, error) {
	children, err := c.convertInlineChildren(n)
	if err != nil {
		return nil, err
	}

	attrs, err := json.Marshal(prosemirror.HeadingAttrs{Level: n.Level})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal heading attrs: %w", err)
	}

	return blocksFromInlines(prosemirror.NodeHeading, attrs, children), nil
}

func (c *converter) convertParagraph(n ast.Node) ([]prosemirror.Node, error) {
	children, err := c.convertInlineChildren(n)
	if err != nil {
		return nil, err
	}

	return blocksFromInlines(prosemirror.NodeParagraph, nil, children), nil
}

func (c *converter) convertBlockquote(n ast.Node) ([]prosemirror.Node, error) {
	children, err := c.convertChildren(n)
	if err != nil {
		return nil, err
	}

	return []prosemirror.Node{{
		Type:    prosemirror.NodeBlockquote,
		Content: children,
	}}, nil
}

func (c *converter) convertFencedCodeBlock(n *ast.FencedCodeBlock) ([]prosemirror.Node, error) {
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

	attrs, err := json.Marshal(prosemirror.CodeBlockAttrs{Language: lang})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []prosemirror.Node
	if content != "" {
		textNodes = []prosemirror.Node{{
			Type: prosemirror.NodeText,
			Text: content,
		}}
	}

	return []prosemirror.Node{{
		Type:    prosemirror.NodeCodeBlock,
		Content: textNodes,
		Attrs:   attrs,
	}}, nil
}

func (c *converter) convertCodeBlock(n *ast.CodeBlock) ([]prosemirror.Node, error) {
	var buf bytes.Buffer

	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(c.source))
	}

	content := normalizeCodeBlockContent(buf.String())

	attrs, err := json.Marshal(prosemirror.CodeBlockAttrs{Language: nil})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []prosemirror.Node
	if content != "" {
		textNodes = []prosemirror.Node{{
			Type: prosemirror.NodeText,
			Text: content,
		}}
	}

	return []prosemirror.Node{{
		Type:    prosemirror.NodeCodeBlock,
		Content: textNodes,
		Attrs:   attrs,
	}}, nil
}

func (c *converter) convertList(n *ast.List) ([]prosemirror.Node, error) {
	items, err := c.markdownListItems(n)
	if err != nil {
		return nil, err
	}

	groups := consecutiveGroups(items, func(item prosemirror.Node) bool {
		return item.Type == prosemirror.NodeTaskItem
	})
	if len(groups) == 0 {
		return wrapPlainList(n, nil, 0)
	}

	return markdownListsForGroups(n, groups)
}

func (c *converter) markdownListItems(n *ast.List) ([]prosemirror.Node, error) {
	var items []prosemirror.Node

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != ast.KindListItem {
			continue
		}

		converted, err := c.convertListItem(child)
		if err != nil {
			return nil, err
		}

		items = append(items, converted...)
	}

	return items, nil
}

func markdownListsForGroups(n *ast.List, groups [][]prosemirror.Node) ([]prosemirror.Node, error) {
	var nodes []prosemirror.Node

	consumed := 0

	for _, group := range groups {
		if group[0].Type == prosemirror.NodeTaskItem {
			nodes = append(nodes, prosemirror.Node{Type: prosemirror.NodeTaskList, Content: group})

			consumed += len(group)

			continue
		}

		plain, err := wrapPlainList(n, group, consumed)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, plain...)
		consumed += len(group)
	}

	return nodes, nil
}

func wrapPlainList(n *ast.List, items []prosemirror.Node, consumed int) ([]prosemirror.Node, error) {
	if n.IsOrdered() {
		attrs, err := json.Marshal(prosemirror.OrderedListAttrs{Start: n.Start + consumed})
		if err != nil {
			return nil, fmt.Errorf("cannot marshal ordered list attrs: %w", err)
		}

		return []prosemirror.Node{{
			Type:    prosemirror.NodeOrderedList,
			Content: items,
			Attrs:   attrs,
		}}, nil
	}

	return []prosemirror.Node{{
		Type:    prosemirror.NodeBulletList,
		Content: items,
	}}, nil
}

func listItemTask(item ast.Node) (bool, bool) {
	first := item.FirstChild()
	if first == nil {
		return false, false
	}

	box, ok := first.FirstChild().(*goldmarkast.TaskCheckBox)
	if !ok || box == nil {
		return false, false
	}

	return box.IsChecked, true
}

func (c *converter) convertListItem(n ast.Node) ([]prosemirror.Node, error) {
	checked, task := listItemTask(n)

	children, err := c.convertChildren(n)
	if err != nil {
		return nil, err
	}

	item := prosemirror.Node{
		Type:    prosemirror.NodeListItem,
		Content: ensureLeadingParagraph(children),
	}

	if !task {
		return []prosemirror.Node{item}, nil
	}

	attrs, err := json.Marshal(prosemirror.TaskItemAttrs{Checked: checked})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal task item attrs: %w", err)
	}

	item.Type = prosemirror.NodeTaskItem
	item.Attrs = attrs

	return []prosemirror.Node{item}, nil
}

func (c *converter) convertImage(n *ast.Image) ([]prosemirror.Node, error) {
	src := string(n.Destination)
	imgAttrs := prosemirror.ImageAttrs{
		Src: src,
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

	return []prosemirror.Node{{
		Type:  prosemirror.NodeImage,
		Attrs: attrs,
	}}, nil
}

func (c *converter) convertText(n *ast.Text) ([]prosemirror.Node, error) {
	content := string(n.Segment.Value(c.source))
	if n.SoftLineBreak() {
		content += " "
	}

	if content == "" {
		return nil, nil
	}

	nodes := []prosemirror.Node{{
		Type:  prosemirror.NodeText,
		Text:  content,
		Marks: copyMarks(c.marks),
	}}

	if n.HardLineBreak() {
		nodes = append(nodes, prosemirror.Node{Type: prosemirror.NodeHardBreak})
	}

	return nodes, nil
}

func (c *converter) convertString(n *ast.String) ([]prosemirror.Node, error) {
	content := string(n.Value)
	if content == "" {
		return nil, nil
	}

	return []prosemirror.Node{{
		Type:  prosemirror.NodeText,
		Text:  content,
		Marks: copyMarks(c.marks),
	}}, nil
}

func (c *converter) convertEmphasis(n *ast.Emphasis) ([]prosemirror.Node, error) {
	var mark prosemirror.Mark
	if n.Level == 2 {
		mark = prosemirror.Mark{Type: prosemirror.MarkStrong}
	} else {
		mark = prosemirror.Mark{Type: prosemirror.MarkEm}
	}

	c.marks = append(c.marks, mark)
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]

	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertCodeSpan(n ast.Node) ([]prosemirror.Node, error) {
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
	marks = append(marks, prosemirror.Mark{Type: prosemirror.MarkCode})

	return []prosemirror.Node{{
		Type:  prosemirror.NodeText,
		Text:  content,
		Marks: marks,
	}}, nil
}

func (c *converter) convertLink(n *ast.Link) ([]prosemirror.Node, error) {
	linkAttrs := prosemirror.LinkAttrs{
		Href: urlsafe.LinkHref(string(n.Destination)),
	}

	if n.Title != nil {
		t := string(n.Title)
		linkAttrs.Title = &t
	}

	attrs, err := json.Marshal(linkAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}

	c.marks = append(c.marks, prosemirror.Mark{Type: prosemirror.MarkLink, Attrs: attrs})
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]

	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertAutoLink(n *ast.AutoLink) ([]prosemirror.Node, error) {
	url := string(n.URL(c.source))

	linkAttrs := prosemirror.LinkAttrs{Href: urlsafe.LinkHref(url)}

	attrs, err := json.Marshal(linkAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}

	label := string(n.Label(c.source))

	return []prosemirror.Node{{
		Type:  prosemirror.NodeText,
		Text:  label,
		Marks: append(copyMarks(c.marks), prosemirror.Mark{Type: prosemirror.MarkLink, Attrs: attrs}),
	}}, nil
}

func (c *converter) convertRawHTML(n ast.Node) ([]prosemirror.Node, error) {
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

func (c *converter) convertHTMLBlock(n *ast.HTMLBlock) ([]prosemirror.Node, error) {
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

func (c *converter) convertStrikethrough(n ast.Node) ([]prosemirror.Node, error) {
	c.marks = append(c.marks, prosemirror.Mark{Type: prosemirror.MarkStrike})
	children, err := c.convertInlineChildren(n)
	c.marks = c.marks[:len(c.marks)-1]

	if err != nil {
		return nil, err
	}

	return children, nil
}

func (c *converter) convertTable(n ast.Node) ([]prosemirror.Node, error) {
	var rows []prosemirror.Node

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

	return []prosemirror.Node{{Type: prosemirror.NodeTable, Content: rows}}, nil
}

func (c *converter) convertTableHeaderRow(n *goldmarkast.TableHeader) ([]prosemirror.Node, error) {
	cells, err := c.convertTableCells(n, prosemirror.NodeTableHeader)
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return nil, nil
	}

	return []prosemirror.Node{{Type: prosemirror.NodeTableRow, Content: cells}}, nil
}

func (c *converter) convertTableDataRow(n ast.Node) ([]prosemirror.Node, error) {
	cells, err := c.convertTableCells(n, prosemirror.NodeTableCell)
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return nil, nil
	}

	return []prosemirror.Node{{Type: prosemirror.NodeTableRow, Content: cells}}, nil
}

func (c *converter) convertTableCells(row ast.Node, cellType prosemirror.NodeType) ([]prosemirror.Node, error) {
	var cells []prosemirror.Node

	for child := row.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != goldmarkast.KindTableCell {
			continue
		}

		content, err := c.convertTableCellContent(child)
		if err != nil {
			return nil, err
		}

		cellAttrs := prosemirror.TableCellAttrs{
			Colspan: 1,
			Rowspan: 1,
		}

		attrs, err := json.Marshal(cellAttrs)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal table cell attrs: %w", err)
		}

		cells = append(
			cells,
			prosemirror.Node{
				Type:    cellType,
				Attrs:   attrs,
				Content: content,
			},
		)
	}

	return cells, nil
}

func (c *converter) convertTableCellContent(cell ast.Node) ([]prosemirror.Node, error) {
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

	return blocksFromInlines(prosemirror.NodeParagraph, nil, inlineContent), nil
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

func copyMarks(marks []prosemirror.Mark) []prosemirror.Mark {
	if len(marks) == 0 {
		return nil
	}

	cp := make([]prosemirror.Mark, len(marks))
	copy(cp, marks)

	return cp
}

// prependOuterMarks applies markdown inline context marks (e.g. emphasis around
// raw HTML) to nodes produced from sanitized HTML. Images are left unchanged.
func prependOuterMarks(outer []prosemirror.Mark, nodes []prosemirror.Node) []prosemirror.Node {
	if len(outer) == 0 {
		return nodes
	}

	for i := range nodes {
		if nodes[i].Type == prosemirror.NodeImage {
			continue
		}

		nodes[i].Marks = append(copyMarks(outer), nodes[i].Marks...)
	}

	return nodes
}
