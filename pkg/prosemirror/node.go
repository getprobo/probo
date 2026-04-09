// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package prosemirror

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type (
	// NodeType identifies the type of a ProseMirror/Tiptap document node.
	NodeType string

	// MarkType identifies the type of an inline mark.
	MarkType string

	// Node is a ProseMirror/Tiptap document node.
	Node struct {
		Type    NodeType        `json:"type"`
		Content []Node          `json:"content,omitempty"`
		Marks   []Mark          `json:"marks,omitempty"`
		Text    *string         `json:"text,omitempty"`
		Attrs   json.RawMessage `json:"attrs,omitempty"`
	}

	// Mark represents inline formatting applied to a text node.
	Mark struct {
		Type  MarkType        `json:"type"`
		Attrs json.RawMessage `json:"attrs,omitempty"`
	}

	// HeadingAttrs contains attributes for heading nodes.
	HeadingAttrs struct {
		Level int `json:"level"`
	}

	// CodeBlockAttrs contains attributes for code block nodes.
	CodeBlockAttrs struct {
		Language *string `json:"language"`
	}

	// OrderedListAttrs contains attributes for ordered list nodes.
	OrderedListAttrs struct {
		Start int     `json:"start"`
		Type  *string `json:"type"`
	}

	// ImageAttrs contains attributes for image nodes.
	ImageAttrs struct {
		Src   string  `json:"src"`
		Alt   *string `json:"alt"`
		Title *string `json:"title"`
	}

	// TableCellAttrs contains attributes for table cell and table header nodes.
	TableCellAttrs struct {
		Colspan  int   `json:"colspan"`
		Rowspan  int   `json:"rowspan"`
		Colwidth []int `json:"colwidth"`
	}

	// LinkAttrs contains attributes for link marks.
	LinkAttrs struct {
		Href   string  `json:"href"`
		Target *string `json:"target"`
		Rel    *string `json:"rel"`
		Class  *string `json:"class"`
		Title  *string `json:"title"`
	}
)

// Node type constants. String values match the Tiptap JSON the frontend produces.
// Mark type constants. Go names follow ProseMirror conventions (Strong, Em);
// string values match Tiptap JSON (bold, italic).
const (
	NodeDoc            NodeType = "doc"
	NodeParagraph      NodeType = "paragraph"
	NodeBlockquote     NodeType = "blockquote"
	NodeHeading        NodeType = "heading"
	NodeCodeBlock      NodeType = "codeBlock"
	NodeHorizontalRule NodeType = "horizontalRule"
	NodeHardBreak      NodeType = "hardBreak"
	NodeText           NodeType = "text"
	NodeImage          NodeType = "image"
	NodeBulletList     NodeType = "bulletList"
	NodeOrderedList    NodeType = "orderedList"
	NodeListItem       NodeType = "listItem"
	NodeTable          NodeType = "table"
	NodeTableRow       NodeType = "tableRow"
	NodeTableCell      NodeType = "tableCell"
	NodeTableHeader    NodeType = "tableHeader"

	MarkStrong    MarkType = "bold"
	MarkEm        MarkType = "italic"
	MarkUnderline MarkType = "underline"
	MarkStrike    MarkType = "strike"
	MarkCode      MarkType = "code"
	MarkLink      MarkType = "link"
)

// Parse unmarshals a ProseMirror/Tiptap JSON document string into a Node tree.
// Nested content and marks are unmarshaled into Go structs recursively.
func Parse(s string) (Node, error) {
	var n Node
	if err := json.Unmarshal([]byte(s), &n); err != nil {
		return Node{}, fmt.Errorf("cannot parse prosemirror node: %w", err)
	}
	return n, nil
}

// HeadingAttrs parses and returns the heading attributes from a heading node.
func (n Node) HeadingAttrs() (HeadingAttrs, error) {
	var a HeadingAttrs
	if err := json.Unmarshal(n.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse heading attrs: %w", err)
	}
	return a, nil
}

// CodeBlockAttrs parses and returns the code block attributes.
func (n Node) CodeBlockAttrs() (CodeBlockAttrs, error) {
	if len(n.Attrs) == 0 {
		return CodeBlockAttrs{}, nil
	}
	var a CodeBlockAttrs
	if err := json.Unmarshal(n.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse code block attrs: %w", err)
	}
	return a, nil
}

// OrderedListAttrs parses and returns the ordered list attributes.
func (n Node) OrderedListAttrs() (OrderedListAttrs, error) {
	var a OrderedListAttrs
	if err := json.Unmarshal(n.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse ordered list attrs: %w", err)
	}
	return a, nil
}

// ImageAttrs parses and returns the image attributes.
func (n Node) ImageAttrs() (ImageAttrs, error) {
	var a ImageAttrs
	if err := json.Unmarshal(n.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse image attrs: %w", err)
	}
	return a, nil
}

// TableCellAttrs parses and returns the table cell/header attributes.
func (n Node) TableCellAttrs() (TableCellAttrs, error) {
	var a TableCellAttrs
	if err := json.Unmarshal(n.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse table cell attrs: %w", err)
	}
	return a, nil
}

// TextLength returns the total number of characters (Unicode code points)
// across all text nodes in the tree. Structural markup is excluded.
func (n Node) TextLength() int {
	length := 0
	if n.Text != nil {
		length += utf8.RuneCountInString(*n.Text)
	}
	for _, child := range n.Content {
		length += child.TextLength()
	}
	return length
}

// LinkAttrs parses and returns the link attributes from a link mark.
func (m Mark) LinkAttrs() (LinkAttrs, error) {
	var a LinkAttrs
	if err := json.Unmarshal(m.Attrs, &a); err != nil {
		return a, fmt.Errorf("cannot parse link attrs: %w", err)
	}
	return a, nil
}
