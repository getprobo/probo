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
	"fmt"
	"html"
	"strconv"
)

// RenderHTML renders a ProseMirror document node tree to an HTML string.
func RenderHTML(node Node) (string, error) {
	var buf bytes.Buffer
	if err := renderNode(&buf, node); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderNode(buf *bytes.Buffer, n Node) error {
	switch n.Type {
	case NodeDoc:
		return renderChildren(buf, n.Content)
	case NodeParagraph:
		buf.WriteString("<p>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</p>")
	case NodeHeading:
		attrs, err := n.HeadingAttrs()
		if err != nil {
			return fmt.Errorf("cannot render heading node: %w", err)
		}
		if attrs.Level < 1 || attrs.Level > 6 {
			return fmt.Errorf("cannot render heading node: invalid level %d", attrs.Level)
		}
		level := strconv.Itoa(attrs.Level)
		buf.WriteString("<h")
		buf.WriteString(level)
		buf.WriteByte('>')
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</h")
		buf.WriteString(level)
		buf.WriteByte('>')
	case NodeBlockquote:
		buf.WriteString("<blockquote>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</blockquote>")
	case NodeCodeBlock:
		attrs, err := n.CodeBlockAttrs()
		if err != nil {
			return fmt.Errorf("cannot render code block node: %w", err)
		}
		buf.WriteString("<pre><code")
		if attrs.Language != nil {
			writeAttr(buf, "class", "language-"+*attrs.Language)
		}
		buf.WriteByte('>')
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</code></pre>")
	case NodeHorizontalRule:
		buf.WriteString("<hr>")
	case NodeHardBreak:
		buf.WriteString("<br>")
	case NodeText:
		return renderText(buf, n)
	case NodeImage:
		attrs, err := n.ImageAttrs()
		if err != nil {
			return fmt.Errorf("cannot render image node: %w", err)
		}
		buf.WriteString("<img")
		writeAttr(buf, "src", attrs.Src)
		if attrs.Alt != nil {
			writeAttr(buf, "alt", *attrs.Alt)
		}
		if attrs.Title != nil {
			writeAttr(buf, "title", *attrs.Title)
		}
		buf.WriteByte('>')
	case NodeBulletList:
		buf.WriteString("<ul>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</ul>")
	case NodeOrderedList:
		attrs, err := n.OrderedListAttrs()
		if err != nil {
			return fmt.Errorf("cannot render ordered list node: %w", err)
		}
		buf.WriteString("<ol")
		if attrs.Start != 1 {
			writeAttr(buf, "start", strconv.Itoa(attrs.Start))
		}
		if attrs.Type != nil {
			writeAttr(buf, "type", *attrs.Type)
		}
		buf.WriteByte('>')
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</ol>")
	case NodeListItem:
		buf.WriteString("<li>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</li>")
	case NodeTable:
		buf.WriteString("<table>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</table>")
	case NodeTableRow:
		buf.WriteString("<tr>")
		if err := renderChildren(buf, n.Content); err != nil {
			return err
		}
		buf.WriteString("</tr>")
	case NodeTableCell:
		return renderTableCell(buf, n, "td")
	case NodeTableHeader:
		return renderTableCell(buf, n, "th")
	default:
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}
	return nil
}

func renderChildren(buf *bytes.Buffer, nodes []Node) error {
	for _, child := range nodes {
		if err := renderNode(buf, child); err != nil {
			return err
		}
	}
	return nil
}

func renderText(buf *bytes.Buffer, n Node) error {
	if n.Text == nil {
		return fmt.Errorf("cannot render text node: text is nil")
	}
	for _, m := range n.Marks {
		if err := openMark(buf, m); err != nil {
			return err
		}
	}
	buf.WriteString(html.EscapeString(*n.Text))
	for i := len(n.Marks) - 1; i >= 0; i-- {
		closeMark(buf, n.Marks[i])
	}
	return nil
}

func openMark(buf *bytes.Buffer, m Mark) error {
	switch m.Type {
	case MarkStrong:
		buf.WriteString("<strong>")
	case MarkEm:
		buf.WriteString("<em>")
	case MarkUnderline:
		buf.WriteString("<u>")
	case MarkStrike:
		buf.WriteString("<s>")
	case MarkCode:
		buf.WriteString("<code>")
	case MarkLink:
		attrs, err := m.LinkAttrs()
		if err != nil {
			return fmt.Errorf("cannot render link mark: %w", err)
		}
		buf.WriteString("<a")
		writeAttr(buf, "href", attrs.Href)
		if attrs.Target != nil {
			writeAttr(buf, "target", *attrs.Target)
		}
		if attrs.Rel != nil {
			writeAttr(buf, "rel", *attrs.Rel)
		}
		if attrs.Class != nil {
			writeAttr(buf, "class", *attrs.Class)
		}
		if attrs.Title != nil {
			writeAttr(buf, "title", *attrs.Title)
		}
		buf.WriteByte('>')
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
	return nil
}

func closeMark(buf *bytes.Buffer, m Mark) {
	switch m.Type {
	case MarkStrong:
		buf.WriteString("</strong>")
	case MarkEm:
		buf.WriteString("</em>")
	case MarkUnderline:
		buf.WriteString("</u>")
	case MarkStrike:
		buf.WriteString("</s>")
	case MarkCode:
		buf.WriteString("</code>")
	case MarkLink:
		buf.WriteString("</a>")
	}
}

func renderTableCell(buf *bytes.Buffer, n Node, tag string) error {
	attrs, err := n.TableCellAttrs()
	if err != nil {
		return fmt.Errorf("cannot render %s node: %w", tag, err)
	}
	buf.WriteByte('<')
	buf.WriteString(tag)
	if attrs.Colspan > 1 {
		writeAttr(buf, "colspan", strconv.Itoa(attrs.Colspan))
	}
	if attrs.Rowspan > 1 {
		writeAttr(buf, "rowspan", strconv.Itoa(attrs.Rowspan))
	}
	if len(attrs.Colwidth) > 0 {
		total := 0
		for _, w := range attrs.Colwidth {
			total += w
		}
		writeAttr(buf, "style", "min-width: "+strconv.Itoa(total)+"px")
	}
	buf.WriteByte('>')
	if err := renderChildren(buf, n.Content); err != nil {
		return err
	}
	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteByte('>')
	return nil
}

func writeAttr(buf *bytes.Buffer, name, value string) {
	buf.WriteByte(' ')
	buf.WriteString(name)
	buf.WriteString(`="`)
	buf.WriteString(html.EscapeString(value))
	buf.WriteByte('"')
}
