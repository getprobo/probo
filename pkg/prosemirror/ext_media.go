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

type imageNode struct{}

// Image is a leaf node referencing an uploaded or remote image.
func Image() NodeSpec {
	return imageNode{}
}

func (imageNode) Name() NodeType {
	return NodeImage
}

func (imageNode) Group() Group {
	return GroupBlock
}

func (imageNode) Content() ContentRule {
	return ContentRule{}
}

func (imageNode) Visible() bool {
	return true
}

func (imageNode) Check(n *Node, mode CheckMode) error {
	if err := validateLeaf(n); err != nil {
		return fmt.Errorf("cannot validate image node: %w", err)
	}

	attrs, err := n.ImageAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate image node: %w", err)
	}

	if attrs.FileID != nil {
		id := strings.TrimSpace(*attrs.FileID)
		if id != "" && !validFileID(id) {
			return fmt.Errorf("cannot validate image node: invalid file id")
		}
	}

	if attrs.Width != nil && (*attrs.Width < 1 || *attrs.Width > 100) {
		return fmt.Errorf("cannot validate image node: invalid width")
	}

	if mode != CheckSanitize {
		return nil
	}

	if attrs.FileID != nil {
		id := strings.TrimSpace(*attrs.FileID)
		if id == "" {
			attrs.FileID = nil
		} else {
			attrs.FileID = &id
			attrs.Src = ""
		}
	}

	if attrs.FileID == nil {
		attrs.Src = persistedImageSrc(attrs.Src)
	}

	if attrs.Width != nil && (*attrs.Width < 1 || *attrs.Width > 100) {
		attrs.Width = nil
	}

	raw, err := json.Marshal(attrs)
	if err != nil {
		n.Attrs = []byte(`{"src":""}`)
		return nil
	}

	n.Attrs = raw

	return nil
}

func (imageNode) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.ImageAttrs()
	if err != nil {
		return fmt.Errorf("cannot render image node: %w", err)
	}

	w.WriteString("<img")
	w.WriteAttr("src", imageSrc(attrs))

	if attrs.Alt != nil {
		w.WriteAttr("alt", *attrs.Alt)
	}

	if attrs.Title != nil {
		w.WriteAttr("title", *attrs.Title)
	}

	if attrs.Width != nil {
		w.WriteAttr("width", fmt.Sprintf("%d%%", *attrs.Width))
	}

	w.WriteByte('>')

	return nil
}

func (imageNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	attrs, err := n.ImageAttrs()
	if err != nil {
		return fmt.Errorf("cannot render image node: %w", err)
	}

	w.ensurePrefix()

	if attrs.Width != nil {
		w.buf.WriteString("<img")
		writeAttr(w.buf, "src", imageSrc(attrs))

		if attrs.Alt != nil {
			writeAttr(w.buf, "alt", *attrs.Alt)
		}

		if attrs.Title != nil {
			writeAttr(w.buf, "title", *attrs.Title)
		}

		writeAttr(w.buf, "width", fmt.Sprintf("%d%%", *attrs.Width))
		w.buf.WriteByte('>')

		return nil
	}

	w.buf.WriteString("![")

	if attrs.Alt != nil {
		w.buf.WriteString(escapeMarkdown(*attrs.Alt))
	}

	w.buf.WriteString("](")
	w.buf.WriteString(imageSrc(attrs))

	if attrs.Title != nil {
		w.buf.WriteString(` "`)
		w.buf.WriteString(strings.ReplaceAll(*attrs.Title, `"`, `\"`))
		w.buf.WriteByte('"')
	}

	w.buf.WriteByte(')')

	return nil
}

type fileNode struct{}

// File is a leaf node referencing an uploaded file attachment.
func File() NodeSpec {
	return fileNode{}
}

func (fileNode) Name() NodeType {
	return NodeFile
}

func (fileNode) Group() Group {
	return GroupBlock
}

func (fileNode) Content() ContentRule {
	return ContentRule{}
}

func (fileNode) Visible() bool {
	return true
}

func (fileNode) Check(n *Node, _ CheckMode) error {
	if err := validateLeaf(n); err != nil {
		return fmt.Errorf("cannot validate file node: %w", err)
	}

	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate file node: %w", err)
	}

	if !validFileID(attrs.FileID) {
		return fmt.Errorf("cannot validate file node: invalid file id")
	}

	if strings.TrimSpace(attrs.FileName) == "" || len(attrs.FileName) > MaxFileNameBytes {
		return fmt.Errorf("cannot validate file node: invalid file name")
	}

	if strings.TrimSpace(attrs.MimeType) == "" || len(attrs.MimeType) > 255 {
		return fmt.Errorf("cannot validate file node: invalid mime type")
	}

	if attrs.Size < 0 {
		return fmt.Errorf("cannot validate file node: invalid size")
	}

	return nil
}

func (fileNode) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot render file node: %w", err)
	}

	w.WriteString("<a")
	w.WriteAttr("href", AttachmentPath(attrs.FileID))
	w.WriteAttr("class", "document-file")
	w.WriteByte('>')
	w.WriteEscaped(attrs.FileName)
	w.WriteString("</a>")

	return nil
}

func (fileNode) RenderMarkdown(w *MarkdownWriter, n Node) error {
	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot render file node: %w", err)
	}

	w.ensurePrefix()
	w.buf.WriteByte('[')
	w.buf.WriteString(strings.ReplaceAll(escapeMarkdown(attrs.FileName), "&", `\&`))
	w.buf.WriteString("](")
	w.buf.WriteString(AttachmentPath(attrs.FileID))
	w.buf.WriteByte(')')

	return nil
}
