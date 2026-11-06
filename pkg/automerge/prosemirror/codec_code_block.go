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

type codeBlockCodec struct{ baseCodec }

func (codeBlockCodec) Node() prosemirror.NodeType { return prosemirror.NodeCodeBlock }

func (codeBlockCodec) Block() string { return blockTypeCode }

func (codeBlockCodec) Role() blockRole { return roleTextblock }

func (codeBlockCodec) ToAttrs(attrs map[string]any) map[string]any {
	if language, ok := attrs["language"].(string); ok {
		return map[string]any{"language": language}
	}

	return map[string]any{}
}

func (codeBlockCodec) Render(source block, children []node) ([]node, string) {
	attrs := map[string]any{"language": nil}
	if language, ok := source.Attrs["language"].(string); ok {
		attrs["language"] = language
	}

	rendered := node{
		Type:    blockNodeNames[blockTypeCode],
		Attrs:   attrs,
		Content: source.Content,
	}

	return append([]node{rendered}, children...), ""
}
