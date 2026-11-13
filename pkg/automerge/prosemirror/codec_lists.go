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

// List items share the same shape: their wrapper is transparent, their leading
// paragraph is always folded into the item on parse, and they render back into a
// single leading paragraph followed by any nested blocks. Role is roleContainer
// so the block accepts inline content and is allowed under any ancestor; the
// wrapper is reached on parse through Outer and recreated on render through the
// returned list type.

type orderedListItemCodec struct{ baseCodec }

func (orderedListItemCodec) Node() prosemirror.NodeType { return prosemirror.NodeListItem }

func (orderedListItemCodec) Block() string { return blockTypeOrderedListItem }

func (orderedListItemCodec) Role() blockRole { return roleContainer }

func (orderedListItemCodec) Outer() prosemirror.NodeType { return prosemirror.NodeOrderedList }

func (c orderedListItemCodec) Render(source block, children []node) ([]node, string) {
	return renderListItem(blockTypeOrderedListItem, nil, source, children), string(c.Outer())
}

type unorderedListItemCodec struct{ baseCodec }

func (unorderedListItemCodec) Node() prosemirror.NodeType { return prosemirror.NodeListItem }

func (unorderedListItemCodec) Block() string { return blockTypeUnorderedListItem }

func (unorderedListItemCodec) Role() blockRole { return roleContainer }

func (unorderedListItemCodec) Outer() prosemirror.NodeType { return prosemirror.NodeBulletList }

func (c unorderedListItemCodec) Render(source block, children []node) ([]node, string) {
	return renderListItem(blockTypeUnorderedListItem, nil, source, children), string(c.Outer())
}

type taskListItemCodec struct{ baseCodec }

func (taskListItemCodec) Node() prosemirror.NodeType { return prosemirror.NodeTaskItem }

func (taskListItemCodec) Block() string { return blockTypeTaskListItem }

func (taskListItemCodec) Role() blockRole { return roleContainer }

func (taskListItemCodec) Outer() prosemirror.NodeType { return prosemirror.NodeTaskList }

func (taskListItemCodec) ToAttrs(attrs map[string]any) map[string]any {
	return map[string]any{"checked": boolAttribute(attrs, "checked")}
}

func (c taskListItemCodec) Render(source block, children []node) ([]node, string) {
	attrs := map[string]any{"checked": boolAttribute(source.Attrs, "checked")}

	return renderListItem(blockTypeTaskListItem, attrs, source, children), string(c.Outer())
}

func renderListItem(blockType string, attrs map[string]any, source block, children []node) []node {
	paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

	return []node{{
		Type:    blockNodeNames[blockType],
		Attrs:   attrs,
		Content: append([]node{paragraph}, children...),
	}}
}
