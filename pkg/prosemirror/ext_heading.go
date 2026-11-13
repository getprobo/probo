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

type headingNode struct{}

// HeadingAttrs contains attributes for heading nodes.
type HeadingAttrs struct {
	Level int `json:"level"`
}

// HeadingAttrs parses and returns the heading attributes from a heading node.
// Attributes are optional in the Tiptap schema: a node that omits them, or
// omits individual keys, takes the schema defaults.
func (n Node) HeadingAttrs() (HeadingAttrs, error) {
	return decodeAttrs(n.Attrs, HeadingAttrs{Level: 1}, "heading")
}

// Heading is a section heading of level 1 to 6 holding inline content.
func Heading() NodeSpec {
	return headingNode{}
}

func (headingNode) Name() NodeType {
	return NodeHeading
}

func (headingNode) Group() Group {
	return GroupBlock
}

func (headingNode) Content() ContentRule {
	return ContentRule{Groups: []Group{GroupInline}}
}

func (headingNode) Check(n *Node, _ CheckMode) error {
	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	attrs, err := n.HeadingAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate heading node: %w", err)
	}

	if attrs.Level < 1 || attrs.Level > 6 {
		return fmt.Errorf("cannot validate heading node: invalid level %d", attrs.Level)
	}

	return nil
}
