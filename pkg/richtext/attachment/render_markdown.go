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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
)

type imageMarkdownRenderer struct{}

// MarkdownImage returns the Markdown renderer for image nodes.
func MarkdownImage() markdown.NodeRenderer { return imageMarkdownRenderer{} }

func (imageMarkdownRenderer) Node() prosemirror.NodeType { return prosemirror.NodeImage }

func (imageMarkdownRenderer) Render(w *markdown.Writer, n prosemirror.Node) error {
	attrs, err := n.ImageAttrs()
	if err != nil {
		return fmt.Errorf("cannot render image node: %w", err)
	}

	w.EnsurePrefix()

	if attrs.Width != nil {
		w.WriteString("<img")
		w.WriteAttr("src", imageSrc(attrs))

		if attrs.Alt != nil {
			w.WriteAttr("alt", *attrs.Alt)
		}

		if attrs.Title != nil {
			w.WriteAttr("title", *attrs.Title)
		}

		w.WriteAttr("width", fmt.Sprintf("%d%%", *attrs.Width))

		if err := w.WriteByte('>'); err != nil {
			return fmt.Errorf("cannot write image tag: %w", err)
		}

		return nil
	}

	w.WriteString("![")

	if attrs.Alt != nil {
		w.WriteEscaped(*attrs.Alt)
	}

	w.WriteString("](")
	w.WriteString(imageSrc(attrs))

	if attrs.Title != nil {
		w.WriteString(` "`)
		w.WriteString(strings.ReplaceAll(*attrs.Title, `"`, `\"`))

		if err := w.WriteByte('"'); err != nil {
			return fmt.Errorf("cannot write image title: %w", err)
		}
	}

	if err := w.WriteByte(')'); err != nil {
		return fmt.Errorf("cannot write image link: %w", err)
	}

	return nil
}

type fileMarkdownRenderer struct{}

// MarkdownFile returns the Markdown renderer for file nodes.
func MarkdownFile() markdown.NodeRenderer { return fileMarkdownRenderer{} }

func (fileMarkdownRenderer) Node() prosemirror.NodeType { return prosemirror.NodeFile }

func (fileMarkdownRenderer) Render(w *markdown.Writer, n prosemirror.Node) error {
	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot render file node: %w", err)
	}

	w.EnsurePrefix()

	if err := w.WriteByte('['); err != nil {
		return fmt.Errorf("cannot write file link start: %w", err)
	}

	w.WriteLinkText(attrs.FileName)
	w.WriteString("](")
	w.WriteString(Path(attrs.FileID))

	if err := w.WriteByte(')'); err != nil {
		return fmt.Errorf("cannot write file link end: %w", err)
	}

	return nil
}

// MarkdownNodes returns Markdown renderers for image and file nodes.
func MarkdownNodes() []markdown.NodeRenderer {
	return []markdown.NodeRenderer{MarkdownImage(), MarkdownFile()}
}
