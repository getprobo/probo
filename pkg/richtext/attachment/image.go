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

type imageNode struct{}

// Image is a leaf node referencing an uploaded or remote image.
func Image() prosemirror.NodeSpec {
	return imageNode{}
}

func (imageNode) Name() prosemirror.NodeType {
	return prosemirror.NodeImage
}

func (imageNode) Group() prosemirror.Group {
	return prosemirror.GroupBlock
}

func (imageNode) Content() prosemirror.ContentRule {
	return prosemirror.ContentRule{}
}

func (imageNode) Visible() bool {
	return true
}

func (imageNode) Check(n *prosemirror.Node, mode prosemirror.CheckMode) error {
	if err := prosemirror.ValidateLeaf(n); err != nil {
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

	if mode != prosemirror.CheckSanitize {
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
		attrs.Src = prosemirror.PersistedImageSrc(attrs.Src)
	}

	if attrs.Width != nil && (*attrs.Width < 1 || *attrs.Width > 100) {
		attrs.Width = nil
	}

	if err := n.SetAttrs(attrs); err != nil {
		return fmt.Errorf("cannot validate image node: %w", err)
	}

	return nil
}
