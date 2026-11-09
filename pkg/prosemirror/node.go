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
	"fmt"
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
		Text    string          `json:"text,omitempty"`
		Attrs   json.RawMessage `json:"attrs,omitempty"`
	}

	// Mark represents inline formatting applied to a text node.
	Mark struct {
		Type  MarkType        `json:"type"`
		Attrs json.RawMessage `json:"attrs,omitempty"`
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
	NodeFile           NodeType = "file"
	NodeMention        NodeType = "mention"
	NodeBulletList     NodeType = "bulletList"
	NodeOrderedList    NodeType = "orderedList"
	NodeListItem       NodeType = "listItem"
	NodeTaskList       NodeType = "taskList"
	NodeTaskItem       NodeType = "taskItem"
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
