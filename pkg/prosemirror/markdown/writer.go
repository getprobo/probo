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
	"fmt"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

// Writer accumulates Markdown for a node tree.
type Writer struct {
	buf         *bytes.Buffer
	renderer    *Renderer
	prefix      string
	tight       bool
	atLineStart bool
	scratch     bytes.Buffer
	linkAttrs   []prosemirror.LinkAttrs
}

func (w *Writer) pushLinkAttrs(attrs prosemirror.LinkAttrs) {
	w.linkAttrs = append(w.linkAttrs, attrs)
}

func (w *Writer) popLinkAttrs() (prosemirror.LinkAttrs, bool) {
	n := len(w.linkAttrs)
	if n == 0 {
		return prosemirror.LinkAttrs{}, false
	}

	attrs := w.linkAttrs[n-1]
	w.linkAttrs = w.linkAttrs[:n-1]

	return attrs, true
}

func (w *Writer) ensurePrefix() {
	if w.atLineStart {
		w.buf.WriteString(w.prefix)
		w.atLineStart = false
	}
}

// EnsurePrefix writes the current block prefix when at a line start.
func (w *Writer) EnsurePrefix() {
	w.ensurePrefix()
}

// WriteString appends raw bytes to the Markdown output.
func (w *Writer) WriteString(s string) {
	w.buf.WriteString(s)
}

// WriteByte appends a single raw byte to the Markdown output. The error
// return satisfies io.ByteWriter and is always nil.
func (w *Writer) WriteByte(b byte) error {
	return w.buf.WriteByte(b)
}

// WriteAttr appends a space, an attribute name, and its HTML-escaped value.
func (w *Writer) WriteAttr(name, value string) {
	writeAttr(w.buf, name, value)
}

// WriteEscaped appends Markdown-escaped text to the output.
func (w *Writer) WriteEscaped(s string) {
	writeMarkdownEscaped(w.buf, s)
}

// WriteLinkText appends Markdown-escaped link text, also escaping ampersands.
func (w *Writer) WriteLinkText(s string) {
	w.buf.WriteString(strings.ReplaceAll(escapeMarkdown(s), "&", `\&`))
}

func (w *Writer) newLine() {
	w.buf.WriteByte('\n')
	w.atLineStart = true
}

// RenderBlocks renders a sequence of block nodes, inserting a blank line
// between them unless the writer is in tight-list mode.
func (w *Writer) RenderBlocks(nodes []prosemirror.Node) error {
	for i, n := range nodes {
		if i > 0 && !w.tight {
			w.ensurePrefix()
			w.newLine()
		}

		if err := w.renderer.render(w, n); err != nil {
			return err
		}
	}

	return nil
}

// RenderInline renders a sequence of inline nodes.
func (w *Writer) RenderInline(nodes []prosemirror.Node) error {
	for _, n := range nodes {
		if err := w.renderer.render(w, n); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) renderText(n prosemirror.Node) error {
	if n.Text == "" {
		return fmt.Errorf("cannot render text node: text is nil")
	}

	w.ensurePrefix()

	if len(n.Marks) == 0 {
		writeMarkdownEscaped(w.buf, n.Text)
		return nil
	}

	var hasCode bool

	for _, m := range n.Marks {
		if m.Type == prosemirror.MarkCode {
			hasCode = true
			break
		}
	}

	if hasCode {
		return w.renderCodeText(n)
	}

	text := n.Text

	var needsTrim bool

	for _, m := range n.Marks {
		switch m.Type {
		case prosemirror.MarkStrong, prosemirror.MarkEm, prosemirror.MarkStrike:
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

	writeMarkdownEscaped(w.buf, text)

	for _, v := range slices.Backward(n.Marks) {
		if err := w.closeMark(v); err != nil {
			return err
		}
	}

	w.buf.WriteString(trailing)

	return nil
}

func (w *Writer) renderCodeText(n prosemirror.Node) error {
	text := n.Text
	fence := strings.Repeat("`", maxConsecutiveBackticks(text)+1)

	var otherMarks []prosemirror.Mark

	for _, m := range n.Marks {
		if m.Type != prosemirror.MarkCode {
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

func (w *Writer) openMark(m prosemirror.Mark) error {
	return w.renderer.openMark(w, m)
}

func (w *Writer) closeMark(m prosemirror.Mark) error {
	return w.renderer.closeMark(w, m)
}

type listItemMarker func(index int, item prosemirror.Node) (marker, indent string, err error)

func (w *Writer) renderListItems(items []prosemirror.Node, tight bool, next listItemMarker) error {
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

func (w *Writer) renderCellInline(cell prosemirror.Node) error {
	if len(cell.Content) == 1 && cell.Content[0].Type == prosemirror.NodeParagraph {
		return w.RenderInline(cell.Content[0].Content)
	}

	for _, child := range cell.Content {
		w.scratch.Reset()

		html, err := w.renderer.htmlRenderer.Render(child)
		if err != nil {
			return fmt.Errorf("cannot render table cell content: %w", err)
		}

		w.scratch.WriteString(html)
		writeEscapingPipes(w.buf, w.scratch.Bytes())
	}

	return nil
}
