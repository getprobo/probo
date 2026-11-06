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
	"fmt"
	"strconv"
	"strings"
)

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

func (bulletListNode) RenderHTML(w *HTMLWriter, n Node) error {
	w.WriteString("<ul>")

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</ul>")

	return nil
}

func (bulletListNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	tight, err := listTightness(n, NodeListItem)
	if err != nil {
		return fmt.Errorf("cannot render bullet list: %w", err)
	}

	return w.renderListItems(n.Content, tight, func(_ int, _ Node) (string, string, error) {
		return "- ", "  ", nil
	})
}

type orderedListNode struct{}

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

func (orderedListNode) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.OrderedListAttrs()
	if err != nil {
		return fmt.Errorf("cannot render ordered list node: %w", err)
	}

	w.WriteString("<ol")

	if attrs.Start != 1 {
		w.WriteAttr("start", strconv.Itoa(attrs.Start))
	}

	if attrs.Type != nil {
		w.WriteAttr("type", *attrs.Type)
	}

	w.WriteByte('>')

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</ol>")

	return nil
}

func (orderedListNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	attrs, err := n.OrderedListAttrs()
	if err != nil {
		return fmt.Errorf("cannot render ordered list node: %w", err)
	}

	tight, err := listTightness(n, NodeListItem)
	if err != nil {
		return fmt.Errorf("cannot render ordered list: %w", err)
	}

	start := max(attrs.Start, 1)

	return w.renderListItems(n.Content, tight, func(i int, _ Node) (string, string, error) {
		num := strconv.Itoa(start + i)

		return num + ". ", strings.Repeat(" ", len(num)+2), nil
	})
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

func (listItemNode) RenderHTML(w *HTMLWriter, n Node) error {
	w.WriteString("<li>")

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</li>")

	return nil
}

func (listItemNode) RenderMarkdown(_ *MarkdownWriter, _ Node) error {
	return fmt.Errorf("cannot render list item outside of list context")
}
