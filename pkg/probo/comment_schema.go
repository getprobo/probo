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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/prosemirror"
)

// CommentSchema is the ProseMirror schema for document review comments. It is a
// subset of DocumentSchema: paragraphs, lists, blockquotes, hard breaks,
// membership-profile mentions, and the bold/italic/strike/code/link marks.
// Headings, images, tables, underline, and other document-only nodes are
// rejected. Keep it aligned with richEditorCommentExtensions on the frontend.
var CommentSchema = prosemirror.NewSchema(
	[]prosemirror.NodeSpec{
		prosemirror.Doc(),
		prosemirror.Paragraph(),
		prosemirror.Text(),
		prosemirror.Blockquote(),
		prosemirror.BulletList(),
		prosemirror.OrderedList(),
		prosemirror.ListItem(),
		prosemirror.HardBreak(),
		prosemirror.Mention(prosemirror.MentionOptions{
			EntityTypes: []uint16{coredata.MembershipProfileEntityType},
		}),
	},
	[]prosemirror.MarkSpec{
		prosemirror.Bold(),
		prosemirror.Italic(),
		prosemirror.Strike(),
		prosemirror.Code(),
		prosemirror.Link(),
	},
)

// DefaultCommentJSON sanitizes comment content with CommentSchema. A nil or
// whitespace-only value becomes an empty document (one empty paragraph).
func DefaultCommentJSON(s *string) (string, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return prosemirror.FromPlainText(""), nil
	}

	return CommentSchema.Sanitize(*s)
}
