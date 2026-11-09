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

import "fmt"

type tableNode struct{}

// Table holds one or more table rows.
func Table() NodeSpec {
	return tableNode{}
}

func (tableNode) Name() NodeType {
	return NodeTable
}

func (tableNode) Group() Group {
	return GroupBlock
}

func (tableNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeTableRow}, Min: 1}
}

func (tableNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type tableRowNode struct{}

// TableRow holds one or more table cells or headers.
func TableRow() NodeSpec {
	return tableRowNode{}
}

func (tableRowNode) Name() NodeType {
	return NodeTableRow
}

func (tableRowNode) Group() Group {
	return GroupNone
}

func (tableRowNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeTableCell, NodeTableHeader}, Min: 1}
}

func (tableRowNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type tableCellNode struct{}

// TableCellAttrs contains attributes for table cell and table header nodes.
type TableCellAttrs struct {
	Colspan  int   `json:"colspan"`
	Rowspan  int   `json:"rowspan"`
	Colwidth []int `json:"colwidth"`
}

// TableCellAttrs parses and returns the table cell/header attributes.
func (n Node) TableCellAttrs() (TableCellAttrs, error) {
	return decodeAttrs(n.Attrs, TableCellAttrs{Colspan: 1, Rowspan: 1}, "table cell")
}

// TableCell is a standard table cell holding block content.
func TableCell() NodeSpec {
	return tableCellNode{}
}

func (tableCellNode) Name() NodeType {
	return NodeTableCell
}

func (tableCellNode) Group() Group {
	return GroupNone
}

func (tableCellNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (tableCellNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type tableHeaderNode struct{}

// TableHeader is a header table cell holding block content.
func TableHeader() NodeSpec {
	return tableHeaderNode{}
}

func (tableHeaderNode) Name() NodeType {
	return NodeTableHeader
}

func (tableHeaderNode) Group() Group {
	return GroupNone
}

func (tableHeaderNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (tableHeaderNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}
