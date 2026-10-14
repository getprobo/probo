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
	"encoding/json"
	"strings"
)

// blocksFromInlines splits images out of inline content. An image is a block,
// so "hello ![alt](url) world" becomes a paragraph, an image, and a paragraph.
func blocksFromInlines(blockType NodeType, attrs json.RawMessage, inlines []Node) []Node {
	if len(inlines) == 0 {
		return []Node{{Type: blockType, Attrs: attrs}}
	}

	var out []Node

	var buf []Node

	// Only the text before the first image keeps the heading. Later runs are
	// paragraphs, so one image does not turn the rest of the line into another heading.
	flushedHeading := false
	flushedText := false

	flush := func() {
		if whitespaceOnlyInlines(buf) {
			buf = nil

			return
		}

		typ := blockType
		nodeAttrs := attrs

		if blockType == NodeHeading && flushedHeading {
			typ = NodeParagraph
			nodeAttrs = nil
		}

		if typ == NodeHeading {
			flushedHeading = true
		}

		flushedText = true

		out = append(out, Node{
			Type:    typ,
			Attrs:   nodeAttrs,
			Content: buf,
		})
		buf = nil
	}

	for _, n := range inlines {
		if n.Type == NodeImage {
			n.Marks = nil

			flush()

			// The image consumes the heading, so text after it is a paragraph.
			if blockType == NodeHeading {
				flushedHeading = true
			}

			out = append(out, n)

			continue
		}

		buf = append(buf, n)
	}

	flush()

	if len(out) == 0 {
		return []Node{{Type: blockType, Attrs: attrs}}
	}

	// Text after an opening image is a paragraph, and an empty heading in
	// front of it keeps the level. An image with no text run is the whole block.
	if blockType == NodeHeading && flushedText {
		for _, n := range out {
			if n.Type == NodeHeading {
				return out
			}
		}

		return append([]Node{{Type: blockType, Attrs: attrs}}, out...)
	}

	return out
}

// ensureLeadingParagraph gives a list item the paragraph the schema requires.
func ensureLeadingParagraph(nodes []Node) []Node {
	if len(nodes) == 0 {
		return []Node{{Type: NodeParagraph}}
	}

	if nodes[0].Type == NodeParagraph {
		return nodes
	}

	return append([]Node{{Type: NodeParagraph}}, nodes...)
}

// whitespaceOnlyInlines reports a run of unmarked spaces. A space between two
// images would otherwise flush as its own paragraph.
func whitespaceOnlyInlines(nodes []Node) bool {
	for _, n := range nodes {
		if n.Type != NodeText || n.Text == nil || len(n.Marks) != 0 || strings.TrimSpace(*n.Text) != "" {
			return false
		}
	}

	return true
}
