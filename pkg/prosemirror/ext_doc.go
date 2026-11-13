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

import "fmt"

type docNode struct{}

// Doc is the top-level document node. It holds one or more block nodes.
func Doc() NodeSpec {
	return docNode{}
}

func (docNode) Name() NodeType {
	return NodeDoc
}

func (docNode) Group() Group {
	return GroupNone
}

func (docNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (docNode) Check(_ *Node, _ CheckMode) error {
	return nil
}

type blockquoteNode struct{}

// Blockquote wraps one or more block nodes in a quotation.
func Blockquote() NodeSpec {
	return blockquoteNode{}
}

func (blockquoteNode) Name() NodeType {
	return NodeBlockquote
}

func (blockquoteNode) Group() Group {
	return GroupBlock
}

func (blockquoteNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupBlock}, Min: 1}
}

func (blockquoteNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}
