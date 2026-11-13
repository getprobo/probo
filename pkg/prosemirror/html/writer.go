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

package html

import (
	"bytes"
	"html"

	"go.probo.inc/probo/pkg/prosemirror"
)

// Writer accumulates HTML for a node tree.
type Writer struct {
	buf      *bytes.Buffer
	renderer *Renderer
}

// WriteString appends raw bytes to the output.
func (w *Writer) WriteString(s string) {
	w.buf.WriteString(s)
}

// WriteByte appends a single raw byte to the output. The error return
// satisfies io.ByteWriter and is always nil.
func (w *Writer) WriteByte(b byte) error {
	return w.buf.WriteByte(b)
}

// WriteEscaped appends HTML-escaped text to the output.
func (w *Writer) WriteEscaped(s string) {
	w.buf.WriteString(html.EscapeString(s))
}

// WriteAttr appends a space, an attribute name, and its HTML-escaped value.
func (w *Writer) WriteAttr(name, value string) {
	writeAttr(w.buf, name, value)
}

// RenderChildren renders each node through the renderer.
func (w *Writer) RenderChildren(nodes []prosemirror.Node) error {
	for _, child := range nodes {
		if err := w.renderer.render(w, child); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) openMark(m prosemirror.Mark) error {
	return w.renderer.openMark(w, m)
}

func (w *Writer) closeMark(m prosemirror.Mark) {
	w.renderer.closeMark(w, m)
}
