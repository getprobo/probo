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

type headingCodec struct{ baseCodec }

func (headingCodec) Node() prosemirror.NodeType { return prosemirror.NodeHeading }

func (headingCodec) Block() string { return blockTypeHeading }

func (headingCodec) Role() blockRole { return roleTextblock }

func (headingCodec) ToAttrs(attrs map[string]any) map[string]any {
	level := intAttribute(attrs, "level", 1)
	if level < 1 || level > 6 {
		level = 1
	}

	return map[string]any{"level": level}
}

func (headingCodec) Render(source block, children []node) ([]node, string) {
	level := intAttribute(source.Attrs, "level", 1)
	if level < 1 || level > 6 {
		level = 1
	}

	rendered := node{
		Type:    blockNodeNames[blockTypeHeading],
		Attrs:   map[string]any{"level": level},
		Content: source.Content,
	}

	return append([]node{rendered}, children...), ""
}
