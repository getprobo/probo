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

package probo

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

const thirdPartyRegisterNotesHeadingDemote = 3

// RewriteThirdPartyRegisterNotesContent rewrites third-party register
// ProseMirror JSON that still embeds risk-assessment notes as a single raw
// Markdown text node into structured blocks (headings, tables, bold, …).
// Returns the rewritten JSON and whether any notes paragraph was converted.
func RewriteThirdPartyRegisterNotesContent(content string) (string, bool, error) {
	doc, err := prosemirror.Parse(content)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse register document: %w", err)
	}

	if doc.Type != prosemirror.NodeDoc {
		return "", false, fmt.Errorf("cannot rewrite register document: root type %q is not doc", doc.Type)
	}

	rewritten, changed, err := rewriteThirdPartyRegisterNotesNodes(doc.Content)
	if err != nil {
		return "", false, err
	}

	if !changed {
		return content, false, nil
	}

	doc.Content = rewritten

	out, err := json.Marshal(doc)
	if err != nil {
		return "", false, fmt.Errorf("cannot marshal rewritten register document: %w", err)
	}

	return string(out), true, nil
}

func rewriteThirdPartyRegisterNotesNodes(nodes []prosemirror.Node) ([]prosemirror.Node, bool, error) {
	out := make([]prosemirror.Node, 0, len(nodes))
	changed := false

	for _, node := range nodes {
		notes, ok := rawMarkdownNotesFromParagraph(node)
		if !ok {
			out = append(out, node)
			continue
		}

		blocks, err := markdownNotesToProseMirrorNodes(notes)
		if err != nil {
			return nil, false, fmt.Errorf("cannot convert notes markdown: %w", err)
		}

		out = append(out, notesLabelParagraph())
		out = append(out, blocks...)
		changed = true
	}

	return out, changed, nil
}

func notesLabelParagraph() prosemirror.Node {
	label := "Notes: "

	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{
				Type:  prosemirror.NodeText,
				Text:  &label,
				Marks: []prosemirror.Mark{{Type: prosemirror.MarkStrong}},
			},
		},
	}
}

// rawMarkdownNotesFromParagraph detects the legacy register shape:
// a paragraph with bold "Notes: " followed by a single Markdown text blob.
func rawMarkdownNotesFromParagraph(node prosemirror.Node) (string, bool) {
	if node.Type != prosemirror.NodeParagraph || len(node.Content) < 2 {
		return "", false
	}

	label := node.Content[0]
	if label.Type != prosemirror.NodeText || label.Text == nil || *label.Text != "Notes: " {
		return "", false
	}

	hasBold := false

	for _, mark := range label.Marks {
		if mark.Type == prosemirror.MarkStrong {
			hasBold = true
			break
		}
	}

	if !hasBold {
		return "", false
	}

	var notes strings.Builder

	for _, child := range node.Content[1:] {
		if child.Type != prosemirror.NodeText || child.Text == nil {
			return "", false
		}

		notes.WriteString(*child.Text)
	}

	text := notes.String()
	if !looksLikeMarkdownNotes(text) {
		return "", false
	}

	return text, true
}

func looksLikeMarkdownNotes(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || trimmed == "—" || trimmed == "Not specified" {
		return false
	}

	if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
		return true
	}

	if strings.Contains(s, "\n# ") || strings.Contains(s, "\n## ") {
		return true
	}

	if strings.Contains(s, "**") {
		return true
	}

	if strings.Contains(s, "\n|") && strings.Contains(s, "|") {
		return true
	}

	if strings.Contains(s, "\n- ") || strings.Contains(s, "\n* ") {
		return true
	}

	return false
}

// proseMirrorBlocksFromMarkdownNotes converts markdown risk-assessment notes
// into a comma-separated sequence of ProseMirror block nodes for splicing
// into the third-party register JSON template. Headings are demoted so they
// nest under the register's h3 "Risk Assessments" sections (h1→h4, …).
func proseMirrorBlocksFromMarkdownNotes(notes string) (string, error) {
	nodes, err := markdownNotesToProseMirrorNodes(notes)
	if err != nil {
		return "", err
	}

	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		raw, err := json.Marshal(node)
		if err != nil {
			return "", fmt.Errorf("cannot marshal notes block: %w", err)
		}

		parts = append(parts, string(raw))
	}

	return strings.Join(parts, ","), nil
}

func markdownNotesToProseMirrorNodes(notes string) ([]prosemirror.Node, error) {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		notes = "—"
	}

	doc, err := prosemirror.ParseMarkdown(notes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse notes markdown: %w", err)
	}

	if len(doc.Content) == 0 {
		fallback := notes
		doc.Content = []prosemirror.Node{
			{
				Type: prosemirror.NodeParagraph,
				Content: []prosemirror.Node{
					{Type: prosemirror.NodeText, Text: &fallback},
				},
			},
		}
	}

	if err := demoteProseMirrorHeadingLevels(doc.Content, thirdPartyRegisterNotesHeadingDemote); err != nil {
		return nil, fmt.Errorf("cannot demote notes headings: %w", err)
	}

	return doc.Content, nil
}

func demoteProseMirrorHeadingLevels(nodes []prosemirror.Node, delta int) error {
	if delta <= 0 {
		return nil
	}

	for i := range nodes {
		if nodes[i].Type == prosemirror.NodeHeading {
			attrs, err := nodes[i].HeadingAttrs()
			if err != nil {
				return fmt.Errorf("cannot parse heading attrs: %w", err)
			}

			level := min(attrs.Level+delta, 6)

			raw, err := json.Marshal(prosemirror.HeadingAttrs{Level: level})
			if err != nil {
				return fmt.Errorf("cannot marshal heading attrs: %w", err)
			}

			nodes[i].Attrs = raw
		}

		if err := demoteProseMirrorHeadingLevels(nodes[i].Content, delta); err != nil {
			return err
		}
	}

	return nil
}
