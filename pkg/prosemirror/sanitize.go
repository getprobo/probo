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

// ValidateDocumentContentJSON returns nil if s is empty or whitespace-only.
// Otherwise s must be valid ProseMirror JSON whose root node has type "doc".
func ValidateDocumentContentJSON(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	_, err := parseDocRoot(s)

	return err
}

func parseDocRoot(s string) (Node, error) {
	n, err := Parse(s)
	if err != nil {
		return Node{}, fmt.Errorf("cannot parse document content as ProseMirror JSON: %w", err)
	}

	if n.Type != NodeDoc {
		return Node{}, fmt.Errorf("document content root must be type %q", NodeDoc)
	}

	return n, nil
}

// SanitizeDocumentJSON parses a ProseMirror/Tiptap JSON document, replaces
// unsafe link mark href values using the same rules as RenderHTML, and
// re-serializes the document. Whitespace-only input is returned unchanged.
// Non-empty content must be valid JSON whose root node has type "doc".
func SanitizeDocumentJSON(s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return s, nil
	}

	n, err := parseDocRoot(s)
	if err != nil {
		return "", err
	}

	sanitizeNode(&n)

	out, err := json.Marshal(n)
	if err != nil {
		return "", fmt.Errorf("cannot marshal sanitized document: %w", err)
	}

	return string(out), nil
}

func sanitizeNode(n *Node) {
	if n.Type == NodeImage {
		sanitizeImageNode(n)
	}

	for i := range n.Marks {
		sanitizeLinkMark(&n.Marks[i])
	}

	for i := range n.Content {
		sanitizeNode(&n.Content[i])
	}
}

func sanitizeImageNode(n *Node) {
	attrs, err := n.ImageAttrs()
	if err != nil {
		n.Attrs = []byte(`{"src":""}`)
		return
	}

	attrs.Src = safeImageSrc(attrs.Src)

	raw, err := json.Marshal(attrs)
	if err != nil {
		n.Attrs = []byte(`{"src":""}`)
		return
	}

	n.Attrs = raw
}

func sanitizeLinkMark(m *Mark) {
	if m.Type != MarkLink {
		return
	}

	attrs, err := m.LinkAttrs()
	if err != nil {
		m.Attrs = []byte(`{"href":"#"}`)
		return
	}

	attrs.Href = safeLinkHref(attrs.Href)

	raw, err := json.Marshal(attrs)
	if err != nil {
		m.Attrs = []byte(`{"href":"#"}`)
		return
	}

	m.Attrs = raw
}
