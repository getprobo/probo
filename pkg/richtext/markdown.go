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
	"fmt"

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
	"go.probo.inc/probo/pkg/richtext/attachment"
)

// ParseMarkdown converts markdown into a document and checks it against
// DocumentSchema. Empty input becomes one empty paragraph.
func ParseMarkdown(src string) (prosemirror.Node, error) {
	node, err := markdown.Parse(src)
	if err != nil {
		return prosemirror.Node{}, err
	}

	markdown.ApplyRules(&node, attachment.ImportRule())

	if len(node.Content) == 0 {
		node.Content = []prosemirror.Node{{Type: prosemirror.NodeParagraph}}
	}

	if err := DocumentSchema.Validate(node); err != nil {
		return prosemirror.Node{}, fmt.Errorf("cannot validate markdown document: %w", err)
	}

	return node, nil
}
