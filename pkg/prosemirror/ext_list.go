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

type bulletListNode struct{}

// BulletList is an unordered list of list items.
func BulletList() NodeSpec {
	return bulletListNode{}
}

func (bulletListNode) Name() NodeType {
	return NodeBulletList
}

func (bulletListNode) Group() Group {
	return GroupBlock
}

func (bulletListNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeListItem}, Min: 1}
}

func (bulletListNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type orderedListNode struct{}

// OrderedListAttrs contains attributes for ordered list nodes.
type OrderedListAttrs struct {
	Start int     `json:"start"`
	Type  *string `json:"type"`
}

// OrderedListAttrs parses and returns the ordered list attributes.
func (n Node) OrderedListAttrs() (OrderedListAttrs, error) {
	return decodeAttrs(n.Attrs, OrderedListAttrs{Start: 1}, "ordered list")
}

// OrderedList is a numbered list of list items.
func OrderedList() NodeSpec {
	return orderedListNode{}
}

func (orderedListNode) Name() NodeType {
	return NodeOrderedList
}

func (orderedListNode) Group() Group {
	return GroupBlock
}

func (orderedListNode) Content() ContentRule {
	return ContentRule{Types: []NodeType{NodeListItem}, Min: 1}
}

func (orderedListNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

type listItemNode struct{}

// ListItem is one entry of a bullet or ordered list. It starts with a
// paragraph and may hold further block content.
func ListItem() NodeSpec {
	return listItemNode{}
}

func (listItemNode) Name() NodeType {
	return NodeListItem
}

func (listItemNode) Group() Group {
	return GroupNone
}

func (listItemNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, First: NodeParagraph}
}

func (listItemNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}
