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
	"strings"
)

type codeBlockNode struct{}

// CodeBlock holds preformatted text, optionally tagged with a language.
func CodeBlock() NodeSpec {
	return codeBlockNode{}
}

func (codeBlockNode) Name() NodeType {
	return NodeCodeBlock
}

func (codeBlockNode) Group() Group {
	return GroupBlock
}

func (codeBlockNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeText}, RejectChildMarks: true}
}

func (codeBlockNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func (codeBlockNode) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.CodeBlockAttrs()
	if err != nil {
		return fmt.Errorf("cannot render code block node: %w", err)
	}

	if attrs.Language != nil && *attrs.Language == "mermaid" {
		w.WriteString(`<pre class="mermaid">`)

		if err := w.RenderChildren(n.Content); err != nil {
			return err
		}

		w.WriteString("</pre>")

		return nil
	}

	w.WriteString("<pre><code")

	if attrs.Language != nil {
		w.WriteAttr("class", "language-"+*attrs.Language)
	}

	w.WriteByte('>')

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</code></pre>")

	return nil
}

func (codeBlockNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	attrs, err := n.CodeBlockAttrs()
	if err != nil {
		return fmt.Errorf("cannot render code block node: %w", err)
	}

	code := collectText(n.Content)
	fence := chooseFence(code)

	w.ensurePrefix()
	w.buf.WriteString(fence)

	if attrs.Language != nil {
		w.buf.WriteString(*attrs.Language)
	}

	w.newLine()

	for line := range strings.SplitSeq(code, "\n") {
		w.ensurePrefix()
		w.buf.WriteString(line)
		w.newLine()
	}

	w.ensurePrefix()
	w.buf.WriteString(fence)
	w.newLine()

	return nil
}
