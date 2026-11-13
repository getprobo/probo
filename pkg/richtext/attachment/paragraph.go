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

package attachment

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/prosemirror"
)

// AttachmentFromParagraph reports whether paragraph is only a file-attachment link.
func AttachmentFromParagraph(n prosemirror.Node) (prosemirror.Node, bool) {
	if n.Type != prosemirror.NodeParagraph || len(n.Content) == 0 {
		return prosemirror.Node{}, false
	}

	var name strings.Builder

	var fileID string

	for _, child := range n.Content {
		if child.Type == prosemirror.NodeText && child.Text != "" && len(child.Marks) == 0 && strings.TrimSpace(child.Text) == "" {
			name.WriteString(child.Text)
			continue
		}

		if child.Type != prosemirror.NodeText || child.Text == "" || len(child.Marks) != 1 || child.Marks[0].Type != prosemirror.MarkLink {
			return prosemirror.Node{}, false
		}

		attrs, err := child.Marks[0].LinkAttrs()
		if err != nil {
			return prosemirror.Node{}, false
		}

		id, ok := ParseID(attrs.Href)
		if !ok || (fileID != "" && id != fileID) {
			return prosemirror.Node{}, false
		}

		fileID = id

		name.WriteString(child.Text)
	}

	if fileID == "" {
		return prosemirror.Node{}, false
	}

	fileName := strings.TrimSpace(name.String())
	if fileName == "" {
		return prosemirror.Node{}, false
	}

	if len(fileName) > prosemirror.MaxFileNameBytes {
		end := prosemirror.MaxFileNameBytes
		for end > 0 && !utf8.RuneStart(fileName[end]) {
			end--
		}

		fileName = fileName[:end]
	}

	raw, err := json.Marshal(prosemirror.FileAttrs{
		FileID:   fileID,
		FileName: fileName,
		MimeType: "application/octet-stream",
		Size:     0,
	})
	if err != nil {
		return prosemirror.Node{}, false
	}

	return prosemirror.Node{Type: prosemirror.NodeFile, Attrs: raw}, true
}
