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

package cmdutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

// FormatRichText renders stored ProseMirror JSON as markdown for human CLI
// output. Empty or whitespace-only values return "".
func FormatRichText(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse content: %w", err)
	}

	markdown, err := prosemirror.RenderMarkdown(node)
	if err != nil {
		return "", fmt.Errorf("cannot render content: %w", err)
	}

	return strings.TrimRight(markdown, "\n"), nil
}

// CLIContent stores editor text. Valid ProseMirror document JSON is kept.
// Plain text keeps one paragraph per line. Markdown, including headings,
// lists, and links to /api/files/v1/{id}, is parsed as markdown.
func CLIContent(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") && prosemirror.ValidateDocumentContentJSON(text) == nil {
		return text, nil
	}

	if trimmed == "" {
		return prosemirror.FromPlainText(""), nil
	}

	node, err := prosemirror.ParseMarkdown(text)
	if err != nil {
		return "", fmt.Errorf("cannot parse markdown: %w", err)
	}

	if cliPlainParagraphs(node) && strings.Contains(text, "\n") {
		return prosemirror.FromPlainText(text), nil
	}

	encoded, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal content: %w", err)
	}

	return string(encoded), nil
}

func cliPlainParagraphs(doc prosemirror.Node) bool {
	for _, block := range doc.Content {
		if block.Type != prosemirror.NodeParagraph {
			return false
		}

		for _, inline := range block.Content {
			if inline.Type != prosemirror.NodeText || len(inline.Marks) > 0 {
				return false
			}
		}
	}

	return true
}
