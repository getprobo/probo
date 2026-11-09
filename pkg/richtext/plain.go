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
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

// FromPlainText wraps plaintext as a ProseMirror document (one paragraph per
// line), matching the task/comment SQL backfill and console editor.
func FromPlainText(text string) (string, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	content := make([]prosemirror.Node, 0, len(lines))

	for _, line := range lines {
		para := prosemirror.Node{Type: prosemirror.NodeParagraph}

		if line != "" {
			para.Content = []prosemirror.Node{{
				Type: prosemirror.NodeText,
				Text: line,
			}}
		}

		content = append(content, para)
	}

	encoded, err := json.Marshal(prosemirror.Node{
		Type:    prosemirror.NodeDoc,
		Content: content,
	})
	if err != nil {
		return "", fmt.Errorf("cannot marshal plaintext document: %w", err)
	}

	return string(encoded), nil
}
