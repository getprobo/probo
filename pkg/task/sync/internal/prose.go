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

package internal

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/prosemirror"
)

func TruncateMentionLabel(label string) string {
	if len(label) <= prosemirror.MaxMentionLabelBytes {
		return label
	}

	cut := label[:prosemirror.MaxMentionLabelBytes]
	for cut != "" && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}

	return strings.TrimSpace(cut)
}

func NewMentionNode(attrs prosemirror.MentionAttrs) (prosemirror.Node, error) {
	raw, err := json.Marshal(attrs)
	if err != nil {
		return prosemirror.Node{}, err
	}

	return prosemirror.Node{Type: prosemirror.NodeMention, Attrs: raw}, nil
}

func TextNode(value string, marks []prosemirror.Mark) prosemirror.Node {
	text := value

	return prosemirror.Node{
		Type:  prosemirror.NodeText,
		Text:  text,
		Marks: append([]prosemirror.Mark(nil), marks...),
	}
}

func NodeHasMark(node prosemirror.Node, kind prosemirror.MarkType) bool {
	for _, mark := range node.Marks {
		if mark.Type == kind {
			return true
		}
	}

	return false
}
