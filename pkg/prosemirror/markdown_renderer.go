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
	"strconv"
	"strings"
)

// RenderMarkdown renders a ProseMirror document node tree to a Markdown string.
func RenderMarkdown(node Node) (string, error) {
	r := &mdRenderer{
		buf:         &bytes.Buffer{},
		atLineStart: true,
	}
	if err := r.renderNode(node); err != nil {
		return "", err
	}
	out := strings.TrimRight(r.buf.String(), "\n")
	if out != "" {
		out += "\n"
	}
	return out, nil
}

type mdRenderer struct {
	buf         *bytes.Buffer
	prefix      string
	tight       bool
	atLineStart bool
}

func (r *mdRenderer) ensurePrefix() {
	if r.atLineStart {
		r.buf.WriteString(r.prefix)
		r.atLineStart = false
	}
}

func (r *mdRenderer) newLine() {
	r.buf.WriteByte('\n')
	r.atLineStart = true
}

func (r *mdRenderer) renderNode(n Node) error {
	switch n.Type {
	case NodeDoc:
		return r.renderBlocks(n.Content)
	case NodeParagraph:
		r.ensurePrefix()
		if err := r.renderInline(n.Content); err != nil {
			return err
		}
		r.newLine()
	case NodeHeading:
		attrs, err := n.HeadingAttrs()
		if err != nil {
			return fmt.Errorf("cannot render heading node: %w", err)
		}
		if attrs.Level < 1 || attrs.Level > 6 {
			return fmt.Errorf("cannot render heading node: invalid level %d", attrs.Level)
		}
		r.ensurePrefix()
		for i := 0; i < attrs.Level; i++ {
			r.buf.WriteByte('#')
		}
		r.buf.WriteByte(' ')
		if err := r.renderInline(n.Content); err != nil {
			return err
		}
		r.newLine()
	case NodeBlockquote:
		oldPrefix := r.prefix
		r.prefix += "> "
		if err := r.renderBlocks(n.Content); err != nil {
			r.prefix = oldPrefix
			return err
		}
		r.prefix = oldPrefix
	case NodeCodeBlock:
		attrs, err := n.CodeBlockAttrs()
		if err != nil {
			return fmt.Errorf("cannot render code block node: %w", err)
		}
		code := collectText(n.Content)
		fence := chooseFence(code)
		r.ensurePrefix()
		r.buf.WriteString(fence)
		if attrs.Language != nil {
			r.buf.WriteString(*attrs.Language)
		}
		r.newLine()
		for line := range strings.SplitSeq(code, "\n") {
			r.ensurePrefix()
			r.buf.WriteString(line)
			r.newLine()
		}
		r.ensurePrefix()
		r.buf.WriteString(fence)
		r.newLine()
	case NodeHorizontalRule:
		r.ensurePrefix()
		r.buf.WriteString("---")
		r.newLine()
	case NodeHardBreak:
		r.buf.WriteByte('\\')
		r.newLine()
	case NodeText:
		return r.renderText(n)
	case NodeImage:
		attrs, err := n.ImageAttrs()
		if err != nil {
			return fmt.Errorf("cannot render image node: %w", err)
		}
		r.ensurePrefix()
		r.buf.WriteString("![")
		if attrs.Alt != nil {
			r.buf.WriteString(escapeMarkdown(*attrs.Alt))
		}
		r.buf.WriteString("](")
		r.buf.WriteString(safeImageSrc(attrs.Src))
		if attrs.Title != nil {
			r.buf.WriteString(` "`)
			r.buf.WriteString(strings.ReplaceAll(*attrs.Title, `"`, `\"`))
			r.buf.WriteByte('"')
		}
		r.buf.WriteByte(')')
	case NodeBulletList:
		return r.renderBulletList(n)
	case NodeOrderedList:
		return r.renderOrderedList(n)
	case NodeListItem:
		return fmt.Errorf("cannot render list item outside of list context")
	case NodeTable:
		return r.renderTable(n)
	case NodeTableRow, NodeTableCell, NodeTableHeader:
		return fmt.Errorf("cannot render %s outside of table context", n.Type)
	default:
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}
	return nil
}

func (r *mdRenderer) renderBlocks(nodes []Node) error {
	for i, n := range nodes {
		if i > 0 && !r.tight {
			r.ensurePrefix()
			r.newLine()
		}
		if err := r.renderNode(n); err != nil {
			return err
		}
	}
	return nil
}

func (r *mdRenderer) renderInline(nodes []Node) error {
	for _, n := range nodes {
		if err := r.renderNode(n); err != nil {
			return err
		}
	}
	return nil
}

func (r *mdRenderer) renderText(n Node) error {
	if n.Text == nil {
		return fmt.Errorf("cannot render text node: text is nil")
	}

	r.ensurePrefix()

	if len(n.Marks) == 0 {
		r.buf.WriteString(escapeMarkdown(*n.Text))
		return nil
	}

	var hasCode bool
	for _, m := range n.Marks {
		if m.Type == MarkCode {
			hasCode = true
			break
		}
	}
	if hasCode {
		return r.renderCodeText(n)
	}

	text := *n.Text

	var needsTrim bool
	for _, m := range n.Marks {
		switch m.Type {
		case MarkStrong, MarkEm, MarkStrike:
			needsTrim = true
		}
	}

	var leading, trailing string
	if needsTrim {
		origLen := len(text)
		text = strings.TrimLeft(text, " ")
		leading = strings.Repeat(" ", origLen-len(text))
		origLen = len(text)
		text = strings.TrimRight(text, " ")
		trailing = strings.Repeat(" ", origLen-len(text))
	}

	if text == "" {
		r.buf.WriteString(leading)
		r.buf.WriteString(trailing)
		return nil
	}

	r.buf.WriteString(leading)

	for _, m := range n.Marks {
		if err := r.openMark(m); err != nil {
			return err
		}
	}

	r.buf.WriteString(escapeMarkdown(text))

	for i := len(n.Marks) - 1; i >= 0; i-- {
		if err := r.closeMark(n.Marks[i]); err != nil {
			return err
		}
	}

	r.buf.WriteString(trailing)

	return nil
}

func (r *mdRenderer) renderCodeText(n Node) error {
	text := *n.Text
	backtick := "`"
	if strings.Contains(text, "`") {
		backtick = "``"
	}

	var otherMarks []Mark
	for _, m := range n.Marks {
		if m.Type != MarkCode {
			otherMarks = append(otherMarks, m)
		}
	}

	for _, m := range otherMarks {
		if err := r.openMark(m); err != nil {
			return err
		}
	}

	r.buf.WriteString(backtick)
	if backtick == "``" {
		r.buf.WriteByte(' ')
	}
	r.buf.WriteString(text)
	if backtick == "``" {
		r.buf.WriteByte(' ')
	}
	r.buf.WriteString(backtick)

	for i := len(otherMarks) - 1; i >= 0; i-- {
		if err := r.closeMark(otherMarks[i]); err != nil {
			return err
		}
	}

	return nil
}

func (r *mdRenderer) openMark(m Mark) error {
	switch m.Type {
	case MarkStrong:
		r.buf.WriteString("**")
	case MarkEm:
		r.buf.WriteByte('*')
	case MarkUnderline:
		r.buf.WriteString("<u>")
	case MarkStrike:
		r.buf.WriteString("~~")
	case MarkLink:
		r.buf.WriteByte('[')
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
	return nil
}

func (r *mdRenderer) closeMark(m Mark) error {
	switch m.Type {
	case MarkStrong:
		r.buf.WriteString("**")
	case MarkEm:
		r.buf.WriteByte('*')
	case MarkUnderline:
		r.buf.WriteString("</u>")
	case MarkStrike:
		r.buf.WriteString("~~")
	case MarkLink:
		attrs, err := m.LinkAttrs()
		if err != nil {
			return fmt.Errorf("cannot render link mark: %w", err)
		}
		r.buf.WriteString("](")
		r.buf.WriteString(safeLinkHref(attrs.Href))
		if attrs.Title != nil {
			r.buf.WriteString(` "`)
			r.buf.WriteString(strings.ReplaceAll(*attrs.Title, `"`, `\"`))
			r.buf.WriteByte('"')
		}
		r.buf.WriteByte(')')
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
	return nil
}

func (r *mdRenderer) renderBulletList(n Node) error {
	tight := isTightList(n)
	for i, item := range n.Content {
		if item.Type != NodeListItem {
			continue
		}
		if i > 0 && !tight {
			r.ensurePrefix()
			r.newLine()
		}
		r.ensurePrefix()
		r.buf.WriteString("- ")
		r.atLineStart = false

		oldPrefix := r.prefix
		oldTight := r.tight
		r.prefix += "  "
		r.tight = tight

		if err := r.renderBlocks(item.Content); err != nil {
			r.prefix = oldPrefix
			r.tight = oldTight
			return err
		}

		r.prefix = oldPrefix
		r.tight = oldTight
	}
	return nil
}

func (r *mdRenderer) renderOrderedList(n Node) error {
	attrs, err := n.OrderedListAttrs()
	if err != nil {
		return fmt.Errorf("cannot render ordered list node: %w", err)
	}
	tight := isTightList(n)
	start := max(attrs.Start, 1)
	for i, item := range n.Content {
		if item.Type != NodeListItem {
			continue
		}
		if i > 0 && !tight {
			r.ensurePrefix()
			r.newLine()
		}
		r.ensurePrefix()
		num := strconv.Itoa(start + i)
		r.buf.WriteString(num)
		r.buf.WriteString(". ")
		r.atLineStart = false

		indent := strings.Repeat(" ", len(num)+2)
		oldPrefix := r.prefix
		oldTight := r.tight
		r.prefix += indent
		r.tight = tight

		if err := r.renderBlocks(item.Content); err != nil {
			r.prefix = oldPrefix
			r.tight = oldTight
			return err
		}

		r.prefix = oldPrefix
		r.tight = oldTight
	}
	return nil
}

func (r *mdRenderer) renderTable(n Node) error {
	return r.renderGFMTable(n)
}

func (r *mdRenderer) renderGFMTable(n Node) error {
	if len(n.Content) == 0 {
		return nil
	}

	headerRow := n.Content[0]
	r.ensurePrefix()
	r.buf.WriteByte('|')
	for _, cell := range headerRow.Content {
		r.buf.WriteByte(' ')
		if err := r.renderCellInline(cell); err != nil {
			return err
		}
		r.buf.WriteString(" |")
	}
	r.newLine()

	r.ensurePrefix()
	r.buf.WriteByte('|')
	for range headerRow.Content {
		r.buf.WriteString(" --- |")
	}
	r.newLine()

	for _, row := range n.Content[1:] {
		r.ensurePrefix()
		r.buf.WriteByte('|')
		for _, cell := range row.Content {
			r.buf.WriteByte(' ')
			if err := r.renderCellInline(cell); err != nil {
				return err
			}
			r.buf.WriteString(" |")
		}
		r.newLine()
	}

	return nil
}

func (r *mdRenderer) renderCellInline(cell Node) error {
	if len(cell.Content) == 1 && cell.Content[0].Type == NodeParagraph {
		return r.renderInline(cell.Content[0].Content)
	}
	for _, child := range cell.Content {
		h, err := RenderHTML(child)
		if err != nil {
			return fmt.Errorf("cannot render table cell content: %w", err)
		}
		r.buf.WriteString(strings.ReplaceAll(h, "|", `\|`))
	}
	return nil
}

func isTightList(n Node) bool {
	for _, item := range n.Content {
		if item.Type != NodeListItem {
			continue
		}
		if len(item.Content) != 1 {
			return false
		}
	}
	return true
}

func collectText(nodes []Node) string {
	var buf strings.Builder
	for _, n := range nodes {
		if n.Text != nil {
			buf.WriteString(*n.Text)
		}
	}
	return buf.String()
}

func chooseFence(code string) string {
	fence := "```"
	for strings.Contains(code, fence) {
		fence += "`"
	}
	return fence
}

func escapeMarkdown(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	for _, c := range s {
		switch c {
		case '\\', '*', '_', '`', '[', ']', '~', '|', '<':
			buf.WriteByte('\\')
		}
		buf.WriteRune(c)
	}
	return buf.String()
}
