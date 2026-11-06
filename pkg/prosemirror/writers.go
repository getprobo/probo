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
	"bytes"
	"fmt"
	"html"
	"slices"
	"strings"
)

// HTMLWriter accumulates the HTML rendering of a node tree. Node specs write
// into it and call RenderChildren to dispatch nested nodes through the schema.
type HTMLWriter struct {
	buf    *bytes.Buffer
	schema *Schema
}

// WriteString appends raw bytes to the output.
func (w *HTMLWriter) WriteString(s string) {
	w.buf.WriteString(s)
}

// WriteByte appends a single raw byte to the output. The error return
// satisfies io.ByteWriter and is always nil.
func (w *HTMLWriter) WriteByte(b byte) error {
	return w.buf.WriteByte(b)
}

// WriteEscaped appends HTML-escaped text to the output.
func (w *HTMLWriter) WriteEscaped(s string) {
	w.buf.WriteString(html.EscapeString(s))
}

// WriteAttr appends a space, an attribute name, and its HTML-escaped value.
func (w *HTMLWriter) WriteAttr(name, value string) {
	writeAttr(w.buf, name, value)
}

// RenderChildren renders each node through the schema.
func (w *HTMLWriter) RenderChildren(nodes []Node) error {
	for _, child := range nodes {
		if err := w.schema.renderHTMLNode(w, child); err != nil {
			return err
		}
	}

	return nil
}

func (w *HTMLWriter) openMark(m Mark) error {
	if w.schema != DocumentSchema {
		if _, ok := w.schema.markEntries[m.Type]; !ok {
			return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
		}
	}

	return openHTMLMarkByType(w, m)
}

func (w *HTMLWriter) closeMark(m Mark) {
	if w.schema != DocumentSchema {
		if _, ok := w.schema.markEntries[m.Type]; !ok {
			return
		}
	}

	closeHTMLMarkByType(w, m)
}

// RenderHTML renders a node tree to an HTML string.
func (s *Schema) RenderHTML(n Node) (string, error) {
	var buf bytes.Buffer
	w := HTMLWriter{buf: &buf, schema: s}
	if err := s.renderHTMLNode(&w, n); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Schema) renderHTMLNode(w *HTMLWriter, n Node) error {
	if s != DocumentSchema {
		if _, ok := s.index[n.Type]; !ok {
			return fmt.Errorf("cannot render node: unknown type %q", n.Type)
		}
	}

	return renderHTMLByType(w, n)
}

// MarkdownWriter accumulates the Markdown rendering of a node tree. It tracks
// the block prefix, list tightness, and line position that Markdown layout
// needs. Node specs write into it and call RenderBlocks/RenderInline to
// dispatch nested nodes through the schema.
type MarkdownWriter struct {
	buf         *bytes.Buffer
	schema      *Schema
	prefix      string
	tight       bool
	atLineStart bool
}

func (w *MarkdownWriter) ensurePrefix() {
	if w.atLineStart {
		w.buf.WriteString(w.prefix)
		w.atLineStart = false
	}
}

func (w *MarkdownWriter) newLine() {
	w.buf.WriteByte('\n')
	w.atLineStart = true
}

// RenderBlocks renders a sequence of block nodes, inserting a blank line
// between them unless the writer is in tight-list mode.
func (w *MarkdownWriter) RenderBlocks(nodes []Node) error {
	for i, n := range nodes {
		if i > 0 && !w.tight {
			w.ensurePrefix()
			w.newLine()
		}

		if err := w.schema.renderMarkdownNode(w, n); err != nil {
			return err
		}
	}

	return nil
}

// RenderInline renders a sequence of inline nodes.
func (w *MarkdownWriter) RenderInline(nodes []Node) error {
	for _, n := range nodes {
		if err := w.schema.renderMarkdownNode(w, n); err != nil {
			return err
		}
	}

	return nil
}

func (w *MarkdownWriter) renderText(n Node) error {
	if n.Text == nil {
		return fmt.Errorf("cannot render text node: text is nil")
	}

	w.ensurePrefix()

	if len(n.Marks) == 0 {
		w.buf.WriteString(escapeMarkdown(*n.Text))
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
		return w.renderCodeText(n)
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
		w.buf.WriteString(leading)
		w.buf.WriteString(trailing)

		return nil
	}

	w.buf.WriteString(leading)

	for _, m := range n.Marks {
		if err := w.openMark(m); err != nil {
			return err
		}
	}

	w.buf.WriteString(escapeMarkdown(text))

	for _, v := range slices.Backward(n.Marks) {
		if err := w.closeMark(v); err != nil {
			return err
		}
	}

	w.buf.WriteString(trailing)

	return nil
}

func (w *MarkdownWriter) renderCodeText(n Node) error {
	text := *n.Text
	fence := strings.Repeat("`", maxConsecutiveBackticks(text)+1)

	var otherMarks []Mark

	for _, m := range n.Marks {
		if m.Type != MarkCode {
			otherMarks = append(otherMarks, m)
		}
	}

	for _, m := range otherMarks {
		if err := w.openMark(m); err != nil {
			return err
		}
	}

	w.buf.WriteString(fence)

	if len(fence) > 1 {
		w.buf.WriteByte(' ')
	}

	w.buf.WriteString(text)

	if len(fence) > 1 {
		w.buf.WriteByte(' ')
	}

	w.buf.WriteString(fence)

	for _, otherMark := range slices.Backward(otherMarks) {
		if err := w.closeMark(otherMark); err != nil {
			return err
		}
	}

	return nil
}

func (w *MarkdownWriter) openMark(m Mark) error {
	if w.schema != DocumentSchema {
		if _, ok := w.schema.markEntries[m.Type]; !ok {
			return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
		}
	}

	return openMarkdownMarkByType(w, m)
}

func (w *MarkdownWriter) closeMark(m Mark) error {
	if w.schema != DocumentSchema {
		if _, ok := w.schema.markEntries[m.Type]; !ok {
			return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
		}
	}

	return closeMarkdownMarkByType(w, m)
}

// listItemMarker returns the marker written before an item's content (such
// as "- " or "2. ") and the prefix used to indent any continuation lines.
type listItemMarker func(index int, item Node) (marker, indent string, err error)

// renderListItems writes each item of a list using next to produce its
// marker, sharing the blank-line, prefix, and tightness handling that bullet,
// ordered, and task lists all need.
func (w *MarkdownWriter) renderListItems(items []Node, tight bool, next listItemMarker) error {
	for i, item := range items {
		if i > 0 && !tight {
			w.ensurePrefix()
			w.newLine()
		}

		marker, indent, err := next(i, item)
		if err != nil {
			return err
		}

		w.ensurePrefix()
		w.buf.WriteString(marker)
		w.atLineStart = false

		oldPrefix := w.prefix
		oldTight := w.tight
		w.prefix += indent
		w.tight = tight

		if err := w.RenderBlocks(item.Content); err != nil {
			w.prefix = oldPrefix
			w.tight = oldTight

			return err
		}

		w.prefix = oldPrefix
		w.tight = oldTight
	}

	return nil
}

func (w *MarkdownWriter) renderCellInline(cell Node) error {
	if len(cell.Content) == 1 && cell.Content[0].Type == NodeParagraph {
		return w.RenderInline(cell.Content[0].Content)
	}

	for _, child := range cell.Content {
		h, err := w.schema.RenderHTML(child)
		if err != nil {
			return fmt.Errorf("cannot render table cell content: %w", err)
		}

		w.buf.WriteString(strings.ReplaceAll(h, "|", `\|`))
	}

	return nil
}

// RenderMarkdown renders a node tree to a Markdown string.
func (s *Schema) RenderMarkdown(n Node) (string, error) {
	var buf bytes.Buffer
	w := MarkdownWriter{
		buf:         &buf,
		schema:      s,
		atLineStart: true,
	}
	if err := s.renderMarkdownNode(&w, n); err != nil {
		return "", err
	}

	out := strings.TrimRight(buf.String(), "\n")
	if out != "" {
		out += "\n"
	}

	return out, nil
}

func (s *Schema) renderMarkdownNode(w *MarkdownWriter, n Node) error {
	if s != DocumentSchema {
		if _, ok := s.index[n.Type]; !ok {
			return fmt.Errorf("cannot render node: unknown type %q", n.Type)
		}
	}

	return renderMarkdownByType(w, n)
}

func writeAttr(buf *bytes.Buffer, name, value string) {
	buf.WriteByte(' ')
	buf.WriteString(name)
	buf.WriteString(`="`)
	buf.WriteString(html.EscapeString(value))
	buf.WriteByte('"')
}

// maxConsecutiveBackticks returns the length of the longest run of '`' in s.
// Inline code fences must be longer than this value (CommonMark).
func maxConsecutiveBackticks(s string) int {
	max, cur := 0, 0

	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			cur++
			if cur > max {
				max = cur
			}
		} else {
			cur = 0
		}
	}

	return max
}

// listTightness returns true when the list is tight (no blank lines between items).
// Every direct child of n must be itemType; otherwise listTightness returns an error.
func listTightness(n Node, itemType NodeType) (tight bool, err error) {
	tight = true

	for _, item := range n.Content {
		if item.Type != itemType {
			return false, fmt.Errorf(
				"invalid child type %q (expected %q)",
				item.Type,
				itemType,
			)
		}

		if len(item.Content) != 1 {
			tight = false
		}
	}

	return tight, nil
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
