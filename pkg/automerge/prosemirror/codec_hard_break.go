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

// hardBreakCodec is an inline embed: a hard break never becomes a standalone
// block. The forward traversal emits it from within a text block (see
// emitInline) and the reverse traversal appends it to the current block while
// collecting spans (see collectBlocks), so Render is never consulted.
type hardBreakCodec struct{ baseCodec }

func (hardBreakCodec) Node() prosemirror.NodeType { return prosemirror.NodeHardBreak }

func (hardBreakCodec) Block() string { return blockTypeHardBreak }

func (hardBreakCodec) Role() blockRole { return roleInlineEmbed }

func (hardBreakCodec) Render(_ block, children []node) ([]node, string) {
	return append([]node{{Type: blockNodeNames[blockTypeHardBreak]}}, children...), ""
}
