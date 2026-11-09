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

package mention

import (
	"fmt"

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/html"
)

type htmlRenderer struct{}

// HTMLNode returns the HTML renderer for mention nodes.
func HTMLNode() html.NodeRenderer { return htmlRenderer{} }

func (htmlRenderer) Node() prosemirror.NodeType { return prosemirror.NodeMention }

func (htmlRenderer) Render(w *html.Writer, n prosemirror.Node) error {
	attrs, err := n.MentionAttrs()
	if err != nil {
		return fmt.Errorf("cannot render mention node: %w", err)
	}

	w.WriteString(`<span class="mention" data-type="mention"`)
	w.WriteAttr("data-id", attrs.ID)
	w.WriteAttr("data-label", attrs.Label)
	w.WriteByte('>')
	w.WriteEscaped("@" + attrs.Label)
	w.WriteString("</span>")

	return nil
}

// HTMLNodes returns HTML renderers for mention nodes.
func HTMLNodes() []html.NodeRenderer {
	return []html.NodeRenderer{HTMLNode()}
}
