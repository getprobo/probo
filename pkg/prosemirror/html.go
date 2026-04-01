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
	"net/url"
	"strconv"
	"strings"
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
		if attrs.Language != nil && *attrs.Language == "mermaid" {
			buf.WriteString(`<pre class="mermaid">`)
			if err := renderChildren(buf, n.Content); err != nil {
				return err
			}
			buf.WriteString("</pre>")
		} else {
			buf.WriteString("<pre><code")
			if attrs.Language != nil {
				writeAttr(buf, "class", "language-"+*attrs.Language)
			}
			buf.WriteByte('>')
			if err := renderChildren(buf, n.Content); err != nil {
				return err
			}
			buf.WriteString("</code></pre>")
		}
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
		writeAttr(buf, "src", safeImageSrc(attrs.Src))
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
		writeAttr(buf, "href", safeLinkHref(attrs.Href))
		if attrs.Target != nil {
			writeAttr(buf, "target", *attrs.Target)
		}
		rel := linkRelToEmit(attrs)
		if rel != "" {
			writeAttr(buf, "rel", rel)
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

const linkRelBlankTargetDefault = "noopener noreferrer"

// linkRelToEmit returns the rel attribute value for a link mark, or empty when
// the attribute should be omitted. When target opens a new browsing context
// (_blank), noopener is always injected to prevent the opened page from
// accessing window.opener, even when the document supplies a custom rel.
func linkRelToEmit(attrs LinkAttrs) string {
	blanksTarget := attrs.Target != nil &&
		strings.EqualFold(strings.TrimSpace(*attrs.Target), "_blank")

	if attrs.Rel != nil {
		if s := strings.TrimSpace(*attrs.Rel); s != "" {
			if blanksTarget {
				return ensureNoopener(s)
			}
			return s
		}
	}

	if blanksTarget {
		return linkRelBlankTargetDefault
	}
	return ""
}

// ensureNoopener returns rel unchanged when it already contains the noopener
// token (case-insensitive check). Otherwise it appends " noopener".
func ensureNoopener(rel string) string {
	for tok := range strings.FieldsSeq(rel) {
		if strings.EqualFold(tok, "noopener") {
			return rel
		}
	}
	return rel + " noopener"
}

func writeAttr(buf *bytes.Buffer, name, value string) {
	buf.WriteByte(' ')
	buf.WriteString(name)
	buf.WriteString(`="`)
	buf.WriteString(html.EscapeString(value))
	buf.WriteByte('"')
}

// safeLinkHref returns a value safe to use in link mark attrs and to emit in an
// HTML href attribute. URLs with disallowed schemes (for example javascript: or
// data:) are replaced with "#" so escaped text content cannot be combined with an
// executable URL.
func safeLinkHref(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return "#"
	}
	if href[0] == '#' {
		return href
	}
	if strings.HasPrefix(href, "/") {
		if len(href) > 1 && (href[1] == '/' || href[1] == '\\') {
			return "#"
		}
		return href
	}
	u, err := url.Parse(href)
	if err != nil {
		return "#"
	}
	if u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https", "mailto", "tel":
			return href
		default:
			return "#"
		}
	}
	if u.Host != "" {
		return "#"
	}
	return href
}

// safeImageSrc returns a value safe to use in an HTML img src attribute.
// Only http, https, and data schemes are permitted; everything else
// (javascript:, vbscript:, etc.) is replaced with an empty string so the
// image simply does not render.
func safeImageSrc(src string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}
	if strings.HasPrefix(src, "/") {
		if len(src) > 1 && (src[1] == '/' || src[1] == '\\') {
			return ""
		}
		return src
	}
	u, err := url.Parse(src)
	if err != nil {
		return ""
	}
	if u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https", "data":
			return src
		default:
			return ""
		}
	}
	if u.Host != "" {
		return ""
	}
	return src
}
