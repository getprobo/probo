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

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/html"
)

type imageHTMLRenderer struct{}

// HTMLImage returns the HTML renderer for image nodes.
func HTMLImage() html.NodeRenderer { return imageHTMLRenderer{} }

func (imageHTMLRenderer) Node() prosemirror.NodeType { return prosemirror.NodeImage }

func (imageHTMLRenderer) Render(w *html.Writer, n prosemirror.Node) error {
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

type fileHTMLRenderer struct{}

// HTMLFile returns the HTML renderer for file nodes.
func HTMLFile() html.NodeRenderer { return fileHTMLRenderer{} }

func (fileHTMLRenderer) Node() prosemirror.NodeType { return prosemirror.NodeFile }

func (fileHTMLRenderer) Render(w *html.Writer, n prosemirror.Node) error {
	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot render file node: %w", err)
	}

	w.WriteString("<a")
	w.WriteAttr("href", Path(attrs.FileID))
	w.WriteAttr("class", "document-file")
	w.WriteByte('>')
	w.WriteEscaped(attrs.FileName)
	w.WriteString("</a>")

	return nil
}

// HTMLNodes returns HTML renderers for image and file nodes.
func HTMLNodes() []html.NodeRenderer {
	return []html.NodeRenderer{HTMLImage(), HTMLFile()}
}
