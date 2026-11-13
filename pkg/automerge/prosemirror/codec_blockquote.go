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

import "go.probo.inc/probo/pkg/prosemirror"

type blockquoteCodec struct{ baseCodec }

func (blockquoteCodec) Node() prosemirror.NodeType { return prosemirror.NodeBlockquote }

func (blockquoteCodec) Block() string { return blockTypeBlockquote }

func (blockquoteCodec) Role() blockRole { return roleContainer }

func (blockquoteCodec) Render(source block, children []node) ([]node, string) {
	content := children
	if len(source.Content) > 0 || len(children) == 0 {
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}
		content = append([]node{paragraph}, children...)
	}

	return []node{{
		Type:    blockNodeNames[blockTypeBlockquote],
		Content: content,
	}}, ""
}
