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
)

func validateDocument(n Node) error {
	if n.Type != NodeDoc {
		return fmt.Errorf("document content root must be type %q", NodeDoc)
	}

	return validateNode(n)
}

func validateNode(n Node) error {
	switch n.Type {
	case NodeDoc:
		return validateChildren(n, "block", isBlock, 1)
	case NodeParagraph:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "inline", isInline, 0)
	case NodeHeading:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		attrs, err := n.HeadingAttrs()
		if err != nil {
			return fmt.Errorf("cannot validate heading node: %w", err)
		}

		if attrs.Level < 1 || attrs.Level > 6 {
			return fmt.Errorf("cannot validate heading node: invalid level %d", attrs.Level)
		}

		return validateChildren(n, "inline", isInline, 0)
	case NodeBlockquote:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "block", isBlock, 1)
	case NodeCodeBlock:
		return validateCodeBlock(n)
	case NodeHorizontalRule, NodeImage, NodeHardBreak:
		return validateLeaf(n)
	case NodeText:
		return validateText(n)
	case NodeBulletList, NodeOrderedList:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "listItem", isListItem, 1)
	case NodeListItem:
		return validateListItem(n)
	case NodeTable:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "tableRow", isTableRow, 1)
	case NodeTableRow:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "tableCell", isTableCell, 1)
	case NodeTableCell, NodeTableHeader:
		if err := rejectNonTextPayload(n); err != nil {
			return err
		}

		return validateChildren(n, "block", isBlock, 1)
	default:
		return fmt.Errorf("cannot validate node: unknown type %q", n.Type)
	}
}

func validateChildren(n Node, expected string, allowed func(NodeType) bool, min int) error {
	if len(n.Content) < min {
		return fmt.Errorf("%s must contain at least %d %s", n.Type, min, expected)
	}

	for _, child := range n.Content {
		if !allowed(child.Type) {
			return fmt.Errorf("invalid child type %q in %s (expected %s)", child.Type, n.Type, expected)
		}

		if err := validateNode(child); err != nil {
			return err
		}
	}

	return nil
}

func validateListItem(n Node) error {
	if err := rejectNonTextPayload(n); err != nil {
		return err
	}

	if len(n.Content) == 0 || n.Content[0].Type != NodeParagraph {
		return fmt.Errorf("listItem must start with a paragraph")
	}

	for i, child := range n.Content {
		if i > 0 && !isBlock(child.Type) {
			return fmt.Errorf("invalid child type %q in listItem (expected block)", child.Type)
		}

		if err := validateNode(child); err != nil {
			return err
		}
	}

	return nil
}

func validateCodeBlock(n Node) error {
	if err := rejectNonTextPayload(n); err != nil {
		return err
	}

	for _, child := range n.Content {
		if child.Type != NodeText {
			return fmt.Errorf("invalid child type %q in codeBlock (expected text)", child.Type)
		}

		if len(child.Marks) > 0 {
			return fmt.Errorf("codeBlock text cannot have marks")
		}

		if err := validateText(child); err != nil {
			return err
		}
	}

	return nil
}

func validateLeaf(n Node) error {
	if len(n.Content) > 0 {
		return fmt.Errorf("%s cannot have children", n.Type)
	}

	return rejectNonTextPayload(n)
}

func validateText(n Node) error {
	if n.Text == nil {
		return fmt.Errorf("text node text is nil")
	}

	if len(n.Content) > 0 {
		return fmt.Errorf("text node cannot have children")
	}

	for _, mark := range n.Marks {
		switch mark.Type {
		case MarkStrong, MarkEm, MarkUnderline, MarkStrike, MarkCode, MarkLink:
		default:
			return fmt.Errorf("cannot validate mark: unknown type %q", mark.Type)
		}
	}

	return nil
}

func rejectNonTextPayload(n Node) error {
	if n.Text != nil {
		return fmt.Errorf("%s cannot have text", n.Type)
	}

	if len(n.Marks) > 0 {
		return fmt.Errorf("%s cannot have marks", n.Type)
	}

	return nil
}

func isBlock(t NodeType) bool {
	switch t {
	case NodeParagraph, NodeHeading, NodeBlockquote, NodeCodeBlock,
		NodeHorizontalRule, NodeBulletList, NodeOrderedList, NodeTable, NodeImage:
		return true
	default:
		return false
	}
}

func isInline(t NodeType) bool {
	return t == NodeText || t == NodeHardBreak
}

func isListItem(t NodeType) bool {
	return t == NodeListItem
}

func isTableRow(t NodeType) bool {
	return t == NodeTableRow
}

func isTableCell(t NodeType) bool {
	return t == NodeTableCell || t == NodeTableHeader
}
