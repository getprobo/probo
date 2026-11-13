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

package richtext

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
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

	if err := DocumentSchema.Validate(n); err != nil {
		return fmt.Errorf("cannot validate comment document: %w", err)
	}

	if err := validateCommentAllowlist(n); err != nil {
		return fmt.Errorf("cannot validate comment allowlist: %w", err)
	}

	return nil
}

func validateCommentAllowlist(n prosemirror.Node) error {
	if !isCommentNode(n.Type) {
		return fmt.Errorf("node type %q is not allowed in comments", n.Type)
	}

	if n.Type == prosemirror.NodeMention {
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

func isCommentNode(t prosemirror.NodeType) bool {
	switch t {
	case prosemirror.NodeDoc, prosemirror.NodeParagraph, prosemirror.NodeBlockquote,
		prosemirror.NodeBulletList, prosemirror.NodeOrderedList,
		prosemirror.NodeListItem, prosemirror.NodeText, prosemirror.NodeHardBreak,
		prosemirror.NodeMention:
		return true
	default:
		return false
	}
}

func isCommentMark(t prosemirror.MarkType) bool {
	switch t {
	case prosemirror.MarkStrong, prosemirror.MarkEm, prosemirror.MarkStrike,
		prosemirror.MarkCode, prosemirror.MarkLink:
		return true
	default:
		return false
	}
}

func parseDocRoot(s string) (prosemirror.Node, error) {
	n, err := DocumentSchema.Parse(s)
	if err != nil {
		return prosemirror.Node{}, fmt.Errorf("cannot parse document content as ProseMirror JSON: %w", err)
	}

	if n.Type != prosemirror.NodeDoc {
		return prosemirror.Node{}, fmt.Errorf("document content root must be type %q", prosemirror.NodeDoc)
	}

	return n, nil
}
