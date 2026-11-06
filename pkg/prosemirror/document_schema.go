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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// DocumentSchema is the Tiptap-compatible schema the console editor produces.
// A mention accepts any GID, matching the current document behavior.
var DocumentSchema = NewSchema(
	[]NodeSpec{
		Doc(),
		Paragraph(),
		Text(),
		Heading(),
		Blockquote(),
		CodeBlock(),
		HorizontalRule(),
		HardBreak(),
		Image(),
		File(),
		Mention(MentionOptions{}),
		BulletList(),
		OrderedList(),
		ListItem(),
		TaskList(),
		TaskItem(),
		Table(),
		TableRow(),
		TableCell(),
		TableHeader(),
	},
	[]MarkSpec{
		Bold(),
		Italic(),
		Underline(),
		Strike(),
		Code(),
		Link(),
	},
)

// ValidateDocumentContentJSON returns nil if s is empty or whitespace-only.
// Otherwise s must be valid ProseMirror JSON whose root node has type "doc"
// and whose node tree matches the Tiptap schema, including parent-child
// content rules.
func ValidateDocumentContentJSON(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	n, err := DocumentSchema.Parse(s)
	if err != nil {
		return err
	}

	if err := DocumentSchema.Validate(n); err != nil {
		return fmt.Errorf("cannot validate document content: %w", err)
	}

	return nil
}

// ValidateCommentContentJSON returns nil if s is empty or whitespace-only.
// Otherwise s must be valid ProseMirror JSON that also stays within the
// comment allowlist: paragraphs, lists, blockquotes, hard breaks, mentions,
// and the bold/italic/strike/code/link marks.
func ValidateCommentContentJSON(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	n, err := parseDocRoot(s)
	if err != nil {
		return err
	}

	if err := validateDocument(n); err != nil {
		return fmt.Errorf("cannot validate comment document: %w", err)
	}

	if err := validateCommentAllowlist(n); err != nil {
		return fmt.Errorf("cannot validate comment allowlist: %w", err)
	}

	return nil
}

func validateCommentAllowlist(n Node) error {
	if !isCommentNode(n.Type) {
		return fmt.Errorf("node type %q is not allowed in comments", n.Type)
	}

	if n.Type == NodeMention {
		attrs, err := n.MentionAttrs()
		if err != nil {
			return fmt.Errorf("cannot validate mention node: %w", err)
		}

		parsed, err := gid.ParseGID(strings.TrimSpace(attrs.ID))
		if err != nil || parsed.EntityType() != coredata.MembershipProfileEntityType {
			return fmt.Errorf("cannot validate mention node: invalid id")
		}
	}

	for _, mark := range n.Marks {
		if !isCommentMark(mark.Type) {
			return fmt.Errorf("mark type %q is not allowed in comments", mark.Type)
		}
	}

	for _, child := range n.Content {
		if err := validateCommentAllowlist(child); err != nil {
			return err
		}
	}

	return nil
}

func isCommentNode(t NodeType) bool {
	switch t {
	case NodeDoc, NodeParagraph, NodeBlockquote, NodeBulletList, NodeOrderedList,
		NodeListItem, NodeText, NodeHardBreak, NodeMention:
		return true
	default:
		return false
	}
}

func isCommentMark(t MarkType) bool {
	switch t {
	case MarkStrong, MarkEm, MarkStrike, MarkCode, MarkLink:
		return true
	default:
		return false
	}
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

// DefaultDocumentJSON sanitizes a ProseMirror document. A nil or
// whitespace-only value becomes an empty document (one empty paragraph).
func DefaultDocumentJSON(s *string) (string, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return FromPlainText(""), nil
	}

	return SanitizeDocumentJSON(*s)
}

// SanitizeDocumentJSON parses a ProseMirror/Tiptap JSON document, replaces
// unsafe link mark href values using the same rules as RenderHTML, and
// re-serializes the document. Whitespace-only input is returned unchanged.
// Non-empty content must be valid JSON whose root node has type "doc".
func SanitizeDocumentJSON(s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return s, nil
	}

	return DocumentSchema.Sanitize(s)
}

// RenderHTML renders a ProseMirror document node tree to an HTML string.
func RenderHTML(node Node) (string, error) {
	return DocumentSchema.RenderHTML(node)
}

// RenderMarkdown renders a ProseMirror document node tree to a Markdown string.
func RenderMarkdown(node Node) (string, error) {
	return DocumentSchema.RenderMarkdown(node)
}

func validateDocument(n Node) error {
	return DocumentSchema.Validate(n)
}
