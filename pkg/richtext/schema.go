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
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/richtext/attachment"
	"go.probo.inc/probo/pkg/richtext/mention"
)

// DocumentSchema is the Tiptap-compatible schema the console editor produces.
// A mention accepts any GID, matching the current document behavior.
var DocumentSchema = prosemirror.NewSchema(
	[]prosemirror.NodeSpec{
		prosemirror.Doc(),
		prosemirror.Paragraph(),
		prosemirror.Text(),
		prosemirror.Heading(),
		prosemirror.Blockquote(),
		prosemirror.CodeBlock(),
		prosemirror.HorizontalRule(),
		prosemirror.HardBreak(),
		attachment.Image(),
		attachment.File(),
		mention.Document(),
		prosemirror.BulletList(),
		prosemirror.OrderedList(),
		prosemirror.ListItem(),
		prosemirror.TaskList(),
		prosemirror.TaskItem(),
		prosemirror.Table(),
		prosemirror.TableRow(),
		prosemirror.TableCell(),
		prosemirror.TableHeader(),
	},
	[]prosemirror.MarkSpec{
		prosemirror.Bold(),
		prosemirror.Italic(),
		prosemirror.Underline(),
		prosemirror.Strike(),
		prosemirror.Code(),
		prosemirror.Link(),
	},
)

// DefaultDocumentJSON sanitizes a ProseMirror document. A nil or
// whitespace-only value becomes an empty document (one empty paragraph).
func DefaultDocumentJSON(s *string) (string, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return FromPlainText("")
	}

	return DocumentSchema.Sanitize(*s)
}
