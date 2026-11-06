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
	"slices"
	"unicode/utf8"
)

type textNode struct{}

// Text is a leaf node carrying a string and inline marks.
func Text() NodeSpec {
	return textNode{}
}

func (textNode) Name() NodeType {
	return NodeText
}

func (textNode) Group() Group {
	return GroupInline
}

func (textNode) Content() ContentRule {
	return ContentRule{}
}

func (textNode) Check(n *Node, _ CheckMode) error {
	if n.Text == nil {
		return fmt.Errorf("text node text is nil")
	}

	if len(n.Content) > 0 {
		return fmt.Errorf("text node cannot have children")
	}

	return nil
}

func (textNode) TextLength(n Node) int {
	if n.Text == nil {
		return 0
	}

	return utf8.RuneCountInString(*n.Text)
}

func (textNode) RenderHTML(w *HTMLWriter, n Node) error {
	if n.Text == nil {
		return fmt.Errorf("cannot render text node: text is nil")
	}

	for _, m := range n.Marks {
		if err := w.openMark(m); err != nil {
			return err
		}
	}

	w.WriteEscaped(*n.Text)

	for _, v := range slices.Backward(n.Marks) {
		w.closeMark(v)
	}

	return nil
}

func (textNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	return w.renderText(n)
}
