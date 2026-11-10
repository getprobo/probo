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
	"strconv"

	"go.probo.inc/probo/pkg/prosemirror"
)

type headingRenderer struct{}

func Heading() NodeRenderer { return headingRenderer{} }

func (headingRenderer) Node() prosemirror.NodeType { return prosemirror.NodeHeading }

func (headingRenderer) Render(w *Writer, n prosemirror.Node) error {
	attrs, err := n.HeadingAttrs()
	if err != nil {
		return fmt.Errorf("cannot render heading node: %w", err)
	}

	if attrs.Level < 1 || attrs.Level > 6 {
		return fmt.Errorf("cannot render heading node: invalid level %d", attrs.Level)
	}

	level := strconv.Itoa(attrs.Level)

	w.WriteString("<h")
	w.WriteString(level)

	if err := w.WriteByte('>'); err != nil {
		return fmt.Errorf("cannot write heading node: %w", err)
	}

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</h")
	w.WriteString(level)

	if err := w.WriteByte('>'); err != nil {
		return fmt.Errorf("cannot close heading node: %w", err)
	}

	return nil
}
