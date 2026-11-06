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

package prosemirror

import (
	"fmt"
	"strconv"
)

type tableNode struct{}

// Table holds one or more table rows.
func Table() NodeSpec {
	return tableNode{}
}

func (tableNode) Name() NodeType {
	return NodeTable
}

func (tableNode) Group() Group {
	return GroupBlock
}

func (tableNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeTableRow}, Min: 1}
}

func (tableNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func (tableNode) RenderHTML(w *HTMLWriter, n Node) error {
	w.WriteString("<table>")

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</table>")

	return nil
}

func (tableNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	if len(n.Content) == 0 {
		return nil
	}

	headerRow := n.Content[0]

	w.ensurePrefix()
	w.buf.WriteByte('|')

	for _, cell := range headerRow.Content {
		w.buf.WriteByte(' ')

		if err := w.renderCellInline(cell); err != nil {
			return err
		}

		w.buf.WriteString(" |")
	}

	w.newLine()

	w.ensurePrefix()
	w.buf.WriteByte('|')

	for range headerRow.Content {
		w.buf.WriteString(" --- |")
	}

	w.newLine()

	for _, row := range n.Content[1:] {
		w.ensurePrefix()
		w.buf.WriteByte('|')

		for _, cell := range row.Content {
			w.buf.WriteByte(' ')

			if err := w.renderCellInline(cell); err != nil {
				return err
			}

			w.buf.WriteString(" |")
		}

		w.newLine()
	}

	return nil
}

type tableRowNode struct{}

// TableRow holds one or more table cells or headers.
func TableRow() NodeSpec {
	return tableRowNode{}
}

func (tableRowNode) Name() NodeType {
	return NodeTableRow
}

func (tableRowNode) Group() Group {
	return GroupNone
}

func (tableRowNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeTableCell, NodeTableHeader}, Min: 1}
}

func (tableRowNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func (tableRowNode) RenderHTML(w *HTMLWriter, n Node) error {
	w.WriteString("<tr>")

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</tr>")

	return nil
}

func (tableRowNode) RenderMarkdown(_ *MarkdownWriter, n Node) error {
	return fmt.Errorf("cannot render %s outside of table context", n.Type)
}

type tableCellNode struct{}

// TableCell is a standard table cell holding block content.
func TableCell() NodeSpec {
	return tableCellNode{}
}

func (tableCellNode) Name() NodeType {
	return NodeTableCell
}

func (tableCellNode) Group() Group {
	return GroupNone
}

func (tableCellNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (tableCellNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func (tableCellNode) RenderHTML(w *HTMLWriter, n Node) error {
	return renderTableCellHTML(w, n, "td")
}

func (tableCellNode) RenderMarkdown(_ *MarkdownWriter, n Node) error {
	return fmt.Errorf("cannot render %s outside of table context", n.Type)
}

type tableHeaderNode struct{}

// TableHeader is a header table cell holding block content.
func TableHeader() NodeSpec {
	return tableHeaderNode{}
}

func (tableHeaderNode) Name() NodeType {
	return NodeTableHeader
}

func (tableHeaderNode) Group() Group {
	return GroupNone
}

func (tableHeaderNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (tableHeaderNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func (tableHeaderNode) RenderHTML(w *HTMLWriter, n Node) error {
	return renderTableCellHTML(w, n, "th")
}

func (tableHeaderNode) RenderMarkdown(_ *MarkdownWriter, n Node) error {
	return fmt.Errorf("cannot render %s outside of table context", n.Type)
}

func renderTableCellHTML(w *HTMLWriter, n Node, tag string) error {
	attrs, err := n.TableCellAttrs()
	if err != nil {
		return fmt.Errorf("cannot render %s node: %w", tag, err)
	}

	w.WriteByte('<')
	w.WriteString(tag)

	if attrs.Colspan > 1 {
		w.WriteAttr("colspan", strconv.Itoa(attrs.Colspan))
	}

	if attrs.Rowspan > 1 {
		w.WriteAttr("rowspan", strconv.Itoa(attrs.Rowspan))
	}

	if len(attrs.Colwidth) > 0 {
		total := 0
		for _, width := range attrs.Colwidth {
			total += width
		}

		w.WriteAttr("style", "min-width: "+strconv.Itoa(total)+"px")
	}

	w.WriteByte('>')

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</")
	w.WriteString(tag)
	w.WriteByte('>')

	return nil
}
