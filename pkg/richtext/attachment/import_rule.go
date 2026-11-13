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

package attachment

import (
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
)

type importRule struct{}

// ImportRule returns a markdown import rule that rewrites uploaded file URLs
// into image FileID attrs and standalone file blocks.
func ImportRule() markdown.Rule {
	return importRule{}
}

func (importRule) Rewrite(n *prosemirror.Node) {
	if n.Type == prosemirror.NodeImage {
		attrs, err := n.ImageAttrs()
		if err != nil {
			return
		}

		if id, ok := ParseID(attrs.Src); ok {
			attrs.FileID = &id
			attrs.Src = ""
			_ = n.SetAttrs(attrs)
		}

		return
	}

	if len(n.Content) == 0 {
		return
	}

	out := make([]prosemirror.Node, 0, len(n.Content))
	for _, child := range n.Content {
		if child.Type == prosemirror.NodeParagraph {
			if file, ok := AttachmentFromParagraph(child); ok {
				out = append(out, file)
				continue
			}
		}

		out = append(out, child)
	}

	n.Content = out
}
