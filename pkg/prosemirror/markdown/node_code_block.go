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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

type codeBlockRenderer struct{}

func CodeBlock() NodeRenderer { return codeBlockRenderer{} }

func (codeBlockRenderer) Node() prosemirror.NodeType { return prosemirror.NodeCodeBlock }

func (codeBlockRenderer) Render(w *Writer, n prosemirror.Node) error {
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
