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
	"fmt"

	"go.probo.inc/probo/pkg/prosemirror"
)

type linkRenderer struct{}

func Link() MarkRenderer { return linkRenderer{} }

func (linkRenderer) Mark() prosemirror.MarkType { return prosemirror.MarkLink }

func (linkRenderer) Open(w *Writer, m prosemirror.Mark) error {
	attrs, err := m.LinkAttrs()
	if err != nil {
		return fmt.Errorf("cannot render link mark: %w", err)
	}

	w.WriteString("<a")
	w.WriteAttr("href", prosemirror.SafeLinkHref(attrs.Href))

	if attrs.Target != nil {
		w.WriteAttr("target", *attrs.Target)
	}

	rel := prosemirror.LinkRelToEmit(attrs)
	if rel != "" {
		w.WriteAttr("rel", rel)
	}

	if attrs.Class != nil {
		w.WriteAttr("class", *attrs.Class)
	}

	if attrs.Title != nil {
		w.WriteAttr("title", *attrs.Title)
	}

	if err := w.WriteByte('>'); err != nil {
		return fmt.Errorf("cannot write link mark: %w", err)
	}

	return nil
}

func (linkRenderer) Close(w *Writer, _ prosemirror.Mark) {
	w.WriteString("</a>")
}
