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

import "go.probo.inc/probo/pkg/prosemirror"

type imageCodec struct{ baseCodec }

func (imageCodec) Node() prosemirror.NodeType { return prosemirror.NodeImage }

func (imageCodec) Block() string { return blockTypeImage }

func (imageCodec) Role() blockRole { return roleEmbed }

func (imageCodec) ToAttrs(attrs map[string]any) map[string]any {
	return attachmentAttributes(attrs, "fileId", "src", "alt", "title", "width")
}

func (imageCodec) Render(source block, children []node) ([]node, string) {
	rendered := []node{{Type: "image", Attrs: source.Attrs}}

	return append(rendered, children...), ""
}

type fileCodec struct{ baseCodec }

func (fileCodec) Node() prosemirror.NodeType { return prosemirror.NodeFile }

func (fileCodec) Block() string { return blockTypeFile }

func (fileCodec) Role() blockRole { return roleEmbed }

func (fileCodec) ToAttrs(attrs map[string]any) map[string]any {
	return attachmentAttributes(attrs, "fileId", "fileName", "mimeType", "size")
}

func (fileCodec) Render(source block, children []node) ([]node, string) {
	rendered := []node{{Type: "file", Attrs: source.Attrs}}

	return append(rendered, children...), ""
}
