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
)

type fileNode struct{}

// File is a leaf node referencing an uploaded file attachment.
func File() prosemirror.NodeSpec {
	return fileNode{}
}

func (fileNode) Name() prosemirror.NodeType {
	return prosemirror.NodeFile
}

func (fileNode) Group() prosemirror.Group {
	return prosemirror.GroupBlock
}

func (fileNode) Content() prosemirror.ContentRule {
	return prosemirror.ContentRule{}
}

func (fileNode) Visible() bool {
	return true
}

func (fileNode) Check(n *prosemirror.Node, _ prosemirror.CheckMode) error {
	if err := prosemirror.ValidateLeaf(n); err != nil {
		return fmt.Errorf("cannot validate file node: %w", err)
	}

	attrs, err := n.FileAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate file node: %w", err)
	}

	if !validFileID(attrs.FileID) {
		return fmt.Errorf("cannot validate file node: invalid file id")
	}

	if strings.TrimSpace(attrs.FileName) == "" || len(attrs.FileName) > prosemirror.MaxFileNameBytes {
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
