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
	for i := range n.Marks {
		sanitizeLinkMark(&n.Marks[i])
	}
	for i := range n.Content {
		sanitizeNode(&n.Content[i])
	}
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
