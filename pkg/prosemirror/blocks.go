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

import "encoding/json"

// blocksFromInlines splits images out of inline content. An image is a block,
// so "hello ![alt](url) world" becomes a paragraph, an image, and a paragraph.
func blocksFromInlines(blockType NodeType, attrs json.RawMessage, inlines []Node) []Node {
	if len(inlines) == 0 {
		return []Node{{Type: blockType, Attrs: attrs}}
	}

	var out []Node

	var buf []Node

	flush := func() {
		if len(buf) == 0 {
			return
		}

		out = append(out, Node{
			Type:    blockType,
			Attrs:   attrs,
			Content: buf,
		})
		buf = nil
	}

	for _, n := range inlines {
		if n.Type == NodeImage {
			n.Marks = nil

			flush()

			out = append(out, n)

			continue
		}

		buf = append(buf, n)
	}

	flush()

	if len(out) == 0 {
		return []Node{{Type: blockType, Attrs: attrs}}
	}

	return out
}

// ensureLeadingParagraph keeps a list item valid when its first block is an image.
func ensureLeadingParagraph(nodes []Node) []Node {
	if len(nodes) == 0 || nodes[0].Type == NodeParagraph {
		return nodes
	}

	return append([]Node{{Type: NodeParagraph}}, nodes...)
}
