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
	"strings"
)

const maxDocumentDecodeDepth = 4

// RecoverDocument parses stored rich text into a ProseMirror document.
//
// Task content is sometimes saved as the JSON of a document: encoded as a
// JSON string, or split into paragraphs and code blocks whose text is that
// JSON. Those shapes display the serialized document. When the payload is
// one of those shapes and the inner value is a valid document,
// RecoverDocument returns that document so callers can render it as text.
func RecoverDocument(s string) (Node, error) {
	node, err := decodeDocument(s, 0)
	if err != nil {
		return Node{}, err
	}

	return unwrapLiteralDocument(node, 0), nil
}

func decodeDocument(s string, depth int) (Node, error) {
	if depth > maxDocumentDecodeDepth {
		return Node{}, fmt.Errorf("cannot parse prosemirror node: document encoding is too deep")
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return Node{}, fmt.Errorf("cannot parse prosemirror node: empty content")
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return Node{}, fmt.Errorf("cannot parse prosemirror node: %w", err)
	}

	raw = trimRawJSONSpace(raw)
	if len(raw) > 0 && raw[0] == '"' {
		var inner string
		if err := json.Unmarshal(raw, &inner); err != nil {
			return Node{}, fmt.Errorf("cannot parse prosemirror node: %w", err)
		}

		return decodeDocument(inner, depth+1)
	}

	var node Node
	if err := json.Unmarshal(raw, &node); err != nil {
		return Node{}, fmt.Errorf("cannot parse prosemirror node: %w", err)
	}

	if node.Type != NodeDoc {
		return Node{}, fmt.Errorf("document content root must be type %q", NodeDoc)
	}

	return node, nil
}

func trimRawJSONSpace(raw json.RawMessage) json.RawMessage {
	return json.RawMessage(strings.TrimSpace(string(raw)))
}

func unwrapLiteralDocument(node Node, depth int) Node {
	if depth >= maxDocumentDecodeDepth {
		return node
	}

	text, ok := literalDocumentText(node)
	if !ok {
		return node
	}

	text = strings.TrimSpace(text)
	if text == "" || (text[0] != '{' && text[0] != '"') {
		return node
	}

	inner, err := decodeDocument(text, 0)
	if err != nil {
		return node
	}

	if err := validateDocument(inner); err != nil {
		return node
	}

	return unwrapLiteralDocument(inner, depth+1)
}

func literalDocumentText(node Node) (string, bool) {
	if node.Type != NodeDoc {
		return "", false
	}

	var b strings.Builder

	for _, child := range node.Content {
		switch child.Type {
		case NodeParagraph, NodeCodeBlock:
			writeLiteralInline(&b, child)
			b.WriteByte('\n')
		default:
			return "", false
		}
	}

	return b.String(), true
}

func writeLiteralInline(b *strings.Builder, node Node) {
	for _, child := range node.Content {
		switch child.Type {
		case NodeText:
			if child.Text != nil {
				b.WriteString(*child.Text)
			}
		case NodeHardBreak:
			b.WriteByte('\n')
		default:
			writeLiteralInline(b, child)
		}
	}
}
