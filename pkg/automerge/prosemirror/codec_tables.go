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

type tableCodec struct{ baseCodec }

func (tableCodec) Node() prosemirror.NodeType { return prosemirror.NodeTable }

func (tableCodec) Block() string { return blockTypeTable }

func (tableCodec) Role() blockRole { return roleStructural }

func (tableCodec) Render(source block, children []node) ([]node, string) {
	rows, spills := partitionNodes(children, blockNodeNames[blockTypeTableRow])
	rendered := make([]node, 0, 2+len(spills))

	if len(rows) > 0 {
		rendered = append(
			rendered,
			node{Type: blockNodeNames[blockTypeTable], Content: rows},
		)
	}

	if len(source.Content) > 0 {
		rendered = append(
			rendered,
			node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
		)
	}

	return append(rendered, hoistNodes(spills)...), ""
}

type tableRowCodec struct{ baseCodec }

func (tableRowCodec) Node() prosemirror.NodeType { return prosemirror.NodeTableRow }

func (tableRowCodec) Block() string { return blockTypeTableRow }

func (tableRowCodec) Role() blockRole { return roleStructural }

func (tableRowCodec) AllowedUnder(parents []string) bool {
	return immediateParent(parents) == blockTypeTable
}

func (tableRowCodec) Render(source block, children []node) ([]node, string) {
	cells, spills := partitionNodes(
		children,
		blockNodeNames[blockTypeTableCell],
		blockNodeNames[blockTypeTableHeader],
	)
	rendered := make([]node, 0, 2+len(spills))

	if len(cells) > 0 {
		rendered = append(
			rendered,
			node{Type: blockNodeNames[blockTypeTableRow], Content: cells},
		)
	}

	if len(source.Content) > 0 {
		rendered = append(
			rendered,
			node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
		)
	}

	return append(rendered, hoistNodes(spills)...), ""
}

type tableCellCodec struct{ baseCodec }

func (tableCellCodec) Node() prosemirror.NodeType { return prosemirror.NodeTableCell }

func (tableCellCodec) Block() string { return blockTypeTableCell }

func (tableCellCodec) Role() blockRole { return roleContainer }

func (tableCellCodec) AllowedUnder(parents []string) bool {
	return immediateParent(parents) == blockTypeTableRow
}

func (tableCellCodec) ToAttrs(attrs map[string]any) map[string]any {
	return cellAttributes(attrs)
}

func (tableCellCodec) Render(source block, children []node) ([]node, string) {
	return renderCell(blockNodeNames[blockTypeTableCell], source, children), ""
}

type tableHeaderCodec struct{ baseCodec }

func (tableHeaderCodec) Node() prosemirror.NodeType { return prosemirror.NodeTableHeader }

func (tableHeaderCodec) Block() string { return blockTypeTableHeader }

func (tableHeaderCodec) Role() blockRole { return roleContainer }

func (tableHeaderCodec) AllowedUnder(parents []string) bool {
	return immediateParent(parents) == blockTypeTableRow
}

func (tableHeaderCodec) ToAttrs(attrs map[string]any) map[string]any {
	return cellAttributes(attrs)
}

func (tableHeaderCodec) Render(source block, children []node) ([]node, string) {
	return renderCell(blockNodeNames[blockTypeTableHeader], source, children), ""
}

func cellAttributes(attrs map[string]any) map[string]any {
	cellAttrs := map[string]any{
		"colspan": intAttribute(attrs, "colspan", 1),
		"rowspan": intAttribute(attrs, "rowspan", 1),
	}

	if colwidth := numberSliceAttribute(attrs, "colwidth"); colwidth != nil {
		cellAttrs["colwidth"] = colwidth
	}

	return cellAttrs
}

func renderCell(nodeType string, source block, children []node) []node {
	content := children
	if len(source.Content) > 0 || len(children) == 0 {
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}
		content = append([]node{paragraph}, children...)
	}

	attrs := map[string]any{
		"colspan":  intAttribute(source.Attrs, "colspan", 1),
		"rowspan":  intAttribute(source.Attrs, "rowspan", 1),
		"colwidth": intSliceAttribute(source.Attrs, "colwidth"),
	}

	return []node{{
		Type:    nodeType,
		Attrs:   attrs,
		Content: content,
	}}
}

func immediateParent(parents []string) string {
	if len(parents) == 0 {
		return ""
	}

	return parents[len(parents)-1]
}
